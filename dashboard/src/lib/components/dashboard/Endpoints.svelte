<script lang="ts">
	import { replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import { ColumnIndex, methodMap } from '$lib/consts';

	type EndpointFreq = Map<string, { path: string; status: number; count: number }>;

	type MapValue<A> = A extends Map<unknown, infer V> ? V : never;

	function endpointFreq(data: RequestsData) {
		const freq: EndpointFreq = new Map();
		for (const row of data) {
			// Create groups of endpoints by path + status
			const path = ignoreParams ? row[ColumnIndex.Path].split('?')[0] : row[ColumnIndex.Path];
			const status = row[ColumnIndex.Status];
			const endpointID = `${path}${status}`;
			let endpoint = freq.get(endpointID);
			if (!endpoint) {
				const method = methodMap[row[ColumnIndex.Method]];
				endpoint = {
					path: `${method}  ${path}`,
					status: status,
					count: 0
				};
				freq.set(endpointID, endpoint);
			}
			endpoint.count++;
		}

		return freq;
	}

	function statusMatch(status: number) {
		return (
			activeFilter === 'all' ||
			(activeFilter === 'success' && status >= 200 && status <= 299) ||
			(activeFilter === 'redirect' && status >= 300 && status <= 399) ||
			(activeFilter === 'client' && status >= 400 && status <= 499) ||
			(activeFilter === 'server' && status >= 500)
		);
	}

	function removeEndpointParams() {
		page.url.searchParams.delete('path');
		page.url.searchParams.delete('status');
		replaceState(page.url, page.state);
	}

	function setPathParam(path: string | null) {
		if (path === null) {
			page.url.searchParams.delete('path');
		} else {
			page.url.searchParams.set('path', path);
		}
		replaceState(page.url, page.state);
	}

	function setStatusParam(status: number | null) {
		if (status === null) {
			page.url.searchParams.delete('status');
		} else {
			page.url.searchParams.set('status', status.toString());
		}
		replaceState(page.url, page.state);
	}

	function setTargetEndpoint(path: string | null, status: number | null) {
		if (path === null || status === null) {
			// Trigger reset if input is null
			targetPath = null;
			targetStatus = null;
			removeEndpointParams();
		} else if (targetPath === null) {
			// At starting state, set the path first
			targetPath = path;
			setPathParam(path);
		} else if (endpoints.length > 1 && targetStatus === null) {
			// Path already set, now narrow down status (if multiple endpoints still exist)
			targetStatus = status;
			setStatusParam(status);
		} else {
			// Path and status already set, reset
			targetPath = null;
			targetStatus = null;
			removeEndpointParams();
		}
	}

	function getEndpoints(data: RequestsData) {
		const freq = endpointFreq(data);

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
			maxCount
		};
	}

	function setFilter(value: EndpointFilter) {
		activeFilter = value;
	}

	let endpoints: {
		path: string;
		status: number;
		count: number;
	}[];
	let maxCount: number;

	type EndpointFilter = 'all' | 'redirect' | 'success' | 'client' | 'server';

	let activeFilter: EndpointFilter = 'all';

	$: if (data && activeFilter) {
		({ endpoints, maxCount } = getEndpoints(data));
	}

	export let data: RequestsData,
		targetPath: string | null,
		targetStatus: number | null,
		ignoreParams: boolean;
</script>

<div class="card">
	<div class="card-title flex">
		Endpoints
		<div class="toggle ml-auto">
			<button
				class:active={activeFilter === 'all'}
				onclick={() => {
					setFilter('all');
				}}>All</button
			>
			<button
				class:active={activeFilter === 'success'}
				onclick={() => {
					setFilter('success');
				}}>Success</button
			>
			<button
				class:redirect-active={activeFilter === 'redirect'}
				onclick={() => {
					setFilter('redirect');
				}}>Redirect</button
			>
			<button
				class:bad-active={activeFilter === 'client'}
				onclick={() => {
					setFilter('client');
				}}>Client</button
			>
			<button
				class:error-active={activeFilter === 'server'}
				onclick={() => {
					setFilter('server');
				}}>Server</button
			>
		</div>
	</div>

	{#if targetPath !== null}
		<div class="flex px-[25px] pt-3 mb-[-8px] text-[13px] text-[#707070]">
			<div class="flex-grow text-left mr-3">
				<div class="">
					{#if targetStatus === null}
						{targetPath}
					{:else}
						{targetPath}, status: {targetStatus}
					{/if}
				</div>
			</div>
			<button
				class="hover:text-[#ededed]"
				onclick={() => {
					setTargetEndpoint(null, null);
				}}>Clear</button
			>
		</div>
	{/if}

	{#if endpoints != undefined}
		<div class="endpoints">
			{#each endpoints as endpoint, i}
				<div class="endpoint-container flex">
					<button
						class="endpoint"
						id="endpoint-{i}"
						title="Status: {endpoint.status}"
						onclick={() => setTargetEndpoint(endpoint.path.split(' ')[2], endpoint.status)}
					>
						<div class="path">
							<span class="font-semibold">{endpoint.count.toLocaleString()}</span>
							{endpoint.path}
						</div>
						<div
							class="background"
							style="width: {(endpoint.count / maxCount) * 100}%"
							class:success={endpoint.status >= 200 && endpoint.status <= 299}
							class:redirect={endpoint.status >= 300 && endpoint.status <= 399}
							class:bad={endpoint.status >= 400 && endpoint.status <= 499}
							class:error={endpoint.status >= 500}
							class:other={endpoint.status <= 100}
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

	.success,
	.toggle > .active,
	.toggle > .active:hover {
		background: var(--highlight);
	}
	.redirect,
	.toggle > .redirect-active,
	.toggle > .redirect-active:hover {
		background: #4c9cff;
		background: #4598ff;
	}
	.bad,
	.toggle > .bad-active,
	.toggle > .bad-active:hover {
		background: rgb(235, 235, 129);
	}
	.error,
	.toggle > .error-active,
	.toggle > .error-active:hover {
		background: var(--red);
	}
	.other {
		background: rgb(241, 164, 20);
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
