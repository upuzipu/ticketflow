import type { Metadata } from 'next';
import { Inter, Unbounded } from 'next/font/google';
import { ThemeProvider } from '@/shared/ui/theme-provider';
import { MocksProvider } from '@/shared/api/mocks/MocksProvider';
import { Header } from '@/widgets/header/ui/header';
import { Footer } from '@/widgets/footer/ui/footer';
import './globals.css';

const inter = Inter({ subsets: ['latin', 'cyrillic'], variable: '--font-inter', display: 'swap' });
const unbounded = Unbounded({
  subsets: ['latin', 'cyrillic'],
  variable: '--font-unbounded',
  display: 'swap',
});

export const metadata: Metadata = {
  title: { default: 'TicketFlow — билеты вживую', template: '%s · TicketFlow' },
  description: 'Покупка билетов с живыми остатками и честным таймером брони.',
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ru" suppressHydrationWarning className={`${inter.variable} ${unbounded.variable}`}>
      <body>
        <ThemeProvider>
          <MocksProvider>
            <div className="flex min-h-dvh flex-col">
              <Header />
              <main className="flex-1">{children}</main>
              <Footer />
            </div>
          </MocksProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
