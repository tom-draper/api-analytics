<script lang="ts">
	import Error from '$components/Error.svelte';
	import { page } from '$app/state';

	const message =
		page.status === 404 ? '' : page.error && page.error.message ? page.error.message : '';
	const status = page.status;
</script>

<Error {status} {message} />
