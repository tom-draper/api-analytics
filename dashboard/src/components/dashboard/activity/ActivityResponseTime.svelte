<script lang="ts">
  import { onMount } from 'svelte';
  import { periodToDays } from '../../../lib/period';
  import { CREATED_AT, RESPONSE_TIME } from '../../../lib/consts';
  import type { Period } from '../../../lib/settings';
  import { initFreqMap } from '../../../lib/activity';

  function defaultLayout() {
    const days = periodToDays(period);
    let periodAgo = new Date();
    if (days != null) {
      periodAgo.setDate(periodAgo.getDate() - days);
    } else {
      periodAgo = null;
    }
    const tomorrow = new Date();

    return {
      title: false,
      autosize: true,
      margin: { r: 35, l: 70, t: 20, b: 20, pad: 10 },
      hovermode: 'closest',
      plot_bgcolor: 'transparent',
      paper_bgcolor: 'transparent',
      height: 170,
      yaxis: {
        title: { text: 'Response time (ms)' },
        gridcolor: 'gray',
        showgrid: false,
        fixedrange: true,
      },
      xaxis: {
        title: { text: 'Date' },
        showgrid: false,
        fixedrange: true,
        range: [periodAgo, tomorrow],
        visible: false,
      },
      dragmode: false,
    };
  }

  function bars() {
    const responseTimesFreq = initFreqMap(period, () => {
      return { totalResponseTime: 0, count: 0 };
    });

    const days = periodToDays(period);
    if (days === null) {
      return;
    }

    for (let i = 1; i < data.length; i++) {
      const date = new Date(data[i][CREATED_AT]);
      if (days <= 7) {
        // Round down to multiple of 5
        date.setMinutes(Math.floor(date.getMinutes() / 5) * 5, 0, 0);
      } else {
        date.setHours(0, 0, 0, 0);
      }
      const dateStr = date.toISOString();
      if (!(dateStr in responseTimesFreq)) {
        responseTimesFreq.set(dateStr, { totalResponseTime: 0, count: 0 });
      }
      responseTimesFreq.get(dateStr).totalResponseTime +=
        data[i][RESPONSE_TIME];
      responseTimesFreq.get(dateStr).count++;
    }

    // Combine date and avg response time into (x, y) tuples for sorting
    const responseTimeArr: { date: Date; avgResponseTime: number }[] = [];
    for (const [date, obj] of responseTimesFreq.entries()) {
      const point = { date: new Date(date), avgResponseTime: 0 };
      if (obj.count > 0) {
        point.avgResponseTime = obj.totalResponseTime / obj.count;
      }
      responseTimeArr.push(point);
    }

    // Sort by date
    responseTimeArr.sort((a, b) => {
      //@ts-ignore
      return a.date - b.date;
    });

    // Split into two lists
    const dates: Date[] = [];
    const responseTimes: number[] = [];
    let minAvgResponseTime = Number.POSITIVE_INFINITY;
    for (let i = 0; i < responseTimeArr.length; i++) {
      dates.push(responseTimeArr[i].date);
      responseTimes.push(responseTimeArr[i].avgResponseTime);
      if (responseTimeArr[i].avgResponseTime < minAvgResponseTime) {
        minAvgResponseTime = responseTimeArr[i].avgResponseTime;
      }
    }

    return [
      {
        x: dates,
        y: responseTimes,
        type: 'bar',
        marker: { color: '#707070' },
        hovertemplate: `<b>%{y:.1f}ms avg</b><br>%{x|%d %b %Y %H:%M}</b><extra></extra>`,
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

  export let data: RequestsData, period: Period;
</script>

<div id="plotly">
  <div id="plotDiv" bind:this={plotDiv}>
    <!-- Plotly chart will be drawn inside this DIV -->
  </div>
</div>
