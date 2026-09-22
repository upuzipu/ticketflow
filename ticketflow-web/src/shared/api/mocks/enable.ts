let ready: Promise<void> | null = null;

function start(): Promise<void> {
  if (typeof window === 'undefined') return Promise.resolve();

  if (!ready) {
    ready = import('./browser')
      .then(({ worker }) => worker.start({ onUnhandledRequest: 'bypass' }))
      .then(() => {
        console.info('[mocks] MSW enabled');
      })
      .catch((e) => {
        console.error('[mocks] failed to start', e);
      });
  }
  return ready;
}

export const mocksReady =
  process.env.NEXT_PUBLIC_API_MOCKING === 'enabled' ? start() : Promise.resolve();
