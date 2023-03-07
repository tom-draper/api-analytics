<script lang="ts">
  import { onMount } from "svelte";

  function successRatePlotLayout() {
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
    for (let i = 0; i < data.length; i++) {
      let idx = Math.floor(i / (data.length / n));
      if (data[i].status >= 200 && data[i].status <= 299) {
        y[idx] += 1;
      }
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

  function successRatePlotData() {
    return {
      data: lines(),
      layout: successRatePlotLayout(),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  function genPlot() {
    let plotData = successRatePlotData();
    //@ts-ignore
    new Plotly.newPlot(
      plotDiv,
      plotData.data,
      plotData.layout,
      plotData.config
    );
  }

  function build() {
    let totalRequests = 0;
    let successfulRequests = 0;
    for (let i = 0; i < data.length; i++) {
      if (data[i].status >= 200 && data[i].status <= 299) {
        successfulRequests++;
      }
      totalRequests++;
    }
    if (totalRequests > 0) {
      successRate = (successfulRequests / totalRequests) * 100;
    } else {
      successRate = 100;
    }
    genPlot();
  }

  let plotDiv: HTMLDivElement;
  let successRate: number;
  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: data && mounted && build();

  export let data: RequestsData;
</script>

<div class="card">
  <div class="card-title">Success rate</div>
  {#if successRate != undefined}
    <div
      class="value"
      class:red={successRate <= 75}
      class:yellow={successRate > 75 && successRate < 90}
      class:green={successRate > 90}
    >
      {successRate.toFixed(1)}%
    </div>
  {/if}
  <div id="plotly">
    <div id="plotDiv" bind:this={plotDiv}>
      <!-- Plotly chart will be drawn inside this DIV -->
    </div>
  </div>
</div>

<style scoped>
  .card {
    width: calc(215px - 1em);
    margin: 0 0 0 1em;
    position: relative;
    overflow: hidden;
  }
  .value {
    margin: 20px auto;
    width: fit-content;
    font-size: 1.8em;
    font-weight: 700;
    color: var(--yellow);
    position: inherit;
    z-index: 2;
  }
  .red {
    color: var(--red);
  }
  .yellow {
    color: var(--yellow);
  }
  .green {
    color: var(--highlight);
  }
  #plotly {
    position: absolute;
    width: 110%;
    bottom: 0;
    overflow: hidden;
    margin: 0 -5%;
    z-index: 0;
  }
  @media screen and (max-width: 1030px) {
    .card {
      width: auto;
      flex: 1;
    }
  }
</style>
