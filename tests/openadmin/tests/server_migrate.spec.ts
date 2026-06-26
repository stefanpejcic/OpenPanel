import { test, expect } from '@playwright/test';

// NOTE: verify-only -- starting a migration is a one-way, long-running server action
// with no cancel/undo in the UI. We only confirm the pre-start form (or, if a migration
// is already running on this environment, the polling/progress UI) renders correctly.

test('migrate page shows start form or in-progress status', async ({ page }) => {
  await page.goto('/server/migrate');
  await expect(page).toHaveURL(/server\/migrate/);

  const startForm = page.locator('#tour-migrate-host');
  const progressOutput = page.locator('#migration_output');

  if (await startForm.isVisible().catch(() => false)) {
    await expect(page.locator('input[name="root"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Start Migration' })).toBeVisible();
    console.log('migrate page: no migration in progress, start form shown');
  } else {
    await expect(progressOutput).toBeVisible();
    await expect(page.locator('#migration_actions')).toBeVisible();
    console.log('migrate page: migration already in progress, status UI shown');
  }
});
