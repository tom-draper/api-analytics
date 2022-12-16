<script lang="ts">
  import { Router, Route } from "svelte-routing";
  import Home from "./routes/Home.svelte";
  import Generate from "./routes/Generate.svelte";
  import SignIn from "./routes/SignIn.svelte";
  import Dashboard from "./routes/Dashboard.svelte";
  import Delete from "./routes/Delete.svelte";

  // Used for SSR. A falsy value is ignored by the Router.
  export let url = "";
</script>

<Router {url}>
    <Route path="/" component={Home} />
    <Route path="/generate" component={Generate} />
    <Route path="/dashboard" component={SignIn} />
    <Route path="/dashboard/:userID" component={Dashboard} />
    <Route path="/delete" component={Delete} />
</Router>

<style>

</style>
