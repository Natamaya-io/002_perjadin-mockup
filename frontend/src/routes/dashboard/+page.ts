import { api } from '$lib/shared/api';
import { dashboardSummaryStore } from '$lib/features/dashboard/store';

export function load({ fetch }) {
    try {
        if (typeof window !== 'undefined' && localStorage.getItem('auth_token')) {
            // 🔥 NON-BLOCKING FETCH (Stale-While-Revalidate)
            // We intentionally do NOT use `await` here.
            // This allows the SvelteKit router to navigate INSTANTLY (0ms lag).
            // The API call runs in the background and updates the reactive store directly.
            api.getDashboardSummary(fetch)
                .then(res => {
                    if (res) dashboardSummaryStore.set(res);
                })
                .catch(e => console.warn("Failed to background fetch dashboard summary:", e));
        }
    } catch (e) {
        console.warn("Failed to prefetch dashboard summary:", e);
    }
    
    // Return empty so the router proceeds without waiting
    return { summary: null };
}
