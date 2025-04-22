<script lang="ts">
	import type { Period } from '$lib/period';
	import ActivityRequests from './ActivityRequests.svelte';
	import ActivityResponseTime from './ActivityResponseTime.svelte';
	import ActivitySuccessRate from './ActivitySuccessRate.svelte';

	export let data: RequestsData, period: Period;
</script>

<div class="card !m-0 !w-full">
	<div class="card-title">Activity</div>
	<ActivityRequests {data} {period} />
	<ActivityResponseTime {data} {period} />
	<ActivitySuccessRate {data} {period} />
</div>
