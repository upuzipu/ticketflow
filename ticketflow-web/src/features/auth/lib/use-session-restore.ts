'use client';

import { useEffect } from 'react';
import { mocksReady } from '@/shared/api/mocks/enable';
import { useAuthStore } from '../model/auth-store';
import { refreshTokens } from './refresh';

let restore: Promise<void> | null = null;

export function sessionRestore(): Promise<void> {
  if (!restore) {
    restore = (async () => {
      console.info('[auth] session restore: requested');
      await mocksReady;
      const { accessToken, refreshToken } = useAuthStore.getState();
      if (accessToken || !refreshToken) return;
      await refreshTokens().catch(() => undefined);
    })();
  }
  return restore;
}

export function useSessionRestore(): void {
  useEffect(() => {
    void sessionRestore();
  }, []);
}
