import { test, expect, Page } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

let killedPid: string;

test.describe('Process Manager', () => {

    test('loads process table with at least one row', async ({ page }) => {
        await page.goto(`/process-manager`);
        await page.waitForLoadState('networkidle');

        const rows = page.locator('tbody tr[x-show]');
        const count = await rows.count();
        expect(count).toBeGreaterThan(0);
    });

    test('kill a random process, expect toast and row removed', async ({ page }) => {
        await page.goto(`/process-manager`);
        await page.waitForLoadState('networkidle');

        const rows = page.locator('tbody tr[x-show]');
        const count = await rows.count();
        expect(count).toBeGreaterThan(0);

        // Pick a random row
        const targetIndex = Math.floor(Math.random() * count);
        const targetRow = rows.nth(targetIndex);

        // Grab the PID from the 3rd td (index 2)
        killedPid = (await targetRow.locator('td').nth(2).innerText()).trim();

        const terminateBtn = targetRow.locator('button', { hasText: 'Terminate' });

        // Listen for the POST response
        const responsePromise = page.waitForResponse(r => r.url().includes('/process-manager') && r.request().method() === 'POST');

        await terminateBtn.click();
        const response = await responsePromise;
        const body = await response.json();
        console.log('Kill response:', JSON.stringify(body));
        expect(body.success).toBe(true);

        // Row should be gone from DOM
        await expect(targetRow).not.toBeAttached();
    });

    test('refresh and confirm killed PID is absent', async ({ page }) => {
        await page.goto(`/process-manager`);
        await page.waitForLoadState('networkidle');

        // Click the Refresh button
        await page.locator('a', { hasText: 'Refresh Processes' }).click();
        await page.waitForLoadState('networkidle');

        const allPids = await page.locator('tbody tr[x-show] td:nth-child(3)').allInnerTexts();
        expect(allPids.map(p => p.trim())).not.toContain(killedPid);
    });

    test('search filters rows correctly', async ({ page }) => {
        await page.goto(`/process-manager`);
        await page.waitForLoadState('networkidle');

        const rows = page.locator('tbody tr[x-show]');
        const firstRow = rows.first();

        // Grab container name from first column to use as search term
        const containerName = (await firstRow.locator('td').first().innerText()).trim();

        const searchInput = page.locator('input[type="search"]');
        await searchInput.fill(containerName);

        // Alpine x-show filters are DOM-based, wait for them to settle
        await page.waitForTimeout(300);

        // All visible rows should contain the search term in container, pid, or cmd
        const visibleRows = page.locator('tbody tr[x-show]:visible');
        const visibleCount = await visibleRows.count();
        expect(visibleCount).toBeGreaterThan(0);

        for (let i = 0; i < visibleCount; i++) {
            const row = visibleRows.nth(i);
            const container = (await row.locator('td').nth(0).innerText()).trim();
            const pid = (await row.locator('td').nth(2).innerText()).trim();
            const cmd = (await row.locator('td').nth(8).innerText()).trim();
            const matchesAny = [container, pid, cmd].some(v =>
                v.toLowerCase().includes(containerName.toLowerCase())
            );
            expect(matchesAny).toBe(true);
        }

        // Clear search and confirm more rows appear
        await searchInput.fill('');
        await page.waitForTimeout(300);
        const totalVisible = await page.locator('tbody tr[x-show]:visible').count();
        expect(totalVisible).toBeGreaterThan(visibleCount);
    });

});
