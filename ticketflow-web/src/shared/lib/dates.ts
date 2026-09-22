// Centralized date/time formatting so the i18n phase only passes the active locale.
export function formatDateTime(iso: string, locale = 'en-US'): string {
  return new Date(iso).toLocaleString(locale, { dateStyle: 'medium', timeStyle: 'short' });
}
