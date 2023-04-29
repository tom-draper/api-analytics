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

    let dates = Array(markers);
    let requests = Array(markers);
    for (let i = 0; i < markers; i++) {
      let now = new Date();
      now.setMinutes(now.getMinutes() - i * 30);
      dates[i] = now.toISOString();
      requests[markers - i] = data[i].responseTime;
    }

    return [
      {
        x: dates,
        y: requests,
        type: "lines",
        marker: { color: "#707070" },
        fill: "tonexty",
        hovertemplate: `<b>%{y:.0f}ms</b><br>%{x|%d %b %Y}</b><extra></extra>`,
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
