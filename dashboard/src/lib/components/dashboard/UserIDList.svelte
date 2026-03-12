<script lang="ts">
	import { setParam } from '$lib/params';
	import type { UserIDBar } from '$lib/aggregate';
	import type { DashboardSettings } from '$lib/settings';

	function selectUserID(userID: string) {
		if (targetUser?.userID === userID && !targetUser?.composite) {
			targetUser = null;
			setParam('userID', null);
		} else {
			targetUser = { ipAddress: '', userID, composite: false };
			setParam('userID', userID);
		}
	}

	let { userIDBars, targetUser = $bindable<DashboardSettings['targetUser']>(null) }: {
		userIDBars: UserIDBar[];
		targetUser: DashboardSettings['targetUser'];
	} = $props();
</script>

<div class="card">
	<div class="card-title">User ID</div>
	<div class="list">
		{#each userIDBars as bar, i}
			<div class="row-container">
				<button
					class="row"
					id="row-{i}"
					class:selected={targetUser?.userID === bar.userID && !targetUser?.composite}
					onclick={() => selectUserID(bar.userID)}
				>
					<div class="label">
						<span class="font-semibold">{bar.count.toLocaleString()}</span>
						{bar.userID}
					</div>
					<div class="background" style="width: {bar.height * 100}%"></div>
				</button>
			</div>
		{/each}
	</div>
</div>

<style scoped>
	.card {
		min-height: 361px;
		margin-top: 2em;
		margin-left: 2em;
	}
	.list {
		margin: 0.9em 20px 0.6em;
	}
	.row {
		border-radius: var(--radius-sm);
		margin: 5px 0;
		color: var(--light-background);
		text-align: left;
		position: relative;
		font-size: 0.85em;
		width: 100%;
		cursor: pointer;
	}
	.row:hover {
		background: var(--fade-right);
	}
	.selected {
		background: var(--fade-right);
	}
	.label {
		position: relative;
		flex-grow: 1;
		z-index: 1;
		pointer-events: none;
		color: var(--muted-text);
		padding: 3px 12px;
		overflow-wrap: break-word;
	}
	.row-container {
		display: flex;
	}
	.background {
		border-radius: var(--radius-sm);
		background: var(--highlight);
		text-align: left;
		position: absolute;
		top: 0;
		height: 100%;
		font-size: 0.85em;
	}
	@media screen and (max-width: 1600px) {
		.card {
			margin-left: 0;
			width: 100%;
			min-height: unset;
		}
	}
	@media screen and (max-width: 1070px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
