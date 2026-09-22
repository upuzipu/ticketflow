'use client';

import { useEffect } from 'react';
import { mocksReady } from './enable';

export function MocksProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    void mocksReady;
  }, []);
  return children;
}
