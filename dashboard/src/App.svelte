<script lang="ts">
	import { Router, Route } from 'svelte-routing';
	import Home from './routes/Home.svelte';
	import Generate from './routes/Generate.svelte';
	import SignIn from './routes/SignIn.svelte';
	import Dashboard from './routes/Dashboard.svelte';
	import Delete from './routes/Delete.svelte';
	import Monitoring from './routes/Monitoring.svelte';
	import PrivacyPolicy from './routes/PrivacyPolicy.svelte';
	import Error from './routes/Error.svelte';

	// Used for SSR. A falsy value is ignored by the Router.
	export let url = '';
</script>

<Router {url}>
	<Route path="/" component={Home} />
	<Route path="/generate" component={Generate} />
	<Route path="/dashboard" component={SignIn} page="dashboard" />
	<Route path="/monitoring" component={SignIn} page="monitoring" />
	<Route
		path="/dashboard/demo"
		component={Dashboard}
		demo={true}
		userID={null}
	/>
	<Route path="/dashboard/:userID" component={Dashboard} demo={false} />
	<Route path="/monitoring/:userID" component={Monitoring} />
	<Route path="/delete" component={Delete} />
	<Route path="/privacy-policy" component={PrivacyPolicy} />
	<Route path="/:err" component={Error} />
</Router>

<style>
</style>
