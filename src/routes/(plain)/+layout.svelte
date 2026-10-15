<script lang="ts">
	interface Props {
		children?: import("svelte").Snippet;
	}

	let { children }: Props = $props();
</script>

<nav>
	<h1>Watcharr</h1>
</nav>

{@render children?.()}

<style lang="scss">
	nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin: 10px 20px 28px 20px;
		position: relative;

		h1 {
			font-family:
				"Shrikhand",
				system-ui,
				-apple-system,
				BlinkMacSystemFont;
			font-size: 32px;
			letter-spacing: -0.01em;
			background: linear-gradient(135deg, $text-color 65%, $brand-color 100%);
			-webkit-background-clip: text;
			background-clip: text;
			-webkit-text-fill-color: transparent;
			user-select: none;
		}
	}
</style>
