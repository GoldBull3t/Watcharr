<script lang="ts">
	import { MediaVideoType, type MediaVideo } from "@/types";
	import MediaLightbox from "./MediaLightbox.svelte";
	import Icon from "@/lib/Icon.svelte";
	import { i18n } from "@/lib/i18n";

	interface Props {
		videos?: MediaVideo[];
		backdropPath?: string;
		posterPath?: string;
		mediaTitle?: string;
	}

	let { videos = [], backdropPath, posterPath, mediaTitle }: Props = $props();

	let hasVideos = $derived(videos && videos.length > 0);
	let trailersCount = $derived(
		videos ? videos.filter((v) => v.type === MediaVideoType.trailer).length : 0,
	);
	let isPt = $derived(i18n.currentLocale === "pt");
	let lightboxOpen = $state(false);
</script>

{#if hasVideos}
	<button
		type="button"
		class="trailer-btn"
		onclick={() => (lightboxOpen = true)}
		title={isPt
			? "Reproduzir trailer oficial e ver galeria de mídia"
			: "Watch official trailer and media gallery"}
	>
		<Icon i="play" wh={14} />
		<span>
			{#if trailersCount > 1}
				{isPt
					? `Ver Trailers (${trailersCount})`
					: `Watch Trailers (${trailersCount})`}
			{:else}
				{isPt ? "Ver Trailer" : "Watch Trailer"}
			{/if}
		</span>
	</button>

	{#if lightboxOpen}
		<MediaLightbox
			{videos}
			{backdropPath}
			{posterPath}
			{mediaTitle}
			initialType="video"
			onclose={() => (lightboxOpen = false)}
		/>
	{/if}
{/if}

<style lang="scss">
	button.trailer-btn {
		display: inline-flex;
		align-items: center;
		gap: 7px;
		max-width: fit-content;
		padding: 7px 16px;
		border-radius: 10px;
		background: rgba(99, 102, 241, 0.16);
		border: 1px solid rgba(99, 102, 241, 0.35);
		color: #ffffff;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: all 160ms cubic-bezier(0.16, 1, 0.3, 1);

		&:hover {
			background: rgba(99, 102, 241, 0.28);
			border-color: #818cf8;
			box-shadow: 0 0 16px rgba(99, 102, 241, 0.35);
			transform: translateY(-1px);
		}
	}
</style>
