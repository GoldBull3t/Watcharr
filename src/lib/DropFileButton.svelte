<script lang="ts">
	import { onMount } from "svelte";
	import type { Icon as IconT } from "@/types";
	import Icon from "./Icon.svelte";

	interface Props {
		text: string;
		icon?: IconT;
		filesSelected: (f?: FileList | null) => void;
		allowSelectMultipleFiles?: boolean;
	}

	let {
		text,
		icon = "document",
		filesSelected,
		allowSelectMultipleFiles = false,
	}: Props = $props();

	let fileInput: HTMLInputElement | undefined = $state();
	let dragEnterTarget: EventTarget | null | undefined = $state();
	let isDragOver = $state(false);

	function importFile() {
		fileInput?.click();
	}

	onMount(() => {
		if (fileInput) {
			fileInput.addEventListener("change", () => {
				filesSelected(fileInput?.files);
			});
		}
	});
</script>

<div class="drop-file-btn">
	<button
		onclick={importFile}
		ondragover={(ev) => {
			ev.preventDefault();
			ev.stopPropagation();
		}}
		ondragenter={(ev) => {
			ev.preventDefault();
			ev.stopPropagation();
			dragEnterTarget = ev.target;
			console.log("enter");
			isDragOver = true;
		}}
		ondragleave={(ev) => {
			ev.preventDefault();
			ev.stopPropagation();
			if (dragEnterTarget === ev.target) {
				console.log("leave");
				isDragOver = false;
			}
		}}
		ondrop={(ev) => {
			ev.preventDefault();
			ev.stopPropagation();
			filesSelected(ev.dataTransfer?.files);
		}}
		class={isDragOver ? "dragging-over" : ""}
	>
		<Icon i={isDragOver ? "add" : icon} wh="100%" />
		<div>
			<h4 class="norm">
				{#if isDragOver}
					Import {text}
				{:else}
					{text}
				{/if}
			</h4>
		</div>
	</button>
	<input
		type="file"
		multiple={allowSelectMultipleFiles}
		bind:this={fileInput}
	/>
</div>

<style lang="scss">
	.drop-file-btn {
		width: 100%;

		button {
			display: flex;
			flex-flow: column;
			justify-content: center;
			align-items: center;
			gap: 14px;
			height: 160px;
			width: 100%;
			padding: 24px;
			background: rgba(255, 255, 255, 0.03);
			border: 1px solid rgba(255, 255, 255, 0.08);
			border-radius: 14px;
			user-select: none;
			transition: all 180ms ease;
			color: rgba(255, 255, 255, 0.9);
			fill: rgba(255, 255, 255, 0.9);
			cursor: pointer;

			:global {
				#reel path {
					transition: 180ms ease-in-out;

					&:first-of-type {
						fill: transparent;
					}

					&:last-of-type {
						fill: currentColor;
					}
				}
			}

			&:hover {
				background: rgba(255, 255, 255, 0.06);
				border-color: rgba(54, 186, 152, 0.4);
				color: #ffffff;
				fill: #ffffff;
				transform: translateY(-2px);
			}

			&.dragging-over {
				background: rgba(54, 186, 152, 0.12);
				border-color: #36ba98;
				color: #36ba98;
				fill: #36ba98;
				box-shadow: 0 0 20px rgba(54, 186, 152, 0.2);
				transform: scale(1.02);
			}

			h4 {
				font-size: 14px;
				font-weight: 600;
				letter-spacing: 0.2px;
				margin: 0;
			}
		}

		input[type="file"] {
			width: 0px;
			overflow: hidden;
			border: unset;
			background-color: transparent;
			position: absolute;
			top: -500px;
		}
	}
</style>
