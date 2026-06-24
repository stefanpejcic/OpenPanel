import { test, expect } from '@playwright/test';

async function navigateToDashboardPage(page: any) {
  await page.goto(`/dashboard`);
  await expect(page).toHaveURL(/dashboard/);
}

// OPEN
test('access dashboard', async ({ page }) => {
  await navigateToDashboardPage(page);
  await expect(page.getByText(/cpu usage/i)).toBeVisible();
  console.log('Dashboard page is accessible');
});


// SIDEBAR OPEN/CLOSE
test('sidebar open/close', async ({ page }) => {
  await navigateToDashboardPage(page);

  const html = page.locator('html');
  const sidebar = page.locator('#sidebar');

  await expect(html).toHaveClass(/sidebar-open/);
  await expect(sidebar).toBeVisible();

  await page.locator('#sidebar_toggle_button').click();
  await expect(html).not.toHaveClass(/sidebar-open/);
  await expect(sidebar).not.toBeVisible();

  await page.locator('#sidebar_toggle_button').click();
  await expect(html).toHaveClass(/sidebar-open/);
  await expect(sidebar).toBeVisible();

  console.log('sidebar toggle working');
});


// THEME SWITCH
test('toggle dark mode', async ({ page }) => {
  await navigateToDashboardPage(page);

  await page.locator('#user-btn-info').click();

  await expect(page.locator('#theme-toggle-dark-icon')).toBeVisible();

  await page.locator('#theme-toggle-dark-icon').click();
  await expect(page.locator('#theme-toggle-light-icon')).toBeVisible();
  await expect(page.locator('html')).toHaveClass(/dark/);

  await page.locator('#theme-toggle-light-icon').click();
  await expect(page.locator('#theme-toggle-dark-icon')).toBeVisible();
  await expect(page.locator('html')).not.toHaveClass(/dark/);

  console.log('Dark mode switch is working');
});

// SEARCH
test('search results', async ({ page }) => {
  test.setTimeout(300_000);
  await navigateToDashboardPage(page);
  await page.waitForFunction(() => typeof (window as any).Alpine !== 'undefined');
  await page.waitForTimeout(500);

  const items = ['Domains', 'DNS', 'Emails'];
  const failures: Array<{ name: string; error: string }> = [];
  let passed = 0;

  for (const name of items) {
    try {
      const openBtn = page.locator('button[aria-label="Open search"]');
      await expect(openBtn).toBeVisible({ timeout: 3000 });
      await openBtn.click();
      const searchInput = page.locator('#searchInput');
      await expect(searchInput).toBeVisible({ timeout: 2000 });
      await expect(searchInput).toBeFocused({ timeout: 2000 });
      await searchInput.pressSequentially(name, { delay: 50 });
      const dropdown = page.locator('#filteredDropdown');
      await expect(dropdown).toBeVisible({ timeout: 5000 });
      const match = dropdown.locator('a').filter({ hasText: name }).first();
      await expect(match).toBeVisible({ timeout: 3000 });
      console.log(`✓ found: "${name}"`);
      passed++;
    } catch (err: any) {
      const msg = err?.message?.split('\n')[0] ?? String(err);
      console.log(`✗ failed: "${name}" — ${msg}`);
      failures.push({ name, error: msg });
    } finally {
      try {
        const closeBtn = page.locator('button[aria-label="Close search"]');
        if (await closeBtn.isVisible({ timeout: 1000 })) {
          await closeBtn.click();
          await page.locator('#searchInput').waitFor({ state: 'hidden', timeout: 3000 });
        }
      } catch {
        await page.keyboard.press('Escape');
        await page.waitForTimeout(300);
      }
    }
  }

  console.log('\n─── Search Test Summary ───────────────────────────');
  console.log(`  Passed : ${passed} / ${items.length}`);
  console.log(`  Failed : ${failures.length} / ${items.length}`);
  if (failures.length > 0) {
    console.log('\n  Failures:');
    for (const f of failures) console.log(`    ✗ "${f.name}": ${f.error}`);
  }
  console.log('───────────────────────────────────────────────────\n');

  if (failures.length > 0) {
    throw new Error(
      `${failures.length} item(s) not found in search:\n` +
      failures.map(f => `  • "${f.name}": ${f.error}`).join('\n')
    );
  }
});


