<script lang="ts">
  import { onMount } from "svelte";
  import frameworkExamples from "../lib/framework";
  import codeStyle from "svelte-highlight/styles/a11y-dark";
  import Footer from "../components/Footer.svelte";

  function setFramework(value: string) {
    currentFramework = value;
  }

  function animate() {
    translation = -translation;
    let el = document.getElementById("hover-1");
    el.style.transform = `translateY(${translation}%)`;
    let el2 = document.getElementById("hover-2");
    el2.style.transform = `translateY(${-translation}%)`;

    setTimeout(animate, 9000);
  }

  let translation = 3.7;
  onMount(() => {
    setTimeout(animate, 10);
  });

  let frameworks = [
    ["python", "FastAPI"],
    ["python", "Flask"],
    ["python", "Django"],
    ["python", "Tornado"],
    ["javascript", "Express"],
    ["javascript", "Fastify"],
    ["javascript", "Koa"],
    ["go", "Gin"],
    ["go", "Echo"],
    ["go", "Fiber"],
    ["go", "Chi"],
    ["rust", "Actix"],
    ["rust", "Axum"],
    ["ruby", "Rails"],
    ["ruby", "Sinatra"],
  ];
  let currentFramework = frameworks[0][1];
</script>

<svelte:head>
  <link href="prism.css" rel="stylesheet" />
  <script src="prism.js"></script>
  {@html codeStyle}
</svelte:head>

