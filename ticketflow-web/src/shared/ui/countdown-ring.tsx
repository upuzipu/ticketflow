import { cn } from '@/shared/lib/cn';

interface CountdownRingProps {
  secondsLeft: number;
  totalSeconds: number;
  size?: number;
  strokeWidth?: number;
  className?: string;
}

export function CountdownRing({
  secondsLeft,
  totalSeconds,
  size = 56,
  strokeWidth = 4,
  className,
}: CountdownRingProps) {
  const clamped = Math.max(0, Math.min(secondsLeft, totalSeconds));
  const progress = totalSeconds > 0 ? clamped / totalSeconds : 0;
  const r = (size - strokeWidth) / 2;
  const c = 2 * Math.PI * r;
  const urgent = secondsLeft <= 60; // последняя минута — краснеет
  const label = `${Math.floor(clamped / 60)}:${String(clamped % 60).padStart(2, '0')}`;

  return (
    <div
      className={cn('relative inline-flex', className)}
      role="timer"
      aria-label={`Осталось ${label}`}
    >
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          fill="none"
          strokeWidth={strokeWidth}
          className="stroke-border"
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          fill="none"
          strokeWidth={strokeWidth}
          strokeLinecap="round"
          strokeDasharray={c}
          strokeDashoffset={c * (1 - progress)}
          className={cn(
            'transition-[stroke-dashoffset] duration-1000 ease-linear',
            urgent ? 'stroke-danger' : 'stroke-accent',
          )}
        />
      </svg>
      <span
        className={cn(
          'absolute inset-0 grid place-items-center text-xs font-semibold tabular-nums',
          urgent ? 'text-danger' : 'text-foreground',
        )}
      >
        {label}
      </span>
    </div>
  );
}
