<script lang="ts">
	import { periodToDays } from '../../lib/period';
	import type { Period } from '../../lib/settings';
	import { ColumnIndex } from '../../lib/consts';

	function usersPlotLayout() {
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
				dragmode: false,
			},
			xaxis: {
				visible: false,
				dragmode: false,
			},
			dragmode: false,
		};
	}

	function getUserIdentifier(request: RequestsData[number]) {
		return (
			request[ColumnIndex.IPAddress] ??
			'' + request[ColumnIndex.UserID].toString() ??
			''
		);
	}

	function lines(data: RequestsData) {
		const n = 5;
		const x = [...Array(n).keys()];
		const uniqueUsers = Array.from({ length: n }, () => new Set());

		if (data.length > 0) {
			const start = data[0][ColumnIndex.CreatedAt].getTime();
			const end = data[data.length - 1][ColumnIndex.CreatedAt].getTime();
			const range = end - start;
			for (let i = 0; i < data.length; i++) {
				const userID = getUserIdentifier(data[i]);
				if (!userID) {
					continue;
				}
				const time = data[i][ColumnIndex.CreatedAt].getTime();
				const diff = time - start;
				const idx = Math.min(n - 1, Math.floor(diff / (range / n)));
				if (idx >= 0 && idx < n) {
					uniqueUsers[idx].add(userID);
				}
			}
		}

		const y = uniqueUsers.map((set) => set.size);

		return [
			{
				x: x,
				y: y,
				type: 'lines',
				marker: { color: 'transparent' },
				showlegend: false,
				line: { shape: 'spline', smoothing: 1, color: '#3FCF8E30' },
				fill: 'tozeroy',
				fillcolor: '#3fcf8e15',
			},
		];
	}

	function usersPlotData(data: RequestsData) {
		return {
			data: lines(data),
			layout: usersPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false,
			},
		};
	}

	function genPlot(data: RequestsData) {
		const plotData = usersPlotData(data);
		//@ts-ignore
		new Plotly.newPlot(
			plotDiv,
			plotData.data,
			plotData.layout,
			plotData.config,
		);
	}

	function togglePeriod() {
		perHour = !perHour;
	}

	function setPercentageChange(now: number, prev: number) {
		if (prev === 0) {
			percentageChange = null;
		} else {
			percentageChange = (now / prev) * 100 - 100;
		}
	}

	function getUsers(data: RequestsData): Set<string> {
		const users: Set<string> = new Set();
		for (let i = 0; i < data.length; i++) {
			const userID = getUserIdentifier(data[i]);
			if (userID) {
				users.add(userID);
			}
		}
		return users;
	}

	function build(data: RequestsData) {
		({size: numUsers} = getUsers(data));

		const prevUsers = getUsers(prevData);
		const prevNumUsers = prevUsers.size;

		setPercentageChange(numUsers, prevNumUsers);

		if (numUsers > 0) {
			const days = periodToDays(period);
			if (days !== null) {
				usersPerHour = numUsers / (24 * days);
			}
		} else {
			usersPerHour = 0;
		}
		genPlot(data);
	}

	let plotDiv: HTMLDivElement;
	let numUsers: number = 0;
	let usersPerHour: number;
	let perHour = false;
	let percentageChange: number;
	// let mounted = false;
	// onMount(() => {
	// 	mounted = true;
	// });

	$: if (plotDiv && data) {
		build(data);
	}

	// $: data && mounted && build();

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
				class="percentage-change"
				class:positive={percentageChange > 0}
				class:negative={percentageChange < 0}
			>
				{#if percentageChange > 0}
					<img class="arrow" src="../img/up.png" alt="" />
				{:else if percentageChange < 0}
					<img class="arrow" src="../img/down.png" alt="" />
				{/if}
				{Math.abs(percentageChange).toFixed(1)}%
			</div>
		{/if}
		<div class="card-title">Users</div>
		<div class="value">{numUsers.toLocaleString()}</div>
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
		padding: 20px 10px;
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
