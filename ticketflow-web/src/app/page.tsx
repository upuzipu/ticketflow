import Link from 'next/link';
import { Button } from '@/shared/ui/button';
import { UpcomingEvents } from '@/features/catalog/ui/upcoming-events';

export default function Home() {
  return (
    <main>
      <section className="mx-auto max-w-6xl space-y-4 px-6 pb-10 pt-20 text-center">
        <h1 className="font-display text-4xl font-bold sm:text-5xl">Tickets, live and honest</h1>
        <p className="mx-auto max-w-xl text-muted">
          Real-time availability, a 10-minute hold and an idempotent checkout. No overselling, no
          surprises.
        </p>
        <div className="flex justify-center">
          <Button size="lg" asChild>
            <Link href="/events">Browse events</Link>
          </Button>
        </div>
      </section>
      <section className="mx-auto max-w-6xl space-y-4 px-6 pb-20">
        <div className="flex items-center justify-between">
          <h2 className="font-display text-xl font-semibold">Upcoming events</h2>
          <Link href="/events" className="text-sm text-accent hover:underline">
            View all →
          </Link>
        </div>
        <UpcomingEvents limit={3} />
      </section>
    </main>
  );
}
