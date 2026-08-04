import { writable } from 'svelte/store';
import { api } from '$lib/shared/api';
import { browser } from '$app/environment';
import { toTitleCase } from '$lib/shared/utils/utils';

export const PROVINCES = [
    { name: "Aceh", luarKota: 360000 },
    { name: "Sumatra Utara", luarKota: 370000 },
    { name: "Riau", luarKota: 370000 },
    { name: "Kepulauan Riau", luarKota: 370000 },
    { name: "Jambi", luarKota: 370000 },
    { name: "Sumatra Barat", luarKota: 380000 },
    { name: "Sumatra Selatan", luarKota: 380000 },
    { name: "Lampung", luarKota: 380000 },
    { name: "Bengkulu", luarKota: 380000 },
    { name: "Bangka Belitung", luarKota: 410000 },
    { name: "Banten", luarKota: 370000 },
    { name: "Jawa Barat", luarKota: 430000 },
    { name: "DKI Jakarta", luarKota: 530000 },
    { name: "Jawa Tengah", luarKota: 370000 },
    { name: "DI Yogyakarta", luarKota: 420000 },
    { name: "Jawa Timur", luarKota: 410000 },
    { name: "Bali", luarKota: 480000 },
    { name: "Nusa Tenggara Barat", luarKota: 440000 },
    { name: "Nusa Tenggara Timur", luarKota: 430000 },
    { name: "Kalimantan Barat", luarKota: 380000 },
    { name: "Kalimantan Tengah", luarKota: 360000 },
    { name: "Kalimantan Selatan", luarKota: 380000 },
    { name: "Kalimantan Timur", luarKota: 430000 },
    { name: "Kalimantan Utara", luarKota: 430000 },
    { name: "Sulawesi Utara", luarKota: 370000 },
    { name: "Gorontalo", luarKota: 370000 },
    { name: "Sulawesi Barat", luarKota: 410000 },
    { name: "Sulawesi Selatan", luarKota: 430000 },
    { name: "Sulawesi Tengah", luarKota: 370000 },
    { name: "Sulawesi Tenggara", luarKota: 380000 },
    { name: "Maluku", luarKota: 380000 },
    { name: "Maluku Utara", luarKota: 430000 },
    { name: "Papua", luarKota: 580000 },
    { name: "Papua Barat", luarKota: 480000 },
    { name: "Papua Barat Daya", luarKota: 480000 },
    { name: "Papua Tengah", luarKota: 580000 },
    { name: "Papua Selatan", luarKota: 580000 },
    { name: "Papua Pegunungan", luarKota: 580000 }
];

export const STAKEHOLDERS = [
    "Menteri Ketenagakerjaan",
    "Wakil Menteri Ketenagakerjaan",
    "Pendamping Menteri Ketenagakerjaan",
    "Pendamping Wakil Menteri Ketenagakerjaan"
];

export const provincesStore = writable(PROVINCES);
export const sbmRatesStore = writable<any[]>([]);
export const stakeholdersStore = writable(STAKEHOLDERS);

export async function loadMasterData() {
    if (typeof window === 'undefined' || !localStorage.getItem('auth_token')) return;
    try {
        const [provinces, rates] = await Promise.all([
            api.getProvinces(),
            api.getSBMRates()
        ]);

        if (provinces && provinces.length > 0) {
            // Map backend Province to frontend format
            const mappedProvinces = provinces.map(p => {
                const rate = rates.find(r => r.province_id === p.id);
                return {
                    id: p.id,
                    name: toTitleCase(p.name),
                    luarKota: (rate && rate.outside_city_rate && rate.outside_city_rate.Valid) ? rate.outside_city_rate.Float64 : 0
                };
            });
            provincesStore.set(mappedProvinces);
        }
        sbmRatesStore.set(rates);
    } catch (e: any) {
        if (e.message === 'Unauthorized') {
             // Silence unauthorized errors during initial load, auth store will handle redirect
             return;
        }
        console.warn("Could not fetch master data from API (backend might be starting), using defaults.", e.message);
    }
}