// ICONS TOP/START
test('icons mode toggle', async ({ page }) => {
  await navigateToDashboardPage(page);

  const iconViewBtn = page.locator('button[title="Top"]');
  const listViewBtn = page.locator('button[title="Start"]');

  await listViewBtn.click();
  const listValue = await page.evaluate(() => localStorage.getItem('dashboard_icon_view'));
  expect(listValue).toBe('list');
  await expect(listViewBtn).toHaveClass(/bg-slate-700/);
  await expect(iconViewBtn).not.toHaveClass(/bg-slate-700/);

  await iconViewBtn.click();
  const iconValue = await page.evaluate(() => localStorage.getItem('dashboard_icon_view'));
  expect(iconValue).toBe('icon');
  await expect(iconViewBtn).toHaveClass(/bg-slate-700/);
  await expect(listViewBtn).not.toHaveClass(/bg-slate-700/);

  console.log('icons toggle working');
});

// SORTABLE
test('icon sections drag&sort', async ({ page }) => {
  await navigateToDashboardPage(page);
  const firstSection = page.locator('#dashboard-sortable-area > [data-id]').nth(0);
  const secondSection = page.locator('#dashboard-sortable-area > [data-id]').nth(1);
  const firstKey = await firstSection.getAttribute('data-id');
  const secondKey = await secondSection.getAttribute('data-id');
  const handle = page.locator(`#section-title-${firstKey}`);
  const targetHandle = page.locator(`#section-title-${secondKey}`);
  const handleBox = await handle.boundingBox();
  const targetBox = await targetHandle.boundingBox();

  // Start drag from center of first handle
  await page.mouse.move(handleBox.x + handleBox.width / 2, handleBox.y + handleBox.height / 2);
  await page.mouse.down();
  await page.waitForTimeout(300);

  // Move past the midpoint of the target (SortableJS swaps at 50% threshold)
  const startY = handleBox.y + handleBox.height / 2;
  const endY = targetBox.y + targetBox.height * 0.75; // past midpoint of target
  const startX = handleBox.x + handleBox.width / 2;
  const steps = 20;

  for (let i = 1; i <= steps; i++) {
    await page.mouse.move(
      startX,
      startY + ((endY - startY) * i) / steps
    );
    await page.waitForTimeout(20);
  }

  // Hover briefly at destination before releasing
  await page.waitForTimeout(200);
  await page.mouse.up();
  await page.waitForTimeout(500);

  // Verify DOM order changed
  const newFirstKey = await page.locator('#dashboard-sortable-area > [data-id]').nth(0).getAttribute('data-id');
  const newSecondKey = await page.locator('#dashboard-sortable-area > [data-id]').nth(1).getAttribute('data-id');

  expect(newFirstKey).toBe(secondKey);
  expect(newSecondKey).toBe(firstKey);

  // Verify localStorage was updated
  const savedOrder = await page.evaluate(() =>
    JSON.parse(localStorage.getItem('dashboard_section_order'))
  );
  expect(savedOrder).not.toBeNull();
  expect(savedOrder[0]).toBe(secondKey);
  expect(savedOrder[1]).toBe(firstKey);

  console.log(`sections drag to sort working: moved '${firstKey}' below '${secondKey}'`);
});

