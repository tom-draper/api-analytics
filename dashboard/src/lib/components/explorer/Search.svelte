<script lang="ts">
	import { onMount } from 'svelte';
	import Lightning from '$components/Lightning.svelte';

	let { value = $bindable(''), loading = false }: { value?: string; loading?: boolean } = $props();

	let inputEl = $state<HTMLInputElement | undefined>(undefined);
	let canvasEl = $state<HTMLCanvasElement | undefined>(undefined);
	let focused = $state(false);

	// Canvas caustic animation — simulates light refracting through moving water
	onMount(() => {
		const W = 280, H = 20;
		let rafId: number;
		let lastTs = 0;
		let accTime = 0;
		let flashT = 0;
		let wasFocused = false;
		let streakAlpha = 0;
		let streakPos = 0;

		// Read highlight colour from CSS custom property so it stays theme-aware
		const cssRgb = canvasEl?.parentElement
			? getComputedStyle(canvasEl.parentElement).getPropertyValue('--highlight-rgb').trim().split(',').map(Number)
			: [];
		const hlR = cssRgb[0] ?? 63;
		const hlG = cssRgb[1] ?? 207;
		const hlB = cssRgb[2] ?? 142;

		function draw(ts: number) {
			if (!canvasEl) return;
			const ctx = canvasEl.getContext('2d');
			if (!ctx) return;

			const dt = lastTs > 0 ? (ts - lastTs) * 0.001 : 0;
			const speed = focused ? 0.55 : 0.32;
			accTime += dt * speed;
			lastTs = ts;
			const t = accTime;

			// Brief brightness burst when focus is gained
			if (focused && !wasFocused) flashT = 1.0;
			wasFocused = focused;
			flashT *= 0.92;
			const brightBoost = 1 + flashT * 0.8;

			// Loading comet: left-to-right chevron with trailing tail
			if (loading) streakAlpha = Math.min(1, streakAlpha + dt * 5);
			else streakAlpha = Math.max(0, streakAlpha - dt * 4);
			streakPos = (streakPos + dt * 0.9) % 1.0; // always left to right

			const img = ctx.createImageData(W, H);
			const d = img.data;

			// Three caustic epicentres drifting on slow Lissajous paths
			const cx1 = 0.22 + Math.sin(t * 0.41) * 0.18 + Math.sin(t * 0.17) * 0.08;
			const cx2 = 0.68 + Math.cos(t * 0.29) * 0.20 + Math.cos(t * 0.13) * 0.07;
			const cx3 = 0.45 + Math.sin(t * 0.19 + 1.2) * 0.14;

			// Global slow breath
			const breath = 0.78 + Math.sin(t * 0.22) * 0.22;

			// Chromatic aberration: R/B planar waves slightly offset in x
			const ca = 0.03;

			for (let y = 0; y < H; y++) {
				const ny = y / H;
				for (let x = 0; x < W; x++) {
					const nx = x / W;

					const r1 = Math.sqrt((nx - cx1) ** 2 + (ny - 0.5) ** 2);
					const r2 = Math.sqrt((nx - cx2) ** 2 + (ny - 0.5) ** 2);
					const r3 = Math.sqrt((nx - cx3) ** 2 + (ny - 0.5) ** 2);

					// Shared radial caustic waves
					const radial =
						Math.sin(r1 * 22 - t * 1.1) * 0.32 +
						Math.sin(r2 * 17 - t * 0.85) * 0.28 +
						Math.sin(r3 * 14 - t * 0.70) * 0.18;

					// Per-channel planar waves with spatial offset (chromatic aberration)
					const pG = Math.sin(nx * 11 + t * 0.55) * 0.14 + Math.sin((nx * 0.6 - ny) * 13 + t * 0.45) * 0.08;
					const pR = Math.sin((nx + ca) * 11 + t * 0.55) * 0.14 + Math.sin(((nx + ca) * 0.6 - ny) * 13 + t * 0.45) * 0.08;
					const pB = Math.sin((nx - ca) * 11 + t * 0.55) * 0.14 + Math.sin(((nx - ca) * 0.6 - ny) * 13 + t * 0.45) * 0.08;

					// Squaring creates the sharp-bright / dark-trough caustic look
					const bG = Math.pow(Math.max(0, radial + pG + 0.08), 2) * breath * brightBoost;
					const bR = Math.pow(Math.max(0, radial + pR + 0.08), 2) * breath * brightBoost;
					const bB = Math.pow(Math.max(0, radial + pB + 0.08), 2) * breath * brightBoost;

					// Loading comet: asymmetric chevron, sharp leading edge + long trailing tail
					let streak = 0;
					if (streakAlpha > 0) {
						// Chevron: tip at centre, wings trail back toward edges
						const chevronX = streakPos - Math.abs(ny - 0.5) * 0.7;
						// Also check wrapped position so tail shows when head just crossed 0
						function cometAt(cx: number) {
							const rel = nx - cx; // negative = behind (tail), positive = ahead
							if (rel <= 0) return Math.exp(rel / 0.18);        // long exponential tail
							else         return Math.exp(-(rel * rel) / 0.002); // sharp leading edge
						}
						streak = Math.max(cometAt(chevronX), cometAt(chevronX - 1)) * streakAlpha * 0.85;
					}

					const i = (y * W + x) * 4;
					// Colour: highlight hue with subtle R/B fringing at caustic edges
					d[i]     = Math.min(255, Math.max(0, Math.round(hlR + (bR - bG) * 160)));
					d[i + 1] = hlG;
					d[i + 2] = Math.min(255, Math.max(0, Math.round(hlB + (bB - bG) * 160)));
					d[i + 3] = Math.min(255, Math.round((bG + streak) * 260));
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
			if (e.key === 'Tab') {
				e.preventDefault();
				if (document.activeElement === inputEl) inputEl?.blur();
				else inputEl?.focus();
				return;
			}
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
		'userAgent:Chrome*',
		'userAgent:Mozilla*',
		'userAgent:*Macintosh*',
		'userAgent:"*Windows NT 10.0*"',
		'userAgent:*iPhone*',
		'userAgent:*Googlebot*',
		'hostname:example.com',
		'hostname:example2.com',
		'hostname:example3.com',
		'location:US',
		'location:FR',
		'location:UK',
		'location:AU',
		'location:IN',
		'userId:72fd8db2-a64d-40a2-9bf7-1149ad0feea7',
		'userId:560b1dc3-b2e2-4919-80f6-776418a1b14d',
		'userId:5a56556c-ee93-4503-82b6-2cdd41206108',
		'userId:2345678901',
		'responseTime:>1000',
		'responseTime:<20',
		'responseTime:0',
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
		height={20}
		class="pointer-events-none absolute inset-0 h-full w-full"
		style="filter: blur(8px); opacity: 0.85;"
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
		filter: drop-shadow(0 0 14px rgba(var(--highlight-rgb), 0.95));
	}
</style>
