import { formatMoney } from '@/shared/lib/money';
import { cn } from '@/shared/lib/cn';

interface PriceTagProps {
  amountMinor: number;
  currency: string;
  prefix?: string; // например «от»
  className?: string;
}

export function PriceTag({ amountMinor, currency, prefix, className }: PriceTagProps) {
  return (
    <span className={cn('font-semibold', className)}>
      {prefix && <span className="mr-1 text-xs font-normal text-muted">{prefix}</span>}
      {formatMoney(amountMinor, currency)}
    </span>
  );
}
