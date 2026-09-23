import Link from 'next/link';
import { ThemeToggle } from '@/shared/ui/theme-toggle';
import { AuthNav } from '@/shared/ui/auth-nav';

export function Header() {
  return (
    <header className="sticky top-0 z-40 border-b border-border bg-background/80 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
        <Link href="/" className="font-display text-lg font-semibold tracking-tight">
          <span className="text-accent">Ticket</span>Flow
        </Link>
        <nav className="ml-6 flex items-center gap-4 text-sm">
          <Link href="/events" className="text-muted transition-colors hover:text-foreground">
            Events
          </Link>
        </nav>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <AuthNav />
        </div>
      </div>
    </header>
  );
}
