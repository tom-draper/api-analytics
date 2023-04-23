<script lang="ts">
  import { onMount } from "svelte";
  import ResponseTime from "./ResponseTime.svelte";

  async function deleteMonitor() {
    delete data[url];
    data = data; // Trigger reactivity to update display

    try {
      let response = await fetch(
        "https://www.apianalytics-server.com/api/monitor/delete",
        {
          method: "POST",
          mode: "no-cors",
          headers: {
            Accept: "application/json",
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            user_id: userID,
            url: url,
          }),
        }
      );
      if (response.status != 201) {
        console.log("Error", response.status);
      }
    } catch (e) {
      console.log(e);
    }
  }

  function setUptime() {
    let success = 0;
    let total = 0;
    for (let i = 0; i < samples.length; i++) {
      if (samples[i].label === "no-request") {
        continue;
      }
      if (samples[i].label === "success" || samples[i].label === "delay") {
        success++;
      }
      total++;
    }

    let per = (success / total) * 100;
    if (per === 100) {
      uptime = "100";
    } else {
      uptime = per.toFixed(1);
    }
  }

  function periodToMarkers(period: string): number {
    if (period === "24h") {
      return 24 * 2;
    } else if (period === "7d") {
      return 12 * 7;
    } else if (period === "30d") {
      return 30 * 4;
    } else if (period === "60d") {
      return 60 * 2;
    } else {
      return null;
    }
  }

  function periodSample() {
    /* Sample ping recordings at regular intervals if number of bars fewer than 
    total recordings the current period length */
    let sample = [];
    if (period == "30d") {
      // Sample 1 in 4
      for (let i = 0; i < data[url].length; i++) {
        if (i % 4 == 0) {
          sample.push(data[url][i]);
        }
      }
    } else if (period == "60d") {
      // Sample 1 in 8
      for (let i = 0; i < data[url].length; i++) {
        if (i % 8 == 0) {
          sample.push(data[url][i]);
        }
      }
    } else {
      // No sampling - use all
      sample = data[url];
    }
    return sample;
  }

  function setSamples() {
    let markers = periodToMarkers(period);
    samples = Array(markers).fill({ label: "no-request", responseTime: 0 });
    let sampledData = periodSample();
    let start = markers - sampledData.length;

    for (let i = 0; i < sampledData.length; i++) {
      samples[i + start] = {
        label: "no-request",
        status: sampledData[i].status,
        responseTime: sampledData[i].responseTime,
      };
      if (sampledData[i].status >= 200 && sampledData[i].status <= 299) {
        samples[i + start].label = "success";
      } else if (sampledData[i].status != null) {
        samples[i + start].label = "error";
      }
    }
  }

  function setError() {
    if (samples[samples.length - 1].label == null) {
      error = null; // Website not live
    } else {
      error = samples[samples.length - 1].label === "error";
    }
    anyError = anyError || error;
  }

  function build() {
    setSamples();
    setError();
    setUptime();
  }

  let uptime = "";
  let error = false;
  let samples: { label: string; status: number; responseTime: string }[];
  onMount(() => {
    build();
  });

  $: period && build();

  export let url: string,
    data: any,
    userID: string,
    period: string,
    anyError: boolean;
</script>

<div class="card" class:card-error={error}>
  <div class="card-text">
    <div class="card-text-left">
      <div class="card-status">
        {#if error == null}
          <div class="indicator grey-light" />
        {:else if error}
          <div class="indicator red-light" />
        {:else}
          <div class="indicator green-light" />
        {/if}
      </div>
      <a href="http://{url}" class="endpoint"
        ><span style="color: var(--dim-text)">http://</span>{url}</a
      >
      <button class="delete" on:click={deleteMonitor}
        ><img class="bin-icon" src="../img/bin.png" alt="" /></button
      >
    </div>
    <div class="card-text-right">
      <div class="uptime">Uptime: {uptime}%</div>
    </div>
  </div>
  {#if samples != undefined}
    <div class="measurements">
      {#each samples as sample}
        <div
          class="measurement {sample.label}"
          title="Status {sample.status}"
        />
      {/each}
    </div>
    <div class="response-time">
      <ResponseTime data={samples} {period} />
    </div>
  {/if}
</div>

<style scoped>
  .card {
    width: min(100%, 1000px);
    border: 1px solid #2e2e2e;
    margin: 2.2em auto;
  }
  .card-error {
    box-shadow: rgba(228, 98, 98, 0.5) 0px 15px 110px 0px,
      rgba(0, 0, 0, 0.4) 0px 30px 60px -30px;
    border: 2px solid rgba(228, 98, 98, 1);
  }
  .card-text {
    display: flex;
    margin: 2em 2em 0;
    font-size: 0.9em;
  }
  .card-text-left {
    flex-grow: 1;
    display: flex;
  }
  .endpoint {
    margin-left: 10px;
    letter-spacing: 0.01em;
    color: white;
  }
  .measurements {
    display: flex;
    padding: 1em 2em 2em;
  }
  .measurement {
    margin: 0 0.1%;
    flex: 1;
    height: 3em;
    border-radius: 1px;
    background: var(--highlight);
    background: rgb(40, 40, 40);
  }
  .success {
    background: var(--highlight);
  }
  .delayed {
    background: rgb(199, 229, 125);
  }
  .error {
    background: rgb(228, 98, 98);
  }
  .no-request {
    color: var(--dim-text);
  }
  .uptime {
    color: var(--dim-text);
  }
  .indicator {
    width: 10px;
    height: 10px;
    border-radius: 5px;
    margin-right: 5px;
    margin-bottom: 1px;
  }
  .card-status {
    display: grid;
    place-items: center;
  }
  .green-light {
    background: var(--highlight);
    box-shadow: 0 1px 1px #fff, 0 0 6px 3px var(--highlight);
  }
  .red-light {
    background: var(--red);
    box-shadow: 0 1px 1px #fff, 0 0 6px 3px var(--red);
  }
  .grey-light {
    background: grey;
    box-shadow: 0 0 1px 1px #fff;
  }
  .delete {
    background: transparent;
    border: none;
    cursor: pointer;
    margin-left: 10px;
    padding: 2px 4px 1px;
    border-radius: 4px;
  }
  .delete:hover {
    background: var(--red);
  }
  .bin-icon {
    width: 12px;
    filter: invert(0.3);
  }
</style>
