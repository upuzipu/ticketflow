import { describe, expect, it } from 'vitest';
import { decodeAccess } from './decode-access';

function makeToken(claims: object): string {
  const b64 = (s: string) => btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  return `hdr.${b64(JSON.stringify(claims))}.sig`;
}

describe('decodeAccess', () => {
  it('decodes valid claims', () => {
    const token = makeToken({ sub: 'u1', role: 'buyer', exp: Math.floor(Date.now() / 1000) + 60 });
    expect(decodeAccess(token)).toEqual({ sub: 'u1', role: 'buyer' });
  });

  it('rejects an expired token', () => {
    const token = makeToken({ sub: 'u1', role: 'buyer', exp: Math.floor(Date.now() / 1000) - 10 });
    expect(decodeAccess(token)).toBeNull();
  });

  it('rejects an unknown role', () => {
    const token = makeToken({ sub: 'u1', role: 'admin', exp: Math.floor(Date.now() / 1000) + 60 });
    expect(decodeAccess(token)).toBeNull();
  });

  it('returns null for garbage input', () => {
    expect(decodeAccess('not-a-token')).toBeNull();
  });
});
