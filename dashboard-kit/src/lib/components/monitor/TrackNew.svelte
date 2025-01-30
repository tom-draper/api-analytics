<script lang="ts">
	import type { NotificationState } from '$lib/notification';
	import { getServerURL } from '$lib/url';
	import Dropdown from '$lib/components/dashboard/Dropdown.svelte';

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

	function getFullURL(url: string, secure: boolean): string {
		url = url.replace(/^https?(:\/\/)?/, '');
		const prefix = secure ? 'https://' : 'http://';
		url = prefix + url;
		return url;
	}

	async function postMonitor() {
		if (monitorURL == null) {
			triggerNotificationMessage('URL is blank.');
			return;
		} else if (monitorCount >= 3) {
			triggerNotificationMessage('Maximum 3 monitors allowed.');
			return;
		}

		try {
			const secure = urlPrefix === 'https';
			const fullURL = getFullURL(monitorURL, secure);
			const serverURL = getServerURL();
			const response = await fetch(`${serverURL}/api/monitor/add`, {
				method: 'POST',
				headers: {},
				body: JSON.stringify({
					user_id: userID,
					url: getFullURL(monitorURL, secure),
					ping: true,
					secure: secure,
				}),
			});
			if (response.status === 201) {
				triggerNotificationMessage('Created successfully', 'success');
				addEmptyMonitor(fullURL);
				showTrackNew = false;
			} else if (response.status === 409) {
				triggerNotificationMessage('URL already monitored', 'warn');
			} else {
				triggerNotificationMessage('Failed to create monitor');
			}
		} catch (e) {
			console.log(e);
			triggerNotificationMessage('Failed to create monitor');
		}
	}

	let monitorURL: string;
	const options = ['https', 'http'];
	let urlPrefix = options[0];

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
				<Dropdown
					{options}
					bind:selected={urlPrefix}
					defaultOption={null}
				/>
			</div>
			<input
				type="text"
				placeholder="www.example.com/endpoint/"
				class="text-sm"
				bind:value={monitorURL}
			/>
			<button class="add" on:click={postMonitor}>Add</button>
		</div>
		<div class="detail">
			Endpoints are pinged by our servers every 30 mins and response <b
				>status</b
			>
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
		padding: 3px 12px;
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
		color: var(--background)
	}
	.add {
		background: var(--highlight);
		padding: 4px 20px;
		margin: 0;
	}
</style>
