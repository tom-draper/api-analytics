<script lang="ts">
	import type { EndpointFilterType } from '$lib/endpoints';

	function setFilter(value: EndpointFilterType): void {
		filterChange?.(value);
	}

	let { activeFilter, filterChange }: { activeFilter: EndpointFilterType; filterChange?: (value: EndpointFilterType) => void } = $props();
</script>

<div class="toggle">
	<button class:active={activeFilter === 'all'} onclick={() => setFilter('all')}> All </button>
	<button class:active={activeFilter === 'success'} onclick={() => setFilter('success')}>
		Success
	</button>
	<button
		class:redirect-active={activeFilter === 'redirect'}
		onclick={() => setFilter('redirect')}
	>
		Redirect
	</button>
	<button class:bad-active={activeFilter === 'client'} onclick={() => setFilter('client')}>
		Client
	</button>
	<button class:error-active={activeFilter === 'server'} onclick={() => setFilter('server')}>
		Server
	</button>
</div>

<style>
	.toggle > button {
		font-size: 13.333px;
		color: #000;
		border: none;
		border-radius: var(--radius-md);
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
		background: var(--redirect-color);
	}

	.toggle > .bad-active,
	.toggle > .bad-active:hover {
		background: var(--yellow);
	}

	.toggle > .error-active,
	.toggle > .error-active:hover {
		background: var(--red);
	}
</style>
