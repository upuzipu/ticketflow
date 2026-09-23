import type { Metadata } from 'next';
import { EventDetail } from '@/features/catalog/ui/event-detail';

export const metadata: Metadata = { title: 'Event' };

export default async function EventPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <EventDetail id={id} />;
}
