// Watcharr Service Worker - PWA & Web Push Notifications
const CACHE_NAME = "watcharr-v1";
const STATIC_ASSETS = [
	"/",
	"/manifest.webmanifest",
	"/favicon.png",
	"/favicon.svg",
	"/logo-sqre-144.png",
	"/logo-sqre-192.png",
	"/logo-sqre.png",
];

// Install: Cache critical static shell
self.addEventListener("install", (event) => {
	event.waitUntil(
		caches.open(CACHE_NAME).then((cache) => {
			return cache.addAll(STATIC_ASSETS).catch((err) => {
				console.warn("SW: pre-caching partial failure", err);
			});
		}),
	);
	self.skipWaiting();
});

// Activate: Clean old caches and claim clients
self.addEventListener("activate", (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) => {
				return Promise.all(
					keys
						.filter((key) => key !== CACHE_NAME)
						.map((key) => caches.delete(key)),
				);
			})
			.then(() => self.clients.claim()),
	);
});

// Fetch: Network-first for dynamic content, cache fallback for offline shell
self.addEventListener("fetch", (event) => {
	const url = new URL(event.request.url);

	// Bypass service worker caching for API calls, image proxying, and non-GET requests
	if (
		event.request.method !== "GET" ||
		url.pathname.startsWith("/api/") ||
		url.pathname.startsWith("/img/")
	) {
		return;
	}

	event.respondWith(
		fetch(event.request)
			.then((response) => {
				// Clone and cache successful response for same-origin static assets
				if (response && response.status === 200 && response.type === "basic") {
					const responseToCache = response.clone();
					caches.open(CACHE_NAME).then((cache) => {
						cache.put(event.request, responseToCache);
					});
				}
				return response;
			})
			.catch(async () => {
				const cachedResponse = await caches.match(event.request);
				if (cachedResponse) {
					return cachedResponse;
				}
				// If navigating to an HTML page while offline, return cached root shell
				if (event.request.mode === "navigate") {
					return caches.match("/");
				}
				return new Response("Network error (offline)", {
					status: 503,
					statusText: "Offline",
				});
			}),
	);
});

// Push Notifications: Display received Web Push payloads
self.addEventListener("push", (event) => {
	let data = {
		title: "Watcharr",
		body: "New notification from your media library",
		icon: "/logo-sqre-192.png",
		badge: "/logo-sqre-144.png",
		url: "/",
	};

	if (event.data) {
		try {
			const parsed = event.data.json();
			data = { ...data, ...parsed };
		} catch (e) {
			data.body = event.data.text();
		}
	}

	const notificationOptions = {
		body: data.body,
		icon: data.icon || "/logo-sqre-192.png",
		badge: data.badge || "/logo-sqre-144.png",
		vibrate: [150, 50, 100],
		data: {
			url: data.url || "/",
		},
		actions: [
			{
				action: "open",
				title: "Open Watcharr",
			},
		],
	};

	event.waitUntil(
		self.registration.showNotification(data.title, notificationOptions),
	);
});

// Notification Click: Focus existing client or navigate to notification target URL
self.addEventListener("notificationclick", (event) => {
	event.notification.close();

	const targetUrl = event.notification.data?.url || "/";

	event.waitUntil(
		clients
			.matchAll({ type: "window", includeUncontrolled: true })
			.then((windowClients) => {
				// If a tab is already open, focus it and navigate
				for (const client of windowClients) {
					if ("focus" in client) {
						client.focus();
						if ("navigate" in client && targetUrl !== "/") {
							return client.navigate(targetUrl);
						}
						return;
					}
				}
				// Otherwise open a new window
				if (clients.openWindow) {
					return clients.openWindow(targetUrl);
				}
			}),
	);
});
