import { describe, expect, it } from 'vitest';
import { computeSecondsLeft } from './hold-clock';

const NOW = Date.parse('2026-01-01T12:00:00Z');

describe('computeSecondsLeft', () => {
  it('subtracts the server clock offset', () => {
    expect(computeSecondsLeft('2026-01-01T12:01:05Z', '2026-01-01T12:00:05Z', NOW)).toBe(60);
  });

  it('works without server time', () => {
    expect(computeSecondsLeft('2026-01-01T12:00:30Z', null, NOW)).toBe(30);
  });

  it('clamps to zero when the deadline has passed', () => {
    expect(computeSecondsLeft('2026-01-01T11:59:00Z', null, NOW)).toBe(0);
  });
});
