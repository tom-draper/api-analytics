<script lang="ts">
  import { onMount } from "svelte";

  function getDevice(userAgent: string): string {
    if (userAgent === null) {
      return "Unknown";
    } else if (userAgent.match(/iPhone/)) {
      return "iPhone";
    } else if (userAgent.match(/Android/)) {
      return "Android";
    } else if (userAgent.match(/Tizen\//)) {
      return "Samsung";
    } else {
      return "Other";
    }
  }

  function devicePlotLayout() {
    return {
      title: false,
      autosize: true,
      margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
      hovermode: "closest",
      plot_bgcolor: "transparent",
      paper_bgcolor: "transparent",
      height: 180,
      width: 411,
      yaxis: {
        title: { text: "Requests" },
        gridcolor: "gray",
        showgrid: false,
        fixedrange: true,
      },
      xaxis: {
        visible: false,
      },
      dragmode: false,
    };
  }

  let colors = [
    "#3FCF8E", // Green
    "#E46161", // Red
    "#EBEB81", // Yellow
  ];

  function pieChart() {
    let deviceCount = {};
    for (let i = 1; i < data.length; i++) {
      let browser = getDevice(data[i][2]);
      if (!(browser in deviceCount)) {
        deviceCount[browser] = 0;
      }
      deviceCount[browser]++;
    }

    let device = [];
    let count = [];
    for (let browser in deviceCount) {
      device.push(browser);
      count.push(deviceCount[browser]);
    }
    return [
      {
        values: count,
        labels: device,
        type: "pie",
        marker: {
          colors: colors,
        },
      },
    ];
  }

  function devicePlotData() {
    return {
      data: pieChart(),
      layout: devicePlotLayout(),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  function genPlot() {
    let plotData = devicePlotData();
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

<div id="plotly">
  <div id="plotDiv" bind:this={plotDiv}>
    <!-- Plotly chart will be drawn inside this DIV -->
  </div>
</div>

<style>
  #plotDiv {
    margin-right: 20px;
  }
</style>
