import { api } from '@/shared/api/client';
import type { CreateOrderRequest, Order } from '@/shared/api/types';

export function createOrder(body: CreateOrderRequest): Promise<Order> {
  return api.post('/orders', body);
}

export function fetchOrder(id: string): Promise<Order> {
  return api.get(`/orders/${id}`);
}

export function payOrder(id: string): Promise<Order> {
  return api.post(`/orders/${id}/pay`);
}
