<script lang="ts">
  import { onMount } from "svelte";

  // Median and quartiles from StackOverflow answer
  // https://stackoverflow.com/a/55297611/8851732
  const asc = (arr) => arr.sort((a, b) => a - b);

  const sum = (arr) => arr.reduce((a, b) => a + b, 0);

  const mean = (arr) => sum(arr) / arr.length;

  // sample standard deviation
  const std = (arr) => {
    const mu = mean(arr);
    const diffArr = arr.map((a) => (a - mu) ** 2);
    return Math.sqrt(sum(diffArr) / (arr.length - 1));
  };

  const quantile = (arr, q) => {
    const sorted = asc(arr);
    const pos = (sorted.length - 1) * q;
    const base = Math.floor(pos);
    const rest = pos - base;
    if (sorted[base + 1] !== undefined) {
      return sorted[base] + rest * (sorted[base + 1] - sorted[base]);
    } else {
      return sorted[base];
    }
  };

  const q25 = (responseTimes) => quantile(responseTimes, 0.25);
  const q50 = (responseTimes) => quantile(responseTimes, 0.5);
  const q75 = (responseTimes) => quantile(responseTimes, 0.75);

  function build() {
    let responseTimes: number[] = [];
    for (let i = 0; i < data.length; i++) {
      responseTimes.push(data[i].response_time);
    }
    median = q50(responseTimes);
    LQ = q25(responseTimes);
    UQ = q75(responseTimes);
    console.log(median, LQ, UQ);
  }

  let median, LQ, UQ
  onMount(() => {
    build()
  })

  export let data: any;
</script>


<div class="card">
    <div class="title">Response Times <span class="milliseconds">(ms)</span></div>
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
        <div class="bar-green"></div>
        <div class="bar-yellow"></div>
        <div class="bar-red"></div>
    </div>
</div>

<style>
    .card {
        background: #232323;
        color: #ededed;
        /* height: 200px; */
        width: 400px;
        border-radius: 6px;
    }
    .title {
        text-align: left;
        padding: 20px 20px 0;
    }
    .values {
        display: flex;
        color: #3fcf8e;
        font-size: 1.8em;
        font-weight: 700;
    }
    .value {
        flex: 1;
        padding: 20px 20px 4px;
    }
    .labels {
        display: flex;
        font-size: 0.8em;
        color: #707070;
    }
    .label {
        flex: 1;
    }

    .milliseconds {
        color: #707070;
        font-size: 0.8em;
        margin-left: 4px;
    }

    .median {
        font-size: 1.5em;
    }
    .upper-quartile,
    .lower-quartile {
        font-size: 0.9em;
        line-height: 53px;
        padding-bottom: 0;
    }

    .bar {
        padding: 30px 0 25px;
        display: flex;
        height: 10px;
        width: 85%;
        margin: auto;
    }
    .bar-green {
        background: #3fcf8e;
        width: 75%;
        height: 10px;
        border-radius: 3px 0 0 3px;
    }
    .bar-yellow {
        width: 15%;
        height: 10px;
        background: rgb(235, 235, 129);
    }
    .bar-red {
        width: 20%;
        height: 10px;
        border-radius: 0 3px 3px 0;
        background: rgb(228, 97, 97);
    }
</style>