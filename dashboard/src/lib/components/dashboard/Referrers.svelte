<script lang="ts">
	import { toggleParam } from '$lib/params';
	import type { ReferrerBar } from '$lib/aggregate';
	import BarList from './BarList.svelte';

	let { referrerBars, targetReferrer = $bindable<string | null>(null) }: {
		referrerBars: ReferrerBar[];
		targetReferrer: string | null;
	} = $props();

	const rows = $derived(
		referrerBars.map((bar) => ({
			value: bar.referrer,
			label: bar.referrer,
			count: bar.count,
			height: bar.height,
			selected: targetReferrer === bar.referrer
		}))
	);

	function select(referrer: string) {
		toggleParam('referrer', referrer, targetReferrer, (v) => (targetReferrer = v));
	}
</script>

<BarList title="Referrer" {rows} onSelect={select} minHeight="300px" />
