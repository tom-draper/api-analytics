<script lang="ts">
	import type { Endpoint } from '$lib/endpoints';

	export let endpoints: Endpoint[];
	export let maxCount: number;
	
	// In Svelte 5, we define events directly in the component props
	export let selectEndpoint: (e: CustomEvent<{ path: string | null; status: number | null }>) => void;

	function handleSelect(path: string, status: number): void {
		selectEndpoint?.(new CustomEvent('selectEndpoint', { 
			detail: { path, status } 
		}));
	}
</script>

<div class="endpoints">
	{#each endpoints as endpoint, i}
		<div class="endpoint-container flex">
			<button
				class="endpoint"
				id="endpoint-{i}"
				title="Status: {endpoint.status}"
				on:click={() => handleSelect(endpoint.path.split(' ')[2], endpoint.status)}
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

<style>
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
	
	.success {
		background: var(--highlight);
	}
	
	.redirect {
		background: #4598ff;
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
</style>