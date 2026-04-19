<script lang="ts">
	import type { Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';
	import ActivityRequests from './ActivityRequests.svelte';
	import ActivityResponseTimes from './ActivityResponseTimes.svelte';
	import ActivitySuccessRate from './ActivitySuccessRate.svelte';

	let { activityBuckets, period }: {
		activityBuckets: ActivityBucket[];
		period: Period;
	} = $props();
</script>

<div class="card !m-0 !w-full">
	<div class="card-title">Activity</div>
	<ActivityRequests {activityBuckets} {period} />
	<ActivityResponseTimes {activityBuckets} {period} />
	<ActivitySuccessRate {activityBuckets} {period} />
</div>
