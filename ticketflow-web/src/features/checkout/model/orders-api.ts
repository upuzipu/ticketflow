import { api } from '@/shared/api/client';
import type { CreateOrderRequest, Order, OrderList } from '@/shared/api/types';

export function createOrder(body: CreateOrderRequest): Promise<Order> {
  return api.post('/orders', body);
}

export function fetchOrder(id: string): Promise<Order> {
  return api.get(`/orders/${id}`);
}

export function payOrder(id: string): Promise<Order> {
  return api.post(`/orders/${id}/pay`);
}

export function fetchMyOrders(limit: number, cursor?: string): Promise<OrderList> {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) params.set('cursor', cursor);
  return api.get(`/orders/mine?${params.toString()}`);
}
