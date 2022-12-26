<script lang="ts">
  async function postMonitor() {
    try {
      const response = await fetch(
        `https://api-analytics-server.vercel.app/api/add-monitor`,
        {
          method: "POST",
          mode: "no-cors",
          headers: {
            Accept: "application/json",
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ api_key: api_key, url: url, ping: pingType == "simple-ping", secure: false }),
        }
      );

      if (response.status != 200) {
        console.log("Error", response.status);
      }
    } catch (e) {
      failed = true;
    }
  }

  function setPingType(value: string) {
    pingType = value;
  }

  let failed = false;
  let url: string;
  let pingType = "simple-ping";

  export let api_key: string;
</script>

<div class="card">
  <div class="card-text">
    <div class="url">
      <div class="start">HTTP://</div>
      <input type="text" placeholder="example.com/endpoint/" bind:value="{url}" />
      <button class="add" on:click="{postMonitor}">Add</button>
    </div>
    <div class="ping-types">
      <button
        class="ping-type"
        on:click={() => setPingType("simple-ping")}
        class:active={pingType == "simple-ping"}
      >
        Simple ping
      </button>
      <button
        class="ping-type"
        on:click={() => setPingType("get-request")}
        class:active={pingType == "get-request"}
      >
        GET request
      </button>
    </div>
    <div class="detail">
      Endpoints are pinged by our servers every 30 mins and response <b
        >status</b
      >
      and response <b>time</b> are logged.
    </div>
  </div>
</div>

<style scoped>
  .card {
    width: min(100%, 1000px);
    border: 1px solid #2e2e2e;
    margin: 2.2em auto;
  }
  .card-text {
    margin: 2em 2em 1.9em;
  }
  input {
    background: #1c1c1c;
    border-radius: 4px;
    border: none;
    margin: 0 10px;
    width: 100%;
    text-align: left;
    padding: 0 10px;
    color: white;
  }
  .url {
    display: flex;
  }
  .start {
    margin: auto;
  }
  .detail {
    margin-top: 30px;
    color: #707070;
    font-weight: 500;
    font-size: 0.9em;
  }
  button {
    border: none;
    border-radius: 4px;
    background: grey;
    cursor: pointer;
  }
  .add {
    background: var(--highlight);
    padding: 5px 20px;
  }
  .ping-types {
    margin-top: 20px;
  }
  .ping-type {
    padding: 5px 20px;
  }
  .active {
    background: var(--highlight);
  }
</style>
