'use client';

import { useEffect, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { ApiError } from '@/shared/api/client';
import { loginUser } from '../lib/auth-api';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { useAuthStore } from '@/features/auth/model/auth-store';

export function LoginForm() {
  const router = useRouter();
  const setTokens = useAuthStore((s) => s.setTokens);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [cooldown, setCooldown] = useState(0);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (cooldown <= 0) return;
    const t = setTimeout(() => setCooldown((s) => s - 1), 1000);
    return () => clearTimeout(t);
  }, [cooldown]);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    if (pending || cooldown > 0) return;
    setError(null);
    setPending(true);
    try {
      const pair = await loginUser({ email: email.trim(), password });
      setTokens(pair);
      router.push('/');
    } catch (err) {
      if (err instanceof ApiError && err.status === 429) {
        const sec = err.retryAfterSec ?? 60;
        setCooldown(sec);
        setError(`Too many attempts. Try again in ${sec} seconds.`);
      } else if (err instanceof ApiError && err.status === 401) {
        setError('Unknown email or wrong password.');
      } else {
        setError(err instanceof Error ? err.message : 'Something went wrong.');
      }
    } finally {
      setPending(false);
    }
  };

  return (
    <form onSubmit={submit} className="space-y-3">
      <Input
        type="email"
        required
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        autoComplete="email"
      />
      <Input
        type="password"
        required
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        autoComplete="current-password"
      />
      {error && <p className="text-sm text-danger">{error}</p>}
      <Button type="submit" className="w-full" disabled={pending || cooldown > 0}>
        {cooldown > 0 ? `Wait ${cooldown}s` : pending ? 'Signing in…' : 'Log in'}
      </Button>
    </form>
  );
}
