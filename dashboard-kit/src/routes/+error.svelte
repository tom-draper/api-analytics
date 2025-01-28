<script lang="ts">
    import Error from '$lib/components/Error.svelte';
    import { page } from '$app/state';

    console.log($page.status)

    export let error, status: number;
    console.log(error, status)
</script>

<Error status={status}  message={error && error.message ? error.message : ''} />