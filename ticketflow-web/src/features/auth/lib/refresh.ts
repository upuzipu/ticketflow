import { useAuthStore } from '../model/auth-store';
import type { TokenPair } from '@/shared/api/types';

const BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? '/api';

let inFlight: Promise<void> | null = null;

export function refreshTokens(): Promise<void> {
  if (inFlight) return inFlight;

  inFlight = (async () => {
    const { refreshToken, setTokens, clear } = useAuthStore.getState();
    if (!refreshToken) throw new Error('no refresh token');

    console.info('[auth] refresh: start');
    const res = await fetch(`${BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    console.info('[auth] refresh:', res.status);

    if (!res.ok) {
      if (res.status === 401) {
        clear();
        console.warn('[auth] refresh rejected (401) — session cleared');
      } else {
        console.warn('[auth] refresh failed transiently — session kept');
      }
      throw new Error(`refresh failed: ${res.status}`);
    }

    setTokens((await res.json()) as TokenPair);
    console.info('[auth] refresh: ok, new pair stored');
  })().finally(() => {
    inFlight = null;
  });

  return inFlight;
}
