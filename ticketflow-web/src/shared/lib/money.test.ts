import { describe, expect, it } from 'vitest';
import { formatMoney } from './money';

describe('formatMoney', () => {
  it('formats whole amounts without kopecks', () => {
    expect(formatMoney(600000, 'RUB', 'en-US')).toBe('₽6,000');
  });

  it('formats minor units with two decimals', () => {
    expect(formatMoney(150550, 'RUB', 'en-US')).toBe('₽1,505.50');
  });

  it('formats zero', () => {
    expect(formatMoney(0, 'RUB', 'en-US')).toBe('₽0');
  });
});
