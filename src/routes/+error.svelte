<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { page } from "$app/state";
	import Icon from "@/lib/Icon.svelte";
</script>

<svelte:head>
	<title>{page.status} Error - Watcharr</title>
</svelte:head>

<div class="error-page-wrapper">
	<div class="error-card">
		<span class="status-code">{page.status}</span>
		<div class="message-wrap">
			{#if page.status === 404}
				<h2>Page Not Found</h2>
				<p>
					The page or media you were looking for doesn't exist or has moved.
				</p>
			{:else if page.status === 500}
				<h2>Internal Server Error</h2>
				<p>
					The server encountered an unexpected error. Please try refreshing.
				</p>
				{#if page?.error?.message}
					<small style="opacity: 0.7; font-family: monospace; display: block; margin-top: 8px;">{page.error.message}</small>
				{/if}
			{:else if page?.error?.message}
				<h2>Something Went Wrong</h2>
				<p>{page.error.message}</p>
			{:else}
				<h2>Unable to Load Page</h2>
				<p>We encountered an issue loading this view.</p>
			{/if}
		</div>

		<div class="actions">
			<button class="btn-home" onclick={() => goto(resolve("/"))}>
				<Icon i="reel" wh={18} />
				<span>Back to Library</span>
			</button>
			<button class="btn-refresh" onclick={() => location.reload()}>
				<Icon i="refresh" wh={18} />
				<span>Refresh</span>
			</button>
		</div>
	</div>
</div>

<style lang="scss">
	.error-page-wrapper {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 100vh;
		padding: 20px;
		background: radial-gradient(
			circle at 50% 30%,
			var(--brand-glow, rgba(245, 158, 11, 0.08)) 0%,
			transparent 60%
		);
	}

	.error-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		max-width: 480px;
		width: 100%;
		padding: 48px 36px;
		border-radius: 24px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);
		box-shadow: 0 24px 60px -12px rgba(0, 0, 0, 0.6);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);

		.status-code {
			font-size: 80px;
			font-weight: 900;
			letter-spacing: -0.04em;
			line-height: 1;
			background: linear-gradient(135deg, $accent-color 0%, #ff8c00 100%);
			background-clip: text;
			-webkit-background-clip: text;
			-webkit-text-fill-color: transparent;
			margin-bottom: 12px;
		}

		.message-wrap {
			display: flex;
			flex-direction: column;
			gap: 8px;
			margin-bottom: 28px;

			h2 {
				font-size: 24px;
				font-weight: 700;
				color: #fff;
				margin: 0;
			}

			p {
				font-size: 14.5px;
				color: $text-muted;
				line-height: 1.5;
				margin: 0;
			}
		}

		.actions {
			display: flex;
			gap: 12px;
			width: 100%;
			justify-content: center;
			flex-wrap: wrap;

			button {
				display: inline-flex;
				align-items: center;
				justify-content: center;
				gap: 8px;
				height: 42px;
				padding: 0 20px;
				border-radius: 12px;
				font-size: 14px;
				font-weight: 600;
				cursor: pointer;
				transition: all 180ms ease;

				&.btn-home {
					background: $accent-color;
					color: #fff;
					border: none;
					box-shadow: 0 4px 14px rgba(0, 0, 0, 0.25);

					&:hover {
						background: $accent-color-hover;
						transform: translateY(-1px);
					}
				}

				&.btn-refresh {
					background: rgba(255, 255, 255, 0.06);
					color: $text-color;
					border: 1px solid rgba(255, 255, 255, 0.12);

					&:hover {
						background: rgba(255, 255, 255, 0.12);
						border-color: rgba(255, 255, 255, 0.2);
						color: #fff;
						transform: translateY(-1px);
					}
				}
			}
		}
	}
</style>
