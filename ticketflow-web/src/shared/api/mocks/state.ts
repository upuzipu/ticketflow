import type {
  Availability,
  AvailabilityFrame,
  Event,
  Money,
  Order,
  OrderListItem,
} from '@/shared/api/types';

export type Uuid = string;

export interface MockUser {
  id: Uuid;
  email: string;
  password: string;
  role: 'buyer' | 'organizer';
  refreshTokens: Set<string>;
}

export interface MockEvent {
  id: Uuid;
  organizerId: Uuid;
  title: string;
  description?: string;
  image_url?: string;
  startsAt: Date;
  status: 'draft' | 'published' | 'cancelled';
  createdAt: Date;
  updatedAt: Date;
}

export interface MockCategory {
  id: Uuid;
  eventId: Uuid;
  name: string;
  priceMinor: number;
  currency: string;
  totalQty: number;
}

export interface MockHold {
  id: Uuid;
  userId: Uuid;
  eventId: Uuid;
  categoryId: Uuid;
  qty: number;
  status: 'active' | 'confirmed' | 'released';
  expiresAt: Date;
  orderId?: Uuid;
}

export interface MockOrder {
  id: Uuid;
  userId: Uuid;
  holdId: Uuid;
  eventId: Uuid;
  categoryId: Uuid;
  qty: number;
  status: 'pending' | 'paid' | 'failed' | 'confirmed';
  totalMinor: number;
  currency: string;
  createdAt: Date;
}

export const db = {
  users: [] as MockUser[],
  events: [] as MockEvent[],
  categories: [] as MockCategory[],
  holds: [] as MockHold[],
  orders: [] as MockOrder[],
  soldByCategory: new Map<Uuid, number>(),
  idempotency: new Map<string, Uuid>(),
};

export const uuid = (): Uuid => crypto.randomUUID();
export const isUuid = (v: string): boolean =>
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(v);

const ORG_ID = '00000000-0000-4000-8000-000000000001';
const BUYER_ID = '00000000-0000-4000-8000-000000000002';

const inMinutes = (m: number): Date => new Date(Date.now() + m * 60_000);

export const soldQty = (categoryId: Uuid): number => db.soldByCategory.get(categoryId) ?? 0;
export const heldQty = (categoryId: Uuid): number =>
  db.holds
    .filter(
      (h) =>
        h.categoryId === categoryId && h.status === 'active' && h.expiresAt.getTime() > Date.now(),
    )
    .reduce((sum, h) => sum + h.qty, 0);
export const availableQty = (c: MockCategory): number => c.totalQty - soldQty(c.id) - heldQty(c.id);

const STATE_KEY = 'mock:db';

type StoredUser = Omit<MockUser, 'refreshTokens'> & { refreshTokens: string[] };
type StoredEvent = Omit<MockEvent, 'startsAt' | 'createdAt' | 'updatedAt'> & {
  startsAt: string;
  createdAt: string;
  updatedAt: string;
};
type StoredHold = Omit<MockHold, 'expiresAt'> & { expiresAt: string };
type StoredOrder = Omit<MockOrder, 'createdAt'> & { createdAt: string };

interface StoredState {
  users: StoredUser[];
  events: StoredEvent[];
  categories: MockCategory[];
  holds: StoredHold[];
  orders: StoredOrder[];
  sold: Array<[Uuid, number]>;
  idempotency: Array<[string, Uuid]>;
}

export function persistAll(): void {
  if (typeof window === 'undefined') return;
  const state: StoredState = {
    users: db.users.map((u) => ({
      id: u.id,
      email: u.email,
      password: u.password,
      role: u.role,
      refreshTokens: [...u.refreshTokens],
    })),
    events: db.events.map((e) => ({
      id: e.id,
      organizerId: e.organizerId,
      title: e.title,
      description: e.description,
      image_url: e.image_url,
      startsAt: e.startsAt.toISOString(),
      status: e.status,
      createdAt: e.createdAt.toISOString(),
      updatedAt: e.updatedAt.toISOString(),
    })),
    categories: db.categories,
    holds: db.holds.map((h) => ({ ...h, expiresAt: h.expiresAt.toISOString() })),
    orders: db.orders.map((o) => ({ ...o, createdAt: o.createdAt.toISOString() })),
    sold: [...db.soldByCategory],
    idempotency: [...db.idempotency],
  };
  try {
    sessionStorage.setItem(STATE_KEY, JSON.stringify(state));
  } catch {}
}