<div class="home">
  <div class="landing-page-container">
    <div class="landing-page">
      <div class="left">
        <h1>API Analytics</h1>
        <h2>Monitoring and analytics for API frameworks.</h2>
        <div class="links">
          <a href="/generate" class="link">
            <div class="text">
              Try now – it's <span class="italic">free</span>
            </div></a
          >
          <a href="/dashboard/demo" class="link secondary">
            <div class="text">Demo</div></a
          >
        </div>
      </div>
      <div style="position: relative" class="right">
        <img class="logo" src="img/home-logo2.png" alt="" />
        <img
          id="hover-1"
          style="position: absolute;"
          class="logo animated"
          src="img/animated5.png"
          alt=""
        />
        <img
          id="hover-2"
          style="position: absolute;"
          class="logo animated"
          src="img/animated6.png"
          alt=""
        />
      </div>
    </div>
  </div>

  <div class="dashboard">
    <div class="dashboard-title-container">
      <img class="lightning-top" src="img/logo.png" alt="" />
      <h1 class="dashboard-title">Dashboard</h1>
    </div>
    <div class="dashboard-content">
      <div class="dashboard-content-text">
        An all-in-one analytics dashboard. Real-time insight into your API's
        usage.
      </div>
      <div class="dashboard-btn-container">
        <a href="/dashboard" class="dashboard-btn secondary">
          <div class="dashboard-btn-text">Open</div>
        </a>
      </div>
    </div>
    <img class="dashboard-img" src="img/dashboard.png" alt="" />
  </div>
  <div class="dashboard">
    <div class="dashboard-title-container">
      <img class="lightning-top" src="img/logo.png" alt="" />
      <h1 class="dashboard-title">Monitoring</h1>
    </div>
    <div class="dashboard-content">
      <div class="dashboard-content-text">
        Active monitoring and error notifications. Peace of mind.
      </div>
      <div class="dashboard-btn-container">
        <a href="/" class="dashboard-btn secondary">
          <div class="dashboard-btn-text">Coming Soon</div>
        </a>
      </div>
    </div>
    <img class="dashboard-img" src="img/monitoring.png" alt="" />
  </div>
  <div class="test-add-middleware-content">
    <div class="test-left-instructions-container"></div>
    <div class="test-right-instructions-container"></div>
    <div class="radial-dimmer"></div>
    <div class="linear-dimmer"></div>
  </div>
  <div class="add-middleware">
    <div class="add-middleware-title">Getting Started</div>
    <div class="frameworks">
      {#each frameworks as [language, framework]}
        <button
          class="framework {language}"
          class:active={currentFramework == framework}
          on:click={() => {
            setFramework(framework);
          }}>{framework}</button
        >
      {/each}
    </div>
    <div class="add-middleware-content">
      <div class="instructions-container">
        <div class="instructions">
          <div class="subtitle">Install</div>
          <code class="installation"
            >{frameworkExamples[currentFramework].install}</code
          >
          <div class="subtitle">Add middleware to API</div>
          <!-- Render all code snippets to apply one-time syntax highlighting -->
          <!-- TODO: dynamic syntax highlight rendering to only render the 
            frameworks clicked on and reduce this code to one line -->
          <div class="code-file">
            {frameworkExamples[currentFramework].codeFile}
          </div>
          {#each frameworks as [language, framework]}
            <code
              id="code"
              class="code language-{language}"
              style="{currentFramework == framework ? 'display: initial' : ''} "
              >{frameworkExamples[framework].example}</code
            >
          {/each}
        </div>
      </div>
    </div>
  </div>
</div>
<Footer />

<style scoped>
  .test-add-middleware-content {
    position: relative;
    display: flex;
    height: 800px;
    z-index: -1;
    margin-top: -15em;
  }
  .test-left-instructions-container {
    width: 100%;
    height: 100%;
    margin-left: auto;
    background: conic-gradient(from 270deg, var(--highlight),  #1c1c1c);
    transform: scale(-1, 1);
    margin-right: -1px;
    filter: brightness(1.1);
  }
  .test-right-instructions-container {
    filter: brightness(1.1);
    width: 100%;
    height: 100%;
    margin-right: auto;
    background: conic-gradient(from 270deg, var(--highlight), #1c1c1c);
    margin-left: -1px;
  }
  .radial-dimmer {
    position: absolute;
    width: 100%;
    height: 100%;
    background: radial-gradient(transparent, #1c1c1c);
  }
  .linear-dimmer {
    position: absolute;
    width: 100%;
    height: 100%;
    background: linear-gradient(#1c1c1c, transparent, #1c1c1c);
  }

  .landing-page {
    display: flex;
    place-items: center;
    padding-bottom: 6em;
  }
  .landing-page-container {
    display: grid;
    margin: 0 12% 0 15%;
    min-height: 100vh;
  }
  .left {
    flex: 1;
    text-align: left;
  }
  .right {
    display: grid;
    place-items: center;
  }
  .logo {
    max-width: 1400px;
    width: 700px;
    margin-bottom: -50px;
  }

  h1 {
    font-size: 3.4em;
  }

  h2 {
    color: white;
    font-size: 1.5em;
  }

  .links {
    color: #707070;
    display: flex;
    margin-top: 30px;
    text-align: left;
  }
  .link {
    width: fit-content;
    margin-right: 20px;
    font-size: 0.9;
  }
  .link:hover {
    background: #31aa73;
  }

  .secondary {
    background: #1c1c1c;
    border: 3px solid var(--highlight);
    color: var(--highlight);
  }
  .secondary:hover {
    background: #081d13;
  }

  .lightning-top {
    height: 65px;
  }

  .dashboard-title-container {
    display: flex;
    margin: 2em 4em;
  }

  a {
    background: var(--highlight);
    color: black;
    padding: 10px 20px;
    border-radius: 4px;
  }
  .italic {
    font-style: italic;
  }
  .dashboard {
    border: 3px solid var(--highlight);
    width: 80%;
    border-radius: 10px;
    margin: auto;
    margin-bottom: 8em;
    overflow: hidden;
    position: relative;
  }
  .dashboard-title {
    font-size: 2.5em;
    text-align: left;
    margin: 0.2em 1em auto 1em;
  }
  .dashboard-title-container {
    place-content: center;
  }
  .dashboard-img {
    width: 81%;
    border-radius: 10px;
    box-shadow: 0px 24px 120px -25px var(--highlight);
    margin-bottom: -1%;
  }
  .dashboard-content {
    text-align: left;
    color: white;
    margin: 0 5em 4em;
  }
  .dashboard-content-text {
    font-size: 1.1em;
    text-align: center;
  }
  .dashboard-btn-container {
    display: flex;
    justify-content: center;
    margin-top: 2em;
  }
  .dashboard-btn-text {
    text-align: center;
  }

  .add-middleware-title {
    color: var(--highlight);
    color: black;
    font-size: 2.5em;
    font-weight: 700;
    margin-bottom: 1.5em;
  }
  .add-middleware {
    margin: auto;
    margin-bottom: 7em;
    border-radius: 6px;
    margin-top: -480px;
  }

  .frameworks {
    margin: 0 25%;
    overflow-x: auto;
  }
  .framework {
    color: #919191;
    /* color: black; */
    background: transparent;
    border: none;
    font-size: 1em;
    cursor: pointer;
    padding: 8px 13px;
    height: auto;
    width: auto;
    border: 3px solid transparent;
    border-radius: 4px;
  }
  .active {
    color: white;
  }
  .active.python {
    border: 3px solid #4b8bbe;
  }
  .active.go {
    border: 3px solid #00a7d0;
  }
  .active.javascript {
    border: 3px solid #edd718;
  }
  .active.rust {
    border: 3px solid #ef4900;
  }
  .active.ruby {
    border: 3px solid #cd0000;
  }
  .active.php {
    border: 3px solid #7377ad;
  }
  .subtitle {
    color: #919191;
    margin: 10px 0 2px 16px;
    font-size: 0.85em;
  }
  .instructions-container {
    padding: 1.5em 2em 2em;
    width: 850px;
    margin: auto;
  }
  .instructions {
    text-align: left;
    display: flex;
    flex-direction: column;
    position: relative;
  }
  code {
    background: #151515;
    padding: 1.4em 2em;
    border-radius: 0.5em;
    margin: 5px;
    color: #dcdfe4;
    white-space: pre-wrap;
    overflow: auto;
  }
  .code {
    display: none;
  }
  .code-file {
    position: absolute;
    font-size: 0.8em;
    top: 160px;
    color: rgb(97, 97, 97);
    text-align: right;
    right: 2.5em;
    margin-bottom: -2em;
  }

  #hover-1 {
    transform: translateY(3.7%);
  }
  #hover-2 {
    transform: translateY(-3.7%);
  }

  .animated {
    transition: all 10s ease-in-out;
  }

  @media screen and (min-width: 1700px) {
    .frameworks {
      margin: 0 10%;
    }
  }

  @media screen and (max-width: 1500px) {
    .landing-page-container {
      margin: 0 6% 0 7%;
    }
  }

  @media screen and (max-width: 1300px) {
    .landing-page-container {
      margin: 0 5% 0 6%;
    }
  }

  @media screen and (max-width: 1200px) {
    .landing-page {
      flex-direction: column-reverse;
    }
    .landing-page-container {
      margin: 0 2em;
    }
    .dashboard {
      width: 90%;
      margin-bottom: 4em;
    }
    .logo {
      width: 100%;
    }
    .instructions-container {
      width: auto;
    }
  }

  @media screen and (max-width: 900px) {
    .home {
      font-size: 0.85em;
    }
    .code-file {
      top: 140px;
    }
    .add-middleware {
      margin-top: -465px;
    }
    .frameworks {
      margin: 0 5%;
    }
    .test-add-middleware-content {
      margin-top: -18em;
    }
  }
  @media screen and (max-width: 800px) {
    .home {
      font-size: 0.8em;
    }
    .right {
      margin-top: 2em;
    }
    .test-add-middleware-content {
      margin-top: -22em;
    }
  }
  @media screen and (max-width: 700px) {
    h1 {
      font-size: 2.5em;
    }
    h2 {
      font-size: 1.2em;
    }
    .landing-page-container {
      min-height: unset;
    }
    .landing-page {
      padding-bottom: 8em;
    }
    .lightning-top {
      height: 40px;
      margin-top: auto;
    }
    .dashboard-title {
      margin: 0.2em 0.5em auto 1em;
    }
    .dashboard-img {
      margin-bottom: -3%;
    }
    .dashboard-content {
      margin: 0 2em 4em;
    }
    .add-middleware-title {
      margin-bottom: 1em;
    }
    .instructions-container {
      padding: 0 4%;
    }
    .logo {
      margin-bottom: 0;
    }
    .test-add-middleware-content {
      margin-top: -24em;
      overflow: hidden;
    }

    .add-middleware {
      margin-top: -455px;
    }
    .test-left-instructions-container {
      width: 90%;
      margin-left: -40%;
    }
    .test-right-instructions-container {
      width: 90%;
      margin-right: -40%;
    }
  }
  
  @media screen and (max-width: 500px) {
    .test-add-middleware-content {
      margin-top: -24em;
      overflow: hidden;
    }
  }
</style>
