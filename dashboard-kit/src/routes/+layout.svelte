<script lang="ts">
	import '../app.css';
	import Footer from '$components/Footer.svelte';
	import { page } from '$app/state';
	let { children } = $props();

	const currentRoute = $derived(page.route.id);
</script>

{@render children()}

<svelte:head>
	<script src="https://cdn.plot.ly/plotly-latest.min.js" type="text/javascript" defer></script>
</svelte:head>

{#if currentRoute !== '/explorer/[uuid]'}
	<footer>
		<Footer
			generic={currentRoute !== '/generate' &&
				currentRoute !== '/dashboard' &&
				currentRoute !== '/monitor'}
		/>
	</footer>
{/if}
