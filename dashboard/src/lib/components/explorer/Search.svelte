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

<div class="relative flex h-full w-full items-center">
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

<style scoped>
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
