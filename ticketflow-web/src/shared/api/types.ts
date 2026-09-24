export type EventStatus = 'draft' | 'published' | 'cancelled';
export type HoldStatus = 'active' | 'confirmed' | 'released' | 'expired';
export type OrderStatus = 'pending' | 'paid' | 'failed' | 'expired' | 'confirmed' | 'refunded';
export type UserRole = 'buyer' | 'organizer';

export interface Money {
  amount: number;
  currency: string;
}

export interface TokenPair {
  access: string;
  refresh: string;
}
export interface Me {
  id: string;
  role: UserRole;
}
export interface Registered {
  id: string;
  email: string;
  role: UserRole;
  access: string;
  refresh: string;
}

export interface TicketCategory {
  id: string;
  event_id: string;
  name: string;
  price: Money;
  total_qty: number;
}

export interface Event {
  id: string;
  organizer_id: string;
  title: string;
  description?: string;
  image_url?: string;
  starts_at: string;
  status: EventStatus;
  categories: TicketCategory[];
  created_at: string;
  updated_at: string;
}

export interface EventList {
  events: Event[];
  next_cursor?: string;
}

export interface CategoryAvailability {
  category_id: string;
  name: string;
  price: Money;
  total_qty: number;
  available: number;
  held: number;
  sold: number;
}

export interface Availability {
  event_id: string;
  server_time?: string;
  categories: CategoryAvailability[];
}

export interface AvailabilityFrame {
  type: 'availability';
  event_id: string;
  server_time: string;
  categories: CategoryAvailability[];
}

export interface HoldCreated {
  hold_id: string;
  status: 'active';
  expires_at: string;
  server_time: string;
  tickets: number;
}

export interface HoldStatusResponse {
  hold_id: string;
  status: HoldStatus;
  expires_at: string;
  server_time: string;
  tickets: number;
  order_id?: string;
}

export interface Order {
  id: string;
  status: OrderStatus;
  total_minor: number;
  currency: string;
  hold_id: string;
  created?: boolean;
  created_at: string;
}

export type OrderListItem = Omit<Order, 'created'>;
export interface OrderList {
  orders: OrderListItem[];
  next_cursor?: string;
}

// ── Request payloads ──
export interface RegisterRequest {
  email: string;
  password: string;
  role: UserRole;
}
export interface LoginRequest {
  email: string;
  password: string;
}
export interface RefreshRequest {
  refresh_token: string;
}
export interface CreateEventRequest {
  title: string;
  starts_at: string;
  description?: string;
  image_url?: string;
  categories: Array<{ name: string; qty: number; price_minor: number; currency: string }>;
}
export interface HoldRequest {
  category_id: string;
  qty: number;
}
export interface CreateOrderRequest {
  hold_id: string;
  idempotency_key: string;
}

export interface ApiErrorBody {
  error: string;
}
