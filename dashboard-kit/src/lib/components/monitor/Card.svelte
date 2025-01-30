<script lang="ts">
	import ResponseTime from './ResponseTime.svelte';
	import { periodToMarkers } from '$lib/period';
	import type { NotificationState } from '$lib/notification';
	import { getServerURL } from '$lib/url';

	function triggerNotificationMessage(
		message: string,
		style: 'error' | 'warn' | 'success' = 'error'
	) {
		notification.message = message;
		notification.style = style;
		notification.show = true;
		setTimeout(() => {
			notification.show = false;
		}, 4000);
	}

	async function deleteMonitor() {
		try {
			const serverURL = getServerURL();
			const response = await fetch(`${serverURL}/api/monitor/delete`, {
				method: 'POST',
				headers: {},
				body: JSON.stringify({
					user_id: userID,
					status: 0,
					url,
				})
			});
			if (response.status === 201) {
				triggerNotificationMessage('Deleted successfully', 'success');
				removeMonitor(url);
			} else {
				triggerNotificationMessage('Failed to delete monitor');
			}
		} catch (e) {
			console.log(e);
			triggerNotificationMessage('Failed to delete monitor');
		}
	}

	function getUptime(samples: Sample[]) {
		let success = 0;
		let total = 0;
		for (let i = 0; i < samples.length; i++) {
			if (samples[i].label === 'no-request') {
				continue;
			}
			if (samples[i].label === 'success' || samples[i].label === 'delay') {
				success++;
			}
			total++;
		}

		if (total === 0) {
			return 'N/A';
		} 

		const per = (success / total) * 100;
		// If 100% display without decimal
		return per === 100 ? '100' : per.toFixed(2);
	}

	function periodSample(period: string): MonitorSample[] {
		/* Sample ping recordings at regular intervals if number of bars fewer than 
		total recordings the current period length */
		let sample: MonitorSample[] = [];
		switch (period) {
			case '30d':
				// Sample 1 in 4
				for (let i = 0; i < data[url].length; i++) {
					if (i % 4 === 0) {
						sample.push(data[url][i]);
					}
				}
				break;
			case '60d':
				// Sample 1 in 8
				for (let i = 0; i < data[url].length; i++) {
					if (i % 8 === 0) {
						sample.push(data[url][i]);
					}
				}
				break;
			default:
				sample = data[url];
		}
		// Ensure final sample is the most recent sample in the data for system down visual highlight
		sample[sample.length - 1] = data[url][data[url].length - 1];
		return sample;
	}

	function getSamples(period: string) {
		const markers = periodToMarkers(period);
		const samples: Sample[] = Array.from({ length: markers }).fill({
			label: 'no-request',
			responseTime: 0,
			status: 0,
			createdAt: null
		});

		if (!markers) {
			return samples;
		}

		const sampledData = periodSample(period);
		const start = markers - sampledData.length;

		for (let i = 0; i < sampledData.length; i++) {
			samples[i + start] = {
				label: 'no-request',
				status: sampledData[i].status,
				responseTime: sampledData[i]['response_time'],
				createdAt: new Date(sampledData[i]['created_at'])
			};
			if (sampledData[i].status >= 200 && sampledData[i].status <= 299) {
				samples[i + start].label = 'success';
			} else if (sampledData[i].status !== null) {
				samples[i + start].label = 'error';
			}
		}

		samples[samples.length-1].label = 'error'

		return samples;
	}

	function getCurrentStatus(samples: Sample[]) {
		const latest = samples[samples.length - 1];
		if (latest.label === null || latest.label === 'no-request') {
			return 'no-request';
		} else if (latest.label === 'error') {
			return 'error';
		} else if (latest.label == 'success') {
			return 'success';
		}
	}

	function separateURL(url: string) {
		let prefix: string = '';
		let body: string = '';
		if (url.startsWith('https://')) {
			prefix = 'https://';
			body = url.replace('https://', '');
		} else if (url.startsWith('http://')) {
			prefix = 'http://';
			body = url.replace('http://', '');
		}
		return { prefix, body };
	}

	// Monitor sample with label for status colour CSS class
	type Sample = MonitorSample & { label: string };

	let uptime = '';
	let currentStatus: 'success' | 'error' | 'no-request' = 'no-request';
	let samples: Sample[];
	let separatedURL = {
		prefix: '',
		body: ''
	};

	// If card period or url changes at any time, rebuild
	$: {
		separatedURL = separateURL(url);
		samples = getSamples(period);
		currentStatus = getCurrentStatus(samples) || currentStatus;
		anyError = anyError || currentStatus === 'error'
		uptime = getUptime(samples);
	}

	export let url: string,
		data: MonitorData,
		userID: string,
		period: string,
		anyError: boolean,
		notification: NotificationState,
		removeMonitor: (url: string) => void;
