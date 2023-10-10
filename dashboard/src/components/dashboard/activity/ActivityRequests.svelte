<script lang="ts">
  import { onMount } from 'svelte';
  import  { periodToDays } from '../../../lib/period';
  import { CREATED_AT } from '../../../lib/consts';
  import type { Period } from '../../../lib/settings';

  function defaultLayout() {
    let periodAgo = new Date();
    const days = periodToDays(period);
    if (days != null) {
      periodAgo.setDate(periodAgo.getDate() - days);
    } else {
      periodAgo = null;
    }
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate());

    return {
      title: false,
      autosize: true,
      margin: { r: 35, l: 70, t: 20, b: 20, pad: 10 },
      hovermode: 'closest',
      plot_bgcolor: 'transparent',
      paper_bgcolor: 'transparent',
      height: 169,
      yaxis: {
        title: { text: 'Requests' },
        gridcolor: 'gray',
        showgrid: false,
      },
      xaxis: {
        title: { text: 'Date' },
        fixedrange: true,
        range: [periodAgo, tomorrow],
        visible: false,
      },
      dragmode: false,
    };
  }

  function initRequestFreq(): { [date: string]: number } {
    // Populate requestFreq with zeros across date period
    const requestFreq = {};
    const days = periodToDays(period);
    if (days) {
      if (days === 1) {
        // Freq count for every 5 minute
        for (let i = 0; i < 288; i++) {
          const date = new Date();
          date.setSeconds(0, 0);
          // Round down to multiple of 5
          date.setMinutes(Math.floor(date.getMinutes() / 5) * 5 - i * 5);
          const dateStr = date.toISOString();
          requestFreq[dateStr] = 0;
        }
      } else {
        // Freq count for every day
        for (let i = 0; i < days; i++) {
          const date = new Date();
          date.setHours(0, 0, 0, 0);
          date.setDate(date.getDate() - i);
          const dateStr = date.toISOString();
          requestFreq[dateStr] = 0;
        }
      }
    }
    return requestFreq;
  }

  function bars() {
    const requestFreq = initRequestFreq();

    const days = periodToDays(period);
    for (let i = 1; i < data.length; i++) {
      const date = new Date(data[i][CREATED_AT]);
      if (days === 1) {
        // Round down to multiple of 5
        date.setMinutes(Math.floor(date.getMinutes() / 5) * 5, 0, 0);
      } else {
        date.setHours(0, 0, 0, 0);
      }
      const dateStr = date.toISOString();
      if (!(dateStr in requestFreq)) {
        requestFreq[dateStr] = 0;
      }
      requestFreq[dateStr]++;
    }

    // Combine date and frequency count into (x, y) tuples for sorting
    const requestFreqArr = [];
    for (const date in requestFreq) {
      requestFreqArr.push([new Date(date), requestFreq[date]]);
    }
    // Sort by date
    requestFreqArr.sort((a, b) => {
      return a[0] - b[0];
    });
    // Split into two lists
    const dates = [];
    const requests = [];
    for (let i = 0; i < requestFreqArr.length; i++) {
      dates.push(requestFreqArr[i][0]);
      requests.push(requestFreqArr[i][1]);
    }

    return [
      {
        x: dates,
        y: requests,
        type: 'bar',
        marker: { color: '#3fcf8e' },
        hovertemplate: `<b>%{y} requests</b><br>%{x|%d %b %Y}</b><extra></extra>`,
        showlegend: false,
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

  let plotDiv: HTMLDivElement;
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

  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: data && mounted && genPlot();

  export let data: RequestsData, period: Period;
</script>

<div id="plotly">
  <div id="plotDiv" bind:this={plotDiv}>
    <!-- Plotly chart will be drawn inside this DIV -->
  </div>
</div>
