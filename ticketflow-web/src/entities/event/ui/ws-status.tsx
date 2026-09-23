import type { WsState } from '../model/use-availability';
import { cn } from '@/shared/lib/cn';

export function WsStatus({ state }: { state: WsState }) {
  const dot = state === 'open' ? 'bg-success' : state === 'connecting' ? 'bg-warning' : 'bg-danger';
  const label =
    state === 'open' ? 'Live' : state === 'connecting' ? 'Connecting…' : 'Offline — retrying';
  return (
    <span className={cn('inline-flex items-center gap-1.5 text-xs font-medium', 'text-muted')}>
      <span className={cn('relative flex h-2 w-2')}>
        {state === 'open' && (
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-success opacity-60" />
        )}
        <span className={cn('relative inline-flex h-2 w-2 rounded-full', dot)} />
      </span>
      {label}
    </span>
  );
}
