'use client';

import { useEffect, useState } from 'react';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Card } from '@/shared/ui/card';
import { Badge } from '@/shared/ui/badge';
import { Skeleton } from '@/shared/ui/skeleton';
import { LiveBadge } from '@/shared/ui/live-badge';
import { CountdownRing } from '@/shared/ui/countdown-ring';
import { QuantityStepper } from '@/shared/ui/quantity-stepper';
import { PriceTag } from '@/shared/ui/price-tag';

const COLORS = [
  ['background', 'bg-background'],
  ['surface', 'bg-surface'],
  ['accent', 'bg-accent'],
  ['success', 'bg-success'],
  ['warning', 'bg-warning'],
  ['danger', 'bg-danger'],
] as const;

function TickingRing() {
  const total = 600;
  const [left, setLeft] = useState(total);
  useEffect(() => {
    const t = setInterval(() => setLeft((l) => (l <= 0 ? total : l - 1)), 1000);
    return () => clearInterval(t);
  }, []);
  return <CountdownRing secondsLeft={left} totalSeconds={total} />;
}

export default function UiShowcasePage() {
  const [qty, setQty] = useState(2);

  return (
    <main className="mx-auto max-w-4xl space-y-10 p-6">
      <header className="space-y-1">
        <h1 className="font-display text-2xl font-semibold">UI-витрина</h1>
        <p className="text-sm text-muted">Дизайн-система TicketFlow. Переключай тему в хедере.</p>
      </header>

      <section className="space-y-3">
        <h2 className="font-display text-sm uppercase tracking-wide text-muted">Кнопки</h2>
        <div className="flex flex-wrap items-center gap-2">
          <Button>Primary</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="outline">Outline</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="danger">Danger</Button>
          <Button disabled>Disabled</Button>
          <Button size="sm">Small</Button>
          <Button size="lg">Large</Button>
        </div>
      </section>

      <section className="space-y-3">
        <h2 className="font-display text-sm uppercase tracking-wide text-muted">Инпуты</h2>
        <div className="grid max-w-sm gap-2">
          <Input placeholder="email@example.com" />
          <Input disabled placeholder="Недоступно" />
        </div>
      </section>

      <section className="space-y-3">
        <h2 className="font-display text-sm uppercase tracking-wide text-muted">
          Пример карточки категории
        </h2>
        <Card className="flex items-center justify-between gap-4 p-4">
          <div className="space-y-1">
            <div className="font-medium">VIP-ложа</div>
            <LiveBadge available={3} total={12} />
          </div>
          <div className="flex items-center gap-4">
            <PriceTag amountMinor={600000} currency="RUB" />
            <QuantityStepper value={qty} onChange={setQty} max={10} />
            <Button size="sm">Забронировать</Button>
          </div>
        </Card>
      </section>

      <section className="space-y-3">
        <h2 className="font-display text-sm uppercase tracking-wide text-muted">
          Бейджи и live-состояния
        </h2>
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="accent">Draft</Badge>
          <Badge variant="success">Confirmed</Badge>
          <Badge variant="warning">Pending</Badge>
          <Badge variant="danger">Failed</Badge>
          <Badge variant="muted">Refunded</Badge>
        </div>
        <div className="flex flex-wrap gap-6">
          <LiveBadge available={42} total={120} />
          <LiveBadge available={3} total={40} />
          <LiveBadge available={0} total={10} />
        </div>
      </section>

      <section className="space-y-3">
        <h2 className="font-display text-sm uppercase tracking-wide text-muted">Таймер холда</h2>
        <div className="flex items-center gap-6">
          <CountdownRing secondsLeft={540} totalSeconds={600} />
          <CountdownRing secondsLeft={45} totalSeconds={600} />
          <TickingRing />
        </div>
      </section>

      <section className="space-y-3">
        <h2 className="font-display text-sm uppercase tracking-wide text-muted">
          Скелетоны и палитра
        </h2>
        <div className="max-w-sm space-y-2">
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-4 w-1/2" />
          <Skeleton className="h-24 w-full" />
        </div>
        <div className="flex flex-wrap gap-3 pt-2">
          {COLORS.map(([name, cls]) => (
            <div key={name} className="space-y-1 text-center">
              <div className={`h-10 w-16 rounded-lg border border-border ${cls}`} />
              <div className="text-xs text-muted">{name}</div>
            </div>
          ))}
        </div>
      </section>
    </main>
  );
}
