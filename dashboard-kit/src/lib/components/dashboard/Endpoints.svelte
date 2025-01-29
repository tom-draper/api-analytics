<script lang="ts">
	import { ColumnIndex } from '$lib/consts';

	// Integer to method string mapping used by server
	const methodMap = [
		'GET',
		'POST',
		'PUT',
		'PATCH',
		'DELETE',
		'OPTIONS',
		'CONNECT',
		'HEAD',
		'TRACE',
	];

	type EndpointFreq = Map<
		string,
		{ path: string; status: number; count: number }
	>;

	type MapValue<A> = A extends Map<unknown, infer V> ? V : never;

	function endpointFreq() {
		const freq: EndpointFreq = new Map();
		for (let i = 0; i < data.length; i++) {
			// Create groups of endpoints by path + status
			const path = ignoreParams ? data[i][ColumnIndex.Path].split('?')[0] : data[i][ColumnIndex.Path];
			const status = data[i][ColumnIndex.Status];
			const endpointID = `${path}${status}`;
			if (!freq.has(endpointID)) {
				const method = methodMap[data[i][ColumnIndex.Method]];
				freq.set(endpointID, {
					path: `${method}  ${path}`,
					status: status,
					count: 0,
				});
			}
			freq.get(endpointID).count++;
		}
		return freq;
	}

	function statusMatch(status: number) {
		return (
			activeBtn === 'all' ||
			(activeBtn === 'success' && status >= 200 && status <= 299) ||
			(activeBtn === 'client' && status >= 400 && status <= 499) ||
			(activeBtn === 'server' && status >= 500)
		);
	}

	function setTargetEndpoint(endpoint: string, status: number) {
		if (endpoint === null || status === null) {
			// Trigger reset if input is null
			targetPath = null;
			targetStatus = null;
		} else if (targetPath === null) {
			// At starting state, set the path first
			targetPath = endpoint;
		} else if (endpoints.length > 1 && targetStatus === null) {
			// Path already set, now narrow down status (if multiple endpoints still exist)
			targetStatus = status;
		} else {
			// Path and status already set, reset
			targetPath = null;
			targetStatus = null;
		}
	}

	function getEndpoints() {
		const freq = endpointFreq();

		// Convert object to list
		const freqArr: MapValue<EndpointFreq>[] = [];
		let maxCount = 0;
		for (const value of freq.values()) {
			if (statusMatch(value.status)) {
				freqArr.push(value);
				if (value.count > maxCount) {
					maxCount = value.count;
				}
			}
		}

		freqArr.sort((a, b) => {
			return b.count - a.count;
		});

		return {
			endpoints: freqArr.slice(0, 50),
			maxCount: maxCount,
		};
	}

	function build() {
		({ endpoints: endpoints, maxCount: maxCount } = getEndpoints());
	}

	function setBtn(value: typeof activeBtn) {
		activeBtn = value;
	}

	let endpoints: {
		path: string;
		status: number;
		count: number;
	}[];
	let maxCount: number;
	let activeBtn: 'all' | 'success' | 'client' | 'server' = 'all';

	$: if (data && activeBtn) {
		build();
	}

	export let data: RequestsData,
		targetPath: string,
		targetStatus: number,
		ignoreParams: boolean;
</script>

<div class="card">
	<div class="card-title">
		Endpoints
		<div class="toggle">
			<button
				class="cancel"
				class:visible={targetPath !== null}
				on:click={() => {
					setTargetEndpoint(null, null);
				}}>Clear</button
			>
			<button
				class:active={activeBtn === 'all'}
				on:click={() => {
					setBtn('all');
				}}>All</button
			>
			<button
				class:active={activeBtn === 'success'}
				on:click={() => {
					setBtn('success');
				}}>Success</button
			>
			<button
				class:bad-active={activeBtn === 'client'}
				on:click={() => {
					setBtn('client');
				}}>Client</button
			>
			<button
				class:error-active={activeBtn === 'server'}
				on:click={() => {
					setBtn('server');
				}}>Server</button
			>
		</div>
	</div>

	{#if endpoints != undefined}
		<div class="endpoints">
			{#each endpoints as endpoint, i}
				<div class="endpoint-container">
					<button
						class="endpoint"
						id="endpoint-{i}"
						title="Status: {endpoint.status}"
						on:click={() =>
							setTargetEndpoint(
								endpoint.path.split(' ')[2],
								endpoint.status,
							)}
					>
						<div class="path">
							<span class="font-semibold">{endpoint.count.toLocaleString()}</span>
							{endpoint.path}
						</div>
						<div
							class="background"
							style="width: {(endpoint.count / maxCount) * 100}%"
							class:success={(endpoint.status >= 200 &&
								endpoint.status <= 299) ||
								endpoint.status === 0}
							class:bad={endpoint.status >= 400 &&
								endpoint.status <= 499}
							class:error={endpoint.status >= 500}
							class:other={endpoint.status >= 300 &&
								endpoint.status <= 399}
						></div>
					</button>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style scoped>
	.card {
		min-height: 361px;
	}
	.card-title {
		display: flex;
	}
	.toggle {
		margin-left: auto;
	}
	.toggle > button {
		font-size: 13.333px;
		color: #000;
		border: none;
		border-radius: 4px;
		background: rgb(68, 68, 68);
		cursor: pointer;
		padding: 1px 6px 0;
		margin-left: 5px;
	}

	.toggle > button:hover {
		background: rgb(88, 88, 88);
	}

	.toggle > .active {
		background: var(--highlight);
	}
	.toggle > .bad-active {
		background: rgb(235, 235, 129);
	}
	.toggle > .error-active {
		background: var(--red);
	}

	.toggle > .active:hover {
		background: var(--highlight);
	}
	.toggle > .bad-active:hover {
		background: rgb(235, 235, 129);
	}
	.toggle > .error-active:hover {
		background: var(--red);
	}
	.endpoints {
		margin: 0.9em 20px 0.6em;
	}
	.endpoint {
		border-radius: 3px;
		margin: 5px 0;
		color: var(--light-background);
		text-align: left;
		position: relative;
		font-size: 0.85em;
		width: 100%;
		cursor: pointer;

	}
	.endpoint:hover {
		background: linear-gradient(270deg, transparent, var(--background));
		background: linear-gradient(270deg, transparent, #444);
	}
	.path {
		position: relative;
		flex-grow: 1;
		z-index: 1;
		pointer-events: none;
		color: #505050;
		padding: 3px 12px;
		overflow-wrap: break-word;

		font-family: 'Noto Sans' !important;
	}
	.endpoint-container {
		display: flex;
	}
	.success {
		background: var(--highlight);
	}
	.bad {
		background: rgb(235, 235, 129);
	}
	.error {
		background: var(--red);
	}
	.other {
		background: rgb(241, 164, 20);
	}
	.cancel {
		display: none;
	}
	.toggle > .cancel {
		background: gold;
	}
	.toggle > .cancel:hover {
		background: gold;
	}
	.visible {
		display: inline;
	}
	.background {
		border-radius: 3px;
		color: var(--light-background);
		text-align: left;
		position: relative;
		font-size: 0.85em;
		height: 100%;
		position: absolute;
		top: 0;
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
