'use client';

import Link from 'next/link';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '@/shared/api/client';
import { mocksReady } from '@/shared/api/mocks/enable';
import { fetchEvent } from '@/entities/event/api/events-api';
import { useAvailability } from '@/entities/event/model/use-availability';
import { WsStatus } from '@/entities/event/ui/ws-status';
import { CategoryRow } from '@/features/booking/ui/category-row';
import { publishEvent } from '@/features/organizer/model/organizer-api';
import { useAuthStore } from '@/features/auth/model/auth-store';
import { coverGradient } from '@/entities/event/lib/cover';
import { formatDateTime } from '@/shared/lib/dates';
import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { Skeleton } from '@/shared/ui/skeleton';
import { AvailabilityAnnouncer } from '@/entities/event/ui/availability-announcer';

export function EventDetail({ id }: { id: string }) {
  const queryClient = useQueryClient();
  const userId = useAuthStore((s) => s.userId);

  const { query: eventQuery, wsState } = useAvailability(id);
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

  const isOwnerDraft =
    event?.status === 'draft' && userId !== null && event.organizer_id === userId;

  const publish = useMutation({
    mutationFn: async () => {
      await mocksReady;
      return publishEvent(id);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['event', id] });
      void queryClient.invalidateQueries({ queryKey: ['events', 'mine'] });
    },
  });

  const availability = eventQuery.data?.categories;

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
      <div className="flex flex-wrap items-center gap-3">
        {event.status === 'draft' && <Badge variant="warning">Draft</Badge>}
        <span className="text-sm text-muted">{formatDateTime(event.starts_at)}</span>
        <WsStatus state={wsState} />
        {isOwnerDraft && (
          <Button size="sm" disabled={publish.isPending} onClick={() => publish.mutate()}>
            {publish.isPending ? 'Publishing…' : 'Publish'}
          </Button>
        )}
      </div>
      {publish.isError && (
        <p className="text-sm text-danger">
          {publish.error instanceof Error ? publish.error.message : 'Failed to publish'}
        </p>
      )}
      <h1 className="font-display text-3xl font-semibold">{event.title}</h1>
      {event.description && <p className="text-muted">{event.description}</p>}
      <section className="space-y-2">
        <h2 className="font-display text-sm uppercase tracking-wide text-muted">Tickets</h2>
        {availability === undefined && (
          <div className="space-y-2">
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
          </div>
        )}
        <AvailabilityAnnouncer categories={availability ?? []} />
        {availability?.map((c) => (
          <CategoryRow key={c.category_id} category={c} eventId={id} eventTitle={event.title} />
        ))}
      </section>
    </main>
  );
}
