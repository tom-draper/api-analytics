<script lang="ts">
  import { onMount } from "svelte";
  import periodToDays from "../../lib/period";
  import type { Period } from "../../lib/settings";

  function requestsPlotLayout() {
    return {
      title: false,
      autosize: true,
      margin: { r: 0, l: 0, t: 0, b: 0, pad: 0 },
      hovermode: false,
      plot_bgcolor: "transparent",
      paper_bgcolor: "transparent",
      height: 60,
      yaxis: {
        gridcolor: "gray",
        showgrid: false,
        fixedrange: true,
        dragmode: false,
      },
      xaxis: {
        visible: false,
        dragmode: false,
      },
      dragmode: false,
    };
  }

  function lines() {
    let n = 5;
    let x = [...Array(n).keys()];
    let y = Array(n).fill(0);
    for (let i = 1; i < data.length; i++) {
      let idx = Math.floor(i / (data.length / n));
      y[idx] += 1;
    }
    return [
      {
        x: x,
        y: y,
        type: "lines",
        marker: { color: "transparent" },
        showlegend: false,
        line: { shape: "spline", smoothing: 1, color: "#3FCF8E30" },
        fill: "tozeroy",
        fillcolor: "#3fcf8e15",
      },
    ];
  }

  function requestsPlotData() {
    return {
      data: lines(),
      layout: requestsPlotLayout(),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  function genPlot() {
    let plotData = requestsPlotData();
    //@ts-ignore
    new Plotly.newPlot(
      plotDiv,
      plotData.data,
      plotData.layout,
      plotData.config
    );
  }

  function setPercentageChange() {
    if (prevData.length == 0) {
      percentageChange = null;
    } else {
      percentageChange = (data.length / prevData.length) * 100 - 100;
    }
  }

  function setRequestsPerHour() {
    if (data.length > 0) {
      let days = periodToDays(period);
      if (days != null) {
        requestsPerHour = (data.length / (24 * days)).toFixed(2);
      }
    } else {
      requestsPerHour = "0";
    }
  }

  function togglePeriod() {
    perHour = !perHour;
  }

  function build() {
    setPercentageChange();
    setRequestsPerHour();
    genPlot();
  }

  let plotDiv: HTMLDivElement;
  let requestsPerHour: string;
  let perHour = false;
  let percentageChange: number;
  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: data && mounted && build();

  export let data: RequestsData, prevData: RequestsData, period: Period;
</script>

<button class="card" on:click={togglePeriod}>
  {#if perHour}
    <div class="card-title">
      Requests <span class="per-hour">/ hour</span>
    </div>
    {#if requestsPerHour != undefined}
      <div class="value">{requestsPerHour}</div>
    {/if}
  {:else}
    {#if percentageChange != null}
      <div
        class="percentage-change"
        class:positive={percentageChange > 0}
        class:negative={percentageChange < 0}
      >
        {#if percentageChange > 0}
          <img class="arrow" src="../img/up.png" alt="" />
        {:else if percentageChange < 0}
          <img class="arrow" src="../img/down.png" alt="" />
        {/if}
        {Math.abs(percentageChange).toFixed(1)}%
      </div>
    {/if}
    <div class="card-title">Requests</div>
    <div class="value">{data.length.toLocaleString()}</div>
  {/if}
  <div id="plotly">
    <div id="plotDiv" bind:this={plotDiv}>
      <!-- Plotly chart will be drawn inside this DIV -->
    </div>
  </div>
</button>

<style scoped>
  .card {
    width: calc(215px - 1em);
    margin: 0 1em 0 0;
    position: relative;
    cursor: pointer;
    padding: 0;
    overflow: hidden;
  }
  .value {
    margin: 20px auto;
    width: fit-content;
    font-size: 1.8em;
    font-weight: 700;
  }
  .percentage-change {
    position: absolute;
    right: 20px;
    top: 20px;
    font-size: 0.8em;
  }
  .positive {
    color: var(--highlight);
  }
  .negative {
    color: rgb(228, 97, 97);
  }

  .per-hour {
    color: var(--dim-text);
    font-size: 0.8em;
    margin-left: 4px;
  }
  button {
    font-size: unset;
    font-family: unset;
  }
  .arrow {
    height: 11px;
  }
  #plotly {
    position: absolute;
    width: 110%;
    bottom: 0;
    overflow: hidden;
    margin: 0 -5%;
  }
  @media screen and (max-width: 1030px) {
    .card {
      width: auto;
      flex: 1;
      margin: 0 1em 0 0;
    }
  }
</style>