</script>

<div class="card" class:card-error={currentStatus === 'error'}>
	<div class="card-text">
		<div class="card-text-left">
			<div class="card-status">
				{#if currentStatus === 'no-request'}
					<div class="indicator grey-light"></div>
				{:else if currentStatus === 'error'}
					<div class="indicator red-light"></div>
				{:else}
					<div class="indicator green-light"></div>
				{/if}
			</div>
			<a href="{separatedURL.prefix}{separatedURL.body}" class="endpoint"
				><span class="text-[var(--dim-text)]">{separatedURL.prefix}</span>{separatedURL.body}</a
			>
			<button class="delete" on:click={deleteMonitor}>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="1.5"
					stroke="currentColor"
					class="size-6"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"
					/>
				</svg>
			</button>
		</div>
		<div class="card-text-right">
			<div class="uptime">
				Uptime: {currentStatus === 'no-request' ? 'Pending' : `${uptime}%`}
			</div>
		</div>
	</div>
	{#if samples !== undefined}
		<div class="measurements">
			{#each samples as sample}
				<div
					class="measurement {sample.label}"
					title={sample.createdAt === null
						? ''
						: sample.status === 0
							? `No response\n${sample.createdAt.toLocaleString()}`
							: `Status: ${sample.status}\n${sample.createdAt.toLocaleString()}`}
				></div>
			{/each}
		</div>
		{#if samples[samples.length - 1].label !== 'no-request'}
			<div class="response-time">
				<ResponseTime bind:data={samples} {period} />
			</div>
		{/if}
	{/if}
</div>

<style scoped>
	.card {
		width: min(100%, 1000px);
		border: 1px solid #2e2e2e;
		margin: 2.2em auto;
	}
	.card-error {
		box-shadow:
			rgba(228, 98, 98, 0.5) 0px 15px 110px 0px,
			rgba(0, 0, 0, 0.4) 0px 30px 60px -30px;
		border: 2px solid rgba(228, 98, 98, 1);
	}
	.card-text {
		display: flex;
		margin: 2em 2em 0;
		font-size: 0.9em;
	}
	.card-text-left {
		flex-grow: 1;
		display: flex;
	}
	.endpoint {
		margin-left: 10px;
		letter-spacing: 0.01em;
		color: white;
	}
	.measurements {
		display: flex;
		padding: 1em 2em 2em;
	}
	.measurement {
		margin: 0 0.1%;
		flex: 1;
		height: 3em;
		border-radius: 1px;
		background: var(--highlight);
		background: rgb(40, 40, 40);
	}
	.success {
		background: var(--highlight);
	}
	.delayed {
		background: rgb(199, 229, 125);
	}
	.error {
		background: var(--red);
	}
	.no-request {
		color: var(--dim-text);
	}
	.uptime {
		color: var(--dim-text);
	}
	.indicator {
		width: 10px;
		height: 10px;
		border-radius: 5px;
		margin-right: 5px;
		margin-bottom: 1px;
	}
	.delete {
		aspect-ratio: 1/1;
		color: var(--dim-text);
	}
	.delete > svg {
		width: 16px;
		height: 16px;
	}
	.card-status {
		display: grid;
		place-items: center;
	}
	.green-light {
		background: var(--highlight);
		box-shadow:
			0 1px 1px #fff,
			0 0 6px 3px var(--highlight);
	}
	.red-light {
		background: var(--red);
		box-shadow:
			0 1px 1px #fff,
			0 0 6px 3px var(--red);
	}
	.grey-light {
		background: grey;
		box-shadow: 0 0 1px 1px #fff;
	}
	.delete {
		background: transparent;
		border: none;
		cursor: pointer;
		margin-left: 10px;
		padding: 2px 4px;
		border-radius: 4px;
	}
	.delete:hover {
		background: var(--red);
		color: var(--background);
	}
	.bin-icon {
		width: 12px;
		filter: invert(0.3);
	}
</style>
