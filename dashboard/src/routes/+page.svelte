<script lang="ts">
	import Landing from '$components/home/Landing.svelte';
	import Highlight from '$components/home/Highlight.svelte';
	import GettingStarted from '$components/home/GettingStarted.svelte';
	import DocLinks from '$components/home/DocumentLinks.svelte';
	import NginxAnalytics from '$components/home/NginxAnalytics.svelte';
	import { page } from '$app/state';
	import { formatPath } from '$lib/path';
	// import Pricing from '$components/home/Pricing.svelte';

	let params: string;
	$: params = page.url.searchParams.toString();
</script>

<div class="home">
	<Landing />
	<Highlight
		title="Dashboard"
		description="A powerful all-in-one analytics dashboard. Real-time usage insight."
		path={formatPath('/dashboard', params)}
		img="dashboard.png"
	/>
	<!-- <Highlight
		title="Explorer"
		description="A deep-dive into your logged requests."
		path={formatPath('/explorer', params)}
		img="dashboard.png"
	/> -->
	<Highlight
		title="Monitor"
		description="Active monitoring and error notifications. Peace of mind."
		path={formatPath('/monitoring', params)}
		img="monitoring.png"
	/>
	<!-- <Pricing /> -->
	<GettingStarted />
	<NginxAnalytics />
	<DocLinks />
</div>
