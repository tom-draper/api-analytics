<script lang="ts">
	import { formatUserID } from '$lib/user';
	import { setParam, setParamNoReplace } from '$lib/params';
	import type { TopUserData } from '$lib/aggregate';

	type Locations = { [location: string]: number };

	function maxLocation(locations: Locations) {
		const max = { location: '', count: 0 };
		for (const location in locations) {
			if (locations[location] > max.count) {
				max.count = locations[location];
				max.location = location;
			}
		}
		return max.location;
	}

	function selectUser(ipAddress: string, userID: string) {
		if (!targetUser) {
			targetUser = {
				composite: !(userID && multipleUserIDOccurances(userID)),
				ipAddress,
				userID
			};
			if (targetUser.composite) {
				setUserParams(ipAddress, userID);
			} else {
				setUserIDParam(userID);
			}
		} else if (targetUser.composite) {
			const formattedUserID = formatUserID(ipAddress, userID);
			const formattedTargetUserID = formatUserID(targetUser.ipAddress, targetUser.userID);
			targetUser = null;
			if (formattedUserID !== formattedTargetUserID) {
				selectUser(ipAddress, userID);
			} else {
				setUserParams(null, null);
			}
		} else {
			if (userID === targetUser.userID) {
				targetUser.composite = true;
				targetUser.ipAddress = ipAddress;
				setIPAddressParam(ipAddress);
			} else {
				targetUser = null;
				selectUser(ipAddress, userID);
			}
		}
	}

	function setUserParams(ipAddress: string | null, userID: string | null) {
		setParamNoReplace('ipAddress', ipAddress);
		setParam('userID', userID);
	}

	function setIPAddressParam(ipAddress: string | null) {
		setParam('ipAddress', ipAddress);
	}

	function setUserIDParam(userID: string | null) {
		setParam('userID', userID);
	}

	function multipleUserIDOccurances(userID: string) {
		const page = getPage();
		let count = 0;
		for (let i = 0; i < page.length; i++) {
			if (page[i].customUserID === userID) count++;
			if (count > 1) return true;
		}
		return false;
	}

	function userTargeted(ipAddress: string, userID: string) {
		if (!targetUser) return false;
		if (targetUser.composite) {
			return formatUserID(ipAddress, userID) === formatUserID(targetUser.ipAddress, targetUser.userID);
		} else {
			return targetUser.userID === userID;
		}
	}

	function totalPages() {
		if (!users) return 0;
		return Math.ceil(users.length / PAGE_SIZE);
	}

	function nextPage() {
		pageNumber += 1;
		dataPage = getPage();
	}

	function prevPage() {
		pageNumber -= 1;
		dataPage = getPage();
	}

	function getPage() {
		if (!users) return [];
		return users.slice((pageNumber - 1) * PAGE_SIZE, pageNumber * PAGE_SIZE);
	}

	type TargetUser = { ipAddress: string; userID: string; composite: boolean };

	let { users, userIDActive, locationsActive, targetUser = $bindable<TargetUser | null>(null) }: {
		users: TopUserData[] | null;
		userIDActive: boolean;
		locationsActive: boolean;
		targetUser: TargetUser | null;
	} = $props();

	const PAGE_SIZE = 10;
	let pageNumber = $state(1);
	let dataPage = $state<TopUserData[] | null>(null);

	$effect(() => {
		pageNumber = 1;
		dataPage = users ? users.slice(0, PAGE_SIZE) : null;
	});
</script>

{#if dataPage !== null}
	<div class="card">
		<h2 class="card-title">Top Users</h2>
		<div class="table-container">
			<table class="table">
				<thead>
					<tr>
						<th>IP Address</th>
						{#if userIDActive}
							<th>User ID</th>
						{/if}
						{#if locationsActive}
							<th>Location</th>
						{/if}
						<th>Last Access</th>
						<th class="align-right">Requests</th>
					</tr>
				</thead>
				<tbody>
					{#each dataPage as { ipAddress, customUserID, requests, locations, lastRequested }, i}
						<tr
							class="highlight-row"
							class:highlighted-row={targetUser !== null && userTargeted(ipAddress, customUserID)}
							class:dim-row={targetUser !== null && !userTargeted(ipAddress, customUserID)}
							class:last-row={i === dataPage.length - 1}
						>
							<td onclick={() => selectUser(ipAddress, customUserID)}>{ipAddress}</td>
							{#if userIDActive}
								<td onclick={() => selectUser(ipAddress, customUserID)}>{customUserID ?? ''}</td>
							{/if}
							{#if locationsActive}
								<td onclick={() => selectUser(ipAddress, customUserID)}
									>{locations ? maxLocation(locations) : ''}</td
								>
							{/if}
							<td onclick={() => selectUser(ipAddress, customUserID)}>
								{lastRequested.toLocaleString()}
							</td>
							<td class="align-right" onclick={() => selectUser(ipAddress, customUserID)}>
								{requests.toLocaleString()}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
		<div class="buttons">
			<div class="current-page">Page {pageNumber} of {totalPages()}</div>
			{#if pageNumber > 1}
				<button
					class:btn-prev={users && users.length / PAGE_SIZE > pageNumber}
					onclick={prevPage}>Previous</button
				>
			{/if}
			{#if users && users.length / PAGE_SIZE > pageNumber}
				<button class="btn-next" onclick={nextPage}>Next</button>
			{/if}
		</div>
	</div>
{/if}

<style scoped>
	.card {
		width: 100%;
		margin-top: 2em;
	}
	.table-container {
		display: flex;
		overflow: auto;
	}
	.table {
		margin: 1em 1.2em 1em;
		text-align: left;
		flex: 1;
	}
	tr {
		border-bottom: 1px solid #2e2e2e;
	}
	.align-right {
		text-align: right;
	}
	th {
		font-weight: 500;
	}
	table {
		border-collapse: collapse;
		font-size: 0.85em;
	}
	tbody {
		color: #707070;
	}
	td,
	th {
		padding: 0.45em 0.5em;
	}

	.highlight-row:hover td {
		color: #707070;
		color: #808080;
	}

	.highlighted-row td {
		color: #ededed !important;
	}

	.highlight-row td {
		cursor: pointer;
	}
	.dim-row {
		color: #505050;
	}

	.last-row {
		border-bottom: none;
	}

	.buttons {
		display: flex;
		justify-content: right;
		margin: 0 1.2em 1em 1.2em;
	}

	.btn-prev {
		margin-right: 1em;
	}

	.buttons button {
		cursor: pointer;
		border-radius: 4px;
		padding: 0.4em 1em;
		background: transparent;
		color: #505050;
		color: #707070;
		border: 1px solid #2e2e2e;
		font-size: 0.85em;
	}

	.buttons button:hover {
		color: #ededed;
	}

	.current-page {
		place-self: center;
		color: #505050;
		font-size: 0.85em;
		margin-right: auto;
	}
</style>
