'use client';

import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { ApiError } from '@/shared/api/client';
import { mocksReady } from '@/shared/api/mocks/enable';
import { fetchEvent } from '@/entities/event/api/events-api';
import { coverGradient } from '@/entities/event/lib/cover';
import { formatDateTime } from '@/shared/lib/dates';
import { formatMoney } from '@/shared/lib/money';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { Skeleton } from '@/shared/ui/skeleton';

export function EventDetail({ id }: { id: string }) {
  const {
    data: event,
    status,
    error,
  } = useQuery({
    queryKey: ['event', id],
    queryFn: async () => {
      await mocksReady;
      return fetchEvent(id);
    },
  });

  if (status === 'pending') {
    return (
      <main className="mx-auto max-w-3xl space-y-4 p-6">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-48 w-full" />
        <Skeleton className="h-8 w-2/3" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-3/4" />
      </main>
    );
  }

  if (status === 'error') {
    const notFound = error instanceof ApiError && error.status === 404;
    return (
      <main className="mx-auto max-w-md space-y-3 p-6 pt-24 text-center">
        <h1 className="font-display text-xl font-semibold">
          {notFound ? 'Event not found' : 'Something went wrong'}
        </h1>
        <p className="text-sm text-muted">
          {notFound ? 'This event does not exist or is not published yet.' : error.message}
        </p>
        <Link href="/events" className="text-sm text-accent hover:underline">
          ← All events
        </Link>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-3xl space-y-5 p-6">
      <Link href="/events" className="text-sm text-muted hover:text-foreground">
        ← All events
      </Link>
      <div className="h-48 w-full rounded-xl" style={{ background: coverGradient(event.id) }} />
      <div className="flex flex-wrap items-center gap-2">
        {event.status === 'draft' && <Badge variant="warning">Draft</Badge>}
        <span className="text-sm text-muted">{formatDateTime(event.starts_at)}</span>
      </div>
      <h1 className="font-display text-3xl font-semibold">{event.title}</h1>
      {event.description && <p className="text-muted">{event.description}</p>}
      <section className="space-y-2">
        <h2 className="font-display text-sm uppercase tracking-wide text-muted">
          Ticket categories
        </h2>
        <div className="space-y-2">
          {event.categories.map((c) => (
            <Card key={c.id} className="flex items-center justify-between p-4">
              <span className="font-medium">{c.name}</span>
              <span className="text-sm text-muted">
                {formatMoney(c.price.amount, c.price.currency)} · {c.total_qty} seats
              </span>
            </Card>
          ))}
        </div>
      </section>
      <p className="text-sm text-muted">
        Live availability and booking arrive with the next phase.
      </p>
    </main>
  );
}
