const { test, expect } = require('@playwright/test');

test('Resource Usage page loads with gauges', async ({ page }) => {
    await page.goto('/server/usage');

    // Header should be visible
    await expect(page.locator('h1')).toContainText('Resource Usage');

    // Link to history page should exist
    const historyLink = page.locator('a[href="/server/usage/history"]');
    await expect(historyLink).toBeVisible();
    await expect(historyLink).toContainText('View Usage History');
});

test('Resource Usage page shows CPU and RAM gauges when data available', async ({ page }) => {
    await page.goto('/server/usage');

    const cpuGauge = page.locator('#cpuGauge');
    const ramGauge = page.locator('#ramGauge');

    // If gauges are present, data was returned
    if (await cpuGauge.isVisible()) {
        await expect(ramGauge).toBeVisible();
        await expect(page.locator('text=Current CPU usage')).toBeVisible();
        await expect(page.locator('text=Used Memory')).toBeVisible();
        await expect(page.locator('text=Last refreshed at')).toBeVisible();
    } else {
        // No data state
        await expect(page.locator('text=No resource usage data available yet')).toBeVisible();
    }
});

test('Usage History page loads', async ({ page }) => {
    await page.goto('/server/usage/history');

    await expect(page.locator('h1')).toContainText('Historical CPU and Memory Usage');

    // Table should be present
    const table = page.locator('table#usage');
    await expect(table).toBeVisible();

    // Table headers
    await expect(table).toContainText('Date');
    await expect(table).toContainText('CPU');
    await expect(table).toContainText('Memory');
    await expect(table).toContainText('Bandwidth Usage');
    await expect(table).toContainText('Warning');
});

test('Usage History search filters rows', async ({ page }) => {
    await page.goto('/server/usage/history');

    const table = page.locator('table#usage tbody');
    const rows = table.locator('tr');
    const rowCount = await rows.count();

    // Only run search test if there is data
    if (rowCount === 0) {
        console.log('No usage data rows to test search against, skipping.');
        return;
    }

    const searchInput = page.locator('[x-model="searchQuery"]');
    await expect(searchInput).toBeVisible();

    // Search for something that won't match
    await searchInput.fill('zzznomatchzzz');
    const visibleRows = page.locator('table#usage tbody tr:visible');
    await expect(visibleRows).toHaveCount(0, { timeout: 3000 });

    // Clear search - rows should return
    await searchInput.fill('');
    await expect(page.locator('table#usage tbody tr').first()).toBeVisible({ timeout: 3000 });
});

test('Usage History show all checkbox loads all data', async ({ page }) => {
    await page.goto('/server/usage/history');

    const checkbox = page.locator('#showAllCheckboxOnHistoryPage');
    await expect(checkbox).toBeVisible();

    // Check the checkbox
    await checkbox.check();

    // URL should update with show_all=true
    await page.waitForURL('**/server/usage/history?show_all=true', { timeout: 5000 });
    await expect(page.locator('text=Showing 1 -')).toBeVisible();
});

test('Usage History show all checkbox can be unchecked', async ({ page }) => {
    await page.goto('/server/usage/history?show_all=true');

    const checkbox = page.locator('#showAllCheckboxOnHistoryPage');
    await expect(checkbox).toBeChecked();

    await checkbox.uncheck();

    // show_all param should be removed from URL
    await page.waitForURL(url => !url.toString().includes('show_all'), { timeout: 5000 });
});

test('Usage History pagination works', async ({ page }) => {
    await page.goto('/server/usage/history');

    const nextLink = page.locator('a', { hasText: 'Next' });

    // Only test pagination if Next exists (i.e. more than one page)
    if (await nextLink.isVisible()) {
        await nextLink.click();
        await page.waitForLoadState('networkidle');

        // URL should now have page=2
        expect(page.url()).toContain('page=2');

        // Previous button should now appear
        await expect(page.locator('a', { hasText: 'Previous' })).toBeVisible();
    } else {
        console.log('Only one page of data, skipping pagination test.');
    }
});

test('View Usage History button navigates correctly', async ({ page }) => {
    await page.goto('/server/usage');

    const historyBtn = page.locator('a[href="/server/usage/history"]');
    await historyBtn.click();

    await page.waitForLoadState('networkidle');
    expect(page.url()).toContain('/server/usage/history');
    await expect(page.locator('h1')).toContainText('Historical CPU and Memory Usage');
});
