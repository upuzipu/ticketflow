import type { Metadata } from 'next';
import Link from 'next/link';
import { RegisterForm } from '@/shared/ui/register-form';

export const metadata: Metadata = { title: 'Sign up' };

export default function RegisterPage() {
  return (
    <main className="mx-auto w-full max-w-sm space-y-4 p-6 pt-16">
      <h1 className="font-display text-xl font-semibold">Sign up</h1>
      <RegisterForm />
      <p className="text-sm text-muted">
        Already have an account?{' '}
        <Link href="/login" className="text-accent hover:underline">
          Log in
        </Link>
      </p>
    </main>
  );
}
