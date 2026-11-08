<script lang="ts">
	import Checkbox from "@/lib/Checkbox.svelte";
	import DropDown from "@/lib/DropDown.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import { req } from "@/lib/util/api";
	import { ReqerError } from "@/lib/util/fetch";
	import { userHasPermission } from "@/lib/util/helpers";
	import { notify } from "@/lib/util/notify";
	import { UserPermission, UserType, type ManagedUser } from "@/types";

	interface UpdateUserRequest {
		permissions?: number;
		type?: UserType;
	}

	interface Props {
		user: ManagedUser;
		onClose: (saved?: boolean) => void;
	}

	let { user, onClose }: Props = $props();

	let error: string | undefined = $state();
	let saving = $state(false);

	const initialPerms = user.permissions;
	const initialType = user.type;

	let adminPerm = $state(
		userHasPermission(user.permissions, UserPermission.PERM_ADMIN),
	);
	let requestContentPerm = $state(
		userHasPermission(user.permissions, UserPermission.PERM_REQUEST_CONTENT),
	);
	let autoApprovePerm = $state(
		userHasPermission(
			user.permissions,
			UserPermission.PERM_REQUEST_CONTENT_AUTO_APPROVE,
		),
	);
	let selectedType = $state<UserType>(user.type);

	function computePermissions(): number {
		let p = UserPermission.PERM_NONE;
		if (adminPerm) p |= UserPermission.PERM_ADMIN;
		if (requestContentPerm) p |= UserPermission.PERM_REQUEST_CONTENT;
		if (autoApprovePerm) p |= UserPermission.PERM_REQUEST_CONTENT_AUTO_APPROVE;
		return p;
	}

	async function save() {
		error = undefined;
		const computedPerms = computePermissions();
		const changedPerms = computedPerms !== initialPerms;
		const changedType = selectedType !== initialType;

		if (!changedPerms && !changedType) {
			error = "Nothing has been changed";
			return;
		}

		saving = true;
		try {
			const toUpdate: UpdateUserRequest = {};
			if (changedPerms) {
				toUpdate.permissions = computedPerms;
			}
			if (changedType) {
				toUpdate.type = selectedType;
			}
			await req.post(`/server/users/${user.id}`, toUpdate);
			notify({
				type: "success",
				text: "Changes saved!",
			});
			onClose(true);
		} catch (err) {
			console.error("Failed to save user!", err);
			error = ReqerError.getMsg(err, "Failed to save");
		} finally {
			saving = false;
		}
	}

	let confirmingDelete = $state(false);
	let deleting = $state(false);

	/** Permanently delete this user and all their data (#473). */
	async function deleteUser() {
		if (deleting) return;
		deleting = true;
		try {
			await req.delete(`/server/users/${user.id}`);
			notify({ type: "success", text: `User ${user.username} deleted` });
			onClose(true);
		} catch (err) {
			console.error("Failed to delete user!", err);
			error = ReqerError.getMsg(err, "Failed to delete user");
			confirmingDelete = false;
		} finally {
			deleting = false;
		}
	}
</script>

<Modal
	title="Edit User"
	desc={`Configuring ${user.username}`}
	maxWidth="500px"
	onClose={() => onClose(false)}
>
	{#if error}
		<span class="error">{error}!</span>
	{/if}

	<SettingsList>
		<h3 class="norm">Permissions</h3>

		<Setting
			title="Admin"
			desc="Give user admin, overrides all other permissions."
			row
		>
			<Checkbox
				name="USER_PERM_ADMIN"
				bind:value={adminPerm}
				toggled={() => {
					error = undefined;
				}}
			/>
		</Setting>

		<Setting
			title="Request Content"
			desc="Give user permission to request content."
			row
		>
			<Checkbox
				name="USER_PERM_REQUEST_CONTENT"
				bind:value={requestContentPerm}
				toggled={() => {
					error = undefined;
				}}
			/>
		</Setting>

		<Setting
			title="Auto Approve Content Request"
			desc="Auto approve user's content requests."
			row
		>
			<Checkbox
				name="PERM_REQUEST_CONTENT_AUTO_APPROVE"
				bind:value={autoApprovePerm}
				toggled={() => {
					error = undefined;
				}}
			/>
		</Setting>

		<h3 class="norm">Other</h3>

		<Setting
			title="User Type"
			desc="The type of this user, affects how they login and certain features. Currently only possible to swap between Watcharr/Proxy types."
		>
			{#if !user.type || user.type === UserType.Proxy}
				<DropDown
					placeholder="Unknown"
					bind:active={selectedType}
					options={[
						{
							id: UserType.Watcharr,
							value: "Watcharr",
						},
						{
							id: UserType.Proxy,
							value: "Proxy",
						},
					]}
					isDropDownItem={true}
					showActiveElementsInOptions={true}
					onChange={() => {
						error = undefined;
					}}
				/>
			{:else}
				<p>This option is not supported for this user type yet.</p>
			{/if}
		</Setting>

		<h3 class="norm">Danger Zone</h3>

		<Setting
			title="Delete User"
			desc="Permanently delete this user and all of their data (library, tags, activity). This cannot be undone."
		>
			{#if confirmingDelete}
				<div class="btns delete-confirm">
					<button
						class="danger"
						disabled={deleting}
						onclick={() => deleteUser()}
					>
						{deleting ? "Deleting..." : `Yes, delete ${user.username}`}
					</button>
					<button
						class="plain"
						disabled={deleting}
						onclick={() => (confirmingDelete = false)}
					>
						Cancel
					</button>
				</div>
			{:else}
				<button class="danger" onclick={() => (confirmingDelete = true)}>
					Delete User
				</button>
			{/if}
		</Setting>

		<div class="btns">
			<button disabled={saving} onclick={() => save()}>
				{saving ? "Saving..." : "Save"}
			</button>
		</div>
	</SettingsList>
</Modal>

<style lang="scss">
	.btns {
		display: flex;
		flex-flow: row;
		gap: 10px;

		:first-child {
			margin-left: auto;
		}

		button {
			width: max-content;
			padding-left: 15px;
			padding-right: 15px;
		}
	}

	button.danger {
		width: max-content;
		padding-left: 15px;
		padding-right: 15px;
		background-color: rgba(220, 38, 38, 0.15);
		border: 1px solid rgba(220, 38, 38, 0.5);
		color: #f87171;

		&:hover:not(:disabled) {
			background-color: rgba(220, 38, 38, 0.3);
		}
	}

	.delete-confirm :first-child {
		margin-left: 0;
	}

	.error {
		position: sticky;
		top: 0;
		display: flex;
		justify-content: center;
		width: 100%;
		padding: 10px;
		background-color: rgb(221, 48, 48);
		text-transform: capitalize;
		color: white;
		margin-bottom: 15px;
	}
</style>