// SECTION CLOSE/OPEN
test('icon sections open/close', async ({ page }) => {
  await navigateToDashboardPage(page);

  const firstSectionWrapper = page.locator('#dashboard-sortable-area > [data-id]').nth(0);
  const sectionKey = await firstSectionWrapper.getAttribute('data-id');

  const toggleBtn = firstSectionWrapper.locator('button[aria-expanded]');
  const iconsGrid = firstSectionWrapper.locator(`#section-icons-${sectionKey}`);

  await expect(toggleBtn).toHaveAttribute('aria-expanded', 'true');
  await expect(iconsGrid).toBeVisible();

  await toggleBtn.click();
  await page.waitForTimeout(350); // alpinejs x-collapse

  await expect(toggleBtn).toHaveAttribute('aria-expanded', 'false');
  await expect(iconsGrid).toBeHidden();

  const closedValue = await page.evaluate((key) =>
    localStorage.getItem(`section_state_${key}`)
  , sectionKey);
  expect(closedValue).toBe('false');

  await toggleBtn.click();
  await page.waitForTimeout(350);

  await expect(toggleBtn).toHaveAttribute('aria-expanded', 'true');
  await expect(iconsGrid).toBeVisible();

  const openedValue = await page.evaluate((key) =>
    localStorage.getItem(`section_state_${key}`)
  , sectionKey);
  expect(openedValue).toBe('true');

  console.log(`section toggle open/close working for section: '${sectionKey}'`);
});


// MENU ITEMS CLOSE/OPEN
test('menu items collapse/expand individually', async ({ page }) => {
  await navigateToDashboardPage(page);

  const mainSidebarGroup = page.locator('[data-sidebar="group"]').nth(1);
  const groupItems = mainSidebarGroup.locator('[data-sidebar="menu"] > li[x-data]');
  const count = await groupItems.count();

  console.log(`Found ${count} sidebar groups`);
  expect(count).toBeGreaterThan(0);

  for (let i = 0; i < count; i++) {
    const li = groupItems.nth(i);
    const button = li.locator(':scope > button');
    const submenu = li.locator(':scope > ul');
    const label = (await button.innerText()).trim().split('\n')[0].trim();

    const isHidden = await submenu.evaluate(el => el.style.display === 'none');
    if (!isHidden) {
      await button.click();
      await expect(submenu).toBeHidden({ timeout: 2000 });
    }

    console.log(`[${label}] starting collapsed ✓`);

    await button.click();
    await submenu.waitFor({ state: 'visible', timeout: 3000 });
    await expect(submenu).toBeVisible();
    console.log(`[${label}] expanded ✓`);

    await button.click();
    await submenu.waitFor({ state: 'hidden', timeout: 3000 });
    await expect(submenu).toBeHidden();
    console.log(`[${label}] collapsed ✓`);
  }

  console.log('All sidebar groups collapse/expand correctly');
});

// MENU ITEMS COLLAPSE/EXPAND ALL
test('menu items collapse/expand all', async ({ page }) => {
  await navigateToDashboardPage(page);

  const toggleButton = page.locator('button[\\@click*="sidebar-toggle-all"]');
  const mainSidebarGroup = page.locator('[data-sidebar="group"]').nth(1);
  const groupItems = mainSidebarGroup.locator('[data-sidebar="menu"] > li[x-data]');
  const count = await groupItems.count();

  expect(count).toBeGreaterThan(0);

  await expect(toggleButton).toContainText('Expand all');
  await toggleButton.click();
  await expect(toggleButton).toContainText('Collapse all');

  for (let i = 0; i < count; i++) {
    const submenu = groupItems.nth(i).locator(':scope > ul');
    const label = (await groupItems.nth(i).locator(':scope > button').innerText()).trim().split('\n')[0].trim();
    await submenu.waitFor({ state: 'visible', timeout: 3000 });
    await expect(submenu).toBeVisible();
    console.log(`[${label}] expanded ✓`);
  }

  await toggleButton.click();
  await expect(toggleButton).toContainText('Expand all');

  for (let i = 0; i < count; i++) {
    const submenu = groupItems.nth(i).locator(':scope > ul');
    const label = (await groupItems.nth(i).locator(':scope > button').innerText()).trim().split('\n')[0].trim();
    await submenu.waitFor({ state: 'hidden', timeout: 3000 });
    await expect(submenu).toBeHidden();
    console.log(`[${label}] collapsed ✓`);
  }

  console.log('Collapse all / Expand all working correctly');
});
