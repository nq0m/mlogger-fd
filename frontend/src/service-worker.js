/// <reference types="@sveltejs/kit" />
import { build, files, version } from '$service-worker';

const CACHE = `fdlogger-${version}`;
const ASSETS = [...build, ...files];

self.addEventListener('install', (event) => {
	self.skipWaiting();
	event.waitUntil(
		caches.open(CACHE).then((cache) => cache.addAll(ASSETS))
	);
});

self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches.keys().then((keys) =>
			Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k)))
		)
	);
});

self.addEventListener('fetch', (event) => {
	if (event.request.method !== 'GET') return;

	const url = new URL(event.request.url);

	if (url.pathname.startsWith('/api/')) {
		return;
	}

	event.respondWith(
		caches.match(event.request).then((cached) => cached || fetch(event.request))
	);
});
