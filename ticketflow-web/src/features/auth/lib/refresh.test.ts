import { beforeEach, describe, expect, it, vi } from 'vitest';

function stubResponse(status: number, body?: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body ?? {}),
  } as unknown as Response;
}

async function loadFreshModules() {
  vi.resetModules();
  const refresh = await import('./refresh');
  const store = await import('../model/auth-store');
  return { refreshTokens: refresh.refreshTokens, useAuthStore: store.useAuthStore };
}

beforeEach(() => {
  vi.unstubAllGlobals();
});

describe('refreshTokens', () => {
  it('collapses concurrent calls into a single request', async () => {
    const fetchMock = vi.fn(() =>
      Promise.resolve(stubResponse(200, { access: 'a', refresh: 'r' })),
    );
    vi.stubGlobal('fetch', fetchMock);

    const { refreshTokens, useAuthStore } = await loadFreshModules();
    useAuthStore.getState().clear();
    useAuthStore.setState({ refreshToken: 'old-token' });

    await Promise.all([refreshTokens(), refreshTokens(), refreshTokens()]);

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(useAuthStore.getState().accessToken).toBe('a');
    expect(useAuthStore.getState().refreshToken).toBe('r');
  });

  it('clears the session when refresh is rejected with 401', async () => {
    const fetchMock = vi.fn(() =>
      Promise.resolve(stubResponse(401, { error: 'domain: invalid token' })),
    );
    vi.stubGlobal('fetch', fetchMock);

    const { refreshTokens, useAuthStore } = await loadFreshModules();
    useAuthStore.getState().clear();
    useAuthStore.setState({ refreshToken: 'revoked' });

    await expect(refreshTokens()).rejects.toThrow('refresh failed: 401');
    expect(useAuthStore.getState().refreshToken).toBeNull();
  });

  it('keeps the session on a transient failure', async () => {
    const fetchMock = vi.fn(() => Promise.resolve(stubResponse(500)));
    vi.stubGlobal('fetch', fetchMock);

    const { refreshTokens, useAuthStore } = await loadFreshModules();
    useAuthStore.getState().clear();
    useAuthStore.setState({ refreshToken: 'keep-me' });

    await expect(refreshTokens()).rejects.toThrow('refresh failed: 500');
    expect(useAuthStore.getState().refreshToken).toBe('keep-me');
  });

  it('fails fast when no refresh token is stored', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);

    const { refreshTokens, useAuthStore } = await loadFreshModules();
    useAuthStore.getState().clear();

    await expect(refreshTokens()).rejects.toThrow('no refresh token');
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
