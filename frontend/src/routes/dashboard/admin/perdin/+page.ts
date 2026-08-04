import { api } from '$lib/shared/api';

export async function load({ fetch }) {
    try {
        if (typeof window !== 'undefined' && localStorage.getItem('auth_token')) {
            // 🔥 NON-BLOCKING ROUTING
            return {
                recordsResponse: null
            };
        }
    } catch (e) {
        console.warn("Failed to prefetch admin records:", e);
    }
    return { recordsResponse: null };
}
