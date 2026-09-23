import type { Metadata } from 'next';
import { Inter, Unbounded } from 'next/font/google';
import { ThemeProvider } from '@/shared/ui/theme-provider';
import { MocksProvider } from '@/shared/api/mocks/MocksProvider';
import { Header } from '@/widgets/header/ui/header';
import { Footer } from '@/widgets/footer/ui/footer';
import './globals.css';
import { QueryProvider } from '@/shared/api/query-provider';

const inter = Inter({ subsets: ['latin', 'cyrillic'], variable: '--font-inter', display: 'swap' });
const unbounded = Unbounded({
  subsets: ['latin', 'cyrillic'],
  variable: '--font-unbounded',
  display: 'swap',
});

export const metadata: Metadata = {
  title: { default: 'TicketFlow — live tickets', template: '%s · TicketFlow' },
  description: 'Buy tickets with live availability and an honest hold timer.',
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ru" suppressHydrationWarning className={`${inter.variable} ${unbounded.variable}`}>
      <body>
        <ThemeProvider>
          <QueryProvider>
            <MocksProvider>
              <div className="flex min-h-dvh flex-col">
                <Header />
                <main className="flex-1">{children}</main>
                <Footer />
              </div>
            </MocksProvider>
          </QueryProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
