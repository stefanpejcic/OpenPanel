import { test, expect } from '@playwright/test';

// The page renders one of three branches depending on whether imunify360-agent is
// installed and whether its GUI is running on :9000. We assert whichever branch is
// actually active rather than assuming a fixed environment state.

test('imunify page renders the branch matching current service state', async ({ page }) => {
  await page.goto('/security/imunify/');
  await expect(page).toHaveURL(/security\/imunify/);

  const notInstalled = page.getByText('Imunify service is not installed.');
  const notRunning = page.getByText('Imunify GUI is not running.');
  const iframe = page.locator('iframe#myiframe');

  if (await notInstalled.isVisible().catch(() => false)) {
    await expect(page.getByText('opencli imunify install')).toBeVisible();
    console.log('imunify: not installed branch rendered correctly');
  } else if (await notRunning.isVisible().catch(() => false)) {
    await expect(page.getByText('opencli imunify start')).toBeVisible();
    console.log('imunify: not running branch rendered correctly');
  } else {
    await expect(iframe).toBeVisible();
    await expect(iframe).toHaveAttribute('src', /\/imav\//);
    console.log('imunify: running branch rendered with iframe to /imav/');
  }
});
