<script lang="ts">
	import { getUserIdentifier } from '../../lib/user';
	import { ColumnIndex } from '../../lib/consts';

	type Users = {
		[userID: string]: {
			ipAddress: string;
			location: string;
			customUserID: string;
			createdAt: Date;
			requests: number;
		};
	};

	function build(data: RequestsData) {
		const users: Users = {};
		for (let i = 0; i < data.length; i++) {
			const userID = getUserIdentifier(data[i]);
			if (!userID) {
				continue;
			}
			const ipAddress = data[i][ColumnIndex.IPAddress];
			const location = data[i][ColumnIndex.Location];
			const customUserID = data[i][ColumnIndex.UserID];
			const createdAt = data[i][ColumnIndex.CreatedAt];
			if (!(userID in users)) {
				users[userID] = {
					ipAddress,
					location,
					customUserID,
					createdAt,
					requests: 0,
				};
			}
			users[userID].requests += 1;
			if (createdAt > users[userID].createdAt) {
				users[userID].createdAt = createdAt;
			}
		}

		const totalUsers = Object.keys(users).length;
		const topUserRequestsCount = getTopUserRequestsCount(users);
		if (totalUsers < 10 || topUserRequestsCount <= 1) {
			topUsers = null;
			return;
		}

		customUserIDActive = userIDActive(users);

		topUsers = Object.values(users)
			.sort((a, b) => b.requests - a.requests)
			.slice(0, 15);
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

	function userIDActive(users: Users) {
		for (const user in users) {
			if (
				users[user].customUserID !== null &&
				users[user].customUserID !== ''
			) {
				return true;
			}
		}
		return false;
	}

	let topUsers = null;
	let customUserIDActive = false;

	$: if (data) {
		build(data);
	}

	export let data: RequestsData;
</script>

{#if topUsers}
	<div class="card">
		<div class="card-title">Top Users</div>
		<div>
			<table class="table">
				<thead>
					<tr>
						<th>IP Address</th>
						{#if customUserIDActive}
							<th>User ID</th>
						{/if}
						<th>Last Access</th>
						<th style="text-align: right;">Requests</th>
					</tr>
				</thead>
				<tbody>
					{#each topUsers as { ipAddress, location, customUserID, requests, createdAt }}
						<tr>
							<td>{ipAddress}</td>
							{#if customUserIDActive}
								<td>{customUserID}</td>
							{/if}
							<td>
								{createdAt.toLocaleString()}
							</td>
							<td style="text-align: right;">
								{requests.toLocaleString()}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
{/if}

<style scoped>
	.card {
		width: 100%;
		margin-top: 2em;
	}
	.table {
		margin: 1em 1.5em 1.8em;
		text-align: left;
		width: -webkit-fill-available;
	}
	tr {
		border-bottom: 1px solid #2e2e2e;
	}
	thead {
		font-weight: 600;
		font-size: 0.85em;
	}
	tbody {
		font-size: 0.85em;
	}
	table {
		border-collapse: collapse;
	}
	tbody {
		color: #707070;
	}
	td,
	th {
		padding: 0.45em 0.5em;
	}
</style>
