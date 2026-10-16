import { store } from "@/store.svelte";
import type { Icon as IconType } from "@/types";

export interface NotificationAction {
	label: string;
	icon?: IconType;
	onClick: () => void | Promise<void>;
}

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
	type?: "error" | "success" | "loading" | "info";

	/**
	 * How long in milliseconds the popup will stay shown for.
	 */
	time?: number;

	/**
	 * Action button (e.g. Undo / Desfazer)
	 */
	action?: NotificationAction;

	/**
	 * Callback triggered when notification dismisses without undo.
	 */
	onDismiss?: () => void;

	/**
	 * Whether to show an animated progress countdown bar.
	 */
	showProgress?: boolean;
}

interface TimerState {
	timeoutId?: ReturnType<typeof setTimeout>;
	startTime: number;
	remainingMs: number;
	isPaused: boolean;
	totalMs: number;
}

let nextNotifId = 1;
const notifTimers = new Map<number, TimerState>();
const notifCallbacks = new Map<number, () => void>();

export function notify(n: Notification): number {
	let targetId = n.id;
	if (targetId) {
		const existingTimer = notifTimers.get(targetId);
		if (existingTimer?.timeoutId) {
			clearTimeout(existingTimer.timeoutId);
			notifTimers.delete(targetId);
		}
		const notif = store.notifications.find((not) => not.id === targetId);
		if (notif) {
			notif.type = n.type;
			notif.text = n.text;
			notif.action = n.action;
			notif.showProgress = n.showProgress;
			notif.time = n.time;
		} else {
			targetId = nextNotifId++;
			store.notifications.push({ ...n, id: targetId });
		}
	} else {
		targetId = nextNotifId++;
		store.notifications.push({ ...n, id: targetId });
	}

	if (n.onDismiss) {
		notifCallbacks.set(targetId, n.onDismiss);
	}

	if (n.type !== "loading" && n.time !== Infinity) {
		const totalMs = n.time ?? 2500;
		const id = targetId;
		const timeoutId = setTimeout(() => unNotify(id), totalMs);
		notifTimers.set(id, {
			timeoutId,
			startTime: Date.now(),
			remainingMs: totalMs,
			isPaused: false,
			totalMs,
		});
	}
	return targetId;
}

export function pauseNotify(id: number) {
	const timer = notifTimers.get(id);
	if (timer && !timer.isPaused && timer.timeoutId) {
		clearTimeout(timer.timeoutId);
		const elapsed = Date.now() - timer.startTime;
		timer.remainingMs = Math.max(200, timer.remainingMs - elapsed);
		timer.isPaused = true;
	}
}

export function resumeNotify(id: number) {
	const timer = notifTimers.get(id);
	if (timer && timer.isPaused && timer.remainingMs > 0) {
		timer.startTime = Date.now();
		timer.isPaused = false;
		timer.timeoutId = setTimeout(() => unNotify(id), timer.remainingMs);
	}
}

export function unNotify(id?: number, triggerDismiss = true) {
	if (!id) return;
	const timer = notifTimers.get(id);
	if (timer?.timeoutId) {
		clearTimeout(timer.timeoutId);
		notifTimers.delete(id);
	}
	if (triggerDismiss) {
		const cb = notifCallbacks.get(id);
		if (cb) {
			notifCallbacks.delete(id);
			try {
				cb();
			} catch (err) {
				console.error("unNotify dismiss callback error:", err);
			}
		}
	} else {
		notifCallbacks.delete(id);
	}
	store.notifications = store.notifications.filter((e) => e.id !== id);
}

export function notifyUndo(options: {
	text: string;
	undoLabel?: string;
	time?: number;
	onUndo: () => void | Promise<void>;
	onDismiss?: () => void;
}): number {
	let isUndone = false;
	const duration = options.time ?? 6000;
	let targetId = 0;

	targetId = notify({
		text: options.text,
		type: "info",
		time: duration,
		showProgress: true,
		action: {
			label: options.undoLabel ?? "Desfazer",
			icon: "undo",
			onClick: async () => {
				isUndone = true;
				notifCallbacks.delete(targetId);
				unNotify(targetId, false);
				try {
					await options.onUndo();
					notify({
						text: "Ação desfeita!",
						type: "success",
						time: 2500,
					});
				} catch (err) {
					console.error("Undo action failed:", err);
					notify({
						text: "Não foi possível desfazer a ação.",
						type: "error",
						time: 3500,
					});
				}
			},
		},
		onDismiss: () => {
			if (!isUndone && options.onDismiss) {
				options.onDismiss();
			}
		},
	});

	return targetId;
}
