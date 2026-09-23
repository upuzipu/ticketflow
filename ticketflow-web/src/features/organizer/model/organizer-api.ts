import { api } from '@/shared/api/client';
import type { CreateEventRequest, EventList } from '@/shared/api/types';

export function fetchMyEvents(limit: number, cursor?: string): Promise<EventList> {
  const params = new URLSearchParams({ limit: String(limit), mine: 'true' });
  if (cursor) params.set('cursor', cursor);
  return api.get(`/events?${params.toString()}`);
}

export function createEvent(
  body: CreateEventRequest,
): Promise<{ id: string; title: string; starts_at: string; status: 'draft' }> {
  return api.post('/events', body);
}

export function publishEvent(id: string): Promise<void> {
  return api.post(`/events/${id}/publish`);
}
