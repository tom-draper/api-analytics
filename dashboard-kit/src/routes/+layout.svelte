<script lang="ts">
	import '../app.css';
	import Footer from '$lib/components/Footer.svelte';
	import { page } from '$app/state';
	let { children } = $props();

	const currentRoute = $derived(page.route.id);
</script>

{@render children()}

<svelte:head>
	<script src="https://cdn.plot.ly/plotly-latest.min.js" type="text/javascript" defer></script>
</svelte:head>

<footer>
	<Footer
		generic={currentRoute !== '/generate' &&
			currentRoute !== '/dashboard' &&
			currentRoute !== '/monitor'}
	/>
</footer>
