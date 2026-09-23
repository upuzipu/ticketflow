'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { mocksReady } from '@/shared/api/mocks/enable';
import { fetchEvents } from '@/entities/event/api/events-api';

export function useEventsCatalog(limit: number) {
  return useInfiniteQuery({
    queryKey: ['events', 'catalog', limit],
    initialPageParam: '',
    queryFn: async ({ pageParam }) => {
      await mocksReady;
      return fetchEvents(limit, pageParam || undefined);
    },
    getNextPageParam: (lastPage) => lastPage.next_cursor || undefined,
  });
}
