<script lang="ts">
  let state: "sign-in" | "loading" = "sign-in";
  let apiKey = "";
  async function genAPIKey() {
    setState("loading");
    try {
      const response = await fetch(
        // `https://api-analytics-server.vercel.app/api/user-id/${apiKey}`
        `https://213.168.248.206/api/user-id/${apiKey}`
      );

      if (response.status === 200) {
        const userID = await response.json();
        window.location.href = `/${page}/${userID.replaceAll("-", "")}`;
      } else {
        setState("sign-in");
      }
    } catch (e) {
      console.log(e);
      setState("sign-in");
    }
  }

  function setState(value: typeof state) {
    state = value;
  }

  export let page: "dashboard" | "monitoring";
</script>

<div class="generate">
  <div class="content">
    {#if page == "dashboard"}
      <h2>Dashboard</h2>
    {:else if page == "monitoring"}
      <h2>Monitoring</h2>
    {/if}
    <input type="text" bind:value={apiKey} placeholder="Enter API key" />
    <button
      id="formBtn"
      on:click={genAPIKey}
      class:no-display={state != "sign-in"}>Load</button
    >
    <button id="formBtn" class:no-display={state != "loading"}>
      <div class="spinner">
        <div class="loader" />
      </div>
    </button>
  </div>
  <div class="details">
    <div class="keep-secure">Keep your API key safe and secure.</div>
    <div class="highlight logo">API Analytics</div>
    <img class="footer-logo" src="img/logo.png" alt="" />
  </div>
</div>

<style scoped>
  .spinner {
    height: auto;
  }
  .loader {
    border: 3px solid #343434;
    border-top: 3px solid var(--highlight);
    height: 10px;
    width: 10px;
  }
</style>
