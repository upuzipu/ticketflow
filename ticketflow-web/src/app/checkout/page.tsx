import { Suspense } from 'react';
import { CheckoutView } from '@/features/checkout/ui/checkout-view';

export const metadata = { title: 'Checkout' };

export default function CheckoutPage() {
  return (
    <Suspense
      fallback={
        <main className="mx-auto max-w-md p-6 pt-20 text-sm text-muted">Loading checkout…</main>
      }
    >
      <CheckoutView />
    </Suspense>
  );
}
