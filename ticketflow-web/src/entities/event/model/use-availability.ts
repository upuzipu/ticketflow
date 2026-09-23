'use client';

import { useEffect, useRef, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import type { AvailabilityFrame } from '@/shared/api/types';
import { mocksReady } from '@/shared/api/mocks/enable';
import { fetchEventAvailability } from '../api/events-api';
import { AvailabilitySocket } from '@/shared/api/ws/client';

export type WsState = 'connecting' | 'open' | 'closed';

export function useAvailability(eventId: string) {
  const queryClient = useQueryClient();
  const socketRef = useRef<AvailabilitySocket | null>(null);
  const [wsState, setWsState] = useState<WsState>('connecting');

  const query = useQuery({
    queryKey: ['availability', eventId],
    queryFn: async () => {
      await mocksReady;
      return fetchEventAvailability(eventId);
    },
  });

  const frameRef = useRef((frame: AvailabilityFrame) => {
    queryClient.setQueryData(['availability', eventId], {
      event_id: eventId,
      categories: frame.categories,
    });
  });

  useEffect(() => {
    let cancelled = false;
    let pollTimer: ReturnType<typeof setInterval> | null = null;

    const startPolling = () => {
      if (pollTimer) return;
      pollTimer = setInterval(() => {
        void queryClient.invalidateQueries({ queryKey: ['availability', eventId] });
      }, 15_000);
    };
    const stopPolling = () => {
      if (pollTimer) clearInterval(pollTimer);
      pollTimer = null;
    };

    void mocksReady.then(() => {
      if (cancelled) return;
      const socket = new AvailabilitySocket(
        eventId,
        (state) => {
          if (!cancelled) setWsState(state);
        },
        (frame) => frameRef.current(frame),
        startPolling,
        stopPolling,
      );
      socket.connect();
      socketRef.current = socket;
    });

    return () => {
      cancelled = true;
      stopPolling();
      socketRef.current?.disconnect();
      socketRef.current = null;
    };
  }, [eventId, queryClient, setWsState]);

  return { query, wsState };
}
