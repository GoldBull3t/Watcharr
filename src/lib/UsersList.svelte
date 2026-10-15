<script lang="ts">
	import { resolve } from "$app/paths";
	import Icon from "./Icon.svelte";
	import type { PublicUser } from "@/types";

	interface Props {
		users: PublicUser[];
	}

	let { users }: Props = $props();
</script>

<div class="users-section">
	<div class="section-title">
		<Icon i="people-nocircle" wh={18} />
		<h3>Users Found</h3>
		<span class="count-badge">{users.length}</span>
	</div>
	<ul class="users-rail small-scrollbar">
		{#each users as user (user.id)}
			<li title={user.username}>
				<a
					href={resolve(
						`/lists/${user.id}/${user.username}` as `/lists/${number}/${string}`,
					)}
					class="user-chip"
				>
					<div class="user-avatar">
						{user.username.charAt(0).toUpperCase()}
					</div>
					<span class="user-name">{user.username}</span>
				</a>
			</li>
		{/each}
	</ul>
</div>

<style lang="scss">
	.users-section {
		display: flex;
		flex-flow: column;
		gap: 10px;
		margin: 10px 0 20px 0;
		padding: 16px 20px;
		border-radius: 14px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.06);

		.section-title {
			display: flex;
			align-items: center;
			gap: 8px;
			color: $text-muted;

			h3 {
				font-size: 14px;
				font-weight: 600;
				text-transform: uppercase;
				letter-spacing: 0.05em;
				margin: 0;
				color: $text-color;
			}

			.count-badge {
				font-size: 11px;
				font-weight: 700;
				padding: 2px 7px;
				border-radius: 10px;
				background: rgba(255, 255, 255, 0.08);
				color: $text-muted;
			}
		}
	}

	ul.users-rail {
		display: flex;
		flex-flow: row;
		gap: 10px;
		list-style: none;
		overflow-x: auto;
		padding: 4px 0;
		margin: 0;

		li {
			display: flex;
			flex-shrink: 0;

			a.user-chip {
				display: flex;
				flex-flow: row;
				align-items: center;
				gap: 10px;
				padding: 8px 14px;
				border-radius: 24px;
				background: rgba(255, 255, 255, 0.05);
				border: 1px solid rgba(255, 255, 255, 0.08);
				color: $text-color;
				text-decoration: none;
				transition: all 180ms ease;

				.user-avatar {
					display: flex;
					align-items: center;
					justify-content: center;
					width: 26px;
					height: 26px;
					border-radius: 50%;
					background: $accent-color;
					color: #fff;
					font-weight: 700;
					font-size: 12px;
				}

				.user-name {
					font-size: 13.5px;
					font-weight: 500;
					max-width: 140px;
					overflow: hidden;
					text-overflow: ellipsis;
					white-space: nowrap;
				}

				&:hover {
					background: rgba(255, 255, 255, 0.1);
					border-color: rgba(255, 255, 255, 0.2);
					transform: translateY(-1px);
				}
			}
		}
	}
</style>
