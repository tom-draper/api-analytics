<script lang="ts">
  import frameworkExamples from "../lib/framework";
  import Footer from "../components/Footer.svelte";

  import codeStyle from "svelte-highlight/styles/a11y-dark";
  import { onMount } from "svelte";

  function setFramework(value: string) {
    framework = value;
  }

  function animate() {
    translation = -translation
    let el = document.getElementById('hover-1');
    el.style.transform = `translateY(${translation}%)`;
    let el2 = document.getElementById('hover-2');
    el2.style.transform = `translateY(${-translation}%)`;

    setTimeout(animate, 9000);
  }
  
  let translation = 3.7;
  onMount(() => {
    setTimeout(animate, 10)
  })

  let framework = "Django";
</script>

<svelte:head>
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
        <img id="hover-1" style="position: absolute;" class="logo" src="img/animated5.png" alt="" />
        <img id="hover-2" style="position: absolute;" class="logo" src="img/animated6.png" alt="" />
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
  <div class="add-middleware">
    <div class="add-middleware-title">Getting Started</div>
    <div class="frameworks">
      <button
        class="framework python"
        class:active={framework == "Django"}
        on:click={() => {
          setFramework("Django");
        }}>Django</button
      >
      <button
        class="framework python"
        class:active={framework == "Flask"}
        on:click={() => {
          setFramework("Flask");
        }}>Flask</button
      >
      <button
        class="framework python"
        class:active={framework == "FastAPI"}
        on:click={() => {
          setFramework("FastAPI");
        }}>FastAPI</button
      >
      <button
        class="framework python"
        class:active={framework == "Tornado"}
        on:click={() => {
          setFramework("Tornado");
        }}>Tornado</button
      >
      <button
        class="framework javascript"
        class:active={framework == "Express"}
        on:click={() => {
          setFramework("Express");
        }}>Express</button
      >
      <button
        class="framework javascript"
        class:active={framework == "Fastify"}
        on:click={() => {
          setFramework("Fastify");
        }}>Fastify</button
      >
      <button
        class="framework javascript"
        class:active={framework == "Koa"}
        on:click={() => {
          setFramework("Koa");
        }}>Koa</button
      >
      <button
        class="framework golang"
        class:active={framework == "Gin"}
        on:click={() => {
          setFramework("Gin");
        }}>Gin</button
      >
      <button
        class="framework golang"
        class:active={framework == "Echo"}
        on:click={() => {
          setFramework("Echo");
        }}>Echo</button
      >
      <button
        class="framework golang"
        class:active={framework == "Fiber"}
        on:click={() => {
          setFramework("Fiber");
        }}>Fiber</button
      >
      <button
        class="framework golang"
        class:active={framework == "Chi"}
        on:click={() => {
          setFramework("Chi");
        }}>Chi</button
      >
      <button
        class="framework rust"
        class:active={framework == "Actix"}
        on:click={() => {
          setFramework("Actix");
        }}>Actix</button
      >
      <button
        class="framework rust"
        class:active={framework == "Axum"}
        on:click={() => {
          setFramework("Axum");
        }}>Axum</button
      >
      <button
        class="framework ruby"
        class:active={framework == "Rails"}
        on:click={() => {
          setFramework("Rails");
        }}>Rails</button
      >
      <button
        class="framework ruby"
        class:active={framework == "Sinatra"}
        on:click={() => {
          setFramework("Sinatra");
        }}>Sinatra</button
      >
    </div>
    <div class="add-middleware-content">
      <div class="instructions-container">
        <div class="instructions">
          <div class="subtitle">Install</div>
          <code class="installation"
            >{frameworkExamples[framework].install}</code
          >
          <div class="subtitle">Add middleware to API</div>
          <!-- Render all code snippets to apply one-time syntax highlighting -->
          <!-- TODO: dynamic syntax highlight rendering to only render the 
            frameworks clicked on and reduce this code to one line -->
          <div class="code-file">{frameworkExamples[framework].codeFile}</div>
          <code
            id="code"
            class="code language-python"
            style="{framework == 'Django' ? 'display: initial' : ''} "
            >{frameworkExamples["Django"].example}</code
          >
          <code
            id="code"
            class="code language-python"
            style="{framework == 'FastAPI' ? 'display: initial' : ''} "
            >{frameworkExamples["FastAPI"].example}</code
          >
          <code
            id="code"
            class="code language-python"
            style="{framework == 'Flask' ? 'display: initial' : ''} "
            >{frameworkExamples["Flask"].example}</code
          >
          <code
            id="code"
            class="code language-python"
            style="{framework == 'Tornado' ? 'display: initial' : ''} "
            >{frameworkExamples["Tornado"].example}</code
          >
          <code
            id="code"
            class="code language-javascript"
            style="{framework == 'Express' ? 'display: initial' : ''} "
            >{frameworkExamples["Express"].example}</code
          >
          <code
            id="code"
            class="code language-javascript"
            style="{framework == 'Fastify' ? 'display: initial' : ''} "
            >{frameworkExamples["Fastify"].example}</code
          >
          <code
            id="code"
            class="code language-javascript"
            style="{framework == 'Koa' ? 'display: initial' : ''} "
            >{frameworkExamples["Koa"].example}</code
          >
          <code
            id="code"
            class="code language-go"
            style="{framework == 'Gin' ? 'display: initial' : ''} "
            >{frameworkExamples["Gin"].example}</code
          >
          <code
            id="code"
            class="code language-go"
            style="{framework == 'Echo' ? 'display: initial' : ''} "
            >{frameworkExamples["Echo"].example}</code
          >
          <code
            id="code"
            class="code language-go"
            style="{framework == 'Fiber' ? 'display: initial' : ''} "
            >{frameworkExamples["Fiber"].example}</code
          >
          <code
            id="code"
            class="code language-go"
            style="{framework == 'Chi' ? 'display: initial' : ''} "
            >{frameworkExamples["Chi"].example}</code
          >
          <code
            id="code"
            class="code language-rust"
            style="{framework == 'Actix' ? 'display: initial' : ''} "
            >{frameworkExamples["Actix"].example}</code
          >
          <code
            id="code"
            class="code language-rust"
            style="{framework == 'Axum' ? 'display: initial' : ''} "
            >{frameworkExamples["Axum"].example}</code
          >
          <code
            id="code"
            class="code language-ruby"
            style="{framework == 'Rails' ? 'display: initial' : ''} "
            >{frameworkExamples["Rails"].example}</code
          >
          <code
            id="code"
            class="code language-ruby"
            style="{framework == 'Sinatra' ? 'display: initial' : ''} "
            >{frameworkExamples["Sinatra"].example}</code
          >
        </div>
      </div>
    </div>
  </div>
