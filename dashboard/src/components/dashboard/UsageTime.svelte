<script lang="ts">
  import { onMount } from 'svelte';
  import { CREATED_AT } from '../../lib/consts';

  function defaultLayout() {
    return {
      font: { size: 12 },
      paper_bgcolor: 'transparent',
      height: 500,
      margin: { r: 35, l: 70, t: 20, b: 50, pad: 0 },
      polar: {
        bargap: 0,
        bgcolor: 'transparent',
        angularaxis: { direction: 'clockwise', showgrid: false },
        radialaxis: { gridcolor: '#303030' },
      },
    };
  }

  function bars() {
    const responseTimes = Array(24).fill(0);

    for (let i = 0; i < data.length; i++) {
      const date = data[i][CREATED_AT];
      const time = date.getHours();
      // @ts-ignore
      responseTimes[time]++;
    }

    const requestFreqArr: { hour: number; responseTime: number }[] = [];
    for (let i = 0; i < 24; i++) {
      requestFreqArr.push({ hour: i, responseTime: responseTimes[i] });
    }
    requestFreqArr.sort((a, b) => {
      return a.hour - b.hour;
    });

    let dates = [];
    let requests = [];
    for (let i = 0; i < requestFreqArr.length; i++) {
      dates.push(requestFreqArr[i].hour.toString() + ':00');
      requests.push(requestFreqArr[i].responseTime);
    }

    // Shift to 12 onwards to make barpolar like clock face
    dates = dates.slice(12).concat(...dates.slice(0, 12));
    requests = requests.slice(12).concat(...requests.slice(0, 12));

    return [
      {
        r: requests,
        theta: dates,
        marker: { color: '#3fcf8e' },
        type: 'barpolar',
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
    const plotData = buildPlotData();
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

<style scoped>
  .card {
    width: 100%;
    margin: 0;
  }
</style>
