import type { UserRole } from '@/shared/api/types';

export type AccessClaims = { sub: string; role: UserRole };

export function decodeAccess(access: string): AccessClaims | null {
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
