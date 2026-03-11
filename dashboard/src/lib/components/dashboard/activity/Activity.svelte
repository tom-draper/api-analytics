<script lang="ts">
	import type { Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';
	import ActivityRequests from './ActivityRequests.svelte';
	import ActivityResponseTime from './ActivityResponseTime.svelte';
	import ActivitySuccessRate from './ActivitySuccessRate.svelte';

	let { activityBuckets, period, firstRequestDate }: {
		activityBuckets: ActivityBucket[];
		period: Period;
		firstRequestDate: Date | null;
	} = $props();
</script>

<div class="card !m-0 !w-full">
	<div class="card-title">Activity</div>
	<ActivityRequests {activityBuckets} {period} />
	<ActivityResponseTime {activityBuckets} {period} />
	<ActivitySuccessRate {activityBuckets} {period} {firstRequestDate} />
</div>
