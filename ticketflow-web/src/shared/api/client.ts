import type { ApiErrorBody } from './types';
import { useAuthStore } from '@/features/auth/model/auth-store';
import { refreshTokens } from '@/features/auth/lib/refresh';

const BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? '/api';

export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
    readonly retryAfterSec?: number,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function toApiError(res: Response): Promise<ApiError> {
  const header = res.headers.get('retry-after');
  const parsed = header === null ? NaN : Number(header);
  const retryAfterSec = Number.isFinite(parsed) ? parsed : undefined;
  let message = res.statusText || `HTTP ${res.status}`;
  const body: unknown = await res.json().catch(() => null);
  if (body !== null && typeof body === 'object' && 'error' in body) {
    const text = (body as ApiErrorBody).error;
    if (text) message = text;
  }
  return new ApiError(res.status, message, retryAfterSec);
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const doFetch = (token: string | null) =>
    fetch(`${BASE}${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(init?.headers as Record<string, string> | undefined),
      },
    });

  let res = await doFetch(useAuthStore.getState().accessToken);

  if (res.status === 401 && !path.startsWith('/auth/')) {
    await refreshTokens().catch(() => undefined);
    res = await doFetch(useAuthStore.getState().accessToken);
  }

  if (!res.ok) throw await toApiError(res);
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export const api = {
  get: <T>(path: string) => apiFetch<T>(path),
  post: <T>(path: string, body?: unknown) =>
    apiFetch<T>(path, {
      method: 'POST',
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  delete: <T>(path: string) => apiFetch<T>(path, { method: 'DELETE' }),
};
