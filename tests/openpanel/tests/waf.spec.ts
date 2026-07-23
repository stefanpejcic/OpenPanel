import { test, expect, Page } from '@playwright/test';

// Toggle lives on the list page, one per domain row → must be scoped
function wafToggle(page: Page, domain: string) {
  return page.getByRole('row').filter({ hasText: domain })
             .locator('button[aria-checked]');
}

async function setWaf(page: Page, domain: string, desiredOn: boolean) {
  await page.goto('/server/waf');
  const toggle = wafToggle(page, domain);
  await expect(toggle).toBeVisible();
  if (await toggle.getAttribute('aria-checked') !== String(desiredOn)) {
    await toggle.click();
    await expect(toggle).toHaveAttribute('aria-checked', String(desiredOn));
  }
}

async function openDomainPage(page: Page, domain: string) {
  await page.goto(`/server/waf/${domain}`);
  await expect(page).toHaveURL(new RegExp(`/server/waf/${domain.replace(/\./g, '\\.')}$`));
}

test('waf status', async ({ page }) => {
  await page.goto('/server/waf');
  await expect(page.getByRole('heading', { name: 'WAF', level: 1 })).toBeVisible();
});

test('waf on/off and disabled rules for domain', async ({ page }) => {
  const domain = 'wp.tests.openpanel.org';

  const blockedUrl = `https://${domain}/?q=<script>alert(1)</script>`;
  const cleanUrl = `https://${domain}/`;

  // ── 1. WAF ON → blocked request should be blocked, clean should pass ───────
  await setWaf(page, domain, true);

  let blocked = await page.request.get(blockedUrl, { failOnStatusCode: false });
  let clean = await page.request.get(cleanUrl, { failOnStatusCode: false });
  expect([403, 406, 409, 422]).toContain(blocked.status());
  expect(clean.status()).toBe(200);

  // ── 2. WAF OFF → both requests should pass ─────────────────────────────────
  await setWaf(page, domain, false);

  blocked = await page.request.get(blockedUrl, { failOnStatusCode: false });
  clean = await page.request.get(cleanUrl, { failOnStatusCode: false });
  expect(blocked.status()).toBe(200);
  expect(clean.status()).toBe(200);

  // ── 3. WAF ON + disable rule by ID → previously blocked request now passes ─
  await setWaf(page, domain, true);

  blocked = await page.request.get(blockedUrl, { failOnStatusCode: false });
  expect([403, 406, 409, 422]).toContain(blocked.status());

  await openDomainPage(page, domain);

  const ruleId = '941100';
  const removedRules = page.locator('#removed_rules');
  const saveButton = page.getByRole('button', { name: 'Save' });

  await removedRules.fill(ruleId);
  await saveButton.click();
  await expect(removedRules).toHaveValue(ruleId);

  blocked = await page.request.get(blockedUrl, { failOnStatusCode: false });
  clean = await page.request.get(cleanUrl, { failOnStatusCode: false });
  expect(blocked.status()).toBe(200);
  expect(clean.status()).toBe(200);

  // Clear disabled rule IDs
  await removedRules.fill('');
  await saveButton.click();
  await expect(removedRules).toHaveValue('');

  blocked = await page.request.get(blockedUrl, { failOnStatusCode: false });
  expect([403, 406, 409, 422]).toContain(blocked.status());

  // ── 4. WAF ON + disable rule by TAG → previously blocked request now passes ─
  const ruleTag = 'attack-xss';
  const removedTags = page.locator('#removed_tags');

  await removedTags.fill(ruleTag);
  await saveButton.click();
  await expect(removedTags).toHaveValue(ruleTag);

  blocked = await page.request.get(blockedUrl, { failOnStatusCode: false });
  clean = await page.request.get(cleanUrl, { failOnStatusCode: false });
  expect(blocked.status()).toBe(200);
  expect(clean.status()).toBe(200);

  // ── 5. Cleanup – clear disabled tags, confirm blocking restored ────────────
  await removedTags.fill('');
  await saveButton.click();
  await expect(removedTags).toHaveValue('');

  blocked = await page.request.get(blockedUrl, { failOnStatusCode: false });
  expect([403, 406, 409, 422]).toContain(blocked.status());
});

test('waf logs show blocked requests for domain', async ({ page }) => {
  const domain = 'wp.tests.openpanel.org';
  const blockedUrl = `https://${domain}/?q=<script>alert(1)</script>`;

  await setWaf(page, domain, true);

  for (let i = 0; i < 3; i++) {
    const res = await page.request.get(blockedUrl, { failOnStatusCode: false });
    expect([403, 406, 409, 422]).toContain(res.status());
  }

  await page.goto(`/server/waf/log/${domain}`);

  const logsTable = page.locator('#waf-logs-table');
  await expect(logsTable).toBeVisible();
  await expect(page.locator('h1')).toContainText(domain);

  const rows = logsTable.locator('tbody tr');
  await expect(rows).not.toHaveCount(0);
  const rowCount = await rows.count();

  // Scoped so it can't collide with the sidebar search input
  const searchInput = page.getByRole('region', { name: /logs/i })
                          .locator('input[type="search"]');
  await searchInput.fill('alert(1)');

  await expect(rows.first()).toContainText('alert(1)');
  const filteredCount = await rows.count();
  expect(filteredCount).toBeGreaterThan(0);

  await searchInput.fill('');
  await expect(rows).toHaveCount(rowCount);
});
