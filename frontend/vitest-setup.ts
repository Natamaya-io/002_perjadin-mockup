import { vi } from 'vitest';

// Mock SvelteKit Navigation
vi.mock('$app/navigation', () => {
    return {
        goto: vi.fn(),
        beforeNavigate: vi.fn(),
        afterNavigate: vi.fn()
    };
});

vi.mock('$app/environment', () => {
    return {
        browser: true
    };
});

// Mock localStorage
Object.defineProperty(global, 'localStorage', {
    value: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
    },
    writable: true,
});
