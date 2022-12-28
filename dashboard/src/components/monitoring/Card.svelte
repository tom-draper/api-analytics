<script lang="ts">
  import { onMount } from "svelte";
  import ResponseTime from "./ResponseTime.svelte";

  function setUptime() {
    let success = 0;
    let total = 0;
    for (let i = 0; i < samples.length; i++) {
      if (samples[i].status == 'no-request') {
        continue;
      }
      if (
        samples[i].status == "success" ||
        samples[i].status == "delay"
      ) {
        success++;
      }
      total++;
    }

    let per = (success / total) * 100;
    if (per == 100) {
      uptime = "100";
    } else {
      uptime = per.toFixed(1);
    }
  }

  function periodToMarkers(period: string): number {
    if (period == "24h") {
      return 24 * 2;
    } else if (period == "7d") {
      return 12 * 7;
    } else if (period == "30d") {
      return 30 * 4;
    } else if (period == "60d") {
      return 60 * 2;
    } else {
      return null;
    }
  }

  function periodSample() {
    /* Sample ping recordings at regular intervals if number of bars fewer than 
    total recordings the current period length */
    let sample = []
    if (period == "30d") {
      // Sample 1 in 4
      for (let i = 0; i < data.length; i++) {
        if (i % 4 == 0) {
          sample.push(data[i])
        }
      }
    } else if (period == "60d") {
      // Sample 1 in 8
      for (let i = 0; i < data.length; i++) {
        if (i % 8 == 0) {
          sample.push(data[i])
        }
      }
    } else {
      // No sampling - use all
      sample = data
    }
    return sample
  }

  function setSamples() {
    let markers = periodToMarkers(period);
    samples = Array(markers).fill({ status: 'no-request', responseTime: 0 });
    let sampledData = periodSample();
    let start = markers - sampledData.length;

    for (let i = 0; i < sampledData.length; i++) {
      samples[i + start] = {status: 'no-request', responseTime: sampledData[i].responseTime};
      if (sampledData[i].status >= 200 && sampledData[i].status <= 299) {
        samples[i + start].status = 'success';
      } else if (sampledData[i].status != null) {
        samples[i + start].status = 'error';
      }
    }
  }

  function setError() {
    error = samples[samples.length - 1].status == "error";
    anyError = anyError || error;
  }

  function build() {
    setSamples();
    setError();
    setUptime();
  }

  let uptime = '';
  let error = false;
  let samples: any[];
  onMount(() => {
    build();
  });

  $: period && build();

  export let url: string, data: any, period: string, anyError: boolean;
</script>

<div class="card" class:card-error={error}>
  <div class="card-text">
    <div class="card-text-left">
      <div class="card-status">
        {#if error}
          <img src="/img/smallcross.png" alt="" />
        {:else}
          <img src="/img/smalltick.png" alt="" />
        {/if}
      </div>
      <div class="endpoint">{url}</div>
    </div>
    <div class="card-text-right">
      <div class="uptime">Uptime: {uptime}%</div>
    </div>
  </div>
  {#if samples != undefined}
    <div class="measurements">
      {#each samples as sample}
        <div class="measurement {sample.status}" />
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
    color: #707070;
  }
  .uptime {
    color: #707070;
  }
</style>
