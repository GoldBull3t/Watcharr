<script lang="ts">
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import {
		getOrdinalSuffix,
		monthsShort,
		userHasPermission,
	} from "@/lib/util/helpers";
	import { UserPermission, type ManagedUser } from "@/types";
	import { req } from "@/lib/util/api";
	import EditUserModal from "./modals/EditUserModal.svelte";
	import UserTypeIcon from "@/lib/user/UserTypeIcon.svelte";
	import Error from "@/lib/Error.svelte";

	const currentYear = new Date(Date.now()).getFullYear();

	let allUsers: ManagedUser[] = $state([]);
	let editingUser: ManagedUser | undefined = $state();

	async function getUsers() {
		allUsers = await req.get<ManagedUser[]>(`/server/users`);
	}
</script>

<svelte:head>
	<title>User Management - Watcharr</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<div class="page-header">
			<div>
				<h2>User Management</h2>
				<p class="subtitle">
					View and manage accounts, roles, and access permissions on your
					server.
				</p>
			</div>
			{#if allUsers?.length > 0}
				<span class="user-count-badge">{allUsers.length} Users</span>
			{/if}
		</div>

		{#await getUsers()}
			<div class="spinner-wrap">
				<Spinner />
			</div>
		{:then}
			<div class="table-card">
				<table>
					<thead>
						<tr>
							<th>User</th>
							<th>Visibility</th>
							<th>Joined</th>
							<th>Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each allUsers as u (u.id)}
							{@const joinDate = new Date(u.createdAt)}
							{@const isAdmin = userHasPermission(
								u.permissions,
								UserPermission.PERM_ADMIN,
							)}
							<tr>
								<td class="username">
									<div class="user-cell">
										<div class="user-avatar-placeholder">
											<UserTypeIcon type={u.type} />
										</div>
										<div class="user-names">
											<span class="name">{u.username}</span>
											<span
												class={`role-pill ${isAdmin ? "role-admin" : "role-user"}`}
											>
												{isAdmin ? "Admin" : "Member"}
											</span>
										</div>
									</div>
								</td>
								<td>
									<span
										class={`status-badge ${u.private ? "badge-private" : "badge-public"}`}
									>
										{u.private ? "Private" : "Public"}
									</span>
								</td>
								<td class="col-date">
									{joinDate.getDate()}{getOrdinalSuffix(joinDate.getDate())}
									{monthsShort[joinDate.getMonth()]}
									{joinDate.getFullYear() === currentYear
										? ""
										: `'${String(joinDate.getFullYear()).substring(2, 4)}`}
								</td>
								<td class="col-action">
									<button
										class="edit-btn"
										onclick={() => (editingUser = u)}
										title="Edit user"
									>
										<Icon i="pencil" wh={16} />
										<span>Edit</span>
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			{#if editingUser}
				<EditUserModal
					bind:user={editingUser}
					onClose={() => {
						editingUser = undefined;
						getUsers();
					}}
				/>
			{/if}
		{:catch err}
			<Error error={err} pretty="Failed to fetch users!" />
		{/await}
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 24px 20px 80px 20px;

		.inner {
			width: 100%;
			max-width: 900px;
			display: flex;
			flex-direction: column;
			gap: 24px;
		}
	}

	.page-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		flex-wrap: wrap;

		h2 {
			font-size: 26px;
			font-weight: 700;
			letter-spacing: -0.02em;
			color: $text-color;
			margin: 0 0 4px 0;
		}

		.subtitle {
			font-size: 14px;
			color: $text-muted;
			margin: 0;
		}

		.user-count-badge {
			font-size: 13px;
			font-weight: 600;
			padding: 4px 14px;
			border-radius: 20px;
			background: rgba(255, 255, 255, 0.06);
			border: 1px solid rgba(255, 255, 255, 0.1);
			color: $text-color;
		}
	}

	.spinner-wrap {
		display: flex;
		justify-content: center;
		margin: 40px 0;
	}

	.table-card {
		overflow-x: auto;
		border-radius: 16px;
		background: rgba(255, 255, 255, 0.025);
		border: 1px solid rgba(255, 255, 255, 0.08);

		table {
			width: 100%;
			border-collapse: collapse;
			text-align: left;

			th {
				padding: 14px 20px;
				font-size: 12px;
				font-weight: 600;
				text-transform: uppercase;
				letter-spacing: 0.05em;
				color: $text-muted;
				border-bottom: 1px solid rgba(255, 255, 255, 0.08);
				background: rgba(255, 255, 255, 0.02);
			}

			td {
				padding: 14px 20px;
				border-bottom: 1px solid rgba(255, 255, 255, 0.04);
				color: $text-color;
				font-size: 14px;
			}

			tbody tr {
				transition: background-color 150ms ease;

				&:hover {
					background-color: rgba(255, 255, 255, 0.03);
				}

				&:last-child td {
					border-bottom: none;
				}
			}
		}
	}

	.user-cell {
		display: flex;
		align-items: center;
		gap: 12px;

		.user-avatar-placeholder {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 34px;
			height: 34px;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.06);
			border: 1px solid rgba(255, 255, 255, 0.1);
		}

		.user-names {
			display: flex;
			align-items: center;
			gap: 8px;

			.name {
				font-weight: 600;
				color: #fff;
			}
		}
	}

	.role-pill {
		font-size: 11px;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 12px;
		text-transform: uppercase;
		letter-spacing: 0.04em;

		&.role-admin {
			background: rgba(99, 102, 241, 0.15);
			color: #a5b4fc;
			border: 1px solid rgba(99, 102, 241, 0.3);
		}

		&.role-user {
			background: rgba(255, 255, 255, 0.06);
			color: $text-muted;
			border: 1px solid rgba(255, 255, 255, 0.08);
		}
	}

	.status-badge {
		font-size: 12px;
		font-weight: 500;
		padding: 3px 10px;
		border-radius: 12px;

		&.badge-private {
			background: rgba(245, 158, 11, 0.12);
			color: #fbbf24;
			border: 1px solid rgba(245, 158, 11, 0.25);
		}

		&.badge-public {
			background: rgba(16, 185, 129, 0.12);
			color: #34d399;
			border: 1px solid rgba(16, 185, 129, 0.25);
		}
	}

	.col-date {
		color: $text-muted;
		font-size: 13.5px;
	}

	.col-action {
		text-align: right;

		.edit-btn {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			padding: 6px 14px;
			border-radius: 8px;
			font-size: 13px;
			font-weight: 500;
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.1);
			color: $text-color;
			cursor: pointer;
			transition: all 150ms ease;

			&:hover {
				background: rgba(255, 255, 255, 0.12);
				border-color: rgba(255, 255, 255, 0.2);
				color: #fff;
				transform: translateY(-1px);
			}
		}
	}
</style>
