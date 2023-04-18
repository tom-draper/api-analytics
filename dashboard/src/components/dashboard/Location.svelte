<script lang="ts">
  import { onMount } from "svelte";

  function getFlagEmoji(countryCode) {
    const codePoints = countryCode
      .toUpperCase()
      .split("")
      .map((char) => 127397 + char.charCodeAt());
    return String.fromCodePoint(...codePoints);
  }

  function countryCodeToName(countryCode) {
    let regionNames = new Intl.DisplayNames(["en"], { type: "region" });
    return regionNames.of(countryCode);
  }

  function build() {
    let max = 0;
    let locationsFreq = {};
    for (let i = 0; i < data.length; i++) {
      if (!(data[i][6] in locationsFreq)) {
        locationsFreq[data[i][6]] = 1;
      } else {
        locationsFreq[data[i][6]] += 1;
      }
      if (locationsFreq[data[i][6]] > max) {
        max = locationsFreq[data[i][6]];
      }
    }

    console.log(locationsFreq);

    locations = [];
    for (let location in locationsFreq) {
      locations.push([
        location,
        locationsFreq[location],
        locationsFreq[location] / max,
      ]);
    }

    locations.sort((a, b) => {
      return b[1] - a[1];
    });
    locations = locations.slice(0, 10);
    console.log(locations);
  }

  let locations: [string, number, number][] = [];
  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: data && mounted && build();

  export let data: RequestsData;
</script>

<div class="card">
  <div class="card-title">Location</div>
  <div class="bars">
    {#each locations as location}
      <div class="bar-container">
        <div
          class="bar"
          title="{countryCodeToName(location[0])}: {location[1]} requests"
        >
          <div class="bar-inner" style="height: {location[2] * 100}%" />
        </div>
        <div class="label">{getFlagEmoji(location[0])}</div>
      </div>
    {/each}
  </div>
</div>

<style scoped>
  .card {
    flex: 1.2;
    margin: 2em 1em 2em 0;
  }
  .bars {
    height: 180px;
    display: flex;
    padding: 2em 2em 1.6em 2em;
  }

  .bar-container {
    flex: 1;
    margin: 0 5px;
    display: flex;
    flex-direction: column;
  }

  .bar {
    position: relative;
    /* height: 100%; */
    flex-grow: 1;
  }

  .bar-inner {
    position: absolute;
    bottom: 0;
    width: 100%;
    background: var(--highlight);
    border-radius: 3px;
  }
  .label {
    padding-top: 10px;
  }

  @media screen and (max-width: 1600px) {
    .card {
      width: 100%;
      margin: 2em 0 2em;
    }
  }
</style>
