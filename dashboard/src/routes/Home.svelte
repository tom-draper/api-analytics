<script lang="ts">
	import Landing from '../components/home/Landing.svelte';
	import Highlight from '../components/home/Highlight.svelte';
	import GettingStarted from '../components/home/GettingStarted.svelte';
	import Footer from '../components/Footer.svelte';
</script>

<div class="home">
	<Landing />
	<Highlight
		title="Dashboard"
		description="A powerful all-in-one analytics dashboard. Real-time usage insight."
		path="/dashboard"
		img="dashboard.png"
	/>
	<Highlight
		title="Monitor"
		description="Active monitoring and error notifications. Peace of mind."
		path="/monitoring"
		img="monitoring.png"
	/>
	<GettingStarted />
</div>
<Footer />
