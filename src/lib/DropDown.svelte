<script lang="ts">
	import type { DropDownItem } from "@/types";
	import Icon from "./Icon.svelte";
	import { onMount } from "svelte";

	interface Props {
		options: string[] | DropDownItem[];
		// If we are using DropDownItems[] as options.
		isDropDownItem?: boolean;
		active?: string | number | undefined;
		placeholder: string;
		blendIn?: boolean;
		disabled?: boolean;
		onChange?: () => void;
		showActiveElementsInOptions?: boolean;
	}

	let {
		options,
		isDropDownItem = false,
		active = $bindable(undefined),
		placeholder,
		blendIn = false,
		disabled = false,
		onChange = () => {},
		showActiveElementsInOptions = false,
	}: Props = $props();

	let activeValue: string | undefined = $state();
	let open = $state(false);
	let ulElement: HTMLUListElement | undefined = $state();
	let mainElement: HTMLDivElement | undefined = $state();

	function handleKeyPress(event: KeyboardEvent) {
		if (!open || disabled) return; // Don't handle if closed or disabled

		const pressedLetter = event.key.toLowerCase();

		// Find first button with text content starting with letter pressed and scroll it into view.
		const btns = ulElement?.querySelectorAll("button");
		if (btns) {
			for (let i = 0; i < btns.length; i++) {
				const btn = btns[i];
				if (btn.textContent?.toLowerCase()?.startsWith(pressedLetter)) {
					btn.scrollIntoView({
						behavior: "smooth",
						block: "nearest",
						inline: "nearest",
					});
					break;
				}
			}
		}
	}

	$effect(() => {
		if (typeof active === "string" && !isDropDownItem) {
			activeValue = active;
		} else {
			const v = options.find((o) =>
				typeof o !== "string" ? o.id === active : false,
			);
			if (v && typeof v !== "string") activeValue = v.value;
		}
	});

	function handleClickOutside(event: MouseEvent) {
		if (open && mainElement && !mainElement.contains(event.target as Node)) {
			open = false;
		}
	}

	function handleEscapeKey(event: KeyboardEvent) {
		if (open && event.key === "Escape") {
			open = false;
		}
	}

	onMount(() => {
		mainElement?.addEventListener("keypress", handleKeyPress);
		document.addEventListener("click", handleClickOutside);
		document.addEventListener("keydown", handleEscapeKey);

		return () => {
			mainElement?.removeEventListener("keypress", handleKeyPress);
			document.removeEventListener("click", handleClickOutside);
			document.removeEventListener("keydown", handleEscapeKey);
		};
	});
</script>

<div
	class={[
		open ? "is-open" : "",
		typeof active === "undefined" ? "placeholder-shown" : "",
		blendIn ? "blend-in" : "",
	].join(" ")}
	bind:this={mainElement}
>
	<button onclick={() => (open = !open)} {disabled}>
		{activeValue ? activeValue : placeholder}
		<Icon i="chevron" facing={open ? "up" : "down"} />
	</button>
	<ul bind:this={ulElement}>
		{#each showActiveElementsInOptions ? options : options.filter( (o) => (typeof o === "string" ? o !== active : o.id !== active) ) as o (typeof o === "string" ? o : o.id)}
			<li>
				<button
					class="plain"
					onclick={() => {
						active = typeof o == "string" ? o : o.id;
						open = false;
						onChange();
					}}
				>
					{typeof o == "string" ? o : o.value}
				</button>
			</li>
		{/each}
	</ul>
</div>

<style lang="scss">
	div {
		position: relative;

		&.placeholder-shown {
			& > button {
				color: #8e8e8e;

				&:hover:not(:disabled) {
					color: $bg-color;
				}
			}
		}

		&.blend-in {
			& > button {
				border-color: transparent;
				background-color: transparent;

				&:hover:not(:disabled),
				&:focus-visible:not(:disabled) {
					background-color: $text-color;
				}
			}
		}

		&.is-open {
			ul {
				display: flex;
			}

			& > button {
				border-color: $brand-color;
				box-shadow: 0 0 0 3px $brand-glow;
			}
		}

		button {
			gap: 6px;
			text-transform: capitalize;

			:global(svg) {
				margin-left: auto;
			}
		}

		ul {
			display: none;
			flex-flow: column;
			position: absolute;
			top: calc(100% + 4px);
			left: 0;
			list-style: none;
			width: 100%;
			font-size: 13.5px;
			border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
			border-radius: 10px;
			background-color: var(--card-bg, #161b26);
			box-shadow: 0 16px 36px rgba(0, 0, 0, 0.45);
			backdrop-filter: blur(16px);
			-webkit-backdrop-filter: blur(16px);
			z-index: 100;
			max-height: 240px;
			overflow-y: auto;
			padding: 4px;

			li {
				width: 100%;

				button {
					padding: 8px 12px;
					width: 100%;
					border-radius: 6px;
					border: none;
					text-align: start;
					text-transform: capitalize;
					font-size: 13px;
					color: $text-color;
					background: transparent;
					box-shadow: none;
					transform: none !important;
					transition: background-color 120ms ease;

					&:hover:not(:disabled),
					&:focus-visible:not(:disabled) {
						background-color: $accent-color;
						color: $text-color;
						opacity: 1;
					}
				}
			}
		}
	}
</style>
