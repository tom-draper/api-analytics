<script lang="ts">
  import { onMount } from 'svelte';
  import { periodToDays } from '../../../lib/period';
  import { CREATED_AT, IP_ADDRESS } from '../../../lib/consts';
  import type { Period } from '../../../lib/settings';
  import { initFreqMap } from '../../../lib/activity';

  function defaultLayout() {
    const days = periodToDays(period);
    let periodAgo = new Date();
    if (days === null) {
      periodAgo = null;
    } else {
      periodAgo.setDate(periodAgo.getDate() - days);
    }
    const now = new Date();

    return {
      title: false,
      autosize: true,
      margin: { r: 35, l: 70, t: 20, b: 20, pad: 10 },
      hovermode: 'closest',
      plot_bgcolor: 'transparent',
      paper_bgcolor: 'transparent',
      height: 169,
      barmode: 'stack',
      yaxis: {
        title: { text: 'Requests' },
        gridcolor: 'gray',
        showgrid: false,
      },
      xaxis: {
        title: { text: 'Date' },
        fixedrange: true,
        range: [periodAgo, now],
        visible: false,
      },
      dragmode: false,
    };
  }

  function bars(data: RequestsData, period: Period) {
    const requestFreq = initFreqMap(period, () => 0);
    const userFreq = initFreqMap(period, () => new Set());

    const days = periodToDays(period);
    for (let i = 1; i < data.length; i++) {
      const date = new Date(data[i][CREATED_AT]);
      if (days !== null && days <= 7) {
        // Round down to multiple of 5
        date.setMinutes(Math.floor(date.getMinutes() / 5) * 5, 0, 0);
      } else {
        date.setHours(0, 0, 0, 0);
      }
      const ipAddress = data[i][IP_ADDRESS];
      const time = date.getTime();
      if (!userFreq.has(time)) {
        userFreq.set(time, new Set());
      }
      userFreq.get(time).add(ipAddress);

      if (!requestFreq.has(time)) {
        requestFreq.set(time, 0);
      }
      requestFreq.set(time, requestFreq.get(time) + 1);
    }

    // Combine date and frequency count into (x, y) tuples for sorting
    const requestFreqArr = [];
    for (const [time, requestsCount] of requestFreq.entries()) {
      const userCount = userFreq.has(time) ? userFreq.get(time).size : 0;
      requestFreqArr.push({
        date: time,
        requestCount: requestsCount,
        userCount: userCount,
      });
    }
    // Sort by date
    requestFreqArr.sort((a, b) => {
      return a.date - b.date;
    });

    // Split into two lists
    const dates = [];
    const requests = [];
    const requestsText = [];
    const users = [];
    const usersText = [];
    for (let i = 0; i < requestFreqArr.length; i++) {
      dates.push(new Date(requestFreqArr[i].date));
      // Subtract users due to bar stacking
      requests.push(
        requestFreqArr[i].requestCount - requestFreqArr[i].userCount
      );
      // Keep actual requests count for hover text
      requestsText.push(`${requestFreqArr[i].requestCount} requests`);
      users.push(requestFreqArr[i].userCount);
      usersText.push(
        `${requestFreqArr[i].userCount} users from ${requestFreqArr[i].requestCount} requests`
      );
    }

    return [
      {
        x: dates,
        y: users,
        text: usersText,
        type: 'bar',
        marker: { color: '#3fcf8e' },
        hovertemplate: `<b>%{text}</b><br>%{x|%d %b %Y %H:%M}</b><extra></extra>`,
        showlegend: false,
      },
      {
        x: dates,
        y: requests,
        text: requestsText,
        type: 'bar',
        marker: { color: '#228458' },
        hovertemplate: `<b>%{text}</b><br>%{x|%d %b %Y %H:%M}</b><extra></extra>`,
        showlegend: false,
      },
    ];
  }

  function buildPlotData(data: RequestsData, period: Period) {
    return {
      data: bars(data, period),
      layout: defaultLayout(),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  let plotDiv: HTMLDivElement;
  function genPlot(data: RequestsData, period: Period) {
    const plotData = buildPlotData(data, period);
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

  $: data && mounted && genPlot(data, period);

  export let data: RequestsData, period: Period;
</script>

<div id="plotly">
  <div id="plotDiv" bind:this={plotDiv}>
    <!-- Plotly chart will be drawn inside this DIV -->
  </div>
</div>
