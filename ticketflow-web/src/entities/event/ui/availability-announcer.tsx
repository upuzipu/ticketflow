'use client';

import { useEffect, useRef, useState } from 'react';
import type { CategoryAvailability } from '@/shared/api/types';

const INTERVAL_MS = 30_000;

export function AvailabilityAnnouncer({ categories }: { categories: CategoryAvailability[] }) {
  const [message, setMessage] = useState('');
  const lastSpokeAt = useRef(0);
  const lastTotal = useRef<number | null>(null);

  useEffect(() => {
    const total = categories.reduce((sum, c) => sum + c.available, 0);
    const now = Date.now();
    if (now - lastSpokeAt.current < INTERVAL_MS) return;
    if (total === lastTotal.current) return;
    lastSpokeAt.current = now;
    lastTotal.current = total;
    setMessage(`${total} tickets available right now`);
  }, [categories]);

  return (
    <div aria-live="polite" className="sr-only">
      {message}
    </div>
  );
}
