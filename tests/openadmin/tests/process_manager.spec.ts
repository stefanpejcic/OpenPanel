import { test, expect } from '@playwright/test';

test.describe('processes table', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/server/processes');
    await expect(page).toHaveURL(/server\/processes/);
  });

  test('search', async ({ page }) => {
    const table = page.locator('#processes_table');
    const rows = table.locator('tbody tr:visible');
    const searchInput = page.getByPlaceholder('Search processes...');

    const initialCount = await rows.count();

    await searchInput.fill('testinguser');
    await expect(rows.first()).toContainText('testinguser');

    const testingUserCount = await rows.count();
    console.log(`search testinguser: ${testingUserCount}`);
    expect(testingUserCount).not.toBe(initialCount);

    await searchInput.fill('dockerd');
    await expect(rows.first()).toContainText('dockerd');

    const dockerdCount = await rows.count();
    console.log(`search dockerd: ${dockerdCount}`);
    expect(dockerdCount).not.toBe(initialCount);

    await expect(table).toBeVisible();
  });

  test('asc/desc sorting', async ({ page }) => {
    const table = page.locator('#processes_table');

    const columns = [
      { name: 'Pid', asc: 'pid', desc: '-pid' },
      { name: 'Owner', asc: 'owner', desc: '-owner' },
      { name: 'Priority', asc: 'priority', desc: '-priority' },
      { name: 'CPU %', asc: 'cpu', desc: '-cpu' },
      { name: 'Memory %', asc: 'memory', desc: '-memory' },
      { name: 'Command', asc: 'command', desc: '-command' },
    ];

    const getColumnValues = async (colIndex: number) =>
      table.locator(`tbody tr td:nth-child(${colIndex + 1})`).allInnerTexts();

    const normalize = (vals: string[]) =>
      vals.map(v => v.replace(/\u00a0/g, '').trim()).filter(Boolean);

    for (let i = 0; i < columns.length; i++) {
      const col = columns[i];
      const header = table.locator('th', { hasText: col.name });

      await header.locator(`a[href*="sort=${col.asc}"]`).click();
      await expect(page).toHaveURL(new RegExp(`sort=${col.asc}`));

      const asc = normalize(await getColumnValues(i));

      await header.locator(`a[href*="sort=${col.desc}"]`).click();
      await expect(page).toHaveURL(new RegExp(`sort=${col.desc}`));

      const desc = normalize(await getColumnValues(i));

      console.log(`[${col.name}] asc: ${asc[0]} → ${asc.at(-1)}, desc: ${desc[0]} → ${desc.at(-1)}`);

      expect(asc.length).toBeGreaterThan(0);
      expect(desc.length).toBeGreaterThan(0);
      expect(asc[0]).not.toBe(desc[0]);
      expect(asc.at(-1)).not.toBe(desc.at(-1));
    }
  });

  test('strace', async ({ page }) => {
    const rows = page.locator('#processes_table tbody tr:visible');

    const userRow = rows.filter({ hasText: 'testinguser' }).first();
    await userRow.getByRole('link', { name: 'Trace' }).click();

    await expect(page).toHaveURL(/\/server\/processes\/\d+\/strace/);
    await expect(page.locator('body')).toContainText(
      /strace:\s*Process\s+\d+\s+attached/
    );
  });

  test('kill', async ({ page }) => {
    const userRow = page
      .locator('#processes_table tbody tr', { hasText: 'testinguser' })
      .first();

    const pid = (await userRow.locator('td').first().innerText()).trim();

    await userRow.getByRole('link', { name: 'Kill' }).click();

    await expect(page.locator('body')).toContainText(
      new RegExp(`Process with PID ${pid} killed successfully`)
    );

    await expect(page.locator('#processes_table')).not.toContainText(pid);
  });
});
