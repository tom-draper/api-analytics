<script lang="ts">
  import { onMount } from "svelte";

  function defaultLayout() {
    return {
      font: { size: 12 },
      paper_bgcolor: "transparent",
      height: 500,
      margin: { r: 35, l: 70, t: 20, b: 50, pad: 0 },
      polar: {
          bargap: 0,
          bgcolor: "transparent",
        angularaxis: { direction: "clockwise", showgrid: false},
        radialaxis: { gridcolor: "#303030"}
      },
    };
  }

  function bars() {
    let responseTimes = Array(24).fill(0);

    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      let time = date.getHours();
      // @ts-ignore
      responseTimes[time]++;
    }

    let requestFreqArr = [];
    for (let i = 0; i < 24; i++) {
      let point = [i, responseTimes[i]];
      requestFreqArr.push(point);
    }
    requestFreqArr.sort((a, b) => {
      return a[0] - b[0];
    });

    let dates = [];
    let requests = [];
    for (let i = 0; i < requestFreqArr.length; i++) {
      dates.push(requestFreqArr[i][0].toString() + ':00');
      requests.push(requestFreqArr[i][1]);
    }

    return [
      {
        r: requests,
        theta: dates,
        marker: { color: "#3fcf8e" },
        type: "barpolar",
        hovertemplate: `<b>%{r}</b> requests at <b>%{theta}</b><extra></extra>`,
      },
    ];
  }

  function buildPlotData() {
    return {
      data: bars(),
      layout: defaultLayout(),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  function genPlot() {
    let plotData = buildPlotData();
    //@ts-ignore
    new Plotly.newPlot(
      plotDiv,
      plotData.data,
      plotData.layout,
      plotData.config
    );
  }

  let plotDiv: HTMLDivElement;
  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: data && mounted && genPlot();
  export let data: RequestsData;
</script>

<div class="card">
    <div class="card-title">Usage time</div>
  <div id="plotly">
    <div id="plotDiv" bind:this={plotDiv}>
      <!-- Plotly chart will be drawn inside this DIV -->
    </div>
  </div>
</div>

<style>
  .card {
    width: 100%;
    margin: 0;
  }
</style>
