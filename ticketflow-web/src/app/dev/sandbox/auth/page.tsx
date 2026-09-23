'use client';

import { useState } from 'react';
import Link from 'next/link';
import { api } from '@/shared/api/client';
import { useAuthStore } from '@/features/auth/model/auth-store';
import { Button } from '@/shared/ui/button';
import { Badge } from '@/shared/ui/badge';
import type { Me } from '@/shared/api/types';
import { loginUser, logoutUser } from '@/shared/lib/auth-api';

export default function SandboxAuthPage() {
  const accessToken = useAuthStore((s) => s.accessToken);
  const refreshToken = useAuthStore((s) => s.refreshToken);
  const userRole = useAuthStore((s) => s.userRole);
  const [log, setLog] = useState<string[]>([]);
  const [parallelResult, setParallelResult] = useState<string | null>(null);

  const append = (line: string) => setLog((l) => [...l.slice(-9), line]);

  const loginAs = async (email: string) => {
    try {
      const pair = await loginUser({ email, password: 'password' });
      useAuthStore.getState().setTokens(pair);
      append(`logged in as ${email}`);
    } catch (err) {
      append(`login failed: ${err instanceof Error ? err.message : 'unknown'}`);
    }
  };

  const logout = async () => {
    const { refreshToken: rt } = useAuthStore.getState();
    if (!rt) return;
    await logoutUser(rt).catch(() => undefined);
    useAuthStore.getState().clear();
    append('logged out');
  };

  const expireAccess = () => {
    useAuthStore.setState({ accessToken: null });
    append('access token cleared (simulated expiry)');
  };

  const fireParallel = async () => {
    useAuthStore.setState({ accessToken: null });
    const results = await Promise.all(
      Array.from({ length: 5 }, () =>
        api
          .get<Me>('/users/me')
          .then(() => 'ok')
          .catch(() => 'fail'),
      ),
    );
    const ok = results.filter((r) => r === 'ok').length;
    setParallelResult(`${ok}/5 succeeded. Network tab: /auth/refresh must appear exactly once.`);
    append(`parallel burst: ${ok}/5 ok`);
  };

  return (
    <main className="mx-auto max-w-2xl space-y-6 p-6">
      <div>
        <Link href="/dev/sandbox" className="text-sm text-muted hover:text-foreground">
          ← sandbox
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">Auth testing panel</h1>
        <p className="text-sm text-muted">
          Demo users: buyer@tf.dev and org@tf.dev, password “password”.
        </p>
      </div>

      <section className="space-y-2 rounded-xl border border-border bg-surface p-4">
        <div className="flex flex-wrap items-center gap-2 text-sm">
          <Badge variant={refreshToken ? 'success' : 'muted'}>
            {refreshToken ? 'refresh token present' : 'anonymous'}
          </Badge>
          <Badge variant={accessToken ? 'success' : 'warning'}>
            {accessToken ? 'access token present' : 'no access token'}
          </Badge>
          {userRole && <Badge variant="accent">{userRole}</Badge>}
        </div>
        <div className="flex flex-wrap gap-2 pt-1">
          <Button size="sm" variant="secondary" onClick={() => void loginAs('buyer@tf.dev')}>
            Login as buyer
          </Button>
          <Button size="sm" variant="secondary" onClick={() => void loginAs('org@tf.dev')}>
            Login as organizer
          </Button>
          <Button size="sm" variant="ghost" onClick={() => void logout()}>
            Logout
          </Button>
          <Button size="sm" variant="ghost" asChild>
            <Link href="/dev/sandbox/protected">Protected page demo →</Link>
          </Button>
        </div>
      </section>

      <section className="space-y-2 rounded-xl border border-border bg-surface p-4">
        <h2 className="text-sm font-medium">Refresh chain</h2>
        <p className="text-sm text-muted">
          Clearing the access token simulates its 15-minute expiry. The next request must
          transparently refresh and retry.
        </p>
        <div className="flex flex-wrap gap-2">
          <Button size="sm" variant="outline" onClick={expireAccess}>
            Expire access token
          </Button>
          <Button size="sm" variant="outline" onClick={() => void fireParallel()}>
            Fire 5 parallel /users/me
          </Button>
        </div>
        {parallelResult && <p className="text-sm">{parallelResult}</p>}
      </section>

      <section className="space-y-1 rounded-xl border border-border bg-surface p-4">
        <h2 className="text-sm font-medium">Event log</h2>
        {log.length === 0 ? (
          <p className="text-sm text-muted">No events yet.</p>
        ) : (
          <ul className="space-y-1 text-sm text-muted">
            {log.map((line, i) => (
              <li key={`${i}-${line}`}>· {line}</li>
            ))}
          </ul>
        )}
      </section>
    </main>
  );
}
