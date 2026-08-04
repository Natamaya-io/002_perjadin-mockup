import { test, expect } from '@playwright/test';

test.describe('Session Expired UX', () => {
  test('should show Session Expired modal when API returns 401 and redirect to login on click', async ({ page }) => {
    // 1. Mock local storage to simulate a logged-in user
    await page.addInitScript(() => {
      window.localStorage.setItem('auth_token', 'fake-jwt-token');
      window.localStorage.setItem('user_session_v2', JSON.stringify({
        id: '123',
        email: 'test@kemnaker.go.id',
        role: 'admin',
        loggedIn: true
      }));
    });

    // 2. Intercept API calls and force them to return 401 Unauthorized
    // We intercept any request to our backend API
    await page.route('**/api/**', async (route) => {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Session Expired: Logged in from another device' }),
      });
    });

    // 3. Navigate to a protected page (e.g., dashboard)
    // SvelteKit will load the layout, initialize the store, and might fetch data.
    // We can just go to the root which is protected.
    await page.goto('/');

    // 4. Verify that the "Sesi Berakhir" modal appears
    // The exact text we expect from our SessionExpiredModal
    await expect(page.getByText('Sesi Berakhir')).toBeVisible({ timeout: 5000 });
    
    // Verify the specific error message is passed to the modal (translated to Indonesian)
    await expect(page.getByText('Akun Anda baru saja digunakan untuk login di perangkat atau browser lain')).toBeVisible();

    // 5. Click the "Login Kembali" button
    const loginButton = page.getByRole('button', { name: 'Login Kembali' });
    await expect(loginButton).toBeVisible();
    await loginButton.click();

    // 6. Verify that it redirects to the login page
    await expect(page).toHaveURL(/\/login/);
    
    // Check if token was removed
    const token = await page.evaluate(() => window.localStorage.getItem('auth_token'));
    expect(token).toBeNull();
  });
});
