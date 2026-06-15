<script lang="ts">
	import { setParam } from '$lib/params';
	import type { UserIDBar } from '$lib/aggregate';
	import type { DashboardSettings } from '$lib/settings';
	import BarList from './BarList.svelte';

	let { userIDBars, targetUser = $bindable<DashboardSettings['targetUser']>(null) }: {
		userIDBars: UserIDBar[];
		targetUser: DashboardSettings['targetUser'];
	} = $props();

	function isSelected(userID: string) {
		return targetUser?.userID === userID && !targetUser?.composite;
	}

	const rows = $derived(
		userIDBars.map((bar) => ({
			value: bar.userID,
			label: bar.userID,
			count: bar.count,
			height: bar.height,
			selected: isSelected(bar.userID)
		}))
	);

	function select(userID: string) {
		if (isSelected(userID)) {
			targetUser = null;
			setParam('userID', null);
		} else {
			targetUser = { ipAddress: '', userID, composite: false };
			setParam('userID', userID);
		}
	}
</script>

<BarList title="User ID" {rows} onSelect={select} minHeight="361px" marginTop="2em" />
