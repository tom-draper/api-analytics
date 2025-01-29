<script lang="ts">
	import '../app.css';
	import Footer from '$lib/components/Footer.svelte';
	let { children } = $props();
</script>

{@render children()}

<svelte:head>
  <script src="https://cdn.plot.ly/plotly-latest.min.js" type="text/javascript" defer></script>
</svelte:head>

<footer>
	<Footer generic={true}/>
</footer>
