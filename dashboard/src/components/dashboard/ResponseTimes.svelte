<script lang="ts">
  import { onMount } from "svelte";

  // Median and quartiles from StackOverflow answer
  // https://stackoverflow.com/a/55297611/8851732
  const asc = (arr) => arr.sort((a, b) => a - b);
  const sum = (arr) => arr.reduce((a, b) => a + b, 0);
  const mean = (arr) => sum(arr) / arr.length;

  // sample standard deviation
  function std(arr: number[]) {
    const mu = mean(arr);
    const diffArr = arr.map((a) => (a - mu) ** 2);
    return Math.sqrt(sum(diffArr) / (arr.length - 1));
  }

  function quantile(arr: number[], q: number) {
    const sorted = asc(arr);
    const pos = (sorted.length - 1) * q;
    const base = Math.floor(pos);
    const rest = pos - base;
    if (sorted[base + 1] != undefined) {
      return sorted[base] + rest * (sorted[base + 1] - sorted[base]);
    } else if (sorted[base] != undefined) {
      return sorted[base];
    } else {
      return 0
    }
  }

  function markerPosition(x: number) {
    let position = Math.log10(x) * 125 - 300;
    if (position < 0) {
      return 0;
    } else if (position > 100) {
      return 100;
    } else {
      return position;
    }
  }

  function setMarkerPosition(median: number) {
    let position = markerPosition(median);
    marker.style.left = `${position}%`;
  }

  function build() {
    let responseTimes: number[] = [];
    for (let i = 0; i < data.length; i++) {
      responseTimes.push(data[i].response_time);
    }
    LQ = quantile(responseTimes, 0.25);
    median = quantile(responseTimes, 0.5);
    UQ = quantile(responseTimes, 0.75);
    setMarkerPosition(median);
  }

  let median: number ;
  let LQ: number;
  let UQ: number ;
  let marker: HTMLDivElement;
  let mounted = false;
  onMount(() => {
    mounted = true;
  });
  
  $: data && mounted && build();

  export let data: RequestsData;
</script>

<div class="card">
  <div class="card-title">
    Response times <span class="milliseconds">(ms)</span>
  </div>
  <div class="values">
    <div class="value lower-quartile">{LQ}</div>
    <div class="value median">{median}</div>
    <div class="value upper-quartile">{UQ}</div>
  </div>
  <div class="labels">
    <div class="label">25%</div>
    <div class="label">Median</div>
    <div class="label">75%</div>
  </div>
  <div class="bar">
    <div class="bar-green" />
    <div class="bar-yellow" />
    <div class="bar-red" />
    <div class="marker" bind:this={marker} />
  </div>
</div>

<style>
  .values {
    display: flex;
    color: var(--highlight);
    font-size: 1.8em;
    font-weight: 700;
  }
  .values,
  .labels {
    margin: 0 0.5rem;
  }
  .value {
    flex: 1;
    font-size: 1.1em;
    padding: 20px 20px 4px;
  }
  .labels {
    display: flex;
    font-size: 0.8em;
    color: var(--dim-text);
  }
  .label {
    flex: 1;
  }

  .milliseconds {
    color: var(--dim-text);
    font-size: 0.8em;
    margin-left: 4px;
  }

  .median {
    font-size: 1em;
  }
  .upper-quartile,
  .lower-quartile {
    font-size: 1em;
    padding-bottom: 0;
  }

  .bar {
    padding: 20px 0 20px;
    display: flex;
    height: 30px;
    width: 85%;
    margin: auto;
    align-items: center;
    position: relative;
  }
  .bar-green {
    background: var(--highlight);
    width: 75%;
    height: 10px;
    border-radius: 2px 0 0 2px;
  }
  .bar-yellow {
    width: 15%;
    height: 10px;
    background: rgb(235, 235, 129);
  }
  .bar-red {
    width: 20%;
    height: 10px;
    border-radius: 0 2px 2px 0;
    background: rgb(228, 97, 97);
  }
  .marker {
    position: absolute;
    height: 30px;
    width: 5px;
    background: white;
    border-radius: 2px;
    left: 0; /* Changed during runtime to reflect median */
  }
</style>
