<script lang="ts">
	import Footer from '../components/Footer.svelte';

</script>

<div class="info-page-container">
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
	h1 {
		margin: 1.2em 0 !important;
	}
</style>
