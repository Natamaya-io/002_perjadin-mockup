/// <reference types="@sveltejs/kit" />
import { build, files, version } from '$service-worker';

const CACHE = `cache-${version}`;

const ASSETS = [
	...build, // the app itself
	...files  // everything in `static`
];

self.addEventListener('install', (event) => {
	// Force the waiting service worker to become the active service worker.
	self.skipWaiting();

	// Create a new cache and add all files to it
	async function addFilesToCache() {
		const cache = await caches.open(CACHE);
		await cache.addAll(ASSETS);
	}

	event.waitUntil(addFilesToCache());
});

self.addEventListener('activate', (event) => {
	// Claim any clients immediately, so that the page will be controlled by the service worker immediately.
	event.waitUntil(clients.claim());

	// Remove previous caches
	async function deleteOldCaches() {
		for (const key of await caches.keys()) {
			if (key !== CACHE) await caches.delete(key);
		}
	}

	event.waitUntil(deleteOldCaches());
});

self.addEventListener('fetch', (event) => {
	// ignore POST requests etc
	if (event.request.method !== 'GET') return;

	const url = new URL(event.request.url);

	// IGNORE non-http(s) schemes completely (e.g. chrome-extension://)
	if (!url.protocol.startsWith('http')) return;

	// IGNORE API requests - let them be handled directly by the network
	if (url.pathname.startsWith('/api')) return;

	async function respond() {
		const cache = await caches.open(CACHE);

		// `build`/`files` can always be served from the cache
		if (ASSETS.includes(url.pathname)) {
			const response = await cache.match(event.request);
			if (response) return response;
		}

		// for everything else, try the network first, but
		// fall back to the cache if we're offline
		try {
			const response = await fetch(event.request);

			// Only cache valid http/https responses
			if (response.status === 200) {
				cache.put(event.request, response.clone());
			}

			return response;
		} catch {
			const response = await cache.match(event.request);
			if (response) return response;

			return new Response('Offline', { status: 503, statusText: 'Service Unavailable' });
		}
	}

	event.respondWith(respond());
});
