'use client';

import { useEffect, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { ApiError } from '@/shared/api/client';
import { registerUser } from '../lib/auth-api';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import type { UserRole } from '@/shared/api/types';
import { useAuthStore } from '@/features/auth/model/auth-store';

const ROLES: Array<{ value: UserRole; label: string; hint: string }> = [
  { value: 'buyer', label: 'Buyer', hint: 'Buy tickets' },
  { value: 'organizer', label: 'Organizer', hint: 'Create and publish events' },
];

export function RegisterForm() {
  const router = useRouter();
  const setTokens = useAuthStore((s) => s.setTokens);
  const [role, setRole] = useState<UserRole>('buyer');
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
    if (password.length < 8) {
      setError('Password must be at least 8 characters.');
      return;
    }
    setPending(true);
    try {
      const registered = await registerUser({ email: email.trim(), password, role });
      setTokens({ access: registered.access, refresh: registered.refresh });
      router.push('/');
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setError('This email is already registered.');
      } else if (err instanceof ApiError && err.status === 429) {
        const sec = err.retryAfterSec ?? 60;
        setCooldown(sec);
        setError(`Too many attempts. Try again in ${sec} seconds.`);
      } else {
        setError(err instanceof Error ? err.message : 'Something went wrong.');
      }
    } finally {
      setPending(false);
    }
  };

  return (
    <form onSubmit={submit} className="space-y-3">
      <div className="grid grid-cols-2 gap-2" role="radiogroup" aria-label="Account type">
        {ROLES.map((r) => (
          <button
            key={r.value}
            type="button"
            role="radio"
            aria-checked={role === r.value}
            onClick={() => setRole(r.value)}
            className={`rounded-lg border p-3 text-left transition-colors ${
              role === r.value ? 'border-accent bg-accent/10' : 'border-border hover:bg-surface'
            }`}
          >
            <div className="text-sm font-medium">{r.label}</div>
            <div className="text-xs text-muted">{r.hint}</div>
          </button>
        ))}
      </div>
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
        placeholder="Password (min 8 characters)"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        autoComplete="new-password"
      />
      {error && <p className="text-sm text-danger">{error}</p>}
      <Button type="submit" className="w-full" disabled={pending || cooldown > 0}>
        {cooldown > 0 ? `Wait ${cooldown}s` : pending ? 'Creating account…' : 'Sign up'}
      </Button>
    </form>
  );
}
