<script lang="ts">
	import { onMount } from 'svelte';
	import Landing from '$lib/components/home/Landing.svelte';
	import Highlight from '$lib/components/home/Highlight.svelte';
	import GettingStarted from '$lib/components/home/GettingStarted.svelte';
	import DocLinks from '$lib/components/home/DocumentLinks.svelte';
	import { formatPath } from '$lib/path';

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
		description="A powerful all-in-one analytics dashboard. Real-time usage insight."
		path={formatPath('/dashboard', queryString)}
		img="dashboard.png"
	/>
	<!-- <Highlight
		title="Explorer"
		description="A deep-dive into your logged requests."
		path={`/dashboard${queryString ? `?${queryString}` : ''}`}
		img="dashboard.png"
	/>
	<Highlight
		title="Showcase"
		description="A public customisable dashboard. Show off your achivements."
		path={`/dashboard${queryString ? `?${queryString}` : ''}`}
		img="dashboard.png"
	/> -->
	<Highlight
		title="Monitor"
		description="Active monitoring and error notifications. Peace of mind."
		path={formatPath('/monitoring', queryString)}
		img="monitoring.png"
	/>
	<GettingStarted />
	<DocLinks />
</div>
