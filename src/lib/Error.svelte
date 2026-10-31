<script lang="ts">
	import { ReqerError } from "./util/fetch";

	interface Props {
		pretty: string;
		error: unknown;
		onRetry?: () => void | Promise<void>;
	}
	let { pretty, error, onRetry = undefined }: Props = $props();

	let errorMessage = $derived(
		error instanceof Error
			? error.message
			: typeof error === "string"
				? error
				: JSON.stringify(error),
	);
	let extraError = $derived(
		error instanceof ReqerError &&
			typeof error.body === "object" &&
			error.body !== null &&
			"error" in error.body
			? String((error.body as { error: unknown }).error)
			: undefined,
	);
</script>

<div class="error-wrap">
	<div>
		<div class="error-text">
			<strong>{pretty}</strong>
			<p>{errorMessage}</p>
			{#if extraError}
				<p>{extraError}</p>
			{/if}
		</div>
		<div>
			{#if onRetry}
				<button onclick={onRetry}>Try Again</button>
			{/if}
		</div>
	</div>
</div>

<style lang="scss">
	div.error-wrap {
		display: flex;
		justify-content: center;
		width: 100%;
		padding: 0 16px;

		& > div {
			display: flex;
			align-items: center;
			justify-content: space-between;
			flex-flow: row wrap;
			width: 100%;
			max-width: 580px;
			gap: 16px;
			border-radius: 12px;
			padding: 16px 20px;
			margin: 20px 0;
			color: #f8fafc;
			background-color: rgba(35, 18, 22, 0.85);
			border: 1px solid rgba(239, 68, 68, 0.35);
			box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
			backdrop-filter: blur(8px);

			.error-text {
				display: flex;
				flex-flow: column;
				gap: 4px;
				flex: 1;
				min-width: 200px;

				strong {
					font-size: 14px;
					font-weight: 700;
					color: #fca5a5;
				}

				p {
					font-size: 13px;
					color: #e2e8f0;
					line-height: 1.4;
					margin: 0;
					word-break: break-word;
				}
			}

			button {
				width: fit-content;
				padding: 7px 16px;
				font-size: 13px;
				font-weight: 600;
				border-radius: 8px;
				background-color: rgba(239, 68, 68, 0.2);
				border: 1px solid rgba(239, 68, 68, 0.4);
				color: #fecaca;
				transition: all 140ms ease;

				&:hover {
					background-color: #ef4444;
					color: #ffffff;
					border-color: #ef4444;
				}
			}
		}
	}
</style>
