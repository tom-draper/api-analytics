<script lang="ts">
	let dashOffset = 0;

	function updateChart(value: number) {
		const circumference = 2 * Math.PI * 44;
		dashOffset = circumference - (value / 100) * circumference; // Adjust by 2 units
	}

	$: updateChart(value);

	export let label: string, value: number;
</script>

<div class="chart-container p-2">
	<svg width="150" height="150" viewBox="0 0 100 100">
		<!-- Background Circle -->
		<circle cx="50" cy="50" r="44" stroke="#282828" stroke-width="12" fill="none" />

		<!-- Progress Circle -->
		<circle
			cx="50"
			cy="50"
			r="44"
			stroke="var(--highlight)"
			stroke-width="12"
			fill="none"
			stroke-linecap="square"
			stroke-dasharray={2 * Math.PI * 44}
			stroke-dashoffset={dashOffset}
			transform="rotate(-90 50 50)"
		/>

		<!-- Score Text -->
		<text x="50" y="50" text-anchor="middle" font-size="20" font-weight="bold" fill="#ededed"
			>{value}</text
		>

		<text x="50" y="62" text-anchor="middle" font-size="8" font-weight="600" fill="#707070"
			>{label}</text
		>
	</svg>
</div>

<style scoped>
	.chart-container {
		display: block;
		margin: auto;
	}
	svg {
		font-family: 'Noto Sans', sans-serif;
	}
</style>
