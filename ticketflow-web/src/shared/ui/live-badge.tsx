import { cn } from '@/shared/lib/cn';

interface LiveBadgeProps {
  available: number;
  total?: number;
  className?: string;
}

export function LiveBadge({ available, total, className }: LiveBadgeProps) {
  const soldOut = available <= 0;
  const low =
    !soldOut && (total == null ? available <= 5 : available <= Math.max(1, Math.ceil(total * 0.1)));
  const dot = soldOut ? 'bg-danger' : low ? 'bg-warning' : 'bg-success';
  const tone = soldOut ? 'text-danger' : low ? 'text-warning' : 'text-success';

  return (
    <span className={cn('inline-flex items-center gap-1.5 text-xs font-medium', tone, className)}>
      <span className="relative flex h-2 w-2">
        {!soldOut && (
          <span
            className={cn(
              'absolute inline-flex h-full w-full animate-ping rounded-full opacity-60',
              dot,
            )}
          />
        )}
        <span
          className={cn('relative inline-flex h-2 w-2 rounded-full', dot, soldOut && 'opacity-60')}
        />
      </span>
      {soldOut ? 'Распродано' : `${available} свободно`}
    </span>
  );
}
