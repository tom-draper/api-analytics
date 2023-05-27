<script lang="ts">
  import { onMount } from "svelte";

  function getFlagEmoji(countryCode: string) {
    const codePoints = countryCode
      .toUpperCase()
      .split("")
      .map((char) => 127397 + char.charCodeAt(undefined));
    return String.fromCodePoint(...codePoints);
  }

  function countryCodeToName(countryCode: string) {
    let regionNames = new Intl.DisplayNames(["en"], { type: "region" });
    return regionNames.of(countryCode);
  }

  function build() {
    let max = 0;
    let locationsFreq = {};
    for (let i = 0; i < data.length; i++) {
      if (data[i][6] === "") {
        continue;
      }
      if (!(data[i][6] in locationsFreq)) {
        locationsFreq[data[i][6]] = 1;
      } else {
        locationsFreq[data[i][6]] += 1;
      }
      if (locationsFreq[data[i][6]] > max) {
        max = locationsFreq[data[i][6]];
      }
    }

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
  {#if locations.length > 0}
    <div class="locations-count">{locations.length} locations</div>
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
  {:else}
    <div class="no-locations">
      <div class="no-locations-text">No Locations Found</div>
    </div>
  {/if}
</div>

<style scoped>
  .card {
    flex: 1.2;
    margin: 2em 1em 2em 0;
    position: relative;
  }
  .bars {
    height: 150px;
    display: flex;
    padding: 1.5em 2em 1em 2em;
  }

  .bar-container {
    flex: 1;
    margin: 0 5px;
    display: flex;
    flex-direction: column;
  }

  .no-locations {
    height: 180px;
    display: grid;
    place-items: center;
  }
  .no-locations-text {
    margin-bottom: 25px;
    color: #707070;
  }

  .bar {
    position: relative;
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
    padding-top: 8px;
  }
  .locations-count {
    position: absolute;
    top: 1.5em;
    right: 2em;
    font-size: 0.9em;
    color: #505050;
  }

  @media screen and (max-width: 1600px) {
    .card {
      width: 100%;
      margin: 2em 0 2em;
    }
  }
</style>
