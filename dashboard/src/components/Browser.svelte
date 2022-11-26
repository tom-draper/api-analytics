<script lang="ts">
  import { onMount } from "svelte";

  function thisWeek(date: Date): boolean {
    let weekAgo = new Date();
    weekAgo.setDate(weekAgo.getDate() - 7);
    return date > weekAgo;
  }

  function getBrowser(userAgent: string): string {
    if (userAgent.match(/Seamonkey\//)) {
        return 'Seamonkey'
    } else if (userAgent.match(/Firefox\//)) {
        return 'Firefox'
    } else if (userAgent.match(/Chrome\//)) {
        return 'Chrome'
    } else if (userAgent.match(/Chromium\//)) {
        return 'Chromium'
    } else if (userAgent.match(/Safari\//)) {
        return 'Safari'
    } else if (userAgent.match(/Edg\//)) {
        return 'Edge'
    } else if (userAgent.match(/OPR\//) || userAgent.match(/Opera\//)) {
        return 'Opera'
    } else if (userAgent.match(/; MSIE /) || userAgent.match(/Trident\//)) {
        return 'Internet Explorer'
    } else if (userAgent.match(/curl\//)) {
        return 'Curl'
    } else if (userAgent.match(/PostmanRuntime\//)) {
        return 'Postman'
    } else if (userAgent.match(/insomnia\//)) {
        return 'Insomnia'
    } else {
        return 'Other'
    }
  }

  function browserPlotLayout() {
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
      height: 180,
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
    "#3FCF8E",  // Green
    "#E46161",   // Red
    "#EBEB81",  // Yellow
  ];

  function pieChart() {
    let browserCount = {}
    for (let i = 0; i < data.length; i++) {
        let browser = getBrowser(data[i].user_agent)
        if (!(browser in browserCount)) {
            browserCount[browser] = 0
        }
        browserCount[browser]++
    }

    let browsers = [];
    let count = [];
    for (let browser in browserCount) {
        browsers.push(browser);
        count.push(browserCount[browser])
    }
    return [{
    values: count,
        labels: browsers,
        type: 'pie',
          marker: {
            colors: colors
        },
    }];
  }

  function browserPlotData() {
    return {
      data: pieChart(),
      layout: browserPlotLayout(),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  function genPlot() {
    let plotData = browserPlotData();
    //@ts-ignore
    new Plotly.newPlot(
      plotDiv,
      plotData.data,
      plotData.layout,
      plotData.config
    );
  }

//   function build() {
//     let totalRequests = 0;
//     for (let i = 0; i < data.length; i++) {
//       let date = new Date(data[i].created_at);
//       if (thisWeek(date)) {
//         totalRequests++;
//       }
//     }
//     requestsPerHour = ((24 * 7) / totalRequests).toFixed(2);
//   }

  let plotDiv;
  onMount(() => {
    genPlot();
  });

  export let data: any;
</script>

<div class="card" title="Last week">
  <div class="card-title">Browser</div>
  <div id="plotly">
    <div id="plotDiv" bind:this={plotDiv}>
      <!-- Plotly chart will be drawn inside this DIV -->
    </div>
  </div>
</div>

<style>
  .card {
    margin: 2em 2em 2em 0;
    padding-bottom: 1em;
  }
  #plotDiv {
    margin: 0 20px;
  }
</style>
