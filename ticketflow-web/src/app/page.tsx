import Link from 'next/link';

export default function Home() {
  return (
    <main className="mx-auto flex min-h-screen max-w-2xl flex-col items-center justify-center gap-4 p-6">
      <h1 className="text-3xl font-semibold">TicketFlow</h1>
      <p className="text-sm opacity-60">
        Phase 0: scaffold, v1.2 contract, mocks. The landing page ships in Phase 3.
      </p>
      <Link
        href="/dev/sandbox"
        className="rounded-md bg-violet-600 px-4 py-2 text-sm text-white hover:bg-violet-500"
      >
        Open sandbox
      </Link>
    </main>
  );
}
