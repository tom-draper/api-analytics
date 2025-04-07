<script lang="ts">
	import { formatUserID, getUserIdentifier } from '$lib/user';
	import { ColumnIndex } from '$lib/consts';
	import { page } from "$app/state";
	import { replaceState } from '$app/navigation';

	type Users = {
		[userID: string]: User;
	};

	type User = {
		ipAddress: string;
		customUserID: string;
		lastRequested: Date;
		requests: number;
		locations: Locations;
	};

	type Locations = {
		[location: string]: number;
	};

	function maxLocation(locations: Locations) {
		const max = {
			location: '',
			count: 0
		};
		for (const location in locations) {
			if (locations[location] > max.count) {
				max.count = locations[location];
				max.location = location;
			}
		}
		return max.location;
	}

	function resetFlags() {
		userIDActive = false;
		locationsActive = false;
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
				selectUser(ipAddress, userID); // If a different user was selected, select this one instead
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
				selectUser(ipAddress, userID); // If a different user was selected, select this one instead
			}
		}
	}

	function setUserParams(ipAddress: string | null, userID: string | null) {
		if (ipAddress === null) {
			page.url.searchParams.delete('ipAddress'); 
		} else {
			page.url.searchParams.set('ipAddress', ipAddress); 
		}
		if (userID === null) {
			page.url.searchParams.delete('userID')
		} else {
			page.url.searchParams.set('userID', userID);
		}
		replaceState(page.url, page.state);
	}

	function setIPAddressParam(ipAddress: string | null) {
		if (ipAddress === null) {
			page.url.searchParams.delete('ipAddress'); 
		} else {
			page.url.searchParams.set('ipAddress', ipAddress); 
		}
		replaceState(page.url, page.state);
	}

	function setUserIDParam(userID: string | null) {
		if (userID === null) {
			page.url.searchParams.delete('userID')
		} else {
			page.url.searchParams.set('userID', userID);
		}
		replaceState(page.url, page.state);
	}

	function build(data: RequestsData) {
		resetFlags();

		const users = getUsers(data);

		const totalUsers = Object.keys(users).length;
		const topUserRequestsCount = getTopUserRequestsCount(users);
		if (totalUsers < 10 || topUserRequestsCount <= 1) {
			topUsers = null;
			return;
		}

		userIDActive = getUserIDActive(users);
		locationsActive = getLocationsActive(users);

		topUsers = Object.values(users)
			.sort((a, b) => b.requests - a.requests)
			.slice(0, pageSize * maxPages);

		resetPage();
	}

	function getUsers(data: RequestsData) {
		const users: Users = {};

		for (const row of data) {
			const userID = getUserIdentifier(row);
			if (!userID) {
				continue;
			}

			const createdAt = row[ColumnIndex.CreatedAt];
			const location = row[ColumnIndex.Location];
			let user = users[userID];

			if (user) {
				user.requests++;
				user.locations[location] = (user.locations[location] || 0) + 1;

				if (createdAt > user.lastRequested) {
					user.lastRequested = createdAt;
				}
			} else {
				users[userID] = {
					ipAddress: row[ColumnIndex.IPAddress],
					customUserID: row[ColumnIndex.UserID],
					lastRequested: createdAt,
					requests: 1,
					locations: location ? { [location]: 1 } : {}
				};
			}
		}

		return users;
	}

	function getTopUserRequestsCount(users: Users) {
		let max = 0;
		for (const user of Object.values(users)) {
			if (user.requests > max) {
				max = user.requests;
			}
		}
		return max;
	}

	function getUserIDActive(users: Users) {
		for (const user of Object.values(users)) {
			if (user.customUserID !== '' && user.customUserID !== null) {
				return true;
			}
		}
		return false;
	}

	function getLocationsActive(users: Users) {
		for (const user of Object.values(users)) {
			if (user.locations) {
				return true;
			}
		}
		return false;
	}

	function multipleUserIDOccurances(userID: string) {
		const users = getPage();

		let count = 0;
		for (let i = 0; i < users.length; i++) {
			if (users[i].customUserID === userID) {
				count++;
			}

			if (count > 1) {
				return true;
			}
		}

		return false;
	}

	function userTargeted(ipAddress: string, userID: string) {
		if (!targetUser) {
			return false;
		}

		if (targetUser.composite) {
			return (
				formatUserID(ipAddress, userID) === formatUserID(targetUser.ipAddress, targetUser.userID)
			);
		} else {
			return targetUser.userID === userID;
		}
	}

	function totalPages() {
		if (!topUsers) {
			return 0;
		}

		return Math.ceil(topUsers.length / pageSize);
	}

	function resetPage() {
		pageNumber = 1;
		dataPage = getPage();
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
		if (!topUsers) {
			return [];
		}

		return topUsers.slice((pageNumber - 1) * pageSize, pageNumber * pageSize);
	}

	let topUsers: User[] | null = null;
	let dataPage: User[] | null = null;
	let userIDActive = false;
	let locationsActive = false;
	let pageNumber = 1;
	const pageSize = 10;
	const maxPages = 100;

	$: if (data && !targetUser) {
		build(data);
	}

	type TargetUser = {
		ipAddress: string;
		userID: string;
		composite: boolean;
	};

	export let data: RequestsData, targetUser: TargetUser | null;
</script>

{#if dataPage}
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
							<td on:click={() => selectUser(ipAddress, customUserID)}>{ipAddress}</td>
							{#if userIDActive}
								<td on:click={() => selectUser(ipAddress, customUserID)}>{customUserID ?? ''}</td>
							{/if}
							{#if locationsActive}
								<td on:click={() => selectUser(ipAddress, customUserID)}
									>{locations ? maxLocation(locations) : ''}</td
								>
							{/if}
							<td on:click={() => selectUser(ipAddress, customUserID)}>
								{lastRequested.toLocaleString()}
							</td>
							<td class="align-right" on:click={() => selectUser(ipAddress, customUserID)}>
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
					class:btn-prev={topUsers && topUsers.length / pageSize > pageNumber}
					on:click={prevPage}>Previous</button
				>
			{/if}
			{#if topUsers && topUsers.length / pageSize > pageNumber}
				<button class="btn-next" on:click={nextPage}>Next</button>
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
		/* color: var(--highlight) !important; */
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
