import type { Metadata } from 'next';
import Link from 'next/link';
import { MyEvents } from '@/features/organizer/ui/my-events';
import { Button } from '@/shared/ui/button';
import { RequireAuth } from '@/shared/ui/require-auth';

export const metadata: Metadata = { title: 'My events' };

export default function OrganizerPage() {
  return (
    <RequireAuth role="organizer">
      <main className="mx-auto max-w-4xl space-y-5 p-6">
        <header className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="font-display text-2xl font-semibold">My events</h1>
            <p className="text-sm text-muted">Drafts, published and cancelled events.</p>
          </div>
          <Button asChild>
            <Link href="/organizer/events/new">New event</Link>
          </Button>
        </header>
        <MyEvents />
      </main>
    </RequireAuth>
  );
}
