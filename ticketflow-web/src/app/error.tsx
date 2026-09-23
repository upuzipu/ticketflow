'use client';

import { useEffect } from 'react';
import { Button } from '@/shared/ui/button';
import Link from 'next/link';

export default function RouteError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <main className="mx-auto max-w-md space-y-3 p-6 pt-24 text-center">
      <h1 className="font-display text-xl font-semibold">Something went wrong</h1>
      <p className="text-sm text-muted">An unexpected error occurred. It has been logged.</p>
      <div className="flex justify-center gap-2">
        <Button onClick={reset}>Try again</Button>
        <Button variant="ghost" asChild>
          <Link href="/">Go home</Link>
        </Button>
      </div>
    </main>
  );
}
