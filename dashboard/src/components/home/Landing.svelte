<script lang="ts">
  import { onMount } from "svelte";

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
</script>

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
    color: var(--dim-text);
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
    background: var(--background);
    border: 3px solid var(--highlight);
    color: var(--highlight);
  }
  .secondary:hover {
    background: var(--dark-green);
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

  #hover-1 {
    transform: translateY(3.7%);
  }
  #hover-2 {
    transform: translateY(-3.7%);
  }

  .animated {
    transition: all 10s ease-in-out;
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

    .logo {
      width: 100%;
    }
  }

  @media screen and (max-width: 800px) {
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

    .logo {
      margin-bottom: 0;
    }
  }
</style>
