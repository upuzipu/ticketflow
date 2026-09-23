'use client';

import { useEffect, useState } from 'react';

export function useHoldTimer(expiresAt: string | null, serverTime: string | null): number | null {
  const [secondsLeft, setSecondsLeft] = useState<number | null>(null);

  useEffect(() => {
    if (!expiresAt) {
      setSecondsLeft(null);
      return;
    }
    const offset = serverTime ? Date.parse(serverTime) - Date.now() : 0;
    const deadline = Date.parse(expiresAt);
    const tick = () => {
      setSecondsLeft(Math.max(0, Math.floor((deadline - (Date.now() + offset)) / 1000)));
    };
    tick();
    const timer = setInterval(tick, 1000);
    return () => clearInterval(timer);
  }, [expiresAt, serverTime]);

  return secondsLeft;
}
