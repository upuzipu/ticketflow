'use client';

import { useEffect, useRef } from 'react';
import { useEventsCatalog } from '../model/use-events-catalog';
import { EventCard, EventCardSkeleton } from '@/entities/event/ui/event-card';
import { Button } from '@/shared/ui/button';

function SkeletonGrid({ count = 6 }: { count?: number }) {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: count }, (_, i) => (
        <EventCardSkeleton key={i} />
      ))}
    </div>
  );
}

export function EventsCatalog({ limit = 9 }: { limit?: number }) {
  const { data, status, error, refetch, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useEventsCatalog(limit);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && hasNextPage && !isFetchingNextPage) {
          void fetchNextPage();
        }
      },
      { rootMargin: '400px' },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [fetchNextPage, hasNextPage, isFetchingNextPage]);

  if (status === 'pending') return <SkeletonGrid count={6} />;

  if (status === 'error') {
    return (
      <div className="space-y-3 rounded-xl border border-border bg-surface p-6 text-center">
        <p className="text-sm text-danger">
          {error instanceof Error ? error.message : 'Failed to load events'}
        </p>
        <Button variant="secondary" size="sm" onClick={() => void refetch()}>
          Try again
        </Button>
      </div>
    );
  }

  const events = data.pages.flatMap((page) => page.events);

  if (events.length === 0) {
    return (
      <div className="rounded-xl border border-border bg-surface p-10 text-center text-sm text-muted">
        No events yet. Check back soon.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {events.map((event) => (
          <EventCard key={event.id} event={event} />
        ))}
      </div>
      {isFetchingNextPage && <SkeletonGrid count={3} />}
      <div ref={sentinelRef} />
      {hasNextPage && !isFetchingNextPage && (
        <div className="flex justify-center">
          <Button variant="secondary" onClick={() => void fetchNextPage()}>
            Load more
          </Button>
        </div>
      )}
    </div>
  );
}
