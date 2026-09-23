import type { Metadata } from 'next';
import { EventForm } from '@/features/organizer/ui/event-form';
import { RequireAuth } from '@/shared/ui/require-auth';

export const metadata: Metadata = { title: 'New event' };

export default function NewEventPage() {
  return (
    <RequireAuth role="organizer">
      <main className="mx-auto max-w-2xl space-y-5 p-6">
        <h1 className="font-display text-2xl font-semibold">New event</h1>
        <p className="text-sm text-muted">
          The event is created as a draft. Publish it from “My events” when ready.
        </p>
        <EventForm />
      </main>
    </RequireAuth>
  );
}
