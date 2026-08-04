import { writable } from 'svelte/store';

export const loadingStore = writable(false);

export function startLoading() {
    loadingStore.set(true);
}

export function stopLoading() {
    loadingStore.set(false);
}
