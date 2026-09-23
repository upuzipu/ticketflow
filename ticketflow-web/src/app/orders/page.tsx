import type { Metadata } from 'next';
import { MyOrders } from '@/features/orders/ui/my-orders';
import { RequireAuth } from '@/shared/ui/require-auth';

export const metadata: Metadata = { title: 'My orders' };

export default function OrdersPage() {
  return (
    <RequireAuth>
      <main className="mx-auto max-w-2xl space-y-5 p-6">
        <h1 className="font-display text-2xl font-semibold">My orders</h1>
        <MyOrders />
      </main>
    </RequireAuth>
  );
}
