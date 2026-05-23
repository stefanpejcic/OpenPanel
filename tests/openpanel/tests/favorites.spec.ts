const { test, expect } = require('@playwright/test');


  test('Left-click to add', async ({ page }) => {
      await page.goto(`/dashboard`);

      // Click the star button (left-click to add)
      const starBtn = page.locator('#addFavoriteBtn');
      await expect(starBtn).toBeVisible();
      await starBtn.click();

      // Expect success toast message
      await expect(page.locator('text=Successfully added to favorites')).toBeVisible({ timeout: 5000 });

      // Expect the favorite to appear in the left sidebar menu
      const sidebar = page.locator('nav, aside, [class*="sidebar"], [class*="menu"]').first();
      await expect(sidebar).toContainText('Dashboard', { timeout: 5000 });
  });

  test('check table', async ({ page }) => {
      // Navigate to favorites page and verify the entry exists
      await page.goto(`/account/favorites`);

      const table = page.locator('table tbody');
      await expect(table).toContainText('Dashboard', { timeout: 5000 });
  });

  test('Yellow star', async ({ page }) => {
      await page.goto(`/dashboard`);

      const starBtn = page.locator('#addFavoriteBtn');
      await expect(starBtn).toBeVisible();

      // The filled star SVG has an orange fill path
      const filledStar = starBtn.locator('svg path[fill="orange"]');
      await expect(filledStar).toBeVisible({ timeout: 5000 });

      await page.waitForLoadState('networkidle');
      const isFavorite = await starBtn.getAttribute('data-is-favorite');
      expect(isFavorite).toBe('true');
  });

  test('Right-click to remove', async ({ page }) => {
      await page.goto(`/dashboard`);

      const starBtn = page.locator('#addFavoriteBtn');
      await expect(starBtn).toBeVisible();

      // Right-click to remove
      await starBtn.click({ button: 'right' });

      // Expect success toast
      await expect(page.locator('text=Successfully removed from favorites')).toBeVisible({ timeout: 5000 });

      // Expect the item to disappear from the left sidebar
      await expect(page.getByText('Dashboard')).toHaveCount(1); // 1 is in header!

      await page.goto(`/account/favorites`);
      const table = page.locator('table tbody');
      await expect(table).not.toContainText('Dashboard', { timeout: 5000 });

  });

  test('search table', async ({ page }) => {
      await page.goto(`/dashboard`);

      const starBtn = page.locator('#addFavoriteBtn');
      await expect(starBtn).toBeVisible();
      await starBtn.click();

      await expect(page.locator('text=Successfully added to favorites')).toBeVisible({ timeout: 5000 });
  
      await page.goto(`/account/favorites`);
      const table = page.locator('table tbody');
      await expect(table).toContainText('Dashboard', { timeout: 5000 });

      const searchInput = page.locator('[x-model="searchQuery"]');
      await expect(searchInput).toBeVisible();

      // Search for "dashboard" - row should be visible
      await searchInput.fill('dashboard');
      const dashboardRow = page.locator('table tbody tr').filter({ hasText: 'Dashboard' });
      await expect(dashboardRow).toBeVisible({ timeout: 3000 });

      // Search for "lazr" - no rows should be visible
      await searchInput.fill('lazr');
      const allRows = page.locator('table tbody tr:visible');
      await expect(allRows).toHaveCount(0, { timeout: 3000 });

      // Clear search
      await searchInput.fill('');

      const link = dashboardRow.locator('a');
      await expect(link).toBeVisible();
      const href = await link.getAttribute('href');
      expect(href).toBe('/dashboard');
    
  });

  test('delete in table', async ({ page }) => {
      await page.goto(`/account/favorites`);

      const dashboardRow = page.locator('table tbody tr').filter({ hasText: 'Dashboard' });
      const deleteBtn = dashboardRow.locator('button', { hasText: 'Delete' });

      await expect(deleteBtn).toBeVisible();
      await deleteBtn.click();

      // Page reloads and row should be gone
      // await page.goto(`/account/favorites`);
     await page.waitForLoadState('networkidle');

      const table = page.locator('table tbody');
      await expect(table).not.toContainText('Dashboard', { timeout: 5000 });

      // Sidebar should not contain the favorite
      await expect(page.getByText('Dashboard')).toHaveCount(0);

      await page.goto(`/dashboard`);
      const starBtn = page.locator('#addFavoriteBtn');
      await expect(starBtn).toBeVisible();

      // Should show outline star (no orange fill), data-is-favorite should be false
      const filledStar = starBtn.locator('svg path[fill="orange"]');
      await expect(filledStar).not.toBeVisible({ timeout: 5000 });

      await page.waitForLoadState('networkidle');
      const isFavorite = await starBtn.getAttribute('data-is-favorite');
      expect(isFavorite === null || isFavorite === "false").toBe(true);

  });
