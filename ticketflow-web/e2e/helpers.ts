import { expect, type Page } from '@playwright/test';

export async function login(page: Page, email: string, role: 'buyer' | 'organizer'): Promise<void> {
  await page.goto('/login');
  await page.getByPlaceholder('Email').fill(email);
  await page.getByPlaceholder('Password').fill('password');
  await page.getByRole('button', { name: 'Log in' }).click();
  await page.waitForURL(/\/$/);
  await expect(page.getByText(role, { exact: true })).toBeVisible();
}

export async function reserveFirstTicket(page: Page): Promise<void> {
  await page.goto('/events');
  await page
    .getByRole('link', { name: /Stand-up/ })
    .first()
    .click();
  await expect(page.getByRole('heading', { name: /Stand-up/ })).toBeVisible();
  await page.getByRole('button', { name: 'Reserve' }).first().click();
  await page.waitForURL('/checkout');
  await expect(page.getByRole('heading', { name: 'Checkout' })).toBeVisible();
}
