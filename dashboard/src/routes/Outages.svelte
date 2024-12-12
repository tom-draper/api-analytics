<script lang="ts">
	import Footer from '../components/Footer.svelte';
</script>

<div class="info-page-container">
	<h1>Outages</h1>

	<p>
		Outages may be reflected in your dashboard. Make sure to check this page
		if you see unexpected downtime.
	</p>

	<table>
		<thead>
			<tr>
				<th>Start</th>
				<th>End</th>
				<th>Reason</th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td>2021-09-17 20:30:00 UTC</td>
				<td>2021-09-18 09:30:00 UTC</td>
				<td>Emergency maintenance by hosting provider.</td>
			</tr>
			<tr>

			</tr>
		</tbody
		>
	</table>
</div>
<Footer />

<style scoped>
	.info-page-container {
		max-width: 80% !important;
		margin: auto;
		padding-bottom: 5em;
	}
	h1 {
		margin: 1.2em 0 1em !important;
	}

	table {
		width: 100%;
		border-collapse: collapse;
	}
	th {
		background: var(--background);
	}
	td,
	th {
		border: 1px solid #2e2e2e;
		padding: 0.3em 0.5em;
	}
	td {
		color: #c3c3c3;
	}
	p {
		margin: 0;
		padding: 0 0em 3em;
		max-width: 100%;
	}
</style>
