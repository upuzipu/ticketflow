'use client';

import Link from 'next/link';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '@/shared/api/client';
import { mocksReady } from '@/shared/api/mocks/enable';
import { fetchOrder, payOrder } from '../model/orders-api';
import { useCheckoutStore } from '../model/checkout-store';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { Badge } from '@/shared/ui/badge';
import { Skeleton } from '@/shared/ui/skeleton';
import { formatMoney } from '@/shared/lib/money';
import { formatDateTime } from '@/shared/lib/dates';
import type { OrderStatus } from '@/shared/api/types';
import { RequireAuth } from '@/shared/ui/require-auth';

const STATUS_META: Record<
  OrderStatus,
  { label: string; variant: 'accent' | 'success' | 'warning' | 'danger' | 'muted' }
> = {
  pending: { label: 'Pending payment', variant: 'warning' },
  paid: { label: 'Paid — issuing tickets', variant: 'accent' },
  confirmed: { label: 'Tickets ready', variant: 'success' },
  failed: { label: 'Payment declined', variant: 'danger' },
  expired: { label: 'Expired', variant: 'muted' },
  refunded: { label: 'Refunded', variant: 'muted' },
};

function OrderContent({ id }: { id: string }) {
  const queryClient = useQueryClient();
  const hold = useCheckoutStore((s) => s.hold);
  const clear = useCheckoutStore((s) => s.clear);

  const {
    data: order,
    status,
    error,
  } = useQuery({
    queryKey: ['order', id],
    queryFn: async () => {
      await mocksReady;
      return fetchOrder(id);
    },
    refetchInterval: (query) => {
      const s = query.state.data?.status;
      return s === 'pending' || s === 'paid' ? 2000 : false;
    },
  });

  const pay = useMutation({
    mutationFn: async () => {
      await mocksReady;
      return payOrder(id);
    },
    onSuccess: (updated) => {
      queryClient.setQueryData(['order', id], updated);
    },
    onError: () => {
      void queryClient.invalidateQueries({ queryKey: ['order', id] });
    },
  });

  if (status === 'pending') {
    return (
      <main className="mx-auto max-w-md space-y-3 p-6 pt-20">
        <Skeleton className="h-8 w-1/2" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-10 w-full" />
      </main>
    );
  }

  if (status === 'error') {
    const notFound = error instanceof ApiError && error.status === 404;
    return (
      <main className="mx-auto max-w-md space-y-3 p-6 pt-20 text-center">
        <h1 className="font-display text-xl font-semibold">
          {notFound ? 'Order not found' : 'Something went wrong'}
        </h1>
        <p className="text-sm text-muted">{error instanceof Error ? error.message : undefined}</p>
        <Link href="/events" className="text-sm text-accent hover:underline">
          ← Browse events
        </Link>
      </main>
    );
  }

  const meta = STATUS_META[order.status];
  const recap = hold && hold.holdId === order.hold_id ? hold : null;

  return (
    <main className="mx-auto w-full max-w-md space-y-4 p-6 pt-16">
      <h1 className="font-display text-xl font-semibold">Order</h1>
      <Card className="space-y-3 p-4 text-sm">
        <div className="flex items-center justify-between">
          <Badge variant={meta.variant}>{meta.label}</Badge>
          <span className="font-semibold">{formatMoney(order.total_minor, order.currency)}</span>
        </div>
        {recap && (
          <div className="space-y-1 border-t border-border pt-3">
            <div className="font-medium">{recap.eventTitle}</div>
            <div className="text-muted">
              {recap.categoryName} × {recap.qty}
            </div>
          </div>
        )}
        <div className="text-xs text-muted">
          Order ID: <code className="text-foreground">{order.id}</code>
        </div>
        <div className="text-xs text-muted">Created: {formatDateTime(order.created_at)}</div>
      </Card>

      {order.status === 'pending' && (
        <div className="space-y-2">
          <Button className="w-full" disabled={pay.isPending} onClick={() => pay.mutate()}>
            {pay.isPending ? 'Charging…' : 'Retry payment'}
          </Button>
          <p className="text-center text-xs text-muted">
            The payment gateway timed out earlier. Retrying is safe.
          </p>
        </div>
      )}
      {order.status === 'paid' && (
        <p className="flex items-center justify-center gap-2 text-sm text-muted">
          <span className="h-4 w-4 animate-spin rounded-full border-2 border-accent border-t-transparent" />
          Issuing your tickets…
        </p>
      )}
      {order.status === 'confirmed' && (
        <div className="space-y-2 text-center">
          <p className="text-sm text-muted">
            Your tickets are issued. Ticket codes will appear here in a future release.
          </p>
          <Button variant="secondary" asChild onClick={() => clear()}>
            <Link href="/events">Book more tickets</Link>
          </Button>
        </div>
      )}
      {order.status === 'failed' && (
        <div className="space-y-2 text-center">
          <p className="text-sm text-muted">The charge failed and the tickets were released.</p>
          <Button variant="secondary" asChild onClick={() => clear()}>
            <Link href="/events">Choose tickets again</Link>
          </Button>
        </div>
      )}
      {(order.status === 'expired' || order.status === 'refunded') && (
        <p className="text-center text-xs text-muted">This order is closed.</p>
      )}
    </main>
  );
}

export function OrderView({ id }: { id: string }) {
  return (
    <RequireAuth>
      <OrderContent id={id} />
    </RequireAuth>
  );
}
