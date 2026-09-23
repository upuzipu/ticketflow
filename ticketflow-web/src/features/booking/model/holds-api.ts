import { api } from '@/shared/api/client';
import type { HoldCreated, HoldRequest, HoldStatusResponse } from '@/shared/api/types';

export function createHold(eventId: string, body: HoldRequest): Promise<HoldCreated> {
  return api.post(`/events/${eventId}/holds`, body);
}

export function fetchHold(id: string): Promise<HoldStatusResponse> {
  return api.get(`/holds/${id}`);
}
