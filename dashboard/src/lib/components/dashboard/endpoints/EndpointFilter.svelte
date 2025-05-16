<script lang="ts">
	import type { EndpointFilterType } from '$lib/endpoints';

	export let activeFilter: EndpointFilterType;
	
	// In Svelte 5, we define events directly in the component props
	export let filterChange: (e: CustomEvent<EndpointFilterType>) => void;
	
	function setFilter(value: EndpointFilterType): void {
		filterChange?.(new CustomEvent('filterChange', { detail: value }));
	}
</script>

<div class="toggle">
	<button
		class:active={activeFilter === 'all'}
		on:click={() => setFilter('all')}
	>
		All
	</button>
	<button
		class:active={activeFilter === 'success'}
		on:click={() => setFilter('success')}
	>
		Success
	</button>
	<button
		class:redirect-active={activeFilter === 'redirect'}
		on:click={() => setFilter('redirect')}
	>
		Redirect
	</button>
	<button
		class:bad-active={activeFilter === 'client'}
		on:click={() => setFilter('client')}
	>
		Client
	</button>
	<button
		class:error-active={activeFilter === 'server'}
		on:click={() => setFilter('server')}
	>
		Server
	</button>
</div>

<style>
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

	.toggle > .active,
	.toggle > .active:hover {
		background: var(--highlight);
	}
	
	.toggle > .redirect-active,
	.toggle > .redirect-active:hover {
		background: #4598ff;
	}
	
	.toggle > .bad-active,
	.toggle > .bad-active:hover {
		background: rgb(235, 235, 129);
	}
	
	.toggle > .error-active,
	.toggle > .error-active:hover {
		background: var(--red);
	}
</style>