import { useAuthStore } from '../model/auth-store';
import type { TokenPair } from '@/shared/api/types';

const BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? '/api';

let inFlight: Promise<void> | null = null;

export function refreshTokens(): Promise<void> {
  if (inFlight) return inFlight;

  inFlight = (async () => {
    const { refreshToken, setTokens, clear } = useAuthStore.getState();
    if (!refreshToken) throw new Error('no refresh token');

    const res = await fetch(`${BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!res.ok) {
      if (res.status === 401) {
        clear();
      }
      throw new Error(`refresh failed: ${res.status}`);
    }

    setTokens((await res.json()) as TokenPair);
  })().finally(() => {
    inFlight = null;
  });

  return inFlight;
}
