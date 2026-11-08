// Watcharr Web Push Notification Handler
// Loaded via Workbox importScripts into the generated Service Worker

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

self.addEventListener("notificationclick", (event) => {
	event.notification.close();

	const targetUrl = event.notification.data?.url || "/";

	event.waitUntil(
		clients
			.matchAll({ type: "window", includeUncontrolled: true })
			.then((windowClients) => {
				for (const client of windowClients) {
					if ("focus" in client) {
						client.focus();
						if ("navigate" in client && targetUrl !== "/") {
							return client.navigate(targetUrl);
						}
						return;
					}
				}
				if (clients.openWindow) {
					return clients.openWindow(targetUrl);
				}
			}),
	);
});