</div>
<Footer />

<style scoped>
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
    height: 70px;
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
    font-size: 2.1em;
    font-weight: 700;
    margin-bottom: 1.5em;
  }
  .add-middleware {
    margin: auto;
    margin-bottom: 7em;
    border-radius: 6px;
  }

  .frameworks {
    margin: 0 10%;
    overflow-x: auto;
  }
  .framework {
    color: grey;
    background: transparent;
    border: none;
    font-size: 1em;
    cursor: pointer;
    padding: 10px 18px;
    border-bottom: 3px solid transparent;
  }
  .active {
    color: white;
  }
  .active.python {
    border-bottom: 3px solid #4b8bbe;
  }
  .active.golang {
    border-bottom: 3px solid #00a7d0;
  }
  .active.javascript {
    border-bottom: 3px solid #edd718;
  }
  .active.rust {
    border-bottom: 3px solid #ef4900;
  }
  .active.ruby {
    border-bottom: 3px solid #cd0000;
  }
  .subtitle {
    color: rgb(110, 110, 110);
    margin: 10px 0 2px 20px;
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

  img {
   transition:all 10s ease-in-out;
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
  }

  @media screen and (max-width: 900px) {
    .home {
      font-size: 0.85em;
    }
    .instructions-container {
      width: auto;
    }
    .code-file {
      top: 140px;
    }
  }
  @media screen and (max-width: 800px) {
    .home {
      font-size: 0.8em;
    }
    .right {
      margin-top: 2em;
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
      height: 55px;
    }
    .dashboard-img {
      margin-bottom: -3%;
    }
    .add-middleware-title {
      margin-bottom: 0.5em;
    }
    .instructions-container {
      padding-top: 0;
    }
  }
</style>
