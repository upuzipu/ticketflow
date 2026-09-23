'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useQueryClient } from '@tanstack/react-query';
import { ApiError } from '@/shared/api/client';
import { mocksReady } from '@/shared/api/mocks/enable';
import { createEvent } from '../model/organizer-api';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { Input } from '@/shared/ui/input';

interface CategoryDraft {
  name: string;
  qty: string;
  price: string;
}

const emptyCategory: CategoryDraft = { name: '', qty: '', price: '' };

export function EventForm() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [startsAt, setStartsAt] = useState('');
  const [imageUrl, setImageUrl] = useState('');
  const [categories, setCategories] = useState<CategoryDraft[]>([{ ...emptyCategory }]);
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  const updateCategory = (index: number, patch: Partial<CategoryDraft>) => {
    setCategories((list) => list.map((c, i) => (i === index ? { ...c, ...patch } : c)));
  };

  const submit = async () => {
    setError(null);
    if (!title.trim()) return setError('Title is required.');
    if (!startsAt || Number.isNaN(Date.parse(startsAt)))
      return setError('A valid start date is required.');
    const parsed = categories.map((c) => ({
      name: c.name.trim(),
      qty: Number(c.qty),
      price_minor: Math.round(Number(c.price) * 100),
      currency: 'RUB',
    }));
    for (const c of parsed) {
      if (!c.name) return setError('Every category needs a name.');
      if (!Number.isInteger(c.qty) || c.qty < 1)
        return setError(`Quantity for “${c.name}” must be a positive integer.`);
      if (!Number.isFinite(c.price_minor) || c.price_minor < 1)
        return setError(`Price for “${c.name}” must be greater than zero.`);
    }
    if (parsed.length < 1) return setError('Add at least one ticket category.');

    setPending(true);
    try {
      await mocksReady;
      await createEvent({
        title: title.trim(),
        description: description.trim() || undefined,
        image_url: imageUrl.trim() || undefined,
        starts_at: new Date(startsAt).toISOString(),
        categories: parsed,
      });
      void queryClient.invalidateQueries({ queryKey: ['events', 'mine'] });
      router.push('/organizer');
    } catch (err) {
      setError(
        err instanceof ApiError
          ? `The backend rejected the event: ${err.message}`
          : err instanceof Error
            ? err.message
            : 'Failed to create the event.',
      );
    } finally {
      setPending(false);
    }
  };

  return (
    <div className="space-y-4">
      <Card className="space-y-3 p-4">
        <div className="space-y-1">
          <label htmlFor="title" className="text-sm font-medium">
            Title
          </label>
          <Input
            id="title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Quarterly Report: Live"
          />
        </div>
        <div className="space-y-1">
          <label htmlFor="description" className="text-sm font-medium">
            Description (optional)
          </label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            maxLength={2000}
            rows={3}
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            placeholder="What is this event about?"
          />
        </div>
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="space-y-1">
            <label htmlFor="starts-at" className="text-sm font-medium">
              Starts at
            </label>
            <Input
              id="starts-at"
              type="datetime-local"
              value={startsAt}
              onChange={(e) => setStartsAt(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <label htmlFor="image-url" className="text-sm font-medium">
              Image URL (optional)
            </label>
            <Input
              id="image-url"
              type="url"
              value={imageUrl}
              onChange={(e) => setImageUrl(e.target.value)}
              placeholder="https://…"
            />
          </div>
        </div>
      </Card>

      <section className="space-y-2">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-medium uppercase tracking-wide text-muted">
            Ticket categories
          </h2>
          <Button
            size="sm"
            variant="secondary"
            onClick={() => setCategories((l) => [...l, { ...emptyCategory }])}
          >
            Add category
          </Button>
        </div>
        {categories.map((c, i) => (
          <Card key={i} className="grid grid-cols-[1fr_5rem_7rem_auto] items-center gap-2 p-3">
            <Input
              value={c.name}
              onChange={(e) => updateCategory(i, { name: e.target.value })}
              placeholder="Category name"
              aria-label={`Category ${i + 1} name`}
            />
            <Input
              type="number"
              min={1}
              value={c.qty}
              onChange={(e) => updateCategory(i, { qty: e.target.value })}
              placeholder="Qty"
              aria-label={`Category ${i + 1} quantity`}
            />
            <Input
              type="number"
              min={0.01}
              step="0.01"
              value={c.price}
              onChange={(e) => updateCategory(i, { price: e.target.value })}
              placeholder="₽"
              aria-label={`Category ${i + 1} price in rubles`}
            />
            <Button
              size="icon"
              variant="ghost"
              disabled={categories.length === 1}
              onClick={() => setCategories((l) => l.filter((_, j) => j !== i))}
              aria-label={`Remove category ${i + 1}`}
            >
              ×
            </Button>
          </Card>
        ))}
      </section>

      {error && <p className="text-sm text-danger">{error}</p>}
      <Button size="lg" className="w-full" disabled={pending} onClick={() => void submit()}>
        {pending ? 'Creating…' : 'Create event (draft)'}
      </Button>
    </div>
  );
}
