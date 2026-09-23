import type { Metadata } from 'next';
import { EventsCatalog } from '@/features/catalog/ui/events-catalog';

export const metadata: Metadata = {
  title: 'Events',
  description: 'Upcoming events with live ticket availability.',
};

export default function EventsPage() {
  return (
    <main className="mx-auto max-w-6xl space-y-6 p-6">
      <header className="space-y-1">
        <h1 className="font-display text-2xl font-semibold">Events</h1>
        <p className="text-sm text-muted">Upcoming events with live availability.</p>
      </header>
      <EventsCatalog />
    </main>
  );
}
