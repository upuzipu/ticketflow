-- TicketFlow demo seed. Idempotent: safe to run multiple times.
-- Users are NOT seeded here — register them via POST /auth/register
-- with password "secret123":
--   anna@demo.io (organizer), petr@demo.io (organizer), ivan@demo.io (buyer)

-- Anna's events
INSERT INTO events (id, organizer_id, title, description, image_url, starts_at, status, created_at, updated_at) VALUES
                                                                                                                    ('aaaaaaaa-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111',
                                                                                                                     'Rock Symphony Night',
                                                                                                                     'Симфонический рок-концерт с оркестром и спецсценографией. Открытие сезона.',
                                                                                                                     'https://images.unsplash.com/photo-1470229722913-7c0e2dbbafd3?w=800',
                                                                                                                     now() + interval '14 days', 'published', now() - interval '10 days', now() - interval '9 days'),
                                                                                                                    ('aaaaaaaa-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111',
                                                                                                                     'Jazz Evening: Cool Quartet',
                                                                                                                     'Камерный джаз-вечер в формате quartet. Свободная рассадка.',
                                                                                                                     'https://images.unsplash.com/photo-1415201364774-f6f0bb35f28f?w=800',
                                                                                                                     now() + interval '21 days', 'published', now() - interval '8 days', now() - interval '7 days'),
                                                                                                                    ('aaaaaaaa-0000-4000-8000-000000000003', '11111111-1111-4111-8111-111111111111',
                                                                                                                     'Indie Fest: Draft Lineup',
                                                                                                                     'Черновик фестиваля — публикуется после подтверждения артистов.',
                                                                                                                     '',
                                                                                                                     now() + interval '45 days', 'draft', now() - interval '2 days', now() - interval '2 days')
    ON CONFLICT (id) DO NOTHING;

-- Petr's event
INSERT INTO events (id, organizer_id, title, description, image_url, starts_at, status, created_at, updated_at) VALUES
    ('bbbbbbbb-0000-4000-8000-000000000001', '22222222-2222-4222-8222-222222222222',
     'Stand-up Open Mic',
     'Открытый микрофон: 12 выступлений, один зал, свободный вход по регистрации.',
     'https://images.unsplash.com/photo-1585699324551-f6c309eedeca?w=800',
     now() + interval '7 days', 'published', now() - interval '5 days', now() - interval '4 days')
    ON CONFLICT (id) DO NOTHING;

-- Anna event 1: VIP 5000 RUB, Regular 1500 RUB
INSERT INTO ticket_categories (id, event_id, name, price_minor, currency, total_qty) VALUES
                                                                                         ('cccccccc-0000-4000-8000-000000000001', 'aaaaaaaa-0000-4000-8000-000000000001', 'VIP',     500000, 'RUB', 50),
                                                                                         ('cccccccc-0000-4000-8000-000000000002', 'aaaaaaaa-0000-4000-8000-000000000001', 'Regular', 150000, 'RUB', 200)
    ON CONFLICT (event_id, name) DO NOTHING;

-- Anna event 2: Table 8000, Balcony 1200
INSERT INTO ticket_categories (id, event_id, name, price_minor, currency, total_qty) VALUES
                                                                                         ('cccccccc-0000-4000-8000-000000000003', 'aaaaaaaa-0000-4000-8000-000000000002', 'Table',   800000, 'RUB', 30),
                                                                                         ('cccccccc-0000-4000-8000-000000000004', 'aaaaaaaa-0000-4000-8000-000000000002', 'Balcony', 120000, 'RUB', 80)
    ON CONFLICT (event_id, name) DO NOTHING;

-- Petr: Entrance 100 RUB (domain requires positive price)
INSERT INTO ticket_categories (id, event_id, name, price_minor, currency, total_qty) VALUES
    ('cccccccc-0000-4000-8000-000000000005', 'bbbbbbbb-0000-4000-8000-000000000001', 'Entrance', 10000, 'RUB', 120)
    ON CONFLICT (event_id, name) DO NOTHING;

-- Draft event: categories exist but not public
INSERT INTO ticket_categories (id, event_id, name, price_minor, currency, total_qty) VALUES
    ('cccccccc-0000-4000-8000-000000000006', 'aaaaaaaa-0000-4000-8000-000000000003', 'Early Bird', 90000, 'RUB', 500)
    ON CONFLICT (event_id, name) DO NOTHING;

-- Tickets: exactly total_qty per category (skipped if already generated)
INSERT INTO tickets (id, event_id, category_id)
SELECT gen_random_uuid(), c.event_id, c.id
FROM ticket_categories c,
     generate_series(1, c.total_qty) g
WHERE NOT EXISTS (SELECT 1 FROM tickets t WHERE t.category_id = c.id);

-- Rock Symphony: 30 Regular sold
WITH target AS (
    SELECT t.id FROM tickets t
                         JOIN ticket_categories c ON t.category_id = c.id
    WHERE c.event_id = 'aaaaaaaa-0000-4000-8000-000000000001'
      AND c.name = 'Regular' AND t.status = 'available'
    LIMIT 30
    )
UPDATE tickets SET status = 'sold', version = version + 1
WHERE id IN (SELECT id FROM target);

-- Rock Symphony: 12 VIP sold
WITH target AS (
    SELECT t.id FROM tickets t
                         JOIN ticket_categories c ON t.category_id = c.id
    WHERE c.event_id = 'aaaaaaaa-0000-4000-8000-000000000001'
      AND c.name = 'VIP' AND t.status = 'available'
    LIMIT 12
    )
UPDATE tickets SET status = 'sold', version = version + 1
WHERE id IN (SELECT id FROM target);

-- Rock Symphony: 5 Regular held (8h TTL so the expirer does not eat it today)
WITH target AS (
    SELECT t.id FROM tickets t
                         JOIN ticket_categories c ON t.category_id = c.id
    WHERE c.event_id = 'aaaaaaaa-0000-4000-8000-000000000001'
      AND c.name = 'Regular' AND t.status = 'available'
    LIMIT 5
    )
UPDATE tickets SET status = 'held',
                   hold_id = 'dddddddd-0000-4000-8000-000000000001',
                   version = version + 1
WHERE id IN (SELECT id FROM target);

INSERT INTO holds (id, user_id, event_id, category_id, ticket_ids, status, expires_at, created_at)
SELECT 'dddddddd-0000-4000-8000-000000000001',
       '33333333-3333-4333-8333-333333333333',
       'aaaaaaaa-0000-4000-8000-000000000001',
       'cccccccc-0000-4000-8000-000000000002',
       array_agg(id), 'active', now() + interval '8 hours', now()
FROM tickets
WHERE hold_id = 'dddddddd-0000-4000-8000-000000000001'
ON CONFLICT (id) DO NOTHING;