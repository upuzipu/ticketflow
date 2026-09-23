export function keysetPage<T>(
  items: T[],
  cursor: string | null,
  limit: number,
  keyOf: (x: T) => [Date, string],
  direction: 'after' | 'before',
): { page: T[]; nextCursor: string } {
  const [iso = '', id = ''] = (cursor ?? '').split('|');
  const t0 = Date.parse(iso);
  const c = cursor && iso && id && !Number.isNaN(t0) ? { t: t0, id } : null;
  const filtered = c
    ? items.filter((x) => {
        const [t, id2] = keyOf(x);
        return direction === 'after'
          ? t.getTime() > c.t || (t.getTime() === c.t && id2 > c.id)
          : t.getTime() < c.t || (t.getTime() === c.t && id2 < c.id);
      })
    : items;
  const page = filtered.slice(0, limit);
  const last = page[page.length - 1];
  const nextCursor =
    page.length && filtered.length > limit
      ? `${keyOf(last)[0].toISOString()}|${keyOf(last)[1]}`
      : '';
  return { page, nextCursor };
}
