export function formatMoney(minor: number, currency: string): string {
  const hasCents = minor % 100 !== 0;
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency,
    minimumFractionDigits: hasCents ? 2 : 0,
    maximumFractionDigits: hasCents ? 2 : 0,
  }).format(minor / 100);
}
