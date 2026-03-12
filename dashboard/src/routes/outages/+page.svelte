<div class="info-page-container">
	<h1>Outages</h1>
	<p class="subtitle">
		Outages may be reflected in your dashboard. Check this page if you notice unexpected gaps in
		your data.
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
				<td>2024-09-17 20:30 UTC</td>
				<td>2024-09-18 09:30 UTC</td>
				<td>Emergency maintenance by hosting provider.</td>
			</tr>
			<tr>
				<td>2024-11-12 01:10 UTC</td>
				<td>2024-11-12 07:35 UTC</td>
				<td>Server resource limit exceeded.</td>
			</tr>
			<tr>
				<td>2025-08-12 01:30 UTC</td>
				<td>2025-08-12 11:05 UTC</td>
				<td>Server resource limit exceeded.</td>
			</tr>
		</tbody>
	</table>
</div>

<style scoped>
	.info-page-container {
		padding-bottom: 6em;
	}
	h1 {
		font-size: 2em;
		font-weight: 700;
		margin: 1.2em 0 0.4em;
	}
	.subtitle {
		color: var(--dim-text);
		font-size: 0.95em;
		padding: 0;
		margin-bottom: 2.5em;
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	th {
		text-align: left;
		font-size: 0.78em;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--dim-text);
		padding: 0 1em 0.8em;
		border-bottom: 1px solid var(--border);
	}
	td {
		padding: 1em;
		font-size: 0.9em;
		color: #c3c3c3;
		border-bottom: 1px solid #2a2a2a;
	}
	tr:last-child td {
		border-bottom: none;
	}
	td:first-child,
	td:nth-child(2) {
		color: var(--faded-text);
		font-family: 'Geist Mono', monospace;
		font-size: 0.82em;
		white-space: nowrap;
	}
</style>
