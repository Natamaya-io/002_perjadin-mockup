import { writable } from 'svelte/store';
import { api } from '$lib/shared/api';
import type { TravelRecord } from '$lib/shared/api/types';

import { browser } from '$app/environment';

// --- Travel Records ---
export const recordsStore = writable<TravelRecord[]>([]);
export const paginatedRecordsStore = writable<TravelRecord[]>([]);
export const paginatedMetadataStore = writable<{totalItems: number, totalRecords: number, nextCursor: string | null, limit: number}>({
    totalItems: 0,
    totalRecords: 0,
    nextCursor: null,
    limit: 50
});

export const isFetchingRecords = writable<boolean>(false);

export async function loadRecords(spd?: string) {
    if (typeof window === 'undefined' || !localStorage.getItem('auth_token')) return;
    isFetchingRecords.set(true);
    try {
        const filters = spd ? { spd } : { limit: 100 }; // Prevent over-fetching by adding a limit if no spd
        const data = await api.getRecords(filters);
        recordsStore.set(data || []);
    } catch (e: any) {
        if (e.message === 'Unauthorized') return;
        console.warn("Failed to load records (backend might be starting):", e.message);
    } finally {
        isFetchingRecords.set(false);
    }
}

export async function loadPaginatedRecords(params: import('$lib/shared/api/types').PaginatedParams, append = false) {
    if (typeof window === 'undefined' || !localStorage.getItem('auth_token')) return;
    
    isFetchingRecords.set(true);
    try {
        const response = await api.getPaginatedRecords(params);
        if (append) {
            paginatedRecordsStore.update(current => [...current, ...(response.data || [])]);
        } else {
            paginatedRecordsStore.set(response.data || []);
        }
        paginatedMetadataStore.set({
            totalItems: response.totalItems,
            totalRecords: response.totalRecords,
            nextCursor: response.nextCursor || null,
            limit: response.limit
        });
    } catch (e: any) {
        if (e.message === 'Unauthorized') return;
        console.warn("Failed to load paginated records:", e.message);
    } finally {
        isFetchingRecords.set(false);
    }
}

// Export a reset function
export function clearStores() {
    recordsStore.set([]);
    paginatedRecordsStore.set([]);
    paginatedMetadataStore.set({
        totalItems: 0,
        totalRecords: 0,
        nextCursor: null,
        limit: 50
    });
}

export async function addRecord(tripData: any) {
    try {
        const newRecords = await api.createRecord(tripData);
        recordsStore.update(current => [...newRecords, ...current]);
        paginatedRecordsStore.update(current => [...newRecords, ...current]);
        paginatedMetadataStore.update(m => ({ ...m, totalItems: m.totalItems + 1, totalRecords: m.totalRecords + newRecords.length }));
        return newRecords;
    } catch (e) {
        console.error("Failed to add record", e);
        throw e;
    }
}

export async function updateRecord(id: string, data: Partial<TravelRecord>) {
    try {
        const updated = await api.updateRecord(id, data);
        // Backend strips costs/reportData from response for performance (to avoid serializing large base64 files).
        // Preserve existing store values when backend returns null for these fields.
        const mergeRecord = (existing: TravelRecord) => {
            if (existing.id !== id) return existing;
            const merged = { ...existing, ...updated };
            if (updated.costs === undefined || updated.costs === null) merged.costs = existing.costs;
            if (updated.reportData === undefined || updated.reportData === null) merged.reportData = existing.reportData;
            if (!updated.employee?.id) merged.employee = existing.employee;
            return merged;
        };
        recordsStore.update(current => current.map(mergeRecord));
        paginatedRecordsStore.update(current => current.map(mergeRecord));
        return updated;
    } catch (e) {
        console.error("Failed to update record", e);
        throw e;
    }
}

export async function updateMultipleRecords(updates: Array<{id: string, data: Partial<TravelRecord>}>) {
    try {
        // Run updates sequentially to avoid potential race conditions in simple backends
        const results = [];
        for (const update of updates) {
            const res = await api.updateRecord(update.id, update.data);
            results.push(res);
        }
        
        // Refresh full store to ensure consistency
        await loadRecords();
        return results;
    } catch (e) {
        console.error("Failed to update multiple records", e);
        throw e;
    }
}

export async function deleteRecord(id: string) {
    try {
        await api.deleteRecord(id);
        recordsStore.update(current => current.filter(r => r.id !== id));
    } catch (e) {
        console.error(`Failed to delete record ${id}`, e);
        throw e;
    }
}

export async function deleteRecordBySpd(spd: string) {
    // Optimistic update: remove from UI immediately
    recordsStore.update(current => current.filter(r => r.spd !== spd));
    
    let deletedCount = 0;
    paginatedRecordsStore.update(current => {
        const filtered = current.filter(r => r.spd !== spd);
        deletedCount = current.length - filtered.length;
        return filtered;
    });
    if (deletedCount > 0) {
        paginatedMetadataStore.update(m => ({ ...m, totalItems: Math.max(0, m.totalItems - 1), totalRecords: Math.max(0, m.totalRecords - deletedCount) }));
    }

    try {
        await api.deleteRecordsBySpd(spd);
    } catch (e: any) {
        // If it's a 502, it might have actually succeeded in the backend
        // We check if the records are actually gone
        if (e.message?.includes('502')) {
             console.warn("Detected 502 during delete, verifying data status...");
             await loadRecords();
             
             let exists = false;
             recordsStore.subscribe(recs => {
                 exists = recs.some(r => r.spd === spd);
             })();

             if (!exists) {
                 // Success! Data is gone despite the 502 error
                 return;
             }
        }

        console.error(`Failed to delete records for SPD ${spd}`, e);
        // Rollback: reload from server if it actually failed
        await loadRecords();
        throw e;
    }
}
