<script lang="ts">
	import { MediaVideoType, type MediaVideo } from "@/types";
	import VideoEmbedModal from "./VideoEmbedModal.svelte";
	import Icon from "@/lib/Icon.svelte";

	interface Props {
		videos?: MediaVideo[];
	}

	let { videos }: Props = $props();

	let trailer: string | undefined = $derived.by(() => {
		if (videos && videos?.length > 0) {
			const t = videos
				.filter((v) => v.type === MediaVideoType.trailer)
				.sort((a, b) => {
					if (a.best === b.best) return 0;
					return a.best ? -1 : 1;
				});
			if (t[0]?.id) {
				return `https://www.youtube-nocookie.com/embed/${t[0]?.id}`;
			}
		}
	});
	let trailerShown = $state(false);
</script>

{#if trailer}
	<button class="trailer-btn" onclick={() => (trailerShown = !trailerShown)}>
		<Icon i="play" wh={14} />
		<span>Watch Trailer</span>
	</button>
	{#if trailerShown}
		<VideoEmbedModal embed={trailer} closed={() => (trailerShown = false)} />
	{/if}
{/if}

<style lang="scss">
	button.trailer-btn {
		display: inline-flex;
		align-items: center;
		gap: 7px;
		max-width: fit-content;
	}
</style>
