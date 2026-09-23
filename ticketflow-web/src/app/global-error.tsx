'use client';

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="en">
      <body
        style={{ fontFamily: 'system-ui, sans-serif', padding: '4rem 1.5rem', textAlign: 'center' }}
      >
        <h1 style={{ fontSize: '1.25rem', fontWeight: 600 }}>Application error</h1>
        <p style={{ color: '#626a7e', fontSize: '0.875rem' }}>
          The application failed to start.
          {error.digest ? ` Reference: ${error.digest}` : undefined}
        </p>
        <button
          onClick={reset}
          style={{
            marginTop: '1rem',
            padding: '0.5rem 1rem',
            borderRadius: '0.5rem',
            background: '#7c5cff',
            color: 'white',
            border: 'none',
            cursor: 'pointer',
          }}
        >
          Try again
        </button>
      </body>
    </html>
  );
}
