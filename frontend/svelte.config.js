import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		csrf: {
			checkOrigin: true,
			trustedOrigins: ['http://localhost:3000', 'http://127.0.0.1:3000', 'http://100.115.101.14:3000', 'https://perjadin.kemnaker.go.id', 'http://perjadin.kemnaker.go.id']
		},
		// Setting biar output folder namanya 'build'
		adapter: adapter({ out: 'build', precompress: false }),
		output: {
			preloadStrategy: 'modulepreload'
		}
	}
};

export default config;

