import type { Metadata } from 'next';
import Link from 'next/link';
import { LoginForm } from '@/shared/ui/login-form';

export const metadata: Metadata = { title: 'Log in' };

export default function LoginPage() {
  return (
    <main className="mx-auto w-full max-w-sm space-y-4 p-6 pt-16">
      <h1 className="font-display text-xl font-semibold">Log in</h1>
      <LoginForm />
      <p className="text-sm text-muted">
        No account?{' '}
        <Link href="/register" className="text-accent hover:underline">
          Sign up
        </Link>
      </p>
    </main>
  );
}
