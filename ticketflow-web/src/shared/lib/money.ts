// Money is stored in minor units (kopecks). Locale-aware formatting;
// the default is en-US until the i18n phase wires in the active locale.
export function formatMoney(minor: number, currency: string, locale = 'en-US'): string {
  const hasCents = minor % 100 !== 0;
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    currencyDisplay: 'narrowSymbol',
    minimumFractionDigits: hasCents ? 2 : 0,
    maximumFractionDigits: hasCents ? 2 : 0,
  }).format(minor / 100);
}
