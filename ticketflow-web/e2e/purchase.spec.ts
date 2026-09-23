import { test, expect } from '@playwright/test';
import { login, reserveFirstTicket } from './helpers';

test.describe('purchase flow', () => {
  test('buyer purchases tickets end to end', async ({ page }) => {
    await login(page, 'buyer@tf.dev', 'buyer');
    await reserveFirstTicket(page);
    await page.getByRole('button', { name: /^Pay/ }).click();
    await page.waitForURL(/\/orders\//);
    await expect(page.getByText('Stand-up: Quarterly Report')).toBeVisible();
    await expect(page.getByText('Tickets ready')).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText('Your tickets are issued')).toBeVisible();
  });

  test('session survives a page reload', async ({ page }) => {
    await login(page, 'buyer@tf.dev', 'buyer');
    await page.reload();
    await expect(page.getByText('buyer', { exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible();
  });

  test('checkout survives a page reload', async ({ page }) => {
    await login(page, 'buyer@tf.dev', 'buyer');
    await reserveFirstTicket(page);
    await page.reload();
    await expect(page.getByRole('heading', { name: 'Checkout' })).toBeVisible();
    await expect(page.getByRole('timer')).toBeVisible();
  });

  test('declined payment releases tickets', async ({ page }) => {
    await page.addInitScript(() => localStorage.setItem('mock:scenario', 'payment-decline'));
    await login(page, 'buyer@tf.dev', 'buyer');
    await reserveFirstTicket(page);
    await page.getByRole('button', { name: /^Pay/ }).click();
    await expect(page.getByRole('heading', { name: 'Payment declined' })).toBeVisible();
    await page.getByRole('button', { name: 'Choose tickets again' }).click();
    await expect(page).toHaveURL(/\/events/);
  });

  test('gateway timeout keeps the order pending and pays on retry', async ({ page }) => {
    await page.addInitScript(() => localStorage.setItem('mock:scenario', 'gateway-timeout'));
    await login(page, 'buyer@tf.dev', 'buyer');
    await reserveFirstTicket(page);
    await page.getByRole('button', { name: /^Pay/ }).click();
    await page.waitForURL(/\/orders\//);
    await expect(page.getByText('Pending payment')).toBeVisible();
    await page.getByRole('button', { name: 'Retry payment' }).click();
    await expect(page.getByText('Tickets ready')).toBeVisible({ timeout: 15_000 });
  });
});
