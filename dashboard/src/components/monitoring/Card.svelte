<script lang="ts">
  import { onMount } from "svelte";
  import ResponseTime from "./ResponseTime.svelte";

  function setUptime() {
    let success = 0;
    let total = 0;
    for (let i = 0; i < measurements.length; i++) {
      if (
        measurements[i].status == "success" ||
        measurements[i].status == "delay"
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

  function setMeasurements() {
    let markers = periodToMarkers(period);
    measurements = Array(markers).fill({ status: null, response_time: 0});
    let start = markers - data.measurements.length;
    for (let i = 0; i < data.measurements.length; i++) {
      measurements[i + start] = data.measurements[i];
    }
  }

  function setError() {
    error = measurements[measurements.length - 1].status == "error";
    anyError = anyError || error;
  }

  function build() {
    setMeasurements();
    setError();
    setUptime();
  }

  let uptime = "";
  let error = false;
  let measurements: any[];
  onMount(() => {
    build();
  });

  $: period && build();

  export let data: {name: string, measurements: any[]}, period: string, anyError: boolean;
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
      <div class="endpoint">{data.name}</div>
    </div>
    <div class="card-text-right">
      <div class="uptime">Uptime: {uptime}%</div>
    </div>
  </div>
  <div class="measurements">
    {#each measurements as measurement}
      <div class="measurement {measurement.status}" />
    {/each}
  </div>
  <div class="response-time">
    <ResponseTime data={measurements} {period} />
  </div>
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
  .null {
    color: #707070;
  }
  .uptime {
    color: #707070;
  }
</style>
