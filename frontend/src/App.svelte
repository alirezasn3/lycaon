<script lang="ts">
  import { onMount } from "svelte";
  import { GetVersion, Trace } from "../wailsjs/go/main/App.js";
  import { EventsOn } from "../wailsjs/runtime/runtime";
  import "./style.css";
  import { fly } from "svelte/transition";

  interface Hop {
    number: number;
    address: string;
    ipeeInfo: IPEEInfo;
    rtt: number;
    isPrivate: boolean;
  }

  interface IPEEInfo {
    asName: string;
    country: string;
    countryCode: string;
    organizationName: string;
  }

  let ip: string = "";
  let maxHops: number = 32;
  let timeout: number = 1000;
  let error: string = "";
  let loading: boolean = false;
  let hopsEelement: HTMLDivElement;
  let hops: { [number: number]: Hop } = {};
  let route = [];
  let version = "";

  async function trace() {
    try {
      error = "";
      hops = {};
      route = [];
      loading = true;
      error = await Trace(ip, maxHops, timeout);
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
    if (
      hop.ipeeInfo.country !== "" &&
      hop.ipeeInfo.country !== route[route.length - 1]
    ) {
      route = [...route, hop.ipeeInfo.country];
    }
  });

  onMount(async () => {
    try {
      version = await GetVersion();
    } catch (error) {
      console.log(error);
    }
  });
</script>

<div class="absolute left-1 bottom-1 text-neutral-600 z-10 font-bold text-sm">
  v{version}
</div>

<main
  class="bg-neutral-950 overflow-hidden p-8 text-neutral-50 w-full h-[100vh] flex flex-col"
>
  <form class="flex items-center text-lg font-bold">
    <input
      class="text-neutral-950 px-3 py-1.5 rounded focus:outline-none w-full outline-none"
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
      class="bg-blue-700 ml-2 hover:bg-blue-800 rounded px-3 py-1.5 {loading &&
        'opacity-50'}">TRACE</button
    >
  </form>
  <div class="flex mt-4 justify-between">
    <div class="flex items-center">
      <div class="mr-2">Hops:</div>
      <input
        class="text-neutral-950 w-full px-3 py-1.5 rounded focus:outline-none outline-none"
        autocomplete="off"
        bind:value={maxHops}
        id="name"
        type="number"
        disabled={loading}
      />
    </div>
    <div class="flex items-center">
      <div class="mr-2">Timeout:</div>
      <input
        class="text-neutral-950 w-full px-3 py-1.5 rounded focus:outline-none outline-none"
        autocomplete="off"
        bind:value={timeout}
        id="name"
        type="number"
        disabled={loading}
      />
    </div>
  </div>
  {#if error}
    <div class="my-2 text-red-700">{error}</div>
  {/if}
  <div
    bind:this={hopsEelement}
    class="w-full overflow-y-scroll my-4 p-4 bg-neutral-900 rounded shadow-inner h-full scroll-smooth min-w-[640px]"
  >
    <div
      class="w-full grid grid-cols-12 grid-flow-col gap-1 border-b-2 pb-2 mb-2 border-neutral-800 tracking-wider"
    >
      <span class="col-span-1">#</span>
      <span class="col-span-3">address</span>
      <span class="col-span-2">rtt</span>
      <span class="col-span-2">country</span>
      <span class="col-span-4">isp</span>
    </div>
    {#each Object.values(hops) as hop}
      <div
        class="w-full grid grid-cols-12 grid-flow-col gap-1 my-1 hover:bg-neutral-800 px-1 {hop.address ===
          ip && 'text-green-500 font-bold'}"
        transition:fly={{ duration: 100, y: 10 }}
      >
        <span class="col-span-1">
          {hop.number}
        </span>
        <span class="col-span-3 flex">
          {hop.address === "timeout" ? "*" : hop.address}
        </span>
        <span class="col-span-2">
          {hop.address === "timeout" ? "*" : `${hop.rtt}ms`}
        </span>
        <span class="col-span-2 truncate flex items-center">
          {#if hop.address === "timeout" || hop.isPrivate}
            *
          {:else if hop.ipeeInfo.countryCode}
            {hop.ipeeInfo.countryCode}
          {:else}
            <span class="relative flex h-2 w-2 ml-1">
              <span
                class="animate-ping absolute inline-flex h-full w-full rounded-full bg-sky-600 opacity-75"
              ></span>
              <span class="relative inline-flex rounded-full h-2 w-2 bg-sky-700"
              ></span>
            </span>
          {/if}
        </span>
        <span class="col-span-4 truncate flex items-center">
          {#if hop.address === "timeout" || hop.isPrivate}
            *
          {:else if hop.ipeeInfo.organizationName}
            {hop.ipeeInfo.organizationName}
          {:else}
            <span class="relative flex h-2 w-2 ml-1">
              <span
                class="animate-ping absolute inline-flex h-full w-full rounded-full bg-sky-600 opacity-75"
              ></span>
              <span class="relative inline-flex rounded-full h-2 w-2 bg-sky-700"
              ></span>
            </span>
          {/if}
        </span>
      </div>
    {/each}
  </div>
  {#if route.length}
    <div class="p-4 bg-neutral-900 rounded">
      {#each route as country}
        <span class="mr-2">→</span><span class="mr-2 text-lg">{country}</span>
      {/each}
    </div>
  {/if}
</main>
