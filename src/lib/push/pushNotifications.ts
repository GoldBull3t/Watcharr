// Push Notifications Helper for Watcharr PWA
import { req } from "@/lib/util/api";
import { notify } from "@/lib/util/notify";

export interface PushStatus {
	supported: boolean;
	permission: NotificationPermission | "unsupported";
	isSubscribed: boolean;
}

function urlBase64ToUint8Array(base64String: string): Uint8Array {
	const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
	const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
	const rawData = window.atob(base64);
	const outputArray = new Uint8Array(rawData.length);
	for (let i = 0; i < rawData.length; ++i) {
		outputArray[i] = rawData.charCodeAt(i);
	}
	return outputArray;
}

export function isPushSupported(): boolean {
	if (typeof window === "undefined" || typeof navigator === "undefined") {
		return false;
	}
	return (
		"serviceWorker" in navigator &&
		"PushManager" in window &&
		"Notification" in window
	);
}

export async function getPushStatus(): Promise<PushStatus> {
	if (!isPushSupported()) {
		return { supported: false, permission: "unsupported", isSubscribed: false };
	}

	const permission = Notification.permission;
	try {
		const reg = await navigator.serviceWorker.ready;
		const sub = await reg.pushManager.getSubscription();
		return {
			supported: true,
			permission,
			isSubscribed: !!sub,
		};
	} catch (err) {
		console.debug("getPushStatus error:", err);
		return {
			supported: true,
			permission,
			isSubscribed: false,
		};
	}
}

export async function subscribeToPush(): Promise<boolean> {
	if (!isPushSupported()) {
		notify({
			text: "Notificações Web Push não são suportadas neste navegador.",
			type: "error",
		});
		return false;
	}

	try {
		// 1. Request Notification permission
		const permission = await Notification.requestPermission();
		if (permission !== "granted") {
			notify({
				text: "Permissão de notificações recusada pelo utilizador.",
				type: "error",
			});
			return false;
		}

		// 2. Fetch VAPID Public Key from backend
		const keyRes = await req.get<{ publicKey: string }>("push/key");
		if (!keyRes || !keyRes.publicKey) {
			notify({
				text: "Falha ao obter chave VAPID do servidor.",
				type: "error",
			});
			return false;
		}

		// 3. Register service worker and subscribe
		const reg = await navigator.serviceWorker.ready;
		let sub = await reg.pushManager.getSubscription();

		if (!sub) {
			const convertedVapidKey = urlBase64ToUint8Array(keyRes.publicKey);
			sub = await reg.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey: convertedVapidKey as any,
			});
		}

		const subJSON = sub.toJSON();
		if (!subJSON.endpoint || !subJSON.keys?.p256dh || !subJSON.keys?.auth) {
			notify({
				text: "Assinatura Push incompleta recebida do navegador.",
				type: "error",
			});
			return false;
		}

		// 4. Send subscription to Watcharr backend
		await req.post("push/subscribe", {
			endpoint: subJSON.endpoint,
			keys: {
				p256dh: subJSON.keys.p256dh,
				auth: subJSON.keys.auth,
			},
			userAgent: navigator.userAgent,
		});

		notify({
			text: "Notificações Web Push ativadas com sucesso!",
			type: "success",
		});
		return true;
	} catch (err: any) {
		console.error("subscribeToPush error:", err);
		notify({
			text: `Erro ao ativar notificações: ${err?.message || "falha desconhecida"}`,
			type: "error",
		});
		return false;
	}
}

export async function unsubscribeFromPush(): Promise<boolean> {
	if (!isPushSupported()) return false;

	try {
		const reg = await navigator.serviceWorker.ready;
		const sub = await reg.pushManager.getSubscription();

		if (sub) {
			const endpoint = sub.endpoint;
			await sub.unsubscribe();
			await req.delete("push/subscribe", { body: { endpoint } });
		}

		notify({ text: "Notificações Web Push desativadas.", type: "success" });
		return true;
	} catch (err: any) {
		console.error("unsubscribeFromPush error:", err);
		notify({
			text: `Erro ao desativar notificações: ${err?.message || "falha"}`,
			type: "error",
		});
		return false;
	}
}

export async function sendTestNotification(): Promise<boolean> {
	try {
		const res = await req.post<{ success: boolean; sent: number }>("push/test");
		if (res && res.success) {
			notify({
				text: "Notificação de teste enviada! Verifica o teu dispositivo.",
				type: "success",
			});
			return true;
		}
		notify({
			text: "Nenhum dispositivo subscrito recebeu a notificação de teste.",
			type: "error",
		});
		return false;
	} catch (err: any) {
		notify({
			text: `Erro ao enviar teste: ${err?.message || "falha"}`,
			type: "error",
		});
		return false;
	}
}
