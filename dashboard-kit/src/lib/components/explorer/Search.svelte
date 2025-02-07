<script lang="ts">
	import Lightning from '$components/Lightning.svelte';
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
		placeholder="status:200 user_agent:Chrome* location:US"
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
