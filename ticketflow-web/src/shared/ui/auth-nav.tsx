'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/features/auth/model/auth-store';
import { sessionRestore } from '@/features/auth/lib/use-session-restore';
import { Button } from '@/shared/ui/button';
import { Badge } from '@/shared/ui/badge';
import { logoutUser } from '@/shared/lib/auth-api';

export function AuthNav() {
  const router = useRouter();
  const refreshToken = useAuthStore((s) => s.refreshToken);
  const userRole = useAuthStore((s) => s.userRole);
  const [ready, setReady] = useState(false);
  const [busy, setBusy] = useState(false);

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
      <Link
        href={userRole === 'organizer' ? '/organizer' : '/orders'}
        className="text-sm text-muted transition-colors hover:text-foreground"
      >
        {userRole === 'organizer' ? 'My events' : 'My orders'}
      </Link>
      <Button variant="ghost" size="sm" disabled={busy} onClick={() => void logout()}>
        Log out
      </Button>
    </>
  );
}
