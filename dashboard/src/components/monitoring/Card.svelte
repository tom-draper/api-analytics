<script lang="ts">
	import ResponseTime from './ResponseTime.svelte';
	import { periodToMarkers } from '../../lib/period';
	import type { NotificationState } from '../../lib/notification';
	import { getServerURL } from '../../lib/url';

	function triggerNotificationMessage(
		message: string,
		style: 'error' | 'warn' | 'success' = 'error',
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
			const url = getServerURL();
			const response = await fetch(
				`${url}/api/monitor/delete`,
				{
					method: 'POST',
					headers: {},
					body: JSON.stringify({
						user_id: userID,
						url: url,
						status: 0,
					}),
				},
			);
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

	function setUptime() {
		let success = 0;
		let total = 0;
		for (let i = 0; i < samples.length; i++) {
			if (samples[i].label === 'no-request') {
				continue;
			}
			if (
				samples[i].label === 'success' ||
				samples[i].label === 'delay'
			) {
				success++;
			}
			total++;
		}

		if (total === 0) {
			uptime = '0';
		} else {
			const per = (success / total) * 100;
			// If 100% display without decimal
			uptime = per === 100 ? '100' : per.toFixed(2);
		}
	}

	function periodSample(): MonitorSample[] {
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

	function setSamples() {
		const markers = periodToMarkers(period);
		samples = Array(markers).fill({
			label: 'no-request',
			responseTime: 0,
			status: 0,
			createdAt: null,
		});
		const sampledData = periodSample();
		const start = markers - sampledData.length;

		for (let i = 0; i < sampledData.length; i++) {
			samples[i + start] = {
				label: 'no-request',
				status: sampledData[i].status,
				responseTime: sampledData[i]['response_time'],
				createdAt: new Date(sampledData[i]['created_at']),
			};
			if (sampledData[i].status >= 200 && sampledData[i].status <= 299) {
				samples[i + start].label = 'success';
			} else if (sampledData[i].status !== null) {
				samples[i + start].label = 'error';
			}
		}
	}

	function setCurrentStatus() {
		const latest = samples[samples.length - 1];
		if (latest.label === null || latest.label === 'no-request') {
			currentStatus = 'no-request';
		} else if (latest.label === 'error') {
			currentStatus = 'error';
		} else if (latest.label == 'success') {
			currentStatus = 'success';
		}
		anyError = anyError || currentStatus === 'error';
	}

	function separateURL() {
		if (url.startsWith('https://')) {
			separatedURL.prefix = 'https://';
			separatedURL.body = url.replace('https://', '');
		} else if (url.startsWith('http://')) {
			separatedURL.prefix = 'http://';
			separatedURL.body = url.replace('http://', '');
		}
	}

	function build() {
		separateURL();
		setSamples();
		setCurrentStatus();
		setUptime();
	}

	// Monitor sample with label for status colour CSS class
	type Sample = MonitorSample & { label: string };

	let uptime = '';
	let currentStatus: 'success' | 'error' | 'no-request' = 'no-request';
	let samples: Sample[];
	const separatedURL = {
		prefix: '',
		body: '',
	};

	// If card period or url changes at any time, rebuild
	$: (period || url) && build();

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
					<div class="indicator grey-light" />
				{:else if currentStatus === 'error'}
					<div class="indicator red-light" />
				{:else}
					<div class="indicator green-light" />
				{/if}
			</div>
			<a href="{separatedURL.prefix}{separatedURL.body}" class="endpoint"
				><span style="color: var(--dim-text)"
					>{separatedURL.prefix}</span
				>{separatedURL.body}</a
			>
			<button class="delete" on:click={deleteMonitor}
				><img class="bin-icon" src="../img/icons/bin.png" alt="" /></button
			>
		</div>
		<div class="card-text-right">
			<div class="uptime">
				Uptime: {currentStatus === 'no-request' ? 'N/A' : `${uptime}%`}
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
				/>
			{/each}
		</div>
		<div class="response-time">
			<ResponseTime bind:data={samples} {period} />
		</div>
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
		background: rgb(228, 98, 98);
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
		padding: 2px 4px 1px;
		border-radius: 4px;
	}
	.delete:hover {
		background: var(--red);
	}
	.bin-icon {
		width: 12px;
		filter: invert(0.3);
	}
</style>
