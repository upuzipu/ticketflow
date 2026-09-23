import { describe, expect, it } from 'vitest';
import { keysetPage } from './keyset';

interface Item {
  id: string;
  at: Date;
}

const keyOf = (x: Item): [Date, string] => [x.at, x.id];

const PAGES: Item[] = [
  { id: 'a', at: new Date('2026-01-01T10:00:00Z') },
  { id: 'b', at: new Date('2026-01-01T11:00:00Z') },
  { id: 'c', at: new Date('2026-01-01T12:00:00Z') },
  { id: 'd', at: new Date('2026-01-01T13:00:00Z') },
];

describe('keysetPage', () => {
  it('returns the first page and a cursor when more pages exist', () => {
    const { page, nextCursor } = keysetPage(PAGES, null, 2, keyOf, 'after');
    expect(page.map((x) => x.id)).toEqual(['a', 'b']);
    expect(nextCursor).toBe('2026-01-01T11:00:00.000Z|b');
  });

  it('pages forward after the cursor', () => {
    const { page, nextCursor } = keysetPage(PAGES, '2026-01-01T11:00:00.000Z|b', 2, keyOf, 'after');
    expect(page.map((x) => x.id)).toEqual(['c', 'd']);
    expect(nextCursor).toBe('');
  });

  it('pages backward before the cursor, newest first', () => {
    const newestFirst: Item[] = [...PAGES].reverse();
    const { page, nextCursor } = keysetPage(
      newestFirst,
      '2026-01-01T13:00:00.000Z|d',
      2,
      keyOf,
      'before',
    );
    expect(page.map((x) => x.id)).toEqual(['c', 'b']);
    expect(nextCursor).toBe('2026-01-01T11:00:00.000Z|b');
  });

  it('treats an invalid cursor as no cursor', () => {
    const { page, nextCursor } = keysetPage(PAGES, 'garbage', 2, keyOf, 'after');
    expect(page.map((x) => x.id)).toEqual(['a', 'b']);
    expect(nextCursor).toBe('2026-01-01T11:00:00.000Z|b');
  });

  it('uses the id as a tie-breaker for equal timestamps', () => {
    const tied: Item[] = [
      { id: 'x2', at: new Date('2026-01-01T10:00:00Z') },
      { id: 'x1', at: new Date('2026-01-01T10:00:00Z') },
      { id: 'x3', at: new Date('2026-01-01T10:00:00Z') },
      { id: 'x4', at: new Date('2026-01-01T10:00:00Z') },
    ];
    const { page, nextCursor } = keysetPage(tied, '2026-01-01T10:00:00.000Z|x1', 2, keyOf, 'after');
    expect(page.map((x) => x.id)).toEqual(['x2', 'x3']);
    expect(nextCursor).toBe('2026-01-01T10:00:00.000Z|x3');
  });

  it('returns an empty cursor when exactly limit items remain', () => {
    const tied: Item[] = [
      { id: 'x2', at: new Date('2026-01-01T10:00:00Z') },
      { id: 'x1', at: new Date('2026-01-01T10:00:00Z') },
      { id: 'x3', at: new Date('2026-01-01T10:00:00Z') },
    ];
    const { page, nextCursor } = keysetPage(tied, '2026-01-01T10:00:00.000Z|x1', 2, keyOf, 'after');
    expect(page.map((x) => x.id)).toEqual(['x2', 'x3']);
    expect(nextCursor).toBe('');
  });
});
