import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		host: true,
		watch: {
			usePolling: true,
			interval: 500
		},
		proxy: {
			'/api': {
				target: process.env.BACKEND_URL || 'http://127.0.0.1:8081',
				changeOrigin: true
			},
			'/uploads': {
				target: process.env.BACKEND_URL || 'http://127.0.0.1:8081',
				changeOrigin: true
			}
		}
	},
	test: {
		environment: 'jsdom',
		setupFiles: ['./vitest-setup.ts'],
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
