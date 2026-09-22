import Link from 'next/link';

export default function Home() {
  return (
    <main className="mx-auto flex min-h-screen max-w-2xl flex-col items-center justify-center gap-4 p-6">
      <h1 className="text-3xl font-semibold">TicketFlow</h1>
      <p className="text-sm opacity-60">Фаза 0: каркас, контракт v1.2, моки. Лендинг — в Фазе 3.</p>
      <Link
        href="/dev/sandbox"
        className="rounded-md bg-violet-600 px-4 py-2 text-sm text-white hover:bg-violet-500"
      >
        Открыть sandbox
      </Link>
    </main>
  );
}
