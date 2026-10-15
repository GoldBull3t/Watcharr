<script lang="ts">
	interface Props {
		src: string;
	}

	let { src }: Props = $props();
</script>

<img {src} alt="" />

<style lang="scss">
	img {
		width: 260px;
		max-width: 100%;
		aspect-ratio: 2 / 3;
		object-fit: cover;
		box-shadow:
			0 20px 48px -10px rgba(0, 0, 0, 0.75),
			0 0 0 1px rgba(255, 255, 255, 0.1);
		border-radius: 18px;
		flex-shrink: 0;
		transition:
			transform 200ms ease,
			box-shadow 200ms ease;

		&:hover {
			transform: scale(1.015);
			box-shadow:
				0 24px 54px -10px rgba(0, 0, 0, 0.85),
				0 0 20px rgba(99, 102, 241, 0.2),
				0 0 0 1px rgba(255, 255, 255, 0.16);
		}
	}
</style>
