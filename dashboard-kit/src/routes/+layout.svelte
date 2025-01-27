<script lang="ts">
	import '../app.css';
	import Footer from '$lib/components/Footer.svelte';
	let { children } = $props();
</script>

{@render children()}

<Footer />