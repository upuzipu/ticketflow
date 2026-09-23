'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { logoutUser } from '../lib/auth-api';
import { Button } from '@/shared/ui/button';
import { Badge } from '@/shared/ui/badge';
import { useAuthStore } from '@/features/auth/model/auth-store';
import { sessionRestore, useSessionRestore } from '@/features/auth/lib/use-session-restore';

export function AuthNav() {
  const router = useRouter();
  const refreshToken = useAuthStore((s) => s.refreshToken);
  const userRole = useAuthStore((s) => s.userRole);
  const [ready, setReady] = useState(false);
  const [busy, setBusy] = useState(false);

  useSessionRestore();
  useEffect(() => {
    let cancelled = false;
    void sessionRestore().finally(() => {
      if (!cancelled) setReady(true);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  if (!ready || !refreshToken) {
    return (
      <>
        <Button variant="ghost" size="sm" asChild>
          <Link href="/login">Log in</Link>
        </Button>
        <Button size="sm" asChild>
          <Link href="/register">Sign up</Link>
        </Button>
      </>
    );
  }

  const logout = async () => {
    setBusy(true);
    const { refreshToken: rt } = useAuthStore.getState();
    if (rt) await logoutUser(rt).catch(() => undefined);
    useAuthStore.getState().clear();
    setBusy(false);
    router.push('/');
  };

  return (
    <>
      {userRole && (
        <Badge variant={userRole === 'organizer' ? 'accent' : 'muted'}>{userRole}</Badge>
      )}
      <Button variant="ghost" size="sm" disabled={busy} onClick={() => void logout()}>
        Log out
      </Button>
    </>
  );
}
