<script lang="ts">
  import { onMount } from 'svelte';
  import { ColumnIndex } from '../../lib/consts';

  // Median and quartiles from StackOverflow answer
  // https://stackoverflow.com/a/55297611/8851732
  const asc = (arr) => arr.sort((a, b) => a - b);
  const sum = (arr) => arr.reduce((a, b) => a + b, 0);
  const mean = (arr) => sum(arr) / arr.length;

  function quantile(arr: number[], q: number) {
    const sorted = asc(arr);
    const pos = (sorted.length - 1) * q;
    const base = Math.floor(pos);
    const rest = pos - base;
    if (sorted[base + 1] != undefined) {
      return sorted[base] + rest * (sorted[base + 1] - sorted[base]);
    } else if (sorted[base] != undefined) {
      return sorted[base];
    }
    return 0;
  }

  function markerPosition(x: number): number {
    // 170.125 ms -> 0
    // 1000 ms -> 100
    const position = Math.log10(x) * 130 - 290;
    if (position < 0) {
      return 0;
    } else if (position > 100) {
      return 100;
    }
    return position;
  }

  function setMarkerPosition(median: number) {
    const position = markerPosition(median);
    marker.style.left = `${position}%`;
  }

  function build() {
    const responseTimes: number[] = new Array(data.length);
    for (let i = 0; i < data.length; i++) {
      responseTimes[i] = data[i][ColumnIndex.ResponseTime];
    }
    LQ = quantile(responseTimes, 0.25);
    median = quantile(responseTimes, 0.5);
    UQ = quantile(responseTimes, 0.75);
    // setMarkerPosition(median);
    genPlot(data);
  }

  function defaultLayout(range: [number, number]) {
    return {
      title: false,
      autosize: true,
      margin: { r: 15, l: 15, t: 5, b: 10, pad: 10 },
      // margin: { r: , l: 0, t: 0, b: 0, pad: 0 },
      hovermode: 'closest',
      plot_bgcolor: 'transparent',
      paper_bgcolor: 'transparent',
      height: 50,
      yaxis: {
        // title: { text: 'Response time (ms)' },
        gridcolor: 'gray',
        showgrid: false,
        fixedrange: true,
        visible: false,
      },
      xaxis: {
        // title: { text: 'Response Time' },
        range: range,
        showgrid: false,
        fixedrange: true,
        visible: false,
      },
      dragmode: false,
    };
  }

  function bars(data: RequestsData) {
    const responseTimesFreq = new Map<number, number>();
    for (let i = 0; i < data.length; i++) {
      const responseTime = Math.round(data[i][ColumnIndex.ResponseTime]) || 0;
      if (responseTimesFreq.has(responseTime)) {
        responseTimesFreq.set(
          responseTime,
          responseTimesFreq.get(responseTime) + 1,
        );
      } else {
        responseTimesFreq.set(responseTime, 0);
      }
    }

    const responseTimes: number[] = [];
    const counts: number[] = [];
    if (responseTimesFreq.size > 0) {
      const minResponseTime = Math.min(...responseTimesFreq.keys());
      const maxResponseTime = Math.max(...responseTimesFreq.keys());

      // Split into two lists
      for (let i = 0; i < maxResponseTime - minResponseTime + 1; i++) {
        responseTimes.push(minResponseTime + i);
        counts.push(responseTimesFreq.get(minResponseTime + i) || 0);
      }
    }

    return [
      {
        x: responseTimes,
        y: counts,
        type: 'bar',
        marker: { color: '#707070' },
        hovertemplate: `<b>%{y}</b><br>%{x:.1f}ms</b><extra></extra>`,
        showlegend: false,
      },
    ];
  }

  function buildPlotData(data: RequestsData) {
    const b = bars(data);
    return {
      data: b,
      layout: defaultLayout([b[0].x[0], b[0].x[b[0].x.length - 1]]),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  function genPlot(data: RequestsData) {
    const plotData = buildPlotData(data);
    //@ts-ignore
    new Plotly.newPlot(
      plotDiv,
      plotData.data,
      plotData.layout,
      plotData.config,
    );
  }

  let median: number;
  let LQ: number;
  let UQ: number;
  let marker: HTMLDivElement;
  let plotDiv: HTMLDivElement;
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
  {#if LQ !== undefined && median !== undefined && UQ !== undefined}
    <div class="values">
      <div class="value lower-quartile">{LQ.toFixed(1)}</div>
      <div class="value median">{median.toFixed(1)}</div>
      <div class="value upper-quartile">{UQ.toFixed(1)}</div>
    </div>
  {/if}
  <div class="labels">
    <div class="label">LQ</div>
    <div class="label">Median</div>
    <div class="label">UQ</div>
  </div>
  <div class="distribution">
    <div id="plotly">
      <div id="plotDiv" bind:this={plotDiv}>
        <!-- Plotly chart will be drawn inside this DIV -->
      </div>
    </div>
  </div>
  <!-- <div class="bar">
    <div class="bar-green" />
    <div class="bar-yellow" />
    <div class="bar-red" />
    <div class="marker" bind:this={marker} />
  </div> -->
</div>

<style scoped>
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

  @media screen and (max-width: 1030px) {
    .card {
      width: auto;
      flex: 1;
      margin: 0 0 2em 0;
    }
  }
</style>
