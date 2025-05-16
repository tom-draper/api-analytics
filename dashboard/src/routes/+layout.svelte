<script lang="ts">
	import '../app.css';
	import Footer from '$components/Footer.svelte';
	import { page } from '$app/state';
	let { children } = $props();

	const currentRoute = $derived(page.route.id);
</script>

{@render children()}

<svelte:head>
	<script src="https://cdn.plot.ly/plotly-1.58.5.min.js" type="text/javascript"></script>
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
