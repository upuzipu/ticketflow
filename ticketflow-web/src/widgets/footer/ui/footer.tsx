import Link from 'next/link';

export function Footer() {
  return (
    <footer className="border-t border-border py-6">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-4 text-sm text-muted">
        <span>© {new Date().getFullYear()} TicketFlow</span>
        <Link href="/dev/sandbox/ui" className="transition-colors hover:text-foreground">
          UI-витрина
        </Link>
      </div>
    </footer>
  );
}
