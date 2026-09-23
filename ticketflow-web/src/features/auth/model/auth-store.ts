import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { TokenPair, UserRole } from '@/shared/api/types';
import { decodeAccess } from '@/features/auth/lib/decode-access';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  userId: string | null;
  userRole: UserRole | null;
  setTokens(pair: TokenPair): void;
  clear(): void;
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
