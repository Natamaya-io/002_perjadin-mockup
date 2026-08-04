import { writable } from 'svelte/store';

export const settingsStore = writable({
    ppk_name: '',
    ppk_nip: '',
    bendahara_name: '',
    bendahara_nip: '',
    isLoaded: false
});
