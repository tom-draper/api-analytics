<script lang="ts">
	import Footer from '../components/Footer.svelte';

</script>

<div class="outages-container">
	<h1>Outages</h1>

	<table>
		<thead>
			<tr>
				<th>Date</th>
				<th>Time</th>
				<th>Duration</th>
				<th>Reason</th>
			</tr>
	</table>
</div>
<Footer />

<style scoped>
	.outages-container {
		max-width: 800px;
		margin: auto;
		padding: 4em 2em;
		text-align: left;
		color: white;
		font-family: 'Noto Sans' sans-serif;
	}
	h1 {
		margin: 1.2em 0 !important;
	}
</style>
