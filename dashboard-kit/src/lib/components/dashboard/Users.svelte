<script lang="ts">
	import { periodToDays, type Period } from '$lib/period';
	import { ColumnIndex } from '$lib/consts';

	function getPlotLayout() {
		return {
			title: false,
			autosize: true,
			margin: { r: 0, l: 0, t: 0, b: 0, pad: 0 },
			hovermode: false,
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 60,
			yaxis: {
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true,
				dragmode: false
			},
			xaxis: {
				visible: false,
				dragmode: false
			},
			dragmode: false
		};
	}

	function getUserIdentifier(request: RequestsData[number]) {
		return request[ColumnIndex.IPAddress] ?? '' + request[ColumnIndex.UserID].toString() ?? '';
	}

	function lines(data: RequestsData) {
		const n = 5;
		const x = [...Array(n).keys()];
		const uniqueUsers: Set<string>[] = Array(n);
		for (let i = 0; i < n; i++) {
			uniqueUsers[i] = new Set();
		}

		if (data.length > 0) {
			const start = data[0][ColumnIndex.CreatedAt].getTime();
			const end = data[data.length - 1][ColumnIndex.CreatedAt].getTime();
			const range = end - start;
			const interval = range > 0 ? range / n : 1; // Avoid division by zero

			for (const row of data) {
				const userID = getUserIdentifier(row);
				if (!userID) continue;

				const time = row[ColumnIndex.CreatedAt].getTime();
				const idx = Math.min(n - 1, Math.floor((time - start) / interval));

				uniqueUsers[idx].add(userID);
			}
		}

		const y = uniqueUsers.map((set) => set.size);

		return [
			{
				x,
				y,
				type: 'lines',
				marker: { color: 'transparent' },
				showlegend: false,
				line: { shape: 'spline', smoothing: 1, color: '#3FCF8E30' },
				fill: 'tozeroy',
				fillcolor: '#3fcf8e15'
			}
		];
	}

	function getPlotData(data: RequestsData) {
		return {
			data: lines(data),
			layout: getPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function generatePlot(data: RequestsData) {
		if (plotDiv.data) {
			refreshPlot(data);
		} else {
			newPlot(data);
		}
	}

	async function newPlot(data: RequestsData) {
		const plotData = getPlotData(data);
		Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	function refreshPlot(data: RequestsData) {
		Plotly.react(plotDiv, lines(data), getPlotLayout());
	}

	function togglePeriod() {
		perHour = !perHour;
	}

	function getPercentageChange(now: number, prev: number) {
		if (prev === 0) {
			return null;
		}

		return (now / prev) * 100 - 100;
	}

	function getUsersPerHour(users: Set<string>, period: Period) {
		if (users.size === 0) {
			return 0;
		}

		const days = periodToDays(period);
		if (days === null) {
			return 0;
		}

		return users.size / (24 * days);
	}

	function getUsers(data: RequestsData): Set<string> {
		const users: Set<string> = new Set();

		for (const row of data) {
			const userID = getUserIdentifier(row);
			if (userID) {
				users.add(userID);
			}
		}

		return users;
	}

	let plotDiv: HTMLDivElement;
	let userCount: number = 0;
	let usersPerHour: number;
	let perHour = false;
	let percentageChange: number | null;

	$: if (data) {
		const users = getUsers(data);
		const prevUsers = getUsers(prevData);

		userCount = users.size;
		percentageChange = getPercentageChange(users.size, prevUsers.size);
		usersPerHour = getUsersPerHour(users, period);
	}

	$: if (plotDiv && data) {
		generatePlot(data);
	}

	export let data: RequestsData, prevData: RequestsData, period: Period;
</script>

<button class="card" on:click={togglePeriod} title="Based on IP address">
	{#if perHour}
		<div class="card-title">
			Users <span class="per-hour">/ hour</span>
		</div>
		{#if usersPerHour}
			<div class="value">
				{usersPerHour === 0 ? '0' : usersPerHour.toFixed(2)}
			</div>
		{/if}
	{:else}
		{#if percentageChange}
			<div
				class="percentage-change flex"
				class:positive={percentageChange > 0}
				class:negative={percentageChange < 0}
			>
				{#if percentageChange > 0}
					<img class="arrow" src="/images/icons/green-up.png" alt="" />
				{:else if percentageChange < 0}
					<img class="arrow" src="/images/icons/red-down.png" alt="" />
				{/if}
				{Math.abs(percentageChange).toFixed(1)}%
			</div>
		{/if}
		<div class="card-title">Users</div>
		<div class="value">{userCount.toLocaleString()}</div>
	{/if}
	<div id="plotly">
		<div id="plotDiv" bind:this={plotDiv}>
			<!-- Plotly chart will be drawn inside this DIV -->
		</div>
	</div>
</button>

<style scoped>
	.card {
		width: calc(215px - 1em);
		margin: 0 0 0 1em;
		position: relative;
		cursor: pointer;
		padding: 0;
		overflow: hidden;
	}
	.value {
		padding: 0.55em 0.2em;
		font-size: 1.8em;
		font-weight: 700;
		position: inherit;
		z-index: 2;
	}
	.percentage-change {
		position: absolute;
		right: 20px;
		top: 20px;
		font-size: 0.8em;
	}
	.positive {
		color: var(--highlight);
	}
	.negative {
		color: rgb(228, 97, 97);
	}

	.per-hour {
		color: var(--dim-text);
		font-size: 0.8em;
		margin-left: 4px;
	}
	button {
		font-size: unset;
		font-family: unset;
		font-family: 'Noto Sans' !important;
	}
	.arrow {
		height: 11px;
		align-self: center;
		margin-right: 0.25em;
	}
	#plotly {
		position: absolute;
		width: 110%;
		bottom: 0;
		overflow: hidden;
		margin: 0 -5%;
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 0 1em;
		}
	}
</style>
