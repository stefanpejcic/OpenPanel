import { test, expect } from '@playwright/test';

// NOTE: verify-only -- per the template copy itself, "Once enabled, it can't be turned
// off from the admin panel" (requires `opencli config update demo_mode off` over SSH).
// Never submit "Enable Demo Mode" from a test.

test('demo mode page renders branch matching current state', async ({ page }) => {
  await page.goto('/server/demo-mode');
  await expect(page).toHaveURL(/server\/demo-mode/);

  const offBranch = page.getByRole('button', { name: 'Enable Demo Mode' });
  const onBranch = page.getByText('Demo mode is active and can only be disabled from the terminal.');
  const unsupported = page.getByText('DEMO MODE IS NOT SUPPORTED ON YOUR INSTALLATION');

  if (await offBranch.isVisible().catch(() => false)) {
    await expect(page.getByText("can't be turned off from the admin panel")).toBeVisible();
    await expect(page.getByRole('link', { name: 'Cancel' })).toBeVisible();
    console.log('demo mode: off branch rendered, enable button present but not clicked');
  } else if (await onBranch.isVisible().catch(() => false)) {
    console.log('demo mode: already enabled, disable-from-terminal notice shown');
  } else {
    await expect(unsupported).toBeVisible();
    console.log('demo mode: unsupported installation branch shown');
  }
});
