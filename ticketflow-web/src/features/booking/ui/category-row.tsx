'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '@/shared/api/client';
import { mocksReady } from '@/shared/api/mocks/enable';
import { useCheckoutStore } from '@/features/checkout/model/checkout-store';
import type { CategoryAvailability, HoldCreated } from '@/shared/api/types';
import { formatMoney } from '@/shared/lib/money';
import { Button } from '@/shared/ui/button';
import { LiveBadge } from '@/shared/ui/live-badge';
import { QuantityStepper } from '@/shared/ui/quantity-stepper';
import { useAuthStore } from '@/features/auth/model/auth-store';
import { createHold } from '@/features/booking/model/holds-api';

interface CategoryRowProps {
  category: CategoryAvailability;
  eventId: string;
  eventTitle: string;
}

export function CategoryRow({ category, eventId, eventTitle }: CategoryRowProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const refreshToken = useAuthStore((s) => s.refreshToken);
  const [qty, setQty] = useState(1);
  const [error, setError] = useState<string | null>(null);

  const hold = useMutation({
    mutationFn: async (input: { category_id: string; qty: number }) => {
      await mocksReady;
      return createHold(eventId, input);
    },
    onSuccess: (data: HoldCreated, input) => {
      useCheckoutStore.getState().setHold({
        holdId: data.hold_id,
        eventId,
        eventTitle,
        categoryName: category.name,
        qty: input.qty,
        expiresAt: data.expires_at,
        serverTime: data.server_time,
        idempotencyKey: crypto.randomUUID(),
        unitPriceMinor: category.price.amount,
        currency: category.price.currency,
      });
      router.push('/checkout');
    },
    onError: (err) => {
      if (err instanceof ApiError && err.status === 409) {
        void queryClient.invalidateQueries({ queryKey: ['availability', eventId] });
        setError('Tickets just sold out — availability will update shortly.');
      } else if (err instanceof ApiError && err.status === 401) {
        setError('Please log in to book tickets.');
      } else {
        setError(err instanceof Error ? err.message : 'Failed to reserve tickets.');
      }
    },
  });

  const soldOut = category.available <= 0;

  return (
    <div className="flex flex-col gap-2 rounded-xl border border-border bg-surface p-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="space-y-1">
        <div className="font-medium">{category.name}</div>
        <div className="flex items-center gap-3">
          <span>{formatMoney(category.price.amount, category.price.currency)}</span>
          <LiveBadge available={category.available} total={category.total_qty} />
        </div>
      </div>
      <div className="flex items-center gap-3">
        {!soldOut && (
          <>
            <QuantityStepper
              value={qty}
              onChange={setQty}
              min={1}
              max={Math.min(10, category.available)}
            />
            <Button
              size="sm"
              disabled={soldOut || !refreshToken || hold.isPending}
              onClick={() => hold.mutate({ category_id: category.category_id, qty })}
            >
              {hold.isPending ? 'Reserving…' : 'Reserve'}
            </Button>
          </>
        )}
        {soldOut && <span className="text-sm text-danger">Sold out</span>}
      </div>
      {error && <p className="text-sm text-danger sm:basis-full">{error}</p>}
    </div>
  );
}
