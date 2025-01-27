<script lang="ts">
    import Error from '$lib/components/Error.svelte';

    export let error, status: number;
</script>

<Error status={status}  message={error && error.message ? error.message : ''} />