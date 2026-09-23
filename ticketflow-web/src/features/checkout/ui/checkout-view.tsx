'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { useMutation } from '@tanstack/react-query';
import { ApiError } from '@/shared/api/client';
import { mocksReady } from '@/shared/api/mocks/enable';
import { fetchHold, releaseHold } from '@/features/booking/model/holds-api';
import { createOrder } from '../model/orders-api';
import { useCheckoutStore, type HoldSnapshot } from '../model/checkout-store';
import { useHoldTimer } from '../model/use-hold-timer';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { CountdownRing } from '@/shared/ui/countdown-ring';
import { formatMoney } from '@/shared/lib/money';
import { RequireAuth } from '@/shared/ui/require-auth';

type Phase =
  'verifying' | 'ready' | 'paying' | 'review' | 'declined' | 'expired' | 'gone' | 'error';

const HOLD_TOTAL_SECONDS = 600;

function CheckoutContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setHold = useCheckoutStore((s) => s.setHold);
  const clear = useCheckoutStore((s) => s.clear);

  const [phase, setPhase] = useState<Phase>('verifying');
  const [holdId, setHoldId] = useState<string | null>(null);
  const [snapshot, setSnapshot] = useState<HoldSnapshot | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const timerActive =
    phase === 'ready' || phase === 'paying' || phase === 'review' || phase === 'error';
  const secondsLeft = useHoldTimer(
    timerActive ? (snapshot?.expiresAt ?? null) : null,
    timerActive ? (snapshot?.serverTime ?? null) : null,
  );

  useEffect(() => {
    if (secondsLeft === 0 && (phase === 'ready' || phase === 'paying' || phase === 'error')) {
      setPhase('expired');
    }
  }, [secondsLeft, phase]);

  useEffect(() => {
    let cancelled = false;
    const verify = async () => {
      await mocksReady;
      const id = searchParams.get('hold') ?? useCheckoutStore.getState().hold?.holdId ?? null;
      if (!id) {
        if (!cancelled) setPhase('gone');
        return;
      }
      setHoldId(id);
      try {
        const status = await fetchHold(id);
        if (cancelled) return;
        const store = useCheckoutStore.getState();
        if (status.status === 'active') {
          const next: HoldSnapshot =
            store.hold && store.hold.holdId === id
              ? { ...store.hold, expiresAt: status.expires_at, serverTime: status.server_time }
              : {
                  holdId: id,
                  eventId: '',
                  eventTitle: 'Your reservation',
                  categoryName: 'Tickets',
                  qty: status.tickets,
                  expiresAt: status.expires_at,
                  serverTime: status.server_time,
                  idempotencyKey: crypto.randomUUID(),
                  unitPriceMinor: null,
                  currency: null,
                };
          setSnapshot(next);
          setHold(next);
          setPhase('ready');
        } else if (status.status === 'confirmed') {
          if (status.order_id) {
            store.setOrderId(status.order_id);
            router.replace(`/orders/${status.order_id}`);
          } else if (store.hold?.holdId === id && store.hold.idempotencyKey) {
            setSnapshot(store.hold);
            setPhase('review');
          } else {
            store.clear();
            setPhase('gone');
          }
        } else {
          store.clear();
          setPhase('expired');
        }
      } catch (err) {
        if (cancelled) return;
        if (err instanceof ApiError && (err.status === 403 || err.status === 404)) {
          useCheckoutStore.getState().clear();
          setPhase('gone');
        } else {
          setErrorMessage(err instanceof Error ? err.message : 'Failed to verify the reservation.');
          setPhase('error');
        }
      }
    };
    void verify();
    return () => {
      cancelled = true;
    };
  }, [searchParams, router, setHold]);

  const pay = useMutation({
    mutationFn: async () => {
      await mocksReady;
      return createOrder({
        hold_id: holdId as string,
        idempotency_key: (snapshot as HoldSnapshot).idempotencyKey,
      });
    },
    onSuccess: (order) => {
      useCheckoutStore.getState().setOrderId(order.id);
      router.push(`/orders/${order.id}`);
    },
    onError: (err) => {
      if (!(err instanceof ApiError)) {
        setErrorMessage('Network error. Your reservation is still held — try again.');
        setPhase('error');
        return;
      }
      if (err.status === 402) setPhase('declined');
      else if (err.status === 424) setPhase('review');
      else if (err.status === 409 || err.status === 410) setPhase('expired');
      else {
        setErrorMessage(err.message);
        setPhase('error');
      }
    },
  });

  useEffect(() => {
    if (phase !== 'review' || !snapshot || !holdId) return;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | null = null;
    const attempt = async () => {
      try {
        await mocksReady;
        const order = await createOrder({
          hold_id: holdId,
          idempotency_key: snapshot.idempotencyKey,
        });
        if (cancelled) return;
        useCheckoutStore.getState().setOrderId(order.id);
        router.replace(`/orders/${order.id}`);
      } catch (err) {
        if (cancelled) return;
        if (err instanceof ApiError && err.status === 424) {
          timer = setTimeout(() => void attempt(), 3000);
        } else if (err instanceof ApiError && (err.status === 409 || err.status === 410)) {
          setPhase('expired');
        } else {
          setErrorMessage(err instanceof Error ? err.message : 'Payment is stuck — try again.');
          setPhase('error');
        }
      }
    };
    void attempt();
    return () => {
      cancelled = true;
      if (timer) clearTimeout(timer);
    };
  }, [phase, snapshot, holdId, router]);

  const cancel = useMutation({
    mutationFn: async () => {
      await mocksReady;
      return releaseHold(holdId as string);
    },
    onSettled: () => {
      clear();
      router.push(snapshot?.eventId ? `/events/${snapshot.eventId}` : '/events');
    },
  });

  if (phase === 'verifying') {
    return (
      <main className="mx-auto max-w-md p-6 pt-20 text-center text-sm text-muted">
        Verifying your reservation…
      </main>
    );
  }

  if (phase === 'expired') {
    return (
      <main className="mx-auto max-w-md space-y-3 p-6 pt-20 text-center">
        <h1 className="font-display text-xl font-semibold">Reservation expired</h1>
        <p className="text-sm text-muted">
          The 10-minute hold is over and the tickets are back on sale. Nothing was charged.
        </p>
        <Button
          onClick={() => router.push(snapshot?.eventId ? `/events/${snapshot.eventId}` : '/events')}
        >
          Choose tickets again
        </Button>
      </main>
    );
  }

  if (phase === 'declined') {
    return (
      <main className="mx-auto max-w-md space-y-3 p-6 pt-20 text-center">
        <h1 className="font-display text-xl font-semibold">Payment declined</h1>
        <p className="text-sm text-muted">
          The charge did not go through and the tickets have been released. Nothing was charged.
        </p>
        <Button
          onClick={() => router.push(snapshot?.eventId ? `/events/${snapshot.eventId}` : '/events')}
        >
          Choose tickets again
        </Button>
      </main>
    );
  }

  if (phase === 'gone' || !snapshot) {
    clear();
    return (
      <main className="mx-auto max-w-md space-y-3 p-6 pt-20 text-center">
        <h1 className="font-display text-xl font-semibold">No active reservation</h1>
        <p className="text-sm text-muted">
          Pick tickets first — the reservation is created automatically.
        </p>
        <Button asChild>
          <Link href="/events">Browse events</Link>
        </Button>
      </main>
    );
  }

  if (phase === 'error') {
    return (
      <main className="mx-auto max-w-md space-y-3 p-6 pt-20 text-center">
        <h1 className="font-display text-xl font-semibold">Something went wrong</h1>
        <p className="text-sm text-danger">{errorMessage}</p>
        <p className="text-sm text-muted">Your reservation is still held while the timer runs.</p>
        <div className="flex justify-center gap-2">
          <Button
            variant="secondary"
            onClick={() => {
              setErrorMessage(null);
              setPhase('ready');
            }}
          >
            Back to checkout
          </Button>
          <Button
            variant="ghost"
            onClick={() =>
              router.push(snapshot.eventId ? `/events/${snapshot.eventId}` : '/events')
            }
          >
            Leave
          </Button>
        </div>
      </main>
    );
  }

  if (phase === 'review') {
    return (
      <main className="mx-auto max-w-md space-y-4 p-6 pt-20 text-center">
        <h1 className="font-display text-xl font-semibold">Confirming your payment</h1>
        <p className="text-sm text-muted">
          The payment gateway is slow to respond. We keep retrying safely — you will not be charged
          twice.
        </p>
        <div className="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-accent border-t-transparent" />
      </main>
    );
  }

  const total =
    snapshot.unitPriceMinor !== null && snapshot.currency
      ? formatMoney(snapshot.unitPriceMinor * snapshot.qty, snapshot.currency)
      : null;

  return (
    <main className="mx-auto w-full max-w-md space-y-4 p-6 pt-16">
      <h1 className="font-display text-xl font-semibold">Checkout</h1>
      <Card className="flex items-center gap-4 p-4">
        <CountdownRing secondsLeft={secondsLeft ?? 0} totalSeconds={HOLD_TOTAL_SECONDS} size={64} />
        <div className="text-sm">
          <div className="font-medium">Tickets are reserved for you</div>
          <p className="text-muted">Complete the purchase before the timer runs out.</p>
        </div>
      </Card>
      <Card className="space-y-2 p-4 text-sm">
        <div className="text-lg font-semibold">{snapshot.eventTitle}</div>
        <div className="flex justify-between">
          <span className="text-muted">
            {snapshot.categoryName} × {snapshot.qty}
          </span>
          {total && <span>{total}</span>}
        </div>
        {total && (
          <div className="flex justify-between border-t border-border pt-2 font-semibold">
            <span>Total</span>
            <span>{total}</span>
          </div>
        )}
      </Card>
      <Button
        className="w-full"
        size="lg"
        disabled={pay.isPending || cancel.isPending}
        onClick={() => {
          setPhase('paying');
          pay.mutate();
        }}
      >
        {pay.isPending ? 'Paying…' : total ? `Pay ${total}` : 'Pay'}
      </Button>
      <Button
        className="w-full"
        variant="ghost"
        disabled={pay.isPending || cancel.isPending}
        onClick={() => cancel.mutate()}
      >
        {cancel.isPending ? 'Releasing…' : 'Cancel reservation'}
      </Button>
      <p className="text-center text-xs text-muted">
        The payment is idempotent: a retry can never charge you twice.
      </p>
    </main>
  );
}

export function CheckoutView() {
  return (
    <RequireAuth>
      <CheckoutContent />
    </RequireAuth>
  );
}
