import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

export interface HoldSnapshot {
  holdId: string;
  eventId: string;
  eventTitle: string;
  categoryName: string;
  qty: number;
  expiresAt: string;
  serverTime: string;
  idempotencyKey: string;
  unitPriceMinor: number | null;
  currency: string | null;
}

interface CheckoutState {
  hold: HoldSnapshot | null;
  orderId: string | null;
  setHold(hold: HoldSnapshot): void;
  setOrderId(id: string | null): void;
  clear(): void;
}

export const useCheckoutStore = create<CheckoutState>()(
  persist(
    (set) => ({
      hold: null,
      orderId: null,
      setHold: (hold) => set({ hold, orderId: null }),
      setOrderId: (orderId) => set({ orderId }),
      clear: () => set({ hold: null, orderId: null }),
    }),
    {
      name: 'ticketflow.checkout',
      storage: createJSONStorage(() => sessionStorage),
    },
  ),
);
