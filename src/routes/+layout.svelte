<script lang="ts">
	import Notifications from "@/lib/Notifications.svelte";
	import { onMount } from "svelte";
	import { pwaInfo } from "virtual:pwa-info";

	interface Props {
		children?: import("svelte").Snippet;
	}

	let { children }: Props = $props();

	import { onNavigate } from "$app/navigation";
	import { viewTransition } from "@/lib/util/viewTransition.svelte";

	onNavigate((navigation) => {
		// Only run when View Transitions API is supported
		if (!document.startViewTransition) return;

		// Respect reduced motion settings
		if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;

		// Automatically detect backward navigation from detail pages to set active media
		const fromRouteId = navigation.from?.route?.id || "";
		const fromParams = navigation.from?.params as { id?: string } | undefined;

		if (fromParams?.id) {
			if (fromRouteId.includes("/movie/")) {
				viewTransition.setActive("movie", fromParams.id);
				viewTransition.direction = "back";
			} else if (fromRouteId.includes("/tv/")) {
				viewTransition.setActive("tv", fromParams.id);
				viewTransition.direction = "back";
			} else if (fromRouteId.includes("/game/")) {
				viewTransition.setActive("game", fromParams.id);
				viewTransition.direction = "back";
			}
		}

		return new Promise<void>((resolve) => {
			const transition = document.startViewTransition(async () => {
				resolve();
				await navigation.complete;
			});

			transition.finished.finally(() => {
				viewTransition.clear();
			});
		});
	});

	function resetTooltipPos() {
		const t = document.getElementById("tooltip");
		if (t) {
			t.style.top = "0";
			t.style.left = "0";
		}
	}

	onMount(() => {
		window.addEventListener("resize", resetTooltipPos);

		return () => {
			window.removeEventListener("resize", resetTooltipPos);
		};
	});
</script>

<svelte:head>
	{#if pwaInfo?.webManifest?.linkTag}
		<!-- eslint-disable-next-line -->
		{@html pwaInfo.webManifest.linkTag}
	{/if}
</svelte:head>

<div id="tooltip"></div>

<Notifications />

{@render children?.()}

<style lang="scss">
	@use "../styles/norm.scss";
</style>
