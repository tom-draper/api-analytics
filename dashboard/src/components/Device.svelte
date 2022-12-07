<script lang="ts">
  import Browser from "./Browser.svelte";
  import OperatingSystem from "./OperatingSystem.svelte";
  import DeviceType from "./DeviceType.svelte";

  function setBtn(target: string) {
    activeBtn = target
  }

  let activeBtn = "os";
  export let data: RequestsData;
</script>

<div class="card" title="Last week">
  <div class="card-title">
    Device
  
    <div class="toggle">
      <button class="{activeBtn == 'os' ? 'active' : ''}" on:click="{() => {setBtn('os')}}">OS</button>
      <button class="{activeBtn == 'browser' ? 'active' : ''}" on:click="{() => {setBtn('browser')}}">Browser</button>
      <button class="{activeBtn == 'device' ? 'active' : ''}" on:click="{() => {setBtn('device')}}">Device</button>
    </div>
  </div>
  <div class="os" style="{activeBtn == 'os' ? 'display: initial' : ''}">
    <OperatingSystem {data} />
  </div>
  <div class="browser" style="{activeBtn == 'browser' ? 'display: initial' : ''}">
    <Browser {data} />
  </div>
  <div class="device" style="{activeBtn == 'device' ? 'display: initial' : ''}">
    <DeviceType {data} />
  </div>
</div>

<style>
  .card {
    margin: 2em 0 2em 1em;
    flex: 1;
    padding-bottom: 1em;
    width: 420px;
  }
  .card-title {
    display: flex;
  }
  .toggle {
    margin-left: auto;
  }
  .active {
    background: var(--highlight);
  }
  .os,
  .browser,
  .device {
    display: none;
  }
  button {
    border: none;
    border-radius: 4px;
    background: rgb(68, 68, 68);
    cursor: pointer;
    padding: 2px 6px;
    margin-left: 5px;
  }
  @media screen and (max-width: 1580px){
    .card {
        margin: 0 0 2em;
        width: 100%;
      }
  }
</style>
