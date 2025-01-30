<script lang="ts">
	import { getUserIdentifier } from '../../lib/user';
	import { ColumnIndex } from '../../lib/consts';

	type Users = {
		[userID: string]: {
			ipAddress: string;
			customUserID: string;
			lastRequested: Date;
			requests: number;
			locations: Locations;
		};
	};

	type Locations = {
		[location: string]: number;
	};

	function maxLocation(locations: Locations) {
		const max = {
			location: '',
			count: 0,
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

	function selectUser(ipAddress: string, customUserID: string) {
		const formattedUserID = formatUserID(ipAddress, customUserID);
		if (targetUser === formattedUserID) {
			targetUser = null;
		} else {
			targetUser = formattedUserID;
		}
	}

	function formatUserID(ipAddress: string, customUserID: string) {
		return `${ipAddress} ${customUserID}`;
	}

	function build(data: RequestsData) {
		resetFlags();

		const users: Users = {};
		for (let i = 0; i < data.length; i++) {
			const userID = getUserIdentifier(data[i]);
			if (!userID) {
				continue;
			}

			const createdAt = data[i][ColumnIndex.CreatedAt];
			const location = data[i][ColumnIndex.Location];
			if (!(userID in users)) {
				const ipAddress = data[i][ColumnIndex.IPAddress];
				const customUserID = data[i][ColumnIndex.UserID];
				users[userID] = {
					ipAddress,
					customUserID,
					lastRequested: createdAt,
					requests: 1,
					locations: location ? { [location]: 1 } : {},
				};
			} else {
				users[userID].requests += 1;
				users[userID].locations[location] ??= 0
				users[userID].locations[location] += 1;
			}

			if (createdAt > users[userID].lastRequested) {
				users[userID].lastRequested = createdAt;
			}
		}

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
		page = topUsers.slice(0, pageSize);
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
			if (
				user.customUserID !== '' &&
					user.customUserID !== null
			) {
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

	function totalPages() {
		return Math.ceil(topUsers.length / pageSize);
	}

	function nextPage() {
		pageNumber += 1;
		setPage();
	}

	function prevPage() {
		pageNumber -= 1;
		setPage();
	}

	function setPage() {
		page = topUsers.slice((pageNumber - 1) * pageSize, pageNumber * pageSize);
	}

	let topUsers = null;
	let page = null;
	let userIDActive = false;
	let locationsActive = false;
	let pageNumber = 1;
	const pageSize = 10;
	const maxPages = 100;

	$: if (data && !targetUser) {
		build(data);
	}

	export let data: RequestsData, targetUser: string;
</script>

{#if topUsers || targetUser}
	<div class="card">
		<div class="card-title">Top Users</div>
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
					{#each page as { ipAddress, customUserID, requests, locations, lastRequested }, i}
						<!-- svelte-ignore a11y-click-events-have-key-events -->
						<tr class="highlight-row" class:highlighted-row={targetUser !== null && targetUser === formatUserID(ipAddress, customUserID)} class:dim-row={targetUser !== null && targetUser !== formatUserID(ipAddress, customUserID)} class:last-row={i === page.length - 1}>
							<td on:click={() => selectUser(ipAddress, customUserID)}>{ipAddress}</td>
							{#if userIDActive}
								<td on:click={() => selectUser(ipAddress, customUserID)}>{customUserID ?? ''}</td>
							{/if}
							{#if locationsActive}
								<td on:click={() => selectUser(ipAddress, customUserID)}>{locations ? maxLocation(locations) : ''}</td>
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
				<button class:btn-prev={(topUsers.length / pageSize) > pageNumber} on:click={prevPage}>Previous</button>
			{/if}
			{#if (topUsers.length / pageSize) > pageNumber}
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
thead {
	font-weight: 600;
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
	color: #EDEDED;
}

.highlighted-row td {
	color: var(--highlight) !important;
}

.dim-row td {
	color: #505050;
}

.highlight-row td {
	cursor: pointer;
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
	border: 1px solid #2e2e2e;
}

.buttons button:hover {
	color: #EDEDED;
}

.current-page {
	place-self: center;
	color: #505050;
	font-size: 0.85em;
	margin-right:auto;
}

</style>