function restoreAll(): boolean {
  if (typeof window === 'undefined') return false;
  const raw = sessionStorage.getItem(STATE_KEY);
  if (!raw) return false;
  let saved: StoredState;
  try {
    saved = JSON.parse(raw) as StoredState;
  } catch {
    sessionStorage.removeItem(STATE_KEY);
    return false;
  }
  db.users = saved.users.map((u) => ({ ...u, refreshTokens: new Set(u.refreshTokens) }));
  db.events = saved.events.map((e) => ({
    ...e,
    startsAt: new Date(e.startsAt),
    createdAt: new Date(e.createdAt),
    updatedAt: new Date(e.updatedAt),
  }));
  db.categories = saved.categories;
  db.holds = saved.holds.map((h) => ({ ...h, expiresAt: new Date(h.expiresAt) }));
  db.orders = saved.orders.map((o) => ({ ...o, createdAt: new Date(o.createdAt) }));
  db.soldByCategory = new Map(saved.sold);
  db.idempotency = new Map(saved.idempotency);
  return true;
}

function seed(): void {
  if (restoreAll()) return;
  const organizer: MockUser = {
    id: ORG_ID,
    email: 'org@tf.dev',
    password: 'password',
    role: 'organizer',
    refreshTokens: new Set(),
  };
  const buyer: MockUser = {
    id: BUYER_ID,
    email: 'buyer@tf.dev',
    password: 'password',
    role: 'buyer',
    refreshTokens: new Set(),
  };
  db.users.push(organizer, buyer);

  const mkEvent = (
    title: string,
    description: string,
    startsAt: Date,
    status: MockEvent['status'],
  ): MockEvent => {
    const e: MockEvent = {
      id: uuid(),
      organizerId: organizer.id,
      title,
      description,
      startsAt,
      status,
      createdAt: new Date(),
      updatedAt: new Date(),
    };
    db.events.push(e);
    return e;
  };
  const mkCategory = (
    eventId: Uuid,
    name: string,
    priceMinor: number,
    totalQty: number,
  ): MockCategory => {
    const c: MockCategory = { id: uuid(), eventId, name, priceMinor, currency: 'RUB', totalQty };
    db.categories.push(c);
    return c;
  };

  const e1 = mkEvent(
    'Stand-up: Quarterly Report',
    'An evening of humor about product teams: sprints, reviews and sleepless releases.',
    inMinutes(3 * 24 * 60),
    'published',
  );
  const e1vip = mkCategory(e1.id, 'VIP Box', 600_000, 12);
  mkCategory(e1.id, 'Fan Zone', 150_000, 120);
  mkCategory(e1.id, 'Dance Floor', 300_000, 80);
  db.soldByCategory.set(e1vip.id, 4);

  const e2 = mkEvent(
    'Chamber Jazz Night',
    'A quartet, a dark hall and two sets of live jazz.',
    inMinutes(7 * 24 * 60),
    'published',
  );
  mkCategory(e2.id, 'Stalls', 250_000, 60);
  mkCategory(e2.id, 'Balcony', 120_000, 40);

  const e3 = mkEvent(
    'Echo Rock Festival',
    'Three stages, twelve bands, one very loud evening.',
    inMinutes(30 * 24 * 60),
    'published',
  );
  for (let i = 1; i <= 9; i++) {
    const mic = mkEvent(
      `Open Mic Night #${i}`,
      'A cozy evening of short sets, new faces and one surprise headliner.',
      inMinutes(21 + i),
      'published',
    );
    mkCategory(mic.id, 'Standard', 80_000 + i * 5_000, 40);
    mkCategory(mic.id, 'Premium', 150_000 + i * 10_000, 10);
  }
  const e3last = mkCategory(e3.id, 'Last Sector', 90_000, 5);
  mkCategory(e3.id, 'Grandstand', 180_000, 200);
  db.soldByCategory.set(e3last.id, 3);

  const e4 = mkEvent(
    'Workshop: Go for Frontend Developers',
    "Draft: how to read other people's Go code.",
    inMinutes(14 * 24 * 60),
    'draft',
  );
  mkCategory(e4.id, 'Main Hall', 100_000, 30);

  const h: MockHold = {
    id: uuid(),
    userId: buyer.id,
    eventId: e1.id,
    categoryId: e1vip.id,
    qty: 2,
    status: 'confirmed',
    expiresAt: inMinutes(-30),
  };
  db.holds.push(h);
  db.soldByCategory.set(e1vip.id, (db.soldByCategory.get(e1vip.id) ?? 0) + 2);
  db.orders.push({
    id: uuid(),
    userId: buyer.id,
    holdId: h.id,
    eventId: e1.id,
    categoryId: e1vip.id,
    qty: 2,
    status: 'confirmed',
    totalMinor: 1_200_000,
    currency: 'RUB',
    createdAt: new Date(Date.now() - 60_000),
  });
}

