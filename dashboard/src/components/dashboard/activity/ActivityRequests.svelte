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

  function bars() {
    let requestFreq = {};
    let days = periodToDays(period);
    if (days == 1) {
      for (let i = 0; i < 60*24; i++) {
        let date = new Date();
        date.setSeconds(0, 0);
        date.setMinutes(date.getMinutes() - i);
        // @ts-ignore
        requestFreq[date] = 0;
      }
    }
    else if (days) {
      for (let i = 0; i < days; i++) {
        let date = new Date();
        date.setHours(0, 0, 0, 0);
        date.setDate(date.getDate() - i);
        // @ts-ignore
        requestFreq[date] = 0;
      }
    }

    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      if (days == 1) {
        date.setSeconds(0, 0);
      } else {
        date.setHours(0, 0, 0, 0);
      }
      // @ts-ignore
      if (!(date in requestFreq)) {
        // @ts-ignore
        requestFreq[date] = 0;
      }
      // @ts-ignore
      requestFreq[date]++;
    }

    let requestFreqArr = [];
    for (let date in requestFreq) {
      requestFreqArr.push([new Date(date), requestFreq[date]]);
    }
    requestFreqArr.sort((a, b) => {
      return a[0] - b[0];
    });

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

  let plotDiv;
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
