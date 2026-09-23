'use client';

import { useQuery } from '@tanstack/react-query';
import { mocksReady } from '@/shared/api/mocks/enable';
import { fetchEvents } from '@/entities/event/api/events-api';
import { EventCard, EventCardSkeleton } from '@/entities/event/ui/event-card';

export function UpcomingEvents({ limit = 3 }: { limit?: number }) {
  const { data, status } = useQuery({
    queryKey: ['events', 'upcoming', limit],
    queryFn: async () => {
      await mocksReady;
      return fetchEvents(limit);
    },
  });

  if (status !== 'success') {
    return (
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: limit }, (_, i) => (
          <EventCardSkeleton key={i} />
        ))}
      </div>
    );
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {data.events.map((event) => (
        <EventCard key={event.id} event={event} />
      ))}
    </div>
  );
}
