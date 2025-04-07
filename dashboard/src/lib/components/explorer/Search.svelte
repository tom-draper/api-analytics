<script lang="ts">
	import Lightning from '$components/Lightning.svelte';

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
	]

	function getPlaceholder() {
		let placeholderExample = [];
		const used = new Set<number>();
		while (placeholderExample.length < 5) {
			const idx = Math.floor(Math.random() * placeholderExamples.length);
			if (!used.has(idx)) {
				placeholderExample.push(placeholderExamples[idx]);
				used.add(idx);
			}
		}
		return placeholderExample.join(' ');
	}
</script>

<div class="p-4 pr-0 text-sm text-[var(--faded-text)]">
	<div class="absolute grid h-[36px] place-items-center px-4">
		<div class="h-[16px] text-[var(--highlight)]">
			<Lightning />
		</div>
	</div>
	<input
		type="text"
		name=""
		id=""
		placeholder={getPlaceholder()}
		class="m-auto h-[36px] w-full border border-solid border-[#2e2e2e] bg-[var(--background)] pl-10 text-left"
	/>
</div>

<style scoped>
	input {
		border-right: none;
		border-radius: 4px 0 0 4px;
	}
	input::placeholder {
		color: var(--dim-text) !important;
	}
</style>
