import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI
    ? [['github'], ['html', { open: 'never' }]]
    : [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: 'http://localhost:3005',
    navigationTimeout: 60_000,
    trace: 'retain-on-failure',
  },
  webServer: {
    command: process.env.CI
      ? 'pnpm build && pnpm exec next start --port 3005'
      : 'pnpm exec next start --port 3005',
    url: 'http://localhost:3005',
    reuseExistingServer: !process.env.CI,
    timeout: 300_000,
    env: {
      NEXT_PUBLIC_API_MOCKING: 'enabled',
      HUSKY: '0',
    },
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
});
