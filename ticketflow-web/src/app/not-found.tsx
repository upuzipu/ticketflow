import Link from 'next/link';
import { Button } from '@/shared/ui/button';

export default function NotFound() {
  return (
    <main className="mx-auto max-w-md space-y-3 p-6 pt-24 text-center">
      <h1 className="font-display text-2xl font-semibold">Page not found</h1>
      <p className="text-sm text-muted">The page you are looking for does not exist.</p>
      <Button asChild>
        <Link href="/">Back to home</Link>
      </Button>
    </main>
  );
}
