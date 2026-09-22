import Link from 'next/link';

export const metadata = { title: 'Registration' };

export default function LoginPage() {
  return (
    <main className="mx-auto flex max-w-md flex-col items-center gap-3 p-6 pt-24 text-center">
      <h1 className="font-display text-xl font-semibold">Registration</h1>
      <p className="text-sm text-muted">Will be later</p>
      <Link href="/" className="text-sm text-accent hover:underline">
        ← Back to home
      </Link>
    </main>
  );
}
