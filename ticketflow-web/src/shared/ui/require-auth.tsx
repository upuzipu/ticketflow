'use client';

import { useEffect, useState, type ReactNode } from 'react';
import Link from 'next/link';
import type { UserRole } from '@/shared/api/types';
import { useAuthStore } from '@/features/auth/model/auth-store';
import { sessionRestore } from '@/features/auth/lib/use-session-restore';

function AccessDenied({ message }: { message: string }) {
  return (
    <main className="mx-auto max-w-md space-y-3 p-6 pt-24 text-center">
      <h1 className="font-display text-xl font-semibold">Access restricted</h1>
      <p className="text-sm text-muted">{message}</p>
      <Link href="/login" className="text-sm text-accent hover:underline">
        Go to log in
      </Link>
    </main>
  );
}

export function RequireAuth({ role, children }: { role?: UserRole; children: ReactNode }) {
  const refreshToken = useAuthStore((s) => s.refreshToken);
  const userId = useAuthStore((s) => s.userId);
  const userRole = useAuthStore((s) => s.userRole);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    let cancelled = false;
    void sessionRestore().finally(() => {
      if (!cancelled) setReady(true);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  if (!ready) {
    return <div className="p-6 text-sm text-muted">Checking session…</div>;
  }
  if (!refreshToken || !userId) {
    return <AccessDenied message="You need to log in to view this page." />;
  }
  if (role && userRole !== role) {
    return <AccessDenied message="This page is only available to organizers." />;
  }
  return <>{children}</>;
}
