<script lang="ts">
	import Metric from './Metric.svelte';

	function computeResiliance(data: RequestsData) {
		const successRate = getSuccessRate(data);
		return successRate;
	}

	function computePerformance(data: RequestsData) {
		const successRate = getSuccessRate(data);
		return successRate;
	}

	function computeAdoption(data: RequestsData) {
		const successRate = getSuccessRate(data);
		return successRate;
	}

	function computeOverall(
		data: RequestsData,
		resiliance: number,
		performance: number,
		adoption: number
	) {
		const resilianceWeight = 0.3;
		const performanceWeight = 0.3;
		const adoptionWeight = 0.4;
		const overall =
			resiliance * resilianceWeight + performance * performanceWeight + adoption * adoptionWeight;
		return overall;
	}

	function getSuccessRate(data: RequestsData) {
		return 75;
	}

	let resiliance = 0;
	let performance = 0;
	let adoption = 0;
	let overall = 0;

	$: if (data) {
		resiliance = computeResiliance(data);
		performance = computePerformance(data);
		adoption = computeAdoption(data);
		overall = computeOverall(data, resiliance, performance, adoption);
	}

	export let data: RequestsData;
</script>

<div class="card">
	<h2 class="card-title">Health</h2>
	<div class="health-container flex flex-row px-4 pb-4">
		<div class="health-item grid flex-1 place-items-center">
			<Metric label="Resiliance" value={resiliance} />
		</div>
		<div class="health-item grid flex-1 place-items-center">
			<Metric label="Performance" value={performance} />
		</div>
		<div class="health-item grid flex-1 place-items-center">
			<Metric label="Adoption" value={adoption} />
		</div>
		<div class="health-item grid flex-1 place-items-center">
			<Metric label="Overall" value={overall} />
		</div>
	</div>
</div>

<style scoped>
	.card {
		width: 100%;
		margin-top: 0;
		min-height: 238px;
	}
</style>
