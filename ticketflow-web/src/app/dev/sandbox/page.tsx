'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import type { Event } from '@/shared/api/types';
import { api } from '@/shared/api/client';
import { mocksReady } from '@/shared/api/mocks/enable';
import { formatMoney } from '@/shared/lib/money';
import { formatDateTime } from '@/shared/lib/dates';

const SCENARIOS = [
  ['default', 'Happy path'],
  ['payment-decline', '402: payment declined'],
  ['gateway-timeout', '424: gateway timeout'],
] as const;

export default function SandboxPage() {
  const [events, setEvents] = useState<Event[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [scenario, setScenario] = useState('default');

  useEffect(() => {
    setScenario(localStorage.getItem('mock:scenario') ?? 'default');
    void mocksReady.then(async () => {
      try {
        const list = await api.get<{ events: Event[]; next_cursor?: string }>('/events?limit=20');
        setEvents(list.events);
      } catch (e) {
        setError(e instanceof Error ? e.message : String(e));
      }
    });
  }, []);

  const applyScenario = (value: string) => {
    setScenario(value);
    if (value === 'default') localStorage.removeItem('mock:scenario');
    else localStorage.setItem('mock:scenario', value);
  };

  return (
    <main className="mx-auto max-w-2xl space-y-4 p-6">
      <h1 className="text-2xl font-semibold">TicketFlow — sandbox</h1>
      <p className="text-sm opacity-70">Data from MSW mocks (v1.2). F5 resets the state.</p>

      <section className="space-y-1 text-sm">
        <div className="opacity-70">Gateway scenario (affects POST /orders and pay):</div>
        {SCENARIOS.map(([value, label]) => (
          <button
            key={value}
            onClick={() => applyScenario(value)}
            className={`mr-2 rounded border px-2 py-1 ${scenario === value ? 'border-violet-500 text-violet-500' : 'opacity-60'}`}
          >
            {label}
          </button>
        ))}
      </section>

      {error && <p className="text-red-500">Error: {error}</p>}
      {!events && !error && <p>Loading…</p>}

      <ul className="divide-y">
        {(events ?? []).map((e) => {
          const min = Math.min(...e.categories.map((c) => c.price.amount));
          return (
            <li key={e.id} className="py-3">
              <Link
                href={`/dev/sandbox/${e.id}`}
                className="font-medium underline-offset-2 hover:underline"
              >
                {e.title}
              </Link>
              <div className="text-sm opacity-70">
                {formatDateTime(e.starts_at)}
                {' · from '}
                {formatMoney(min, 'RUB')}
              </div>
            </li>
          );
        })}
      </ul>
    </main>
  );
}
