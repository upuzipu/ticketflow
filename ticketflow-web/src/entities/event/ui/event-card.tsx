import Image from 'next/image';
import Link from 'next/link';
import type { Event } from '@/shared/api/types';
import { coverGradient } from '../lib/cover';
import { formatDateTime } from '@/shared/lib/dates';
import { PriceTag } from '@/shared/ui/price-tag';

export function EventCard({ event }: { event: Event }) {
  const currency = event.categories[0]?.price.currency ?? 'RUB';
  const prices = event.categories.map((c) => c.price.amount);
  const min = prices.length > 0 ? Math.min(...prices) : null;

  return (
    <Link
      href={`/events/${event.id}`}
      className="group flex flex-col overflow-hidden rounded-xl border border-border bg-surface transition-colors hover:border-accent/50"
    >
      <div
        className="relative h-36 w-full"
        style={event.image_url ? undefined : { background: coverGradient(event.id) }}
      >
        {event.image_url ? (
          <Image
            src={event.image_url}
            alt={event.title}
            fill
            className="object-cover"
            sizes="(max-width: 640px) 100vw, 33vw"
          />
        ) : (
          <span className="absolute bottom-2 left-3 font-display text-3xl font-bold text-white/90">
            {event.title.charAt(0)}
          </span>
        )}
      </div>
      <div className="flex flex-1 flex-col gap-1.5 p-4">
        <h3 className="font-medium leading-snug group-hover:text-accent">{event.title}</h3>
        {event.description && (
          <p className="line-clamp-2 text-sm text-muted">{event.description}</p>
        )}
        <div className="mt-auto flex items-center justify-between gap-2 pt-2 text-sm">
          <span className="text-muted">{formatDateTime(event.starts_at)}</span>
          {min !== null && <PriceTag amountMinor={min} currency={currency} prefix="from" />}
        </div>
      </div>
    </Link>
  );
}

export function EventCardSkeleton() {
  return (
    <div className="flex flex-col overflow-hidden rounded-xl border border-border bg-surface">
      <div className="h-36 w-full animate-pulse bg-surface" />
      <div className="space-y-2 p-4">
        <div className="h-4 w-3/4 animate-pulse rounded bg-surface" />
        <div className="h-3 w-1/2 animate-pulse rounded bg-surface" />
        <div className="h-3 w-2/3 animate-pulse rounded bg-surface" />
      </div>
    </div>
  );
}
