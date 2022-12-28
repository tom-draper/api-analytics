<script lang="ts">
  import { onMount } from "svelte";

  function periodToMarkers(period: string): number {
    if (period == "24h") {
      return 24 * 2;
    } else if (period == "7d") {
      return 12 * 7;
    } else if (period == "30d") {
      return 30 * 4;
    } else if (period == "60d") {
      return 60 * 2;
    } else {
      return null;
    }
  }

  function defaultLayout() {
    return {
      title: false,
      autosize: true,
      margin: { r: 35, l: 45, t: 0, b: 30, pad: 0 },
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

    let dates = [];
    for (let i = 0; i < markers; i++) {
      dates.push(i);
    }

    let requests = [];
    for (let i = 0; i < markers; i++) {
      requests.push(data[i].responseTime);
    }

    return [
      {
        x: dates,
        y: requests,
        type: "lines",
        marker: { color: "#707070" },
        fill: "tonexty",
        hovertemplate: `<b>%{y:.1f}ms avg</b><br>%{x|%d %b %Y}</b><extra></extra>`,
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
