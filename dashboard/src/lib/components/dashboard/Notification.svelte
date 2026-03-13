<script lang="ts">
	import type { NotificationState } from '$lib/notification';

	let { state }: { state: NotificationState } = $props();
</script>

<div
	class="notification"
	class:show={state.show}
	class:error={state.style === 'error'}
	class:warn={state.style === 'warn'}
	class:success={state.style === 'success'}
>
	<div>{state.message}</div>
</div>

<style scoped>
	.notification {
		font-size: 14px;
		position: fixed;
		bottom: 0;
		display: grid;
		place-items: center;
		z-index: 100;
		padding: 0.3em 2em;
		width: 100%;
		opacity: 0;
		transition: opacity 0.2s linear;
		font-weight: 500;
		user-select: none;
		filter: contrast(1.2);
	}

	.show {
		opacity: 1 !important;
	}

	.error, .warn, .success {
		color: var(--dark-background);
	}

	.error {
		background: var(--error-bg);
	}
	.warn {
		background: var(--warn-bg);
	}
	.success {
		background: var(--success-bg);
	}
</style>
