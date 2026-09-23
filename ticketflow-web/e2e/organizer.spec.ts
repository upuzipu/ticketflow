import { test, expect } from '@playwright/test';
import { login } from './helpers';

test.describe('organizer', () => {
  test('publishes a draft event and it appears in the catalog', async ({ page }) => {
    await login(page, 'org@tf.dev', 'organizer');
    await page.getByRole('link', { name: 'My events' }).click();
    await expect(page).toHaveURL(/\/organizer/);
    await expect(page.getByText('Workshop: Go for Frontend Developers')).toBeVisible();
    await page.getByRole('button', { name: 'Publish' }).click();
    await expect(page.getByRole('button', { name: 'Publish' })).toHaveCount(0, { timeout: 15_000 });
    await page.goto('/events');
    await expect(
      page.getByRole('link', { name: /Workshop: Go for Frontend Developers/ }),
    ).toBeVisible();
  });

  test('buyer is denied the organizer area', async ({ page }) => {
    await login(page, 'buyer@tf.dev', 'buyer');
    await page.goto('/organizer');
    await expect(page.getByRole('heading', { name: 'Access restricted' })).toBeVisible();
  });
});
