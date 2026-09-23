'use client';

import Link from 'next/link';
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { mocksReady } from '@/shared/api/mocks/enable';
import { fetchMyEvents, publishEvent } from '../model/organizer-api';
import { coverGradient } from '@/entities/event/lib/cover';
import { formatDateTime } from '@/shared/lib/dates';
import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { Skeleton } from '@/shared/ui/skeleton';

const STATUS_BADGE = {
  draft: { label: 'Draft', variant: 'warning' },
  published: { label: 'Published', variant: 'success' },
  cancelled: { label: 'Cancelled', variant: 'danger' },
} as const;

export function MyEvents() {
  const queryClient = useQueryClient();
  const { data, status, error, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: ['events', 'mine'],
    initialPageParam: '',
    queryFn: async ({ pageParam }) => {
      await mocksReady;
      return fetchMyEvents(20, pageParam || undefined);
    },
    getNextPageParam: (lastPage) => lastPage.next_cursor || undefined,
  });

  const publish = useMutation({
    mutationFn: async (id: string) => {
      await mocksReady;
      return publishEvent(id);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['events', 'mine'] });
      void queryClient.invalidateQueries({ queryKey: ['events', 'catalog'] });
      void queryClient.invalidateQueries({ queryKey: ['events', 'upcoming'] });
    },
  });

  if (status === 'pending') {
    return (
      <div className="space-y-2">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
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

  const events = data.pages.flatMap((page) => page.events);

  if (events.length === 0) {
    return (
      <Card className="p-10 text-center text-sm text-muted">
        No events yet. Create your first one.
      </Card>
    );
  }

  return (
    <div className="space-y-3">
      {events.map((event) => {
        const badge = STATUS_BADGE[event.status];
        return (
          <Card key={event.id} className="flex items-center gap-4 p-4">
            <div
              className="hidden h-16 w-24 shrink-0 rounded-lg sm:block"
              style={{ background: coverGradient(event.id) }}
            />
            <div className="min-w-0 flex-1 space-y-1">
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant={badge.variant}>{badge.label}</Badge>
                <span className="truncate font-medium">{event.title}</span>
              </div>
              <div className="text-sm text-muted">
                {formatDateTime(event.starts_at)} · {event.categories.length} categories
              </div>
            </div>
            <div className="flex shrink-0 items-center gap-2">
              {event.status === 'draft' && (
                <Button
                  size="sm"
                  disabled={publish.isPending}
                  onClick={() => publish.mutate(event.id)}
                >
                  {publish.isPending ? 'Publishing…' : 'Publish'}
                </Button>
              )}
              <Button size="sm" variant="secondary" asChild>
                <Link href={`/events/${event.id}`}>View</Link>
              </Button>
            </div>
          </Card>
        );
      })}
      {publish.isError && (
        <p className="text-sm text-danger">
          {publish.error instanceof Error ? publish.error.message : 'Failed to publish'}
        </p>
      )}
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
