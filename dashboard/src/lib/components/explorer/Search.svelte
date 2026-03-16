<script lang="ts">
	import { onMount } from 'svelte';
	import Lightning from '$components/Lightning.svelte';

	let { value = $bindable('') }: { value?: string } = $props();

	let inputEl = $state<HTMLInputElement | undefined>(undefined);

	onMount(() => {
		function handleKeydown(e: KeyboardEvent) {
			if (e.key === 'Escape') {
				inputEl?.blur();
				return;
			}
			if (document.activeElement === inputEl) return;
			if (e.metaKey || e.ctrlKey || e.altKey) return;
			if (e.key.length !== 1) return;
			if (!inputEl) return;
			inputEl.focus();
			value += e.key;
			e.preventDefault();
		}
		window.addEventListener('keydown', handleKeydown);
		return () => window.removeEventListener('keydown', handleKeydown);
	});

	const placeholderExamples = [
		'status:200',
		'status:201',
		'status:400',
		'status:404',
		'status:500',
		'method:GET',
		'method:POST',
		'method:PUT',
		'method:DELETE',
		'user_agent:Chrome*',
		'user_agent:Mozilla*',
		'user_agent:*Macintosh*',
		'user_agent:"*Windows NT 10.0*"',
		'user_agent:*iPhone*',
		'user_agent:*Googlebot*',
		'hostname:example.com',
		'hostname:example2.com',
		'hostname:example3.com',
		'location:US',
		'location:FR',
		'location:UK',
		'location:AU',
		'location:IN',
		'user_id:72fd8db2-a64d-40a2-9bf7-1149ad0feea7',
		'user_id:560b1dc3-b2e2-4919-80f6-776418a1b14d',
		'user_id:5a56556c-ee93-4503-82b6-2cdd41206108',
		'user_id:2345678901',
		'response_time:>1000',
		'response_time:<20',
		'response_time:0',
	];

	function getPlaceholder() {
		const used = new Set<number>();
		const examples: string[] = [];
		while (examples.length < 5) {
			const idx = Math.floor(Math.random() * placeholderExamples.length);
			if (!used.has(idx)) {
				examples.push(placeholderExamples[idx]);
				used.add(idx);
			}
		}
		return examples.join('  ');
	}
</script>

<div class="relative flex h-full w-full items-center overflow-hidden">
	<!-- Ambient animation layer -->
	<div class="shimmer-layer" aria-hidden="true">
		<div class="orb orb-a"></div>
		<div class="orb orb-b"></div>
		<div class="orb orb-c"></div>
		<div class="orb orb-d"></div>
	</div>

	<div class="icon-wrap pointer-events-none absolute left-4 flex h-[18px] items-center text-[var(--highlight)]">
		<Lightning />
	</div>
	<input
		bind:this={inputEl}
		bind:value
		type="text"
		placeholder={getPlaceholder()}
		class="!mb-0 !bg-transparent !w-full !text-left !text-[14px] !pl-11 !pr-6 h-full text-[var(--faded-text)] focus:outline-none"
	/>
</div>

<style>
	input::placeholder {
		color: var(--muted-text);
	}

	.icon-wrap {
		filter: drop-shadow(0 0 4px rgba(var(--highlight-rgb), 0.4));
		transition: filter 0.2s ease;
	}

	div:focus-within .icon-wrap {
		filter: drop-shadow(0 0 10px rgba(var(--highlight-rgb), 0.85));
	}

	/* Animation layer */
	.shimmer-layer {
		position: absolute;
		inset: 0;
		pointer-events: none;
		overflow: hidden;
	}

	.orb {
		position: absolute;
		top: -80%;
		height: 260%;
		border-radius: 50%;
		will-change: transform;
	}

	.orb-a {
		width: 38%;
		left: 5%;
		background: radial-gradient(ellipse at center, rgba(var(--highlight-rgb), 0.09) 0%, transparent 65%);
		filter: blur(18px);
		animation: drift-a 26s ease-in-out infinite;
	}

	.orb-b {
		width: 28%;
		left: 30%;
		background: radial-gradient(ellipse at center, rgba(var(--highlight-rgb), 0.06) 0%, transparent 60%);
		filter: blur(22px);
		animation: drift-b 34s ease-in-out infinite;
		animation-delay: -8s;
	}

	.orb-c {
		width: 42%;
		left: 50%;
		background: radial-gradient(ellipse at center, rgba(var(--highlight-rgb), 0.07) 0%, transparent 65%);
		filter: blur(20px);
		animation: drift-c 22s ease-in-out infinite;
		animation-delay: -14s;
	}

	.orb-d {
		width: 24%;
		left: 75%;
		background: radial-gradient(ellipse at center, rgba(var(--highlight-rgb), 0.05) 0%, transparent 60%);
		filter: blur(16px);
		animation: drift-a 40s ease-in-out infinite;
		animation-delay: -20s;
	}

	@keyframes drift-a {
		0%, 100% { transform: translateX(0%) scaleX(1); }
		30%       { transform: translateX(18%) scaleX(1.08); }
		60%       { transform: translateX(-8%) scaleX(0.95); }
	}

	@keyframes drift-b {
		0%, 100% { transform: translateX(0%) scaleX(1); }
		40%       { transform: translateX(-22%) scaleX(0.92); }
		75%       { transform: translateX(14%) scaleX(1.06); }
	}

	@keyframes drift-c {
		0%, 100% { transform: translateX(0%) scaleX(1); }
		50%       { transform: translateX(-30%) scaleX(1.1); }
	}

	/* Breathe slightly more alive on focus */
	div:focus-within .orb-a { animation-duration: 18s; }
	div:focus-within .orb-b { animation-duration: 24s; }
	div:focus-within .orb-c { animation-duration: 16s; }
	div:focus-within .orb-d { animation-duration: 28s; }
</style>
