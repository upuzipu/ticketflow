'use client';

import { useEffect, useState } from 'react';
import { computeSecondsLeft } from '../lib/hold-clock';

export function useHoldTimer(expiresAt: string | null, serverTime: string | null): number | null {
  const [secondsLeft, setSecondsLeft] = useState<number | null>(null);

  useEffect(() => {
    if (!expiresAt) {
      setSecondsLeft(null);
      return;
    }
    const tick = () => {
      setSecondsLeft(computeSecondsLeft(expiresAt, serverTime, Date.now()));
    };
    tick();
    const timer = setInterval(tick, 1000);
    return () => clearInterval(timer);
  }, [expiresAt, serverTime]);

  return secondsLeft;
}
