import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
    id: number;
    message: string;
    type: ToastType;
    duration?: number;
}

function createToastStore() {
    const { subscribe, update } = writable<Toast[]>([]);

    let count = 0;

    return {
        subscribe,
        send: (message: string, type: ToastType = 'info', duration = 3000) => {
            const id = count++;
            update(toasts => [...toasts, { id, message, type, duration }]);

            if (duration > 0) {
                setTimeout(() => {
                    update(toasts => toasts.filter(t => t.id !== id));
                }, duration);
            }
        },
        remove: (id: number) => {
            update(toasts => toasts.filter(t => t.id !== id));
        },
        success: (msg: string, duration = 3000) => toast.send(msg, 'success', duration),
        error: (msg: string, duration = 4000) => toast.send(msg, 'error', duration),
        info: (msg: string, duration = 3000) => toast.send(msg, 'info', duration),
        warning: (msg: string, duration = 3000) => toast.send(msg, 'warning', duration)
    };
}

export const toast = createToastStore();
