<script lang="ts">
	import { untrustedBackendOrigin } from '$lib/url';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	// The non-default backend origin selected by ?source=, or null when trusted.
	const origin = $derived(untrustedBackendOrigin());

	// The origin the user has acknowledged for this session (once acknowledged we
	// stop warning, but a *different* untrusted origin re-prompts).
	let dismissedOrigin = $state<string | null>(null);

	// Honour a prior acknowledgement stored for this session.
	$effect(() => {
		if (
			origin &&
			typeof sessionStorage !== 'undefined' &&
			sessionStorage.getItem(`backend-ack:${origin}`) === '1'
		) {
			dismissedOrigin = origin;
		}
	});

	const show = $derived(origin !== null && origin !== dismissedOrigin);

	function acknowledge() {
		if (origin && typeof sessionStorage !== 'undefined') {
			sessionStorage.setItem(`backend-ack:${origin}`, '1');
		}
		dismissedOrigin = origin;
	}

	function useDefault() {
		const url = new URL(page.url);
		url.searchParams.delete('source');
		goto(url.pathname + url.search + url.hash);
	}
</script>

{#if show}
	<div
		class="backend-overlay"
		role="alertdialog"
		aria-modal="true"
		aria-labelledby="backend-title"
	>
		<div class="backend-card">
			<h2 id="backend-title">Custom backend detected</h2>
			<p>
				This dashboard is set to send your data &mdash; including your API key &mdash; to:
			</p>
			<p class="backend-origin">{origin}</p>
			<p>
				Only continue if this is a server <strong>you</strong> control. If you didn't set this
				up, someone may be trying to capture your API key &mdash; use the default backend instead.
			</p>
			<div class="backend-actions">
				<button class="backend-btn default" onclick={useDefault}>Use default backend</button
				>
				<button class="backend-btn danger" onclick={acknowledge}>Continue anyway</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.backend-overlay {
		position: fixed;
		inset: 0;
		z-index: 10000;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5em;
		background: rgba(0, 0, 0, 0.8);
		backdrop-filter: blur(4px);
	}
	.backend-card {
		max-width: 34em;
		width: 100%;
		background: #14151a;
		border: 1px solid #3a2320;
		border-radius: var(--radius-md, 10px);
		padding: 2em;
		color: #d8dae0;
		box-shadow: 0 0 120px 2px rgba(228, 97, 97, 0.25);
		font-size: 0.95em;
		line-height: 1.55;
	}
	.backend-card h2 {
		margin: 0 0 0.6em;
		color: #e0a561;
		font-size: 1.25em;
	}
	.backend-card p {
		margin: 0 0 0.9em;
	}
	.backend-origin {
		font-family: 'Geist Mono', 'Fira Code', monospace;
		word-break: break-all;
		background: #2b1a0d;
		color: #e0a561;
		border: 1px solid #4a3320;
		border-radius: var(--radius-md, 8px);
		padding: 0.6em 0.9em;
	}
	.backend-actions {
		display: flex;
		gap: 0.75em;
		margin-top: 1.4em;
	}
	.backend-btn {
		flex: 1;
		height: 42px;
		border: none;
		border-radius: var(--radius-md, 8px);
		cursor: pointer;
		font-size: 0.95em;
		font-family: inherit;
	}
	.backend-btn.default {
		background: var(--highlight, #3fcf8e);
		color: #06281a;
		font-weight: 500;
	}
	.backend-btn.danger {
		background: #333333;
		color: #cfcfcf;
	}
	.backend-btn.danger:hover {
		background: #444444;
	}
	.backend-btn.default:hover {
		filter: brightness(1.08);
	}
</style>
