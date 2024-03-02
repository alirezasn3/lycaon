<script lang="ts">
  import { Trace } from "../wailsjs/go/main/App.js";
  import { EventsOn } from "../wailsjs/runtime/runtime";
  import "./style.css";
  import { fly } from "svelte/transition";

  interface Hop {
    number: number;
    address: string;
    ipeeInfo: IPEEInfo;
  }

  interface IPEEInfo {
    asName: string;
    country: string;
  }

  let ip: string = "";
  let error: string = "";
  let loading: boolean = false;
  let hopsEelement: HTMLDivElement;
  let hops: { [number: number]: Hop } = {};
  let routes = [];

  async function trace() {
    try {
      error = ""
      hops = {};
      routes = [];
      loading = true;
      error = await Trace(ip);
      if (!error) {
        const values = Object.values(hops).sort((a, b) => a.number - b.number);
        let country = values[0].ipeeInfo.country;
        let count = 1;
        const temp = [];
        temp.push({ country, count });
        for (let i = 1; i < values.length; i++) {
          count++;
          if (values[i].ipeeInfo.country !== country) {
            country = values[i].ipeeInfo.country;
            temp.push({ country, count });
            count = 1;
          }
        }
        routes = temp;
      }
    } catch (error) {
      console.log(error);
      error = error.message;
    } finally {
      loading = false;
    }
  }

  EventsOn("hop", (hop: Hop) => {
    hops[hop.number] = hop;
  });
  EventsOn("hop info", (hop: Hop) => {
    hops[hop.number] = hop;
  });
</script>

<main
  class="font-bold bg-neutral-950 overflow-hidden p-8 text-neutral-50 w-full h-[100vh] flex flex-col"
>
  <div class="mb-4 flex flex-col items-center">
    <span class="text-[#ffff00] text-4xl">IPEE</span>
    <span class="tracking-wide">TRACER</span>
  </div>
  <div class="flex items-center">
    <input
      class="text-neutral-950 px-2 py-1 rounded focus:outline-none w-full"
      autocomplete="off"
      bind:value={ip}
      id="name"
      type="text"
      placeholder="IP ADDRESS"
      disabled={loading}
    />
    <button
      on:click={trace}
      disabled={loading}
      class="bg-blue-700 rounded px-2 py-1 ml-1 {loading && 'opacity-50'}"
      >TRACE</button
    >
  </div>
  {#if error}
    <div class="my-2 text-red-700">{error}</div>
  {/if}
  <div
    bind:this={hopsEelement}
    class="w-full overflow-y-scroll my-4 p-4 bg-neutral-900 rounded shadow-inner h-full scroll-smooth"
  >
    <div class="w-full grid grid-cols-10 grid-flow-col gap-1">
      <span class="col-span-1">#</span>
      <span class="col-span-3">ADDRESS</span>
      <span class="col-span-3">COUNTRY</span>
      <span class="col-span-3">AS NAME</span>
    </div>
    {#each Object.values(hops) as hop}
      <div
        class="w-full grid grid-cols-10 grid-flow-col gap-1 font-normal"
        transition:fly={{
          duration: 100,
          y: 10,
        }}
      >
        <span class="col-span-1">
          {hop.number}
        </span>
        <span class="col-span-3">
          {hop.address}
        </span>
        <span class="col-span-3">
          {hop.ipeeInfo.country}
        </span>
        <span class="col-span-3">
          {hop.ipeeInfo.asName}
        </span>
      </div>
    {/each}
  </div>
  {#if routes.length}
    <div class="p-4 bg-neutral-900 rounded shadow-inner">
      {#each routes as route}
        {#if route.country}
          <span>→{route.country}</span>
        {/if}
      {/each}
    </div>
  {/if}
</main>
