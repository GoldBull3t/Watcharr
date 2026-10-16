import { store } from "@/store.svelte";

export interface Notification {
	/**
	 * Notification ID.
	 * Used to reference an exiting notification.
	 */
	id?: number;

	/**
	 * Text shown in popup;
	 */
	text: string;

	/**
	 * Type of notification, controls the style.
	 * Loading notifs never hide, so after request completion
	 * or failure, the notif must be updated or `unNotify`ed.
	 */
	type?: "error" | "success" | "loading";

	/**
	 * How long in milliseconds the popup will stay shown for.
	 */
	time?: number;
}

let nextNotifId = 1;
const notifTimeouts = new Map<number, ReturnType<typeof setTimeout>>();

export function notify(n: Notification): number {
	let targetId = n.id;
	if (targetId) {
		const existingTimeout = notifTimeouts.get(targetId);
		if (existingTimeout) {
			clearTimeout(existingTimeout);
			notifTimeouts.delete(targetId);
		}
		const notif = store.notifications.find((not) => not.id === targetId);
		if (notif) {
			notif.type = n.type;
			notif.text = n.text;
		} else {
			console.warn("notify: Can't update notif that doesn't exist, creating new", n);
			targetId = nextNotifId++;
			store.notifications.push({ ...n, id: targetId });
		}
	} else {
		targetId = nextNotifId++;
		store.notifications.push({ ...n, id: targetId });
	}

	if (n.type !== "loading" && n.time !== Infinity) {
		const timeout = setTimeout(() => unNotify(targetId), n.time ?? 2500);
		notifTimeouts.set(targetId, timeout);
	}
	return targetId;
}

export function unNotify(id?: number) {
	if (!id) {
		console.warn("unNotify: Tried removing a notification without an id.");
		return;
	}
	const timeout = notifTimeouts.get(id);
	if (timeout) {
		clearTimeout(timeout);
		notifTimeouts.delete(id);
	}
	store.notifications = store.notifications.filter((e) => e.id !== id);
}
