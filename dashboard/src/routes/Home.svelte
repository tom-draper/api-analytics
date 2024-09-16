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
		description="An all-in-one analytics dashboard. Real-time insight into your API's usage."
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
	<div class="docs">
		<!-- <div class="doc-link-container">
			<a class="doc-link" href="/privacy-policy">FAQ</a>
		</div> -->
		<div class="doc-link-container">
			<a class="doc-link" href="/privacy-policy">Privacy Policy</a>
		</div>
		<!-- <div class="doc-link-container">
			<a class="doc-link" href="/privacy-policy">Terms of Service</a>
		</div> -->
	</div>
</div>
<Footer />

<style scoped>
	.docs {
		display: flex;
		width: 50%;
		margin: auto;
		padding-bottom: 4em;
	}
	.doc-link-container {
		flex: 1;
		display: grid; 
		place-items: center;
	}
	.doc-link {
		flex: 1;
		color: #5f5f5f;
		background: #181818;
		border-radius: 4px;
		min-width: 10em;
		max-width: 15em;
		margin: auto;
		padding: 0.8em 1.5em;
		display: flex;
		margin: 0 1em;
		text-align: center;
		justify-content: center;
	}

	a:hover {
		background: #161616;
	}

	/* if under 800px */
	@media (max-width: 800px) {
		.docs {
			width: 100%;
		}
	}
</style>