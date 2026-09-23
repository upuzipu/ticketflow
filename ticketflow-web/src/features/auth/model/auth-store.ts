import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { TokenPair, UserRole } from '@/shared/api/types';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  userId: string | null;
  userRole: UserRole | null;
  setTokens(pair: TokenPair): void;
  clear(): void;
}

type AccessClaims = { sub: string; role: UserRole };

function decodeAccess(access: string): AccessClaims | null {
  try {
    const [, payload] = access.split('.');
    const json = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    const claims = JSON.parse(json) as { sub?: string; role?: string; exp?: number };
    if (!claims.sub || (claims.role !== 'buyer' && claims.role !== 'organizer')) return null;
    if ((claims.exp ?? 0) * 1000 < Date.now()) return null;
    return { sub: claims.sub, role: claims.role };
  } catch {
    return null;
  }
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      userId: null,
      userRole: null,
      setTokens: (pair) => {
        const claims = decodeAccess(pair.access);
        set({
          accessToken: pair.access,
          refreshToken: pair.refresh,
          userId: claims?.sub ?? null,
          userRole: claims?.role ?? null,
        });
      },
      clear: () => set({ accessToken: null, refreshToken: null, userId: null, userRole: null }),
    }),
    {
      name: 'ticketflow.auth',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        refreshToken: state.refreshToken,
        userRole: state.userRole,
      }),
    },
  ),
);
