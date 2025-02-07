<script lang="ts">
	import Checkbox from "$components/explorer/navigation/Checkbox.svelte";

</script>

<div class="flex flex-col">
    <div class="flex items-center border-b border-solid border-[#2e2e2e]">
        <Checkbox checked={true} label="Success" color="var(--highlight)" />
    </div>
    <div class="flex items-center border-b border-solid border-[#2e2e2e]">
        <Checkbox checked={true} label="Client error" color="var(--yellow)" />
    </div>
    <div class="flex items-center">
        <Checkbox checked={true} label="Server error" color="var(--red)" />
    </div>
</div>