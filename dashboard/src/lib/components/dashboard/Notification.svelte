<script lang="ts">
	import type { NotificationState } from '$lib/notification';

	export let state: NotificationState;
</script>

<div
	class="notification"
	class:show={state.show}
	class:error={state.style === 'error'}
	class:warn={state.style === 'warn'}
	class:success={state.style === 'success'}
>
	<div class="message">{state.message}</div>
</div>

<style scoped>
	.notification {
		font-size: 14px;
		position: fixed;
		/* right: 2em; */
		bottom: 0;
		display: grid;
		place-items: center;
		z-index:100;
		padding: 0.3em 2em;
		/* border-radius: 6px; */
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

	.error {
		background: #521e24c7;
		background: rgba(76, 27, 31, 0.95);
		color: #cc3444;
		background: var(--red);
		background: rgba(228, 97, 97, 0.8);
		color: var(--dark-background);
	}
	.warn {
		background: rgba(186, 186, 100, 0.95);
		color: yellow;
		background: rgba(235, 235, 129, 0.8);
		color: var(--dark-background);
	}
	.success {
		background: rgba(47, 66, 58, 0.95);
		color: var(--highlight);
		background: var(--highlight);
		background: rgba(63, 207, 142, 0.8);
		color: var(--dark-background);
	}
</style>