export function toEventDto(e: MockEvent): Event {
  return {
    id: e.id,
    organizer_id: e.organizerId,
    title: e.title,
    ...(e.description ? { description: e.description } : {}),
    ...(e.image_url ? { image_url: e.image_url } : {}),
    starts_at: e.startsAt.toISOString(),
    status: e.status,
    categories: db.categories
      .filter((c) => c.eventId === e.id)
      .map((c) => ({
        id: c.id,
        event_id: c.eventId,
        name: c.name,
        price: { amount: c.priceMinor, currency: c.currency } satisfies Money,
        total_qty: c.totalQty,
      })),
    created_at: e.createdAt.toISOString(),
    updated_at: e.updatedAt.toISOString(),
  };
}

export function toAvailability(e: MockEvent): Availability {
  return {
    event_id: e.id,
    categories: db.categories
      .filter((c) => c.eventId === e.id)
      .map((c) => ({
        category_id: c.id,
        name: c.name,
        price: { amount: c.priceMinor, currency: c.currency },
        total_qty: c.totalQty,
        available: availableQty(c),
        held: heldQty(c.id),
        sold: soldQty(c.id),
      })),
  };
}

export function toFrame(e: MockEvent): AvailabilityFrame {
  return {
    type: 'availability',
    event_id: e.id,
    server_time: new Date().toISOString(),
    categories: toAvailability(e).categories,
  };
}

export function toOrderDto(o: MockOrder, created: boolean): Order {
  return {
    id: o.id,
    status: o.status,
    total_minor: o.totalMinor,
    currency: o.currency,
    hold_id: o.holdId,
    created,
    created_at: o.createdAt.toISOString(),
  };
}

export function toOrderListItem(o: MockOrder): OrderListItem {
  const dto = toOrderDto(o, true);
  return {
    id: dto.id,
    status: dto.status,
    total_minor: dto.total_minor,
    currency: dto.currency,
    hold_id: dto.hold_id,
    created_at: dto.created_at,
  };
}

const b64url = (s: string): string =>
  btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
const unb64url = (s: string): string => atob(s.replace(/-/g, '+').replace(/_/g, '/'));

export function issueTokens(user: MockUser): { access: string; refresh: string } {
  const sign = (ttlSec: number): string => {
    const payload = { sub: user.id, role: user.role, exp: Math.floor(Date.now() / 1000) + ttlSec };
    return `${b64url(JSON.stringify({ alg: 'none', typ: 'JWT' }))}.${b64url(JSON.stringify(payload))}.mock`;
  };
  const refresh = sign(7 * 24 * 60 * 60);
  user.refreshTokens.add(refresh);
  persistAll();
  return { access: sign(15 * 60), refresh };
}

export function userFromRequest(req: Request): MockUser | null {
  const header = req.headers.get('authorization');
  if (!header?.startsWith('Bearer ')) return null;
  const [, payload] = header.slice(7).split('.');
  try {
    const claims = JSON.parse(unb64url(payload)) as { sub?: string; exp?: number };
    if (!claims.sub || (claims.exp ?? 0) * 1000 < Date.now()) return null;
    return db.users.find((u) => u.id === claims.sub) ?? null;
  } catch {
    return null;
  }
}

const buckets = new Map<string, number[]>();

export function rateLimit(
  key: string,
  max: number,
  windowMs = 60_000,
): { ok: boolean; retryAfterSec: number } {
  const now = Date.now();
  const hits = (buckets.get(key) ?? []).filter((t) => now - t < windowMs);
  const retryAfterSec =
    hits.length >= max ? Math.max(1, Math.ceil((windowMs - (now - hits[0])) / 1000)) : 0;
  if (retryAfterSec === 0) hits.push(now);
  buckets.set(key, hits);
  return { ok: retryAfterSec === 0, retryAfterSec };
}

export type MockScenario = 'default' | 'payment-decline' | 'gateway-timeout';

export function currentScenario(): MockScenario {
  const v = typeof window !== 'undefined' ? window.localStorage.getItem('mock:scenario') : null;
  return v === 'payment-decline' || v === 'gateway-timeout' ? v : 'default';
}

export function marketMovement(eventId: Uuid): void {
  const cats = db.categories.filter((c) => c.eventId === eventId && availableQty(c) > 0);
  if (!cats.length) return;
  const c = cats[Math.floor(Math.random() * cats.length)];
  const qty = Math.min(availableQty(c), 1 + Math.floor(Math.random() * 3));
  db.soldByCategory.set(c.id, soldQty(c.id) + qty);
}

seed();
if (typeof window !== 'undefined') {
  setInterval(persistAll, 1000);
}
