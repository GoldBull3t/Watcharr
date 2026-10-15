<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import {
		addClassToParent,
		calculateTransformOrigin,
	} from "@/lib/util/helpers";

	interface Props {
		id: number | undefined;
		name: string | undefined;
		path: string | undefined;
		role?: string | undefined;
		zoomOnHover?: boolean;
	}

	let {
		id,
		name,
		path,
		role = undefined,
		zoomOnHover = true,
	}: Props = $props();

	const poster = $derived(
		path ? `https://image.tmdb.org/t/p/w300_and_h450_bestv2${path}` : undefined,
	);
	const link: `/person/${number}` | undefined = $derived(
		id ? (`/person/${id}` as `/person/${number}`) : undefined,
	);
</script>

<!-- Quick fix to ignore error, should be fixed -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<li
	onmouseenter={(e) => calculateTransformOrigin(e)}
	onfocusin={(e) => calculateTransformOrigin(e)}
	onclick={() => {
		if (link) goto(resolve(link));
	}}
	onkeypress={() => console.log("on kpress")}
>
	<div
		class={`container${!poster ? " details-shown" : ""}${!zoomOnHover ? " no-zoom" : ""}`}
	>
		{#if poster}
			<div class="img-loader"></div>
			<img
				loading="lazy"
				src={poster}
				alt=""
				onload={(e) => {
					addClassToParent(e, "img-loaded");
				}}
				onerror={(e) => {
					console.log("on err");
					addClassToParent(e, "details-shown");
				}}
			/>
		{/if}
		<div class="inner">
			<h2>
				{#if link}
					<a data-sveltekit-preload-data="tap" href={resolve(link)}>
						{name}
					</a>
				{:else}
					{name}
				{/if}
			</h2>
			{#if role}
				<h3>{role}</h3>
			{/if}
		</div>
	</div>
</li>

<style lang="scss">
	.container {
		display: flex;
		flex-flow: column;
		background-color: var(--card-bg, rgb(22, 27, 38));
		overflow: hidden;
		flex: 1 1;
		border-radius: 14px;
		width: 170px;
		height: 100%;
		min-height: 256.367px;
		position: relative;
		border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
		box-shadow: 0 4px 14px rgba(0, 0, 0, 0.25);
		transition:
			transform 180ms ease,
			box-shadow 180ms ease,
			border-color 180ms ease;
		cursor: pointer;

		&:hover {
			border-color: rgba(99, 102, 241, 0.4);
		}

		img {
			width: 100%;
			height: 100%;
			object-fit: cover;
		}

		&.img-loaded .img-loader,
		&.details-shown .img-loader {
			display: none;
		}

		.img-loader {
			position: absolute;
			width: 100%;
			height: 100%;
			background-color: gray;
			background: linear-gradient(359deg, #5c5c5c, #2c2929, #2c2424);
			background-size: 400% 400%;
			animation: imgloader 4s ease infinite;

			@keyframes imgloader {
				0% {
					background-position: 50% 0%;
				}
				50% {
					background-position: 50% 100%;
				}
				100% {
					background-position: 50% 0%;
				}
			}
		}

		.inner {
			position: absolute;
			display: flex;
			flex-flow: column;
			top: 0;
			height: 100%;
			width: 100%;
			padding: 12px;
			background: linear-gradient(
				to top,
				rgba(0, 0, 0, 0.85) 0%,
				rgba(0, 0, 0, 0.2) 50%,
				transparent 100%
			);

			h2,
			h3 {
				font-family:
					sans-serif,
					system-ui,
					-apple-system,
					BlinkMacSystemFont;
				color: white;
				word-wrap: break-word;
			}

			h2 {
				font-size: 16px;
				font-weight: 600;
				margin-top: auto;
				text-shadow: 0 2px 4px rgba(0, 0, 0, 0.6);

				a {
					color: white;
				}
			}

			h3 {
				font-size: 12px;
				font-weight: 400;
				color: rgba(255, 255, 255, 0.8);
				text-shadow: 0 1px 2px rgba(0, 0, 0, 0.6);
			}
		}

		&:not(.no-zoom) {
			&:hover,
			&:has(:global(:focus-visible)) {
				transform: scale(1.08);
				box-shadow: 0 12px 30px rgba(0, 0, 0, 0.5);
				z-index: 99;
			}
		}

		&:not(:not(.no-zoom)) {
			&:hover,
			&:has(:global(:focus-visible)) {
				outline: 2px solid $brand-color;
			}
		}

		&:hover,
		&:has(:global(:focus-visible)),
		&:global(.details-shown) {
			.inner {
				color: white;
			}
		}
	}
</style>
