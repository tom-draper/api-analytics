<script lang="ts">
  import { onMount } from "svelte";
  import periodToDays from "../../../lib/period"

  function defaultLayout() {
    let periodAgo = new Date();
    let days = periodToDays(period);
    if (days != null) {
      periodAgo.setDate(periodAgo.getDate() - days);
    } else {
      periodAgo = null;
    }
    let tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate());

    return {
      title: false,
      autosize: true,
      margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
      hovermode: "closest",
      plot_bgcolor: "transparent",
      paper_bgcolor: "transparent",
      height: 169,
      yaxis: {
        title: { text: "Requests" },
        gridcolor: "gray",
        showgrid: false,
      },
      xaxis: {
        title: { text: "Date" },
        fixedrange: true,
        range: [periodAgo, tomorrow],
        visible: false,
      },
      dragmode: false,
    };
  }

  function initRequestFreq(): {[date: string]: number} {
    // Populate requestFreq with zeros across date period
    let requestFreq = {};
    let days = periodToDays(period);
    if (days) {
      if (days == 1) {
        // Freq count for every minute
        for (let i = 0; i < 60*24; i++) {
          let date = new Date();
          date.setSeconds(0, 0);
          date.setMinutes(date.getMinutes() - i);
          let dateStr = date.toDateString();
          requestFreq[dateStr] = 0;
        }
      } else {
        // Freq count for every day
        for (let i = 0; i < days; i++) {
          let date = new Date();
          date.setHours(0, 0, 0, 0);
          date.setDate(date.getDate() - i);
          let dateStr = date.toDateString();
          requestFreq[dateStr] = 0;
        }
      }
    }
    return requestFreq;
  }

  function bars() {
    let requestFreq = initRequestFreq();

    let days = periodToDays(period);
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      if (days == 1) {
        date.setSeconds(0, 0);
      } else {
        date.setHours(0, 0, 0, 0);
      }
      let dateStr = date.toDateString();
      if (!(dateStr in requestFreq)) {
        requestFreq[dateStr] = 0;
      }
      requestFreq[dateStr]++;
    }

    // Combine date and frequency count into (x, y) tuples for sorting
    let requestFreqArr = [];
    for (let date in requestFreq) {
      requestFreqArr.push([new Date(date), requestFreq[date]]);
    }
    // Sort by date
    requestFreqArr.sort((a, b) => {
      return a[0] - b[0];
    });
    // Split into two lists
    let dates = [];
    let requests = [];
    for (let i = 0; i < requestFreqArr.length; i++) {
      dates.push(requestFreqArr[i][0]);
      requests.push(requestFreqArr[i][1]);
    }

    return [
      {
        x: dates,
        y: requests,
        type: "bar",
        marker: { color: "#3fcf8e" },
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
    let plotData = buildPlotData();
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

  export let data: RequestsData, period: string;
</script>

<div id="plotly">
  <div id="plotDiv" bind:this={plotDiv}>
    <!-- Plotly chart will be drawn inside this DIV -->
  </div>
</div>
