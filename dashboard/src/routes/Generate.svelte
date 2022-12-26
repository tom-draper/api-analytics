<script lang="ts">
  let loading = false;
  let generatedKey = false;
  let apiKey = "";
  let generateBtn: HTMLButtonElement;
  let copyBtn: HTMLButtonElement;
  let copiedNotification: HTMLDivElement;
  async function genAPIKey() {
    if (!generatedKey) {
      loading = true;
      const response = await fetch(
        "https://api-analytics-server.vercel.app/api/generate-api-key"
      );
      if (response.status == 200) {
        const data = await response.json();
        generatedKey = true;
        apiKey = data;
        generateBtn.style.display = "none";
        copyBtn.style.display = "grid";
      }
      loading = false;
    }
  }

  function copyToClipboard() {
    navigator.clipboard.writeText(apiKey);
    copyBtn.value = "Copied";
    copiedNotification.style.visibility = "visible";
  }
</script>

<div class="generate">
  <div class="content">
    <h2>Generate API key</h2>
    <input type="text" readonly bind:value={apiKey} />
    <button id="formBtn" on:click={genAPIKey} bind:this={generateBtn}
      >Generate</button
    >
    <button id="copyBtn" on:click={copyToClipboard} bind:this={copyBtn}
      ><img class="copy-icon" src="img/copy.png" alt="" /></button
    >
    <div id="copied" bind:this={copiedNotification}>Copied!</div>

    <div class="spinner">
      <div class="loader" style="display: {loading ? 'initial' : 'none'}" />
    </div>
  </div>
  <div class="details">
    <div class="keep-secure">Keep your API key safe and secure.</div>
    <div class="highlight logo">API Analytics</div>
    <img class="footer-logo" src="img/logo.png" alt="" />
  </div>
</div>

<style scoped>

  #copyBtn {
    background: #1c1c1c;
    display: none;
    background: #343434;
    place-items: center;
    margin: auto;
  }
  .copy-icon {
    filter: contrast(0.3);
    height: 20px;
  }
  #copied {
    color: var(--highlight);
    margin: 2em auto auto;
    visibility: hidden;
    height: 1em;
  }
  /* .spinner {
    height: 7em;
  } */
</style>
