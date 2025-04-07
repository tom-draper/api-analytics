<script lang="ts">
	import type { NotificationState } from '$lib/notification';
	import { getServerURL } from '$lib/url';
	import Dropdown from '$components/dashboard/Dropdown.svelte';

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

	function getFullURL(url: string, secure: boolean): string {
		url = url.replace(/^https?(:\/\/)?/, '');
		const prefix = secure ? 'https://' : 'http://';
		url = prefix + url;
		return url;
	}

	async function postMonitor() {
		if (monitorURL == null) {
			triggerNotificationMessage('Monitor URL is blank.');
			return;
		} else if (monitorCount >= monitorLimit) {
			triggerNotificationMessage('Monitor limit reached.');
			return;
		}

		const secure = urlPrefix === 'https';
		const serverURL = getServerURL();

		try {
			const response = await fetch(`${serverURL}/api/monitor/add`, {
				method: 'POST',
				body: JSON.stringify({
					user_id: userID,
					url: getFullURL(monitorURL, secure),
					ping: true,
					secure: secure
				})
			});
			if (response.status === 201) {
				triggerNotificationMessage(
					`Monitor ${monitorCount + 1}/${monitorLimit} created successfully`,
					'success'
				);
				const fullURL = getFullURL(monitorURL, secure);
				addEmptyMonitor(fullURL);
				showTrackNew = false; // Collapse controls for adding new monitor
			} else if (response.status === 409) {
				triggerNotificationMessage('Endpoint already monitored.', 'warn');
			} else {
				triggerNotificationMessage('Failed to create monitor.');
			}
		} catch (e) {
			console.log(e);
			triggerNotificationMessage('Failed to create monitor.');
		}
	}

	let monitorURL: string;
	const options = ['https', 'http'];
	let urlPrefix = options[0];

	const monitorLimit = 3;

	export let userID: string,
		showTrackNew: boolean,
		monitorCount: number,
		notification: NotificationState,
		addEmptyMonitor: (url: string) => void;
</script>

<div class="card">
	<div class="card-text">
		<div class="url">
			<div class="text-sm">
				<Dropdown {options} bind:selected={urlPrefix} defaultOption={null} />
			</div>
			<input
				type="text"
				placeholder="www.example.com/endpoint/"
				class="text-sm font-normal"
				bind:value={monitorURL}
			/>
			<button class="add" on:click={postMonitor}>Add</button>
		</div>
		<div class="detail">
			Endpoints are pinged by our servers every 30 mins and response <b>status</b>
			and response <b>time</b> are logged.
		</div>
	</div>
</div>

<style scoped>
	.card {
		width: min(100%, 1000px);
		border: 1px solid #2e2e2e;
		margin: 2.2em auto 4em;
	}
	.card-text {
		margin: 2em 2em 1.9em;
	}
	input {
		background: var(--background);
		border-radius: 4px;
		border: none;
		margin: 0 10px 0 8px;
		width: 100%;
		text-align: left;
		height: auto;
		padding: 4px 12px;
		font-family: 'Geist';
		border: 1px solid var(--background);
	}
	input::placeholder {
		color: var(--dim-text);
	}
	.url {
		display: flex;
	}
	.detail {
		margin-top: 30px;
		color: var(--dim-text);
		font-weight: 400;
		font-size: 0.85em;
	}
	button {
		border: none;
		border-radius: 4px;
		background: var(--light-background);
		cursor: pointer;
		font-size: 0.85em;
		color: var(--background);
	}
	.add {
		background: var(--highlight);
		padding: 4px 20px;
		margin: 0;
	}
</style>
