<script lang="ts">
	import { formatUserID, formatDisplayUserID, userTargeted } from '$lib/user';
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
		let count = 0;
		for (let i = 0; i < dataPage.length; i++) {
			if (dataPage[i].customUserID === userID) count++;
			if (count > 1) return true;
		}
		return false;
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
	const totalPages = $derived(Math.ceil((users?.length ?? 0) / PAGE_SIZE));
	const dataPage = $derived(users ? users.slice((pageNumber - 1) * PAGE_SIZE, pageNumber * PAGE_SIZE) : []);

	$effect(() => {
		users;
		pageNumber = 1;
	});

	function nextPage() { pageNumber += 1; }
	function prevPage() { pageNumber -= 1; }
</script>

{#if dataPage.length > 0 || targetUser !== null}
	<div class="card">
		<h2 class="card-title">Top users</h2>
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
							class:highlighted-row={targetUser !== null && userTargeted(targetUser, ipAddress, customUserID)}
							class:dim-row={targetUser !== null && !userTargeted(targetUser, ipAddress, customUserID)}
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
			{#if dataPage.length === 0 && targetUser !== null}
				<div class="no-results">Filtered by: <span class="no-results-user">{formatDisplayUserID(targetUser)}</span> — <button onclick={() => { targetUser = null; setUserParams(null, null); }}>Clear filter</button></div>
			{/if}
		</div>
		<div class="buttons">
			<div class="current-page">Page {pageNumber} of {totalPages}</div>
			{#if pageNumber > 1}
				<button
					class:btn-prev={pageNumber < totalPages}
					onclick={prevPage}>Previous</button
				>
			{/if}
			{#if pageNumber < totalPages}
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
		border-bottom: 1px solid var(--border);
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
		color: var(--dim-text);
	}
	td,
	th {
		padding: 0.45em 0.5em;
	}

	.highlight-row:hover td {
		color: var(--dim-text);
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
		color: var(--dim-text);
		border: 1px solid var(--border);
		font-size: 0.85em;
	}

	.buttons button:hover {
		color: #ededed;
	}

	.no-results {
		font-size: 0.85em;
		color: #505050;
		padding: 0.8em 1.2em 0.4em;
	}
	.no-results-user {
		color: var(--dim-text);
	}
	.no-results button {
		color: var(--highlight);
		background: none;
		border: none;
		cursor: pointer;
		font-size: inherit;
		padding: 0;
	}
	.current-page {
		place-self: center;
		color: #505050;
		font-size: 0.85em;
		margin-right: auto;
	}
</style>
