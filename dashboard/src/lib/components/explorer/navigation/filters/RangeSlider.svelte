<script lang="ts">
	let {
		min = 0,
		max = 100,
		values = $bindable<[number, number]>([min, max]),
		onstop
	}: {
		min?: number;
		max?: number;
		values?: [number, number];
		onstop?: (handle: 0 | 1, value: number) => void;
	} = $props();

	let trackEl = $state<HTMLDivElement | undefined>(undefined);
	let active = $state<0 | 1 | null>(null);

	const loPct = $derived(((values[0] - min) / (max - min)) * 100);
	const hiPct = $derived(((values[1] - min) / (max - min)) * 100);

	function clamp(v: number, lo: number, hi: number) {
		return Math.max(lo, Math.min(hi, v));
	}

	function clientToValue(clientX: number): number {
		if (!trackEl) return min;
		const { left, width } = trackEl.getBoundingClientRect();
		return min + clamp((clientX - left) / width, 0, 1) * (max - min);
	}

	function startDrag(handle: 0 | 1, initialClientX: number) {
		active = handle;
		let last = values[handle];

		function move(clientX: number) {
			const raw = clientToValue(clientX);
			const v = handle === 0 ? clamp(raw, min, values[1]) : clamp(raw, values[0], max);
			last = v;
			values = handle === 0 ? [v, values[1]] : [values[0], v];
		}

		function stop() {
			onstop?.(handle, last);
			active = null;
			window.removeEventListener('mousemove', onMouseMove);
			window.removeEventListener('mouseup', onMouseUp);
			window.removeEventListener('touchmove', onTouchMove);
			window.removeEventListener('touchend', onTouchEnd);
		}

		function onMouseMove(e: MouseEvent) { move(e.clientX); }
		function onTouchMove(e: TouchEvent) { e.preventDefault(); move(e.touches[0].clientX); }
		function onMouseUp() { stop(); }
		function onTouchEnd() { stop(); }

		move(initialClientX);
		window.addEventListener('mousemove', onMouseMove);
		window.addEventListener('mouseup', onMouseUp);
		window.addEventListener('touchmove', onTouchMove, { passive: false });
		window.addEventListener('touchend', onTouchEnd);
	}

	function onTrackPointerDown(e: MouseEvent) {
		const v = clientToValue(e.clientX);
		const handle: 0 | 1 = Math.abs(v - values[0]) <= Math.abs(v - values[1]) ? 0 : 1;
		startDrag(handle, e.clientX);
	}
</script>

<div class="px-2 py-3 select-none">
	<div
		bind:this={trackEl}
		class="relative h-[3px] rounded-full cursor-pointer"
		style="background: var(--border)"
		onmousedown={onTrackPointerDown}
		role="presentation"
	>
		<!-- Active range -->
		<div
			class="pointer-events-none absolute h-full rounded-full"
			style="left: {loPct}%; right: {100 - hiPct}%; background: rgba(var(--highlight-rgb), 0.55)"
		></div>

		<!-- Low handle -->
		<div
			role="slider"
			tabindex="0"
			aria-valuenow={values[0]}
			aria-valuemin={min}
			aria-valuemax={values[1]}
			class="absolute top-1/2 size-3 -translate-x-1/2 -translate-y-1/2 rounded-full"
			class:cursor-grab={active !== 0}
			class:cursor-grabbing={active === 0}
			style="left: {loPct}%; background: var(--highlight); box-shadow: 0 1px 4px rgba(0,0,0,0.4)"
			onmousedown={(e) => { e.stopPropagation(); startDrag(0, e.clientX); }}
			ontouchstart={(e) => { e.stopPropagation(); startDrag(0, e.touches[0].clientX); }}
		></div>

		<!-- High handle -->
		<div
			role="slider"
			tabindex="0"
			aria-valuenow={values[1]}
			aria-valuemin={values[0]}
			aria-valuemax={max}
			class="absolute top-1/2 size-3 -translate-x-1/2 -translate-y-1/2 rounded-full"
			class:cursor-grab={active !== 1}
			class:cursor-grabbing={active === 1}
			style="left: {hiPct}%; background: var(--highlight); box-shadow: 0 1px 4px rgba(0,0,0,0.4)"
			onmousedown={(e) => { e.stopPropagation(); startDrag(1, e.clientX); }}
			ontouchstart={(e) => { e.stopPropagation(); startDrag(1, e.touches[0].clientX); }}
		></div>
	</div>
</div>
