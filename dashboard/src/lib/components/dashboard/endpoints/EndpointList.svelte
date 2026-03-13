<script lang="ts">
	import type { Endpoint } from '$lib/endpoints';
	import { statusSuccess, statusRedirect, statusBad, statusError } from '$lib/status';

	function handleSelect(path: string, status: number): void {
		selectEndpoint?.(path, status);
	}

	let { endpoints, maxCount, selectEndpoint }: { endpoints: Endpoint[]; maxCount: number; selectEndpoint?: (path: string | null, status: number | null) => void } = $props();
</script>

<div class="endpoints">
	{#each endpoints as endpoint, i}
		<div class="endpoint-container flex">
			<button
				class="endpoint"
				id="endpoint-{i}"
				title="Status: {endpoint.status}"
				onclick={() => handleSelect(endpoint.path.split(' ')[2], endpoint.status)}
			>
				<div class="path">
					<span class="font-semibold">{endpoint.count.toLocaleString()}</span>
					{endpoint.path}
				</div>
				<div
					class="background"
					style="width: {(endpoint.count / maxCount) * 100}%"
					class:success={statusSuccess(endpoint.status)}
					class:redirect={statusRedirect(endpoint.status)}
					class:bad={statusBad(endpoint.status)}
					class:error={statusError(endpoint.status)}
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
		border-radius: var(--radius-sm);
		margin: 5px 0;
		color: var(--light-background);
		text-align: left;
		position: relative;
		font-size: 0.85em;
		width: 100%;
		cursor: pointer;
	}

	.endpoint:hover {
		background: var(--fade-right);
	}

	.path {
		position: relative;
		flex-grow: 1;
		z-index: 1;
		pointer-events: none;
		color: var(--muted-text);
		padding: 3px 12px;
		overflow-wrap: break-word;
	}

	.background {
		border-radius: var(--radius-sm);
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
		background: var(--redirect-color);
	}

	.bad {
		background: var(--yellow);
	}

	.error {
		background: var(--red);
	}

	.other {
		background: rgb(241, 164, 20);
	}
</style>
