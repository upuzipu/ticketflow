import { api } from '@/shared/api/client';
import type { Availability, Event, EventList } from '@/shared/api/types';

export function fetchEvents(limit: number, cursor?: string): Promise<EventList> {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) params.set('cursor', cursor);
  return api.get(`/events?${params.toString()}`);
}

export function fetchEvent(id: string): Promise<Event> {
  return api.get(`/events/${id}`);
}

export function fetchEventAvailability(id: string): Promise<Availability> {
  return api.get(`/events/${id}/availability`);
}
