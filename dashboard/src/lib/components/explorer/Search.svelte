<script lang="ts">
	import { onMount } from 'svelte';
	import Lightning from '$components/Lightning.svelte';

	let { value = $bindable('') }: { value?: string } = $props();

	let inputEl = $state<HTMLInputElement | undefined>(undefined);
	let canvasEl = $state<HTMLCanvasElement | undefined>(undefined);
	let focused = $state(false);

	// Canvas caustic animation — simulates light refracting through moving water
	onMount(() => {
		const W = 280, H = 10;
		let rafId: number;
		let lastTs = 0;
		let accTime = 0;

		function draw(ts: number) {
			if (!canvasEl) return;
			const ctx = canvasEl.getContext('2d');
			if (!ctx) return;

			const speed = focused ? 0.55 : 0.32;
			if (lastTs > 0) accTime += (ts - lastTs) * 0.001 * speed;
			lastTs = ts;
			const t = accTime;

			const img = ctx.createImageData(W, H);
			const d = img.data;

			// Two caustic epicentres drifting on slow Lissajous paths
			const cx1 = 0.22 + Math.sin(t * 0.41) * 0.18 + Math.sin(t * 0.17) * 0.08;
			const cx2 = 0.68 + Math.cos(t * 0.29) * 0.20 + Math.cos(t * 0.13) * 0.07;
			const cx3 = 0.45 + Math.sin(t * 0.19 + 1.2) * 0.14;

			// Global slow breath
			const breath = 0.78 + Math.sin(t * 0.22) * 0.22;

			for (let y = 0; y < H; y++) {
				const ny = y / H;
				for (let x = 0; x < W; x++) {
					const nx = x / W;

					const r1 = Math.sqrt((nx - cx1) ** 2 + (ny - 0.5) ** 2);
					const r2 = Math.sqrt((nx - cx2) ** 2 + (ny - 0.5) ** 2);
					const r3 = Math.sqrt((nx - cx3) ** 2 + (ny - 0.5) ** 2);

					// Caustic interference: overlapping radial + planar waves
					const v =
						Math.sin(r1 * 22 - t * 1.1) * 0.32 +
						Math.sin(r2 * 17 - t * 0.85) * 0.28 +
						Math.sin(r3 * 14 - t * 0.70) * 0.18 +
						Math.sin(nx * 11 + t * 0.55) * 0.14 +
						Math.sin((nx * 0.6 - ny) * 13 + t * 0.45) * 0.08;

					// Squaring creates the sharp-bright / dark-trough caustic look
					const bright = Math.pow(Math.max(0, v + 0.08), 2) * breath;
					const alpha = Math.min(255, bright * 260);

					const i = (y * W + x) * 4;
					d[i]     = 63;
					d[i + 1] = 207;
					d[i + 2] = 142;
					d[i + 3] = alpha;
				}
			}

			ctx.putImageData(img, 0, 0);
			rafId = requestAnimationFrame(draw);
		}

		rafId = requestAnimationFrame(draw);
		return () => cancelAnimationFrame(rafId);
	});

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
	<canvas
		bind:this={canvasEl}
		width={280}
		height={10}
		class="pointer-events-none absolute inset-0 h-full w-full"
		style="filter: blur(11px); opacity: 0.85;"
		aria-hidden="true"
	></canvas>

	<div class="icon-wrap pointer-events-none absolute left-4 flex h-[18px] items-center text-[var(--highlight)]">
		<Lightning />
	</div>
	<input
		bind:this={inputEl}
		bind:value
		type="text"
		placeholder={getPlaceholder()}
		class="!mb-0 !bg-transparent !w-full !text-left !text-[14px] !pl-11 !pr-6 h-full text-[var(--faded-text)] focus:outline-none"
		onfocus={() => (focused = true)}
		onblur={() => (focused = false)}
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
</style>
