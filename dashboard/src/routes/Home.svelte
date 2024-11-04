<script lang="ts">
	import { onMount } from 'svelte';
	import Landing from '../components/home/Landing.svelte';
	import Highlight from '../components/home/Highlight.svelte';
	import GettingStarted from '../components/home/GettingStarted.svelte';
	import DocLinks from '../components/home/DocLinks.svelte';
	import Footer from '../components/Footer.svelte';

	let queryString: string = '';

	onMount(() => {
		// Retrieve current URL parameters
		const params = new URLSearchParams(window.location.search);
		queryString = params.toString();
	});
</script>

<div class="home">
	<Landing />
	<Highlight
		title="Dashboard"
		description="An all-in-one analytics dashboard. Real-time insight into your API's usage."
		path={`/dashboard${queryString ? `?${queryString}` : ''}`}
		img="dashboard.png"
	/>
	<Highlight
		title="Monitor"
		description="Active monitoring and error notifications. Peace of mind."
		path={`/monitoring${queryString ? `?${queryString}` : ''}`}
		img="monitoring.png"
	/>
	<GettingStarted />
	<DocLinks />
</div>
<Footer />

