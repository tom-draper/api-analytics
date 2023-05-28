<script lang="ts">
  import { onMount } from "svelte";

  function periodToMarkers(period: string): number {
    switch (period) {
      case "24h":
        return 38;
      case "7d":
        return 84;
      case "30d":
      case "60d":
        return 120;
      default:
        return null;
    }
  }

  function defaultLayout() {
    return {
      title: false,
      autosize: true,
      margin: { r: 35, l: 55, t: 10, b: 30, pad: 10 },
      hovermode: "closest",
      plot_bgcolor: "transparent",
      paper_bgcolor: "transparent",
      height: 120,
      yaxis: {
        title: null,
        gridcolor: "gray",
        showgrid: false,
        fixedrange: true,
      },
      xaxis: {
        title: { text: "Date" },
        showgrid: false,
        fixedrange: true,
        visible: false,
      },
      dragmode: false,
    };
  }

  function bars() {
    let markers = periodToMarkers(period);

    let dates: Date[] = Array(markers);
    let requests: number[] = Array(markers);
    for (let i = 0; i < markers; i++) {
      requests[markers - i - 1] = data[i].responseTime;
      dates[markers - i - 1] = data[i].createdAt;
    }

    for (let i = 0; i < dates.length; i++) {
      if (dates[i] === null) {
        if (i === 0) {
          dates[i] = new Date();
        } else {
          // 30 mins from previous date
          dates[i] = new Date(dates[i - 1]);
          dates[i].setMinutes(dates[i].getMinutes() - 30);
        }
      }
    }

    return [
      {
        x: dates,
        y: requests,
        type: "lines",
        marker: { color: "#707070" },
        fill: "tonexty",
        hovertemplate: `<b>%{y:.0f}ms</b><br>%{x|%d %b %Y %H:%M:%S}</b><extra></extra>`,
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
  let setup = false;
  onMount(() => {
    genPlot();
    setup = true;
  });

  $: period && setup && genPlot();

  export let data: any[], period: string;
</script>

<div id="plotly">
  <div id="plotDiv" bind:this={plotDiv}>
    <!-- Plotly chart will be drawn inside this DIV -->
  </div>
</div>
