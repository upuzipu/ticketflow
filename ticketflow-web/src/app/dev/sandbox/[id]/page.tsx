'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import type { AvailabilityFrame, CategoryAvailability, Event } from '@/shared/api/types';
import { api } from '@/shared/api/client';
import { mocksReady } from '@/shared/api/mocks/enable';
import { formatMoney } from '@/shared/lib/money';

export default function SandboxEventPage() {
  const { id } = useParams<{ id: string }>();
  const [event, setEvent] = useState<Event | null>(null);
  const [categories, setCategories] = useState<CategoryAvailability[]>([]);
  const [serverTime, setServerTime] = useState<string | null>(null);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    let socket: WebSocket | null = null;
    let cancelled = false;

    void mocksReady.then(async () => {
      try {
        const detail = await api.get<Event>(`/events/${id}`);
        if (!cancelled) setEvent(detail);
        const avail = await api.get<{ categories: CategoryAvailability[] }>(
          `/events/${id}/availability`,
        );
        if (!cancelled) setCategories(avail.categories);
      } catch {
        // 404 в sandbox — просто пустая страница
      }

      const origin = process.env.NEXT_PUBLIC_WS_ORIGIN ?? 'ws://localhost:8080';
      socket = new WebSocket(`${origin}/ws/events/${id}`);
      socket.onopen = () => setConnected(true);
      socket.onclose = () => setConnected(false);
      socket.onmessage = (e) => {
        const frame = JSON.parse(e.data) as AvailabilityFrame;
        if (frame.type === 'availability') {
          setCategories(frame.categories);
          setServerTime(frame.server_time);
        }
      };
    });

    return () => {
      cancelled = true;
      socket?.close();
    };
  }, [id]);

  return (
    <main className="mx-auto max-w-2xl space-y-4 p-6">
      <Link href="/dev/sandbox" className="text-sm opacity-60 hover:opacity-100">
        ← назад
      </Link>
      {event ? (
        <>
          <h1 className="text-2xl font-semibold">{event.title}</h1>
          {event.description && <p className="text-sm opacity-70">{event.description}</p>}
          <p className="flex items-center gap-2 text-sm">
            <span
              className={`inline-block h-2 w-2 rounded-full ${connected ? 'bg-emerald-500' : 'bg-gray-500'}`}
            />
            {connected ? 'WS: подключено' : 'WS: нет соединения'}
            {serverTime && <span className="opacity-50">· server_time: {serverTime}</span>}
          </p>
          <ul className="divide-y">
            {categories.map((c) => (
              <li key={c.category_id} className="flex items-center justify-between py-3">
                <div>
                  <div className="font-medium">{c.name}</div>
                  <div className="text-sm opacity-70">{formatMoney(c.price_minor, c.currency)}</div>
                </div>
                <div className="text-right text-sm">
                  <div>
                    свободно: <b>{c.available}</b> / {c.total_qty}
                  </div>
                  <div className="opacity-50">
                    held: {c.held} · sold: {c.sold}
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </>
      ) : (
        <p>Загрузка…</p>
      )}
    </main>
  );
}
