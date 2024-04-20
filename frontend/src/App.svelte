<script lang="ts">
  import { onMount } from "svelte";
  import { GetVersion, Trace, GetPublicIP } from "../wailsjs/go/main/App.js";
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
  let publicIP = "";

  async function trace() {
    try {
      loading = true;
      ip = ip.trim();
      error = "";
      for (const key in hops) delete hops[key];
      route = [];
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
    hopsEelement.lastElementChild.scrollIntoView({ behavior: "smooth" });
  });

  EventsOn("hop info", (hop: Hop) => {
    console.log(hop);
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
      const temp = await GetPublicIP();
      if (temp.startsWith("error")) {
        error = temp.slice(5);
        publicIP = "Unknown";
      } else publicIP = temp;
    } catch (error) {
      console.log(error);
    }
  });
</script>

<div class="absolute left-4 bottom-1 z-10 text-neutral-700 font-bold">
  v{version}
</div>
<div class="absolute right-4 bottom-1 z-10 text-neutral-700 font-bold">
  Public IP: {publicIP}
</div>

<main
  class="bg-neutral-950 overflow-hidden pt-4 px-4 pb-8 text-neutral-50 w-full h-[100vh] flex flex-col"
>
  <form class="flex items-center mb-4">
    <input
      class="border-2 border-neutral-800 placeholder:text-neutral-600 font-bold bg-neutral-950 text-xl tracking-wide px-4 py-2 rounded focus:outline-none w-full outline-none"
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
      class="bg-blue-700 text-xl tracking-wide font-bold ml-2 rounded px-4 py-2 {loading
        ? 'opacity-50'
        : 'hover:bg-blue-800'}">TRACE</button
    >
  </form>
  <div class="grid grid-cols-2 gap-4 mb-4">
    <div class="flex items-center relative">
      <div
        class="absolute right-4 text-neutral-600 font-bold text-lg select-none"
      >
        MAX HOPS
      </div>
      <input
        class="border-2 text-lg font-bold border-neutral-800 bg-neutral-950 w-full px-4 py-2 rounded focus:outline-none outline-none"
        autocomplete="off"
        bind:value={maxHops}
        id="name"
        type="number"
        disabled={loading}
      />
    </div>
    <div class="flex items-center relative">
      <div
        class="absolute right-4 text-neutral-600 font-bold text-lg select-none"
      >
        TIMEOUT(ms)
      </div>
      <input
        class="border-2 text-lg font-bold border-neutral-800 bg-neutral-950 w-full px-4 py-2 rounded focus:outline-none outline-none"
        autocomplete="off"
        bind:value={timeout}
        id="name"
        type="number"
        disabled={loading}
      />
    </div>
  </div>
  {#if error}
    <div class="text-red-600 font-semibold mb-4">{error}</div>
  {/if}
  <div
    bind:this={hopsEelement}
    class="w-full overflow-y-scroll h-full p-4 bg-neutral-900 rounded"
  >
    <div
      class="w-full grid grid-cols-12 font-semibold tracking-wide grid-flow-col gap-1 border-b-2 pb-2 mb-2 border-neutral-800"
    >
      <span class="col-span-1">#</span>
      <span class="col-span-3">Address</span>
      <span class="col-span-2">RTT</span>
      <span class="col-span-2">Country</span>
      <span class="col-span-4">ISP</span>
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
    {#if loading}
      {#key Object.keys(hops).length}
        <div class="w-full mt-6 h-1"></div>
      {/key}
    {/if}
  </div>

  {#if route.length}
    <div class="p-4 bg-neutral-900 rounded mt-4 font-semibold">
      {#each route as country}
        <span class="mr-2">→</span><span class="mr-2 text-lg tracking-wide"
          >{country}</span
        >
      {/each}
    </div>
  {/if}
</main>
