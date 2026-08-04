import { Buffer } from 'buffer';
import type { Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

export const handle: Handle = async ({ event, resolve }) => {
	console.log(`[CSRF Debug] Request path: ${event.url.pathname}`);
	console.log(`[CSRF Debug] URL Origin: ${event.url.origin}`);
	console.log(`[CSRF Debug] Request Origin: ${event.request.headers.get('origin')}`);
	console.log(`[CSRF Debug] Headers: ${JSON.stringify(Object.fromEntries(event.request.headers))}`);
	
	// Proxy request ke backend untuk path /api dan /uploads
	if (event.url.pathname.startsWith('/api') || event.url.pathname.startsWith('/uploads')) {
		const target = env.INTERNAL_API_URL || 'http://backend:8081';
		const url = `${target}${event.url.pathname}${event.url.search}`;

		// Persiapkan headers, hapus header yang bisa menyebabkan masalah proxy
		const headers = new Headers(event.request.headers);
		headers.delete('host');
		headers.delete('connection');
		headers.delete('content-length');
		headers.delete('transfer-encoding');
		headers.delete('expect');

		try {
			// Gunakan clone() agar request asli tetap tersedia jika dibutuhkan SvelteKit
			const requestClone = event.request.clone();
			
			const fetchOptions: RequestInit = {
				method: event.request.method,
				headers: headers
			};

			if (event.request.method !== 'GET' && event.request.method !== 'HEAD') {
				console.log("[Proxy Debug] Converting to arrayBuffer...");
				fetchOptions.body = await requestClone.arrayBuffer();
				console.log("[Proxy Debug] Successfully converted to arrayBuffer, size:", fetchOptions.body.byteLength);
			}

			console.log("[Proxy Debug] Calling fetch to", url);
			const response = await fetch(url, fetchOptions);
			console.log("[Proxy Debug] Fetch success with status", response.status);

			// Ambil body sebagai Buffer (Node.js native) agar lebih stabil saat dikirim balik
			const responseData = await response.arrayBuffer();
			const buffer = Buffer.from(responseData);
			

			const proxyHeaders = new Headers();
			let contentType = response.headers.get('content-type');
			
			// --- IDM CLOAKING LOGIC ---
			// Jika tipe adalah PDF, kita samarkan jadi octet-stream 
			// supaya tidak dicuri oleh Internet Download Manager (IDM)
			if (contentType && contentType.includes('application/pdf')) {
				proxyHeaders.set('content-type', 'application/octet-stream');
				// JANGAN teruskan content-disposition (attachment; filename=...pdf) 
				// karena IDM sangat agresif melacak file ber-ekstensi .pdf dari header ini!
			} else {
				if (contentType) proxyHeaders.set('content-type', contentType);
				
				const contentDisposition = response.headers.get('content-disposition');
				if (contentDisposition) proxyHeaders.set('content-disposition', contentDisposition);
			}

			// Paksa no-cache agar browser tidak menyimpan respons error 0-byte sebelumnya
			proxyHeaders.set('cache-control', 'no-cache, no-store, must-revalidate');
			proxyHeaders.set('pragma', 'no-cache');
			proxyHeaders.set('expires', '0');

			// Response dengan status tertentu tidak boleh memiliki body (harus null) di Node.js
			const nullBodyStatuses = [101, 204, 205, 304];
			const responseBody = nullBodyStatuses.includes(response.status) ? null : buffer;

			return new Response(responseBody, {
				status: response.status,
				statusText: response.statusText,
				headers: proxyHeaders
			});
		} catch (err: any) {
			// Log ini akan muncul di 'podman logs <frontend_container>'
			console.error(`[Proxy Error] ${event.request.method} ${url} ->`, err.message);
			return new Response(`Proxy Error: ${err.message}`, { status: 502 });
		}
	}

	return resolve(event);
};
