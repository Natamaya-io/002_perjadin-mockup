import { api } from '$lib/shared/api';

export async function load({ fetch }) {
    try {
        if (typeof window !== 'undefined' && localStorage.getItem('auth_token')) {
            // 🔥 NON-BLOCKING ROUTING
            // We do not await the API here. We let the page render instantly (0ms)
            // and allow the component's reactive filters to fetch data in the background.
            return {
                recordsResponse: null
            };
        }
    } catch (e) {
        console.warn("Failed to prefetch pengajuan records:", e);
    }
    return { recordsResponse: null };
}
