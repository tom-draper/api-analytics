<script lang="ts">
	import '../app.css';
	import Footer from '$components/Footer.svelte';
	import { page } from '$app/state';
	let { children } = $props();

	const currentRoute = $derived(page.route.id);
</script>

{@render children()}

<svelte:head>
	<script src="https://cdn.plot.ly/plotly-3.4.0.min.js" type="text/javascript"></script>
</svelte:head>

{#if currentRoute !== '/explorer/[uuid]'}
	<footer>
		<Footer
			generic={currentRoute !== '/sign-up' &&
				currentRoute !== '/dashboard' &&
				currentRoute !== '/monitor' &&
				currentRoute !== '/explorer' &&
				currentRoute !== '/regenerate' &&
				currentRoute !== '/delete'}
		/>
	</footer>
{/if}
