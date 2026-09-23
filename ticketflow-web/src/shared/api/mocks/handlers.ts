import { http, HttpResponse, delay, type JsonBodyType } from 'msw';
import type {
  CreateEventRequest,
  CreateOrderRequest,
  HoldRequest,
  LoginRequest,
  RefreshRequest,
  RegisterRequest,
} from '@/shared/api/types';
import {
  availableQty,
  currentScenario,
  db,
  isUuid,
  issueTokens,
  keysetPage,
  rateLimit,
  soldQty,
  toAvailability,
  toEventDto,
  toOrderDto,
  toOrderListItem,
  userFromRequest,
  uuid,
  type MockOrder,
  persistAll,
} from './state';

const err = (status: number, msg: string, headers?: Record<string, string>) =>
  HttpResponse.json({ error: `domain: ${msg}` }, { status, headers });
const retryAfter = (sec: number) => ({ 'Retry-After': String(sec) });
const json = (data: JsonBodyType, status = 200) => HttpResponse.json(data, { status });

function holdStatus(h: { status: 'active' | 'confirmed' | 'released'; expiresAt: Date }) {
  if (h.status === 'active')
    return h.expiresAt.getTime() > Date.now() ? ('active' as const) : ('expired' as const);
  return h.status;
}

export const handlers = [
  // ── auth ──
  http.post('/api/auth/register', async ({ request }) => {
    await delay(300);
    const rl = rateLimit('register', 3);
    if (!rl.ok) return err(429, 'rate limit exceeded', retryAfter(rl.retryAfterSec));

    const body = (await request.json().catch(() => null)) as RegisterRequest | null;
    if (
      !body?.email ||
      !body.password ||
      body.password.length < 8 ||
      !['buyer', 'organizer'].includes(body.role)
    )
      return err(422, 'validation failed');
    if (db.users.some((u) => u.email === body.email.toLowerCase()))
      return err(409, 'email already taken');

    const user = {
      id: uuid(),
      email: body.email.toLowerCase(),
      password: body.password,
      role: body.role,
      refreshTokens: new Set<string>(),
    };
    db.users.push(user);
    return json({ id: user.id, email: user.email, role: user.role, ...issueTokens(user) }, 201);
  }),

  http.post('/api/auth/login', async ({ request }) => {
    await delay(300);
    const rl = rateLimit('login', 5);
    if (!rl.ok) return err(429, 'rate limit exceeded', retryAfter(rl.retryAfterSec));

    const body = (await request.json().catch(() => null)) as LoginRequest | null;
    const user = db.users.find((u) => u.email === body?.email?.toLowerCase());
    if (!user || user.password !== body?.password) return err(401, 'invalid credentials'); // неотличимо — как контракт
    return json(issueTokens(user));
  }),

  http.post('/api/auth/refresh', async ({ request }) => {
    const body = (await request.json().catch(() => null)) as RefreshRequest | null;
    const token = body?.refresh_token ?? '';
    const user = db.users.find((u) => u.refreshTokens.has(token));
    if (!user) return err(401, 'invalid token'); // reuse detection: ротированный уже отозван
    user.refreshTokens.delete(token); // ротация
    return json(issueTokens(user));
  }),

  http.post('/api/auth/logout', async ({ request }) => {
    const body = (await request.json().catch(() => null)) as RefreshRequest | null;
    for (const u of db.users) u.refreshTokens.delete(body?.refresh_token ?? '');
    persistAll();
    return new HttpResponse(null, { status: 204 });
  }),

  http.get('/api/users/me', ({ request }) => {
    const user = userFromRequest(request);
    if (!user) return err(401, 'missing or invalid token');
    return json({ id: user.id, role: user.role });
  }),

  // ── events ──
  http.post('/api/events', async ({ request }) => {
    const user = userFromRequest(request);
    if (!user) return err(401, 'missing or invalid token');
    if (user.role !== 'organizer') return err(403, 'organizer only');

    const body = (await request.json().catch(() => null)) as CreateEventRequest | null;
    const valid =
      !!body &&
      !!body.title &&
      !Number.isNaN(Date.parse(body.starts_at ?? '')) &&
      Array.isArray(body.categories) &&
      body.categories.length >= 1 &&
      body.categories.every(
        (c) =>
          !!c.name &&
          Number.isInteger(c.qty) &&
          c.qty >= 1 &&
          Number.isInteger(c.price_minor) &&
          c.price_minor >= 1 &&
          !!c.currency,
      );
    if (!valid) return err(422, 'validation failed');

    const event = {
      id: uuid(),
      organizerId: user.id,
      title: body.title,
      description: body.description,
      startsAt: new Date(body.starts_at),
      status: 'draft' as const,
      createdAt: new Date(),
      updatedAt: new Date(),
    };
    db.events.push(event);
    for (const c of body.categories) {
      db.categories.push({
        id: uuid(),
        eventId: event.id,
        name: c.name,
        priceMinor: c.price_minor,
        currency: c.currency,
        totalQty: c.qty,
      });
    }
    return json(
      {
        id: event.id,
        title: event.title,
        starts_at: event.startsAt.toISOString(),
        status: 'draft',
      },
      201,
    );
  }),

  http.get('/api/events', ({ request }) => {
    const url = new URL(request.url);
    const limit = Math.min(Number(url.searchParams.get('limit') ?? 20) || 20, 100);
    const cursor = url.searchParams.get('cursor');

    if (url.searchParams.get('mine') === 'true') {
      const user = userFromRequest(request);
      if (!user || user.role !== 'organizer') return err(403, 'organizer only'); // явный отказ — как согласовано
      const items = db.events
        .filter((e) => e.organizerId === user.id)
        .sort((a, b) => a.startsAt.getTime() - b.startsAt.getTime() || a.id.localeCompare(b.id));
      const { page, nextCursor } = keysetPage(
        items,
        cursor,
        limit,
        (e) => [e.startsAt, e.id],
        'after',
      );
      return json({ events: page.map(toEventDto), next_cursor: nextCursor });
    }

    const items = db.events
      .filter((e) => e.status === 'published')
      .sort((a, b) => a.startsAt.getTime() - b.startsAt.getTime() || a.id.localeCompare(b.id));
    const { page, nextCursor } = keysetPage(
      items,
      cursor,
      limit,
      (e) => [e.startsAt, e.id],
      'after',
    );
    return json({ events: page.map(toEventDto), next_cursor: nextCursor });
  }),

  http.get('/api/events/:id', ({ request, params }) => {
    const id = params.id as string;
    const event = db.events.find((e) => e.id === id);
    if (!event || !isUuid(id)) return err(404, 'not found');
    if (event.status !== 'published') {
      const user = userFromRequest(request);
      if (!user || user.id !== event.organizerId) return err(404, 'not found'); // чужой draft не раскрываем
    }
    return json(toEventDto(event));
  }),

  http.post('/api/events/:id/publish', ({ request, params }) => {
    const user = userFromRequest(request);
    if (!user) return err(401, 'missing or invalid token');
    const event = db.events.find((e) => e.id === params.id);
    if (!event) return err(404, 'not found');
    if (user.id !== event.organizerId) return err(403, 'not the owner');
    if (event.status !== 'draft') return err(409, 'invalid transition'); // FSM
    event.status = 'published';
    event.updatedAt = new Date();
    return new HttpResponse(null, { status: 204 });
  }),

  http.get('/api/events/:id/availability', ({ params }) => {
    const event = db.events.find((e) => e.id === params.id);
    if (!event) return err(404, 'not found');
    return json(toAvailability(event));
  }),

  // ── holds ──
  http.post('/api/events/:id/holds', async ({ request }) => {
    await delay(200);
    const rl = rateLimit('holds', 30);
    if (!rl.ok) return err(429, 'rate limit exceeded', retryAfter(rl.retryAfterSec));

    const user = userFromRequest(request);
    if (!user) return err(401, 'missing or invalid token');

    const body = (await request.json().catch(() => null)) as HoldRequest | null;
    if (
      !body ||
      !isUuid(body.category_id ?? '') ||
      !Number.isInteger(body.qty) ||
      body.qty < 1 ||
      body.qty > 10
    )
      return err(422, 'validation failed');

    const category = db.categories.find((c) => c.id === body.category_id);
    if (!category || availableQty(category) < body.qty) return err(409, 'no tickets available');

    const hold = {
      id: uuid(),
      userId: user.id,
      eventId: category.eventId,
      categoryId: category.id,
      qty: body.qty,
      status: 'active' as const,
      expiresAt: new Date(Date.now() + 10 * 60_000),
    };
    db.holds.push(hold);
    return json(
      {
        hold_id: hold.id,
        status: 'active',
        expires_at: hold.expiresAt.toISOString(),
        server_time: new Date().toISOString(),
        tickets: hold.qty,
      },
      201,
    );
  }),

  http.get('/api/holds/:id', ({ request, params }) => {
    const user = userFromRequest(request);
    if (!user) return err(401, 'missing or invalid token');
    const hold = db.holds.find((h) => h.id === params.id);
    if (!hold) return err(404, 'not found');
    if (hold.userId !== user.id) return err(403, 'not the owner');

    const order = db.orders.find(
      (o) => o.holdId === hold.id && (o.status === 'paid' || o.status === 'confirmed'),
    );
    return json({
      hold_id: hold.id,
      status: holdStatus(hold),
      expires_at: hold.expiresAt.toISOString(),
      server_time: new Date().toISOString(),
      tickets: hold.qty,
      ...(order ? { order_id: order.id } : {}),
    });
  }),

  http.delete('/api/holds/:id', ({ request, params }) => {
    const user = userFromRequest(request);
    if (!user) return err(401, 'missing or invalid token');
    const hold = db.holds.find((h) => h.id === params.id);
    if (hold && hold.userId !== user.id) return err(403, 'not the owner');
    if (hold && hold.status === 'active') hold.status = 'released'; // идемпотентно
    return new HttpResponse(null, { status: 204 });
  }),

  // ── orders ──
  http.post('/api/orders', async ({ request }) => {
    await delay(400);
    const rl = rateLimit('orders', 20);
    if (!rl.ok) return err(429, 'rate limit exceeded', retryAfter(rl.retryAfterSec));

    const user = userFromRequest(request);
    if (!user) return err(401, 'missing or invalid token');

    const body = (await request.json().catch(() => null)) as CreateOrderRequest | null;
    if (!body || !isUuid(body.hold_id ?? '') || !isUuid(body.idempotency_key ?? ''))
      return err(422, 'bad idempotency key / hold id');

    const idemKey = `${user.id}:${body.idempotency_key}`;
    const existingId = db.idempotency.get(idemKey);
    if (existingId) {
      const existing = db.orders.find((o) => o.id === existingId);
      if (existing) return json(toOrderDto(existing, false), 200);
    }

    const hold = db.holds.find((h) => h.id === body.hold_id);
    if (!hold) return err(404, 'hold not found');
    if (hold.userId !== user.id) return err(403, 'hold belongs to another user');
    if (hold.status !== 'active') return err(409, 'hold not active');
    if (hold.expiresAt.getTime() <= Date.now()) return err(410, 'hold expired');

    const category = db.categories.find((c) => c.id === hold.categoryId)!;
    const mkOrder = (status: 'pending' | 'paid' | 'failed'): MockOrder => {
      const order = {
        id: uuid(),
        userId: user.id,
        holdId: hold.id,
        eventId: hold.eventId,
        categoryId: hold.categoryId,
        qty: hold.qty,
        status,
        totalMinor: hold.qty * category.priceMinor,
        currency: category.currency,
        createdAt: new Date(),
      };
      db.orders.push(order);
      db.idempotency.set(idemKey, order.id);
      return order;
    };

    const scenario = currentScenario();
    if (scenario === 'payment-decline') {
      hold.status = 'released'; // компенсация: билеты возвращаются в продажу
      mkOrder('failed');
      return err(402, 'payment declined');
    }
    if (scenario === 'gateway-timeout') {
      mkOrder('pending'); // заказ существует, id вернёт ретрай с тем же ключом
      return err(424, 'gateway timeout');
    }

    hold.status = 'confirmed';
    hold.orderId = undefined;
    db.soldByCategory.set(category.id, soldQty(category.id) + hold.qty);
    const order = mkOrder('paid');
    setTimeout(() => {
      order.status = 'confirmed';
    }, 1500);
    return json(toOrderDto(order, true), 201);
  }),

  http.post('/api/orders/:id/pay', async ({ request, params }) => {
    await delay(400);
    const user = userFromRequest(request);
    if (!user) return err(401, 'missing or invalid token');
    const order = db.orders.find((o) => o.id === params.id);
    if (!order) return err(404, 'not found');
    if (order.userId !== user.id) return err(403, 'not the owner');
    if (order.status !== 'pending') return err(409, 'order is not payable');

    if (currentScenario() === 'payment-decline') {
      order.status = 'failed';
      const hold = db.holds.find((h) => h.id === order.holdId);
      if (hold && hold.status === 'active') hold.status = 'released';
      return err(402, 'payment declined');
    }

    order.status = 'paid';
    const hold = db.holds.find((h) => h.id === order.holdId);
    if (hold && hold.status === 'active') {
      hold.status = 'confirmed';
      hold.orderId = order.id;
      db.soldByCategory.set(order.categoryId, soldQty(order.categoryId) + order.qty);
    }
    setTimeout(() => {
      order.status = 'confirmed';
    }, 1500);
    return json(toOrderDto(order, true));
  }),

  http.get('/api/orders/mine', ({ request }) => {
    const user = userFromRequest(request);
    if (!user) return err(401, 'missing or invalid token');
    const url = new URL(request.url);
    const limit = Math.min(Number(url.searchParams.get('limit') ?? 20) || 20, 100);
    const items = db.orders
      .filter((o) => o.userId === user.id)
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime() || b.id.localeCompare(a.id)); // created_at DESC, id DESC
    const { page, nextCursor } = keysetPage(
      items,
      url.searchParams.get('cursor'),
      limit,
      (o) => [o.createdAt, o.id],
      'before',
    );
    return json({ orders: page.map(toOrderListItem), next_cursor: nextCursor });
  }),

  http.get('/api/orders/:id', ({ request, params }) => {
    const user = userFromRequest(request);
    if (!user) return err(401, 'missing or invalid token');
    const order = db.orders.find((o) => o.id === params.id);
    if (!order) return err(404, 'not found');
    if (order.userId !== user.id) return err(403, 'not the owner');
    return json(toOrderDto(order, true));
  }),
];
