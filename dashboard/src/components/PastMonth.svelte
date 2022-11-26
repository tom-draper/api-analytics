<script lang="ts">
  import { onMount } from "svelte";

  function pastMonth(date: Date): boolean {
    let monthAgo = new Date();
    monthAgo.setDate(monthAgo.getDate() - 30);
    return date > monthAgo;
  }

  let colors = [
    "#444444",  // Grey (no requests)
    "#E46161",   // Red
    "#F18359",
    "#F5A65A",
    "#F3C966",
    "#EBEB81",  // Yellow
    "#C7E57D",
    "#A1DF7E",
    "#77D884",
    "#3FCF8E",  // Green
  ];

  function daysAgo(date: Date): number {
    let now = new Date();
    return Math.floor((now.getTime() - date.getTime()) / (24 * 60 * 60 * 1000));
  }

  function setSuccessRate() {
    let success = {};
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      if (pastMonth(date)) {
        date.setHours(0, 0, 0, 0);
        // @ts-ignore
        if (!(date in success)) {
          // @ts-ignore
          success[date] = { total: 0, successful: 0 };
        }
        if (data[i].status >= 200 && data[i].status <= 299) {
          // @ts-ignore
          success[date].successful++;
        }
        // @ts-ignore
        success[date].total++;
      }
    }

    let successArr = new Array(60).fill(-0.1);  // -0.1 -> 0
    for (let date in success) {
      let idx = daysAgo(new Date(date));
      successArr[successArr.length - idx] = success[date].successful / success[date].total;
    }
    console.log(successArr)
    successRate = successArr;
  }

  function responseTimePlotLayout() {
    let monthAgo = new Date();
    monthAgo.setDate(monthAgo.getDate() - 30);
    let tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    return {
      title: false,
      autosize: true,
      margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
      hovermode: "closest",
      plot_bgcolor: "transparent",
      paper_bgcolor: "transparent",
      height: 150,
      yaxis: {
        title: { text: "Response time (ms)" },
        gridcolor: "gray",
        showgrid: false,
        // showline: false,
        // zeroline: false,
        fixedrange: true,
        // visible: false,
      },
      xaxis: {
        title: { text: "Date" },
        // linecolor: "black",
        showgrid: false,
        // showline: false,
        fixedrange: true,
        range: [monthAgo, tomorrow],
        visible: false,
      },
      dragmode: false,
    };
  }

  function responseTimeLine() {
    let points = [];
    for (let i = 0; i < data.length; i++) {
      points.push([new Date(data[i].created_at), data[i].response_time]);
    }
    points.sort((a, b) => {
      return a[0] - b[0];
    });
    let dates = [];
    let responses = [];
    for (let i = 0; i < points.length; i++) {
      dates.push(points[i][0]);
      responses.push(points[i][1]);
    }

    return [
      {
        x: dates,
        y: responses,
        mode: "lines",
        line: { color: "#3fcf8e" },
        hovertemplate: `<b>%{y:.1f}ms avg</b><br>%{x|%d %b %Y}</b><extra></extra>`,
        showlegend: false,
      },
    ];
  }

  function responseTimePlotData() {
    return {
      data: responseTimeLine(),
      layout: responseTimePlotLayout(),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  function genResponseTimePlot() {
    let plotData = responseTimePlotData();
    console.log(plotData);
    //@ts-ignore
    new Plotly.newPlot(
      responseTimePlotDiv,
      plotData.data,
      plotData.layout,
      plotData.config
    );
  }

  function requestsFreqPlotLayout() {
    let monthAgo = new Date();
    monthAgo.setDate(monthAgo.getDate() - 30);
    let tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    return {
      title: false,
      autosize: true,
      margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
      hovermode: "closest",
      plot_bgcolor: "transparent",
      paper_bgcolor: "transparent",
      height: 150,
      yaxis: {
        title: { text: "Requests" },
        gridcolor: "gray",
        showgrid: false,
        // showline: false,
        // zeroline: false,
        fixedrange: true,
      },
      xaxis: {
        title: { text: "Date" },
        // linecolor: "black",
        // showgrid: false,
        // showline: false,
        fixedrange: true,
        range: [monthAgo, tomorrow],
        visible: false,
      },
      dragmode: false,
    };
  }

  function requestsFreqLine() {
    let requestFreq = {};
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      date.setHours(0, 0, 0, 0);
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

  function requestsFreqPlotData() {
    return {
      data: requestsFreqLine(),
      layout: requestsFreqPlotLayout(),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  function genRequestsFreqPlot() {
    let plotData = requestsFreqPlotData();
    console.log(plotData);
    //@ts-ignore
    new Plotly.newPlot(
      requestsFreqPlotDiv,
      plotData.data,
      plotData.layout,
      plotData.config
    );
  }

  function build() {
    setSuccessRate();
    genResponseTimePlot();
    genRequestsFreqPlot();
  }

  let successRate: any[];
  let responseTimePlotDiv: HTMLDivElement;
  let requestsFreqPlotDiv: HTMLDivElement;
  onMount(() => {
    build();
  });

  export let data: any;
</script>

<div class="card">
  <div class="card-title">Past Month</div>
  <div id="plotly">
    <div id="requestsFreqPlotDiv" bind:this={requestsFreqPlotDiv}>
      <!-- Plotly chart will be drawn inside this DIV -->
    </div>
  </div>
  <div id="plotlyy">
    <div id="responseTimePlotDiv" bind:this={responseTimePlotDiv}>
      <!-- Plotly chart will be drawn inside this DIV -->
    </div>
  </div>
  <div class="success-rate-container">
    {#if successRate != undefined}
      <div class="success-rate-title">Success rate</div>
      <div class="errors">
        {#each successRate as value, i}
          <div
            class="error"
            style="background: {colors[Math.floor(value * 10) + 1]}"
            title="{(value * 100).toFixed(1)}%"
          />
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .card {
    margin: 0;
    width: 100%;
  }
  .errors {
    display: flex;
    margin-top: 8px;
  }
  .error {
    background: var(--highlight);
    flex: 1;
    height: 40px;
    margin: 0 1px;
    border-radius: 1px;
  }
  .success-rate-container {
    text-align: left;
    font-size: 0.9em;
    color: #707070;
    /* color: rgb(68, 68, 68); */
  }
  .success-rate-container {
    margin: 1.5em 2em 2em;
  }
</style>
