'use client';

import Link from 'next/link';
import { useInfiniteQuery } from '@tanstack/react-query';
import { mocksReady } from '@/shared/api/mocks/enable';
import { fetchMyOrders } from '@/features/checkout/model/orders-api';
import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { Skeleton } from '@/shared/ui/skeleton';
import { formatMoney } from '@/shared/lib/money';
import { formatDateTime } from '@/shared/lib/dates';
import type { OrderStatus } from '@/shared/api/types';

const STATUS_BADGE: Record<
  OrderStatus,
  { label: string; variant: 'accent' | 'success' | 'warning' | 'danger' | 'muted' }
> = {
  pending: { label: 'Pending payment', variant: 'warning' },
  paid: { label: 'Paid — issuing', variant: 'accent' },
  confirmed: { label: 'Tickets ready', variant: 'success' },
  failed: { label: 'Declined', variant: 'danger' },
  expired: { label: 'Expired', variant: 'muted' },
  refunded: { label: 'Refunded', variant: 'muted' },
};

export function MyOrders() {
  const { data, status, error, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: ['orders', 'mine'],
    initialPageParam: '',
    queryFn: async ({ pageParam }) => {
      await mocksReady;
      return fetchMyOrders(20, pageParam || undefined);
    },
    getNextPageParam: (lastPage) => lastPage.next_cursor || undefined,
  });

  if (status === 'pending') {
    return (
      <div className="space-y-2">
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-20 w-full" />
      </div>
    );
  }

  if (status === 'error') {
    return (
      <Card className="p-10 text-center text-sm text-danger">
        {error instanceof Error ? error.message : 'Failed to load your events'}
      </Card>
    );
  }

  const orders = data.pages.flatMap((page) => page.orders);

  if (orders.length === 0) {
    return <Card className="p-10 text-center text-sm text-muted">No orders yet.</Card>;
  }

  return (
    <div className="space-y-3">
      {orders.map((order) => {
        const badge = STATUS_BADGE[order.status];
        return (
          <Link key={order.id} href={`/orders/${order.id}`} className="block">
            <Card className="flex items-center justify-between gap-3 p-4 transition-colors hover:border-accent/50">
              <div className="space-y-1">
                <Badge variant={badge.variant}>{badge.label}</Badge>
                <div className="text-xs text-muted">
                  {formatDateTime(order.created_at)} · {order.id.slice(0, 8)}…
                </div>
              </div>
              <span className="font-semibold">
                {formatMoney(order.total_minor, order.currency)}
              </span>
            </Card>
          </Link>
        );
      })}
      {hasNextPage && (
        <div className="flex justify-center">
          <Button
            variant="secondary"
            onClick={() => void fetchNextPage()}
            disabled={isFetchingNextPage}
          >
            {isFetchingNextPage ? 'Loading…' : 'Load more'}
          </Button>
        </div>
      )}
    </div>
  );
}
