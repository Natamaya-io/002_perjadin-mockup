import { writable, get } from 'svelte/store';
import { api } from '$lib/shared/api';
import type { User } from '$lib/shared/api/types';
import { browser } from '$app/environment';
import { loadRecords, clearStores } from '$lib/features/pengajuan/store';
import { dashboardSummaryStore } from '$lib/features/dashboard/store';
import { loadMasterData } from '$lib/shared/stores/master-data';

const isBrowser = typeof window !== 'undefined';

// --- Session Management ---
// We initialize from localStorage directly for immediate hydration.
// But ideally, we should check `api.getCurrentUser()` on mount.
const storedUser = isBrowser ? localStorage.getItem('user_session_v2') : null;
const initialUser = storedUser ? JSON.parse(storedUser) : {
    id: null,
    email: null,
    role: null,
    loggedIn: false,
    requirePasswordChange: false,
    passwordModalDismissed: false
};

export const userStore = writable(initialUser);

// Store to track if session is expired and the reason why
export const sessionExpiredReason = writable<string | null>(null);

// Persist to localStorage whenever userStore changes
userStore.subscribe(val => {
    if (isBrowser) {
        localStorage.setItem('user_session_v2', JSON.stringify(val));
    }
});

// Auto-load data if session exists on init
if (initialUser.loggedIn && isBrowser && !initialUser.requirePasswordChange) {
    (async () => {
        await refreshUserProfile(); // Mutlak: Tunggu profil refresh agar ID sinkron
        await Promise.all([
            loadMasterData()
        ]);
    })();
}

export async function refreshUserProfile() {
    if (!isBrowser || !localStorage.getItem('auth_token')) return;
    try {
        const user = await api.getCurrentUser();
        if (user) {
            userStore.update(u => ({ ...u, ...user, loggedIn: true }));
        }
    } catch (e) {
        console.error("Failed to refresh user profile", e);
    }
}

export const login = async (email: string, password?: string) => {
    try {
        const { user, require_password_change } = await api.login(email, password);
        userStore.set({ 
            ...user, 
            loggedIn: true, 
            requirePasswordChange: require_password_change,
            passwordModalDismissed: false // Reset on new login
        });
        
        // Load data on login
        if (!require_password_change) {
            loadMasterData();
        }
        
        return { success: true, user, require_password_change };
    } catch (e: any) {
        console.error("Login failed", e);
        return { success: false, error: e.message };
    }
};

export const logout = async () => {
    await api.logout();
    userStore.set({ email: null, role: null, loggedIn: false });
    clearStores();
    dashboardSummaryStore.set(null);
    usersStore.set([]);
};

// Auto-logout on 401 Unauthorized from any API call
api.setUnauthorizedHandler((reason) => {
    console.warn("Session expired or unauthorized:", reason);
    sessionExpiredReason.set(reason || "Sesi Anda telah berakhir. Silakan login kembali.");
    
    // Clear user state but DO NOT redirect yet. Let the UI show the modal.
    // The user will be redirected to /login when they click the button in the modal.
    userStore.set({ email: null, role: null, loggedIn: false });
    clearStores();
    dashboardSummaryStore.set(null);
    usersStore.set([]);
    if (isBrowser) {
        localStorage.removeItem('auth_token');
    }
});

// --- User Database Management (Admin) ---
export const usersStore = writable<User[]>([]);
export const isFetchingUsers = writable(false);

export const loadUsers = async () => {
    if (!isBrowser || !localStorage.getItem('auth_token')) return;
    isFetchingUsers.set(true);
    try {
        const users = await api.getUsers();
        usersStore.set(users);
    } catch (e: any) {
        if (e.message === 'Unauthorized') return;
        console.warn("Failed to load users (backend might be starting):", e.message);
    } finally {
        isFetchingUsers.set(false);
    }
};


export async function addUser(user: Omit<User, 'id'>) {
    try {
        const newUser = await api.createUser(user);
        usersStore.update(users => [...users, newUser]);
        return newUser;
    } catch (e) {
        console.error("Failed to add user", e);
        throw e;
    }
}

export async function updateUser(id: string | number, updatedData: Partial<User>) {
    try {
        const updated = await api.updateUser(id, updatedData);
        usersStore.update(users => users.map(u => u.id === id ? updated : u));
        return updated;
    } catch (e) {
        console.error("Failed to update user", e);
        throw e;
    }
}

export async function removeUser(id: string | number) {
    try {
        await api.deleteUser(id);
        usersStore.update(users => users.filter(u => u.id !== id));
    } catch (e) {
        console.error("Failed to remove user", e);
        throw e;
    }
}
