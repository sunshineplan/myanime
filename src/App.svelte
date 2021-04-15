<script lang="ts">
  import { onMount } from "svelte";

  interface anime {
    id: string;
    name: string;
    url: string;
    image: string;
    playlist: play[];
  }

  interface play {
    aid: string;
    url: string;
    index: string;
    ep: string;
    title: string;
  }

  let query = "";
  let list: anime[] = [];
  let loading = 0;

  const search = async (query?: string) => {
    let url: string;
    if (query) url = "/list?q=" + query;
    else url = "/list";
    loading++;
    const resp = await fetch(url);
    loading--;
    if (resp.ok) {
      const json = await resp.json();
      if (Array.isArray(json)) {
        list = json as anime[];
        return;
      } else if (!json) {
        alert("No anime found");
        return;
      }
    }
    alert("Failed to get list");
  };

  const open = async (play: play) => {
    const resp = await fetch("/play", {
      method: "post",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(play),
    });
    if (resp.ok) {
      const url = await resp.text();
      window.open(url);
    } else alert("Failed to get play");
  };

  onMount(async () => {
    await search();
  });
</script>

<header class="navbar navbar-expand flex-column flex-md-row">
  <a class="navbar-brand text-primary m-0 mr-md-3" href="/">My Anime</a>
  <div class="input-group">
    <input
      class="form-control"
      type="search"
      bind:value={query}
      placeholder="Search Anime"
      on:keydown={async (e) => {
        if (e.key === "Escape") query = "";
        else if (e.key === "Enter") await search(query);
      }}
    />
    <button
      class="btn btn-outline-primary"
      on:click={async () => await search(query)}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="16"
        height="16"
        fill="currentColor"
        viewBox="0 0 16 16"
      >
        <path
          d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5 5.5 0 0 1 11 0z"
        />
      </svg>
    </button>
  </div>
</header>
<div class="content" style="opacity: {loading ? 0.5 : 1}">
  {#each list as anime (anime.id)}
    <div style="display:flex">
      <div class="anime" on:click={() => window.open(anime.url)}>
        <img src={anime.image} alt={anime.name} width="150px" height="208px" />
        {anime.name}
      </div>
      <div class="playlist">
        {#each anime.playlist as play (play.index + play.ep)}
          <li on:click={() => open(play)}>
            <span class="play">{play.title}</span>
          </li>
        {/each}
      </div>
    </div>
  {/each}
</div>
<div class="loading" hidden={!loading}>
  <div class="sk-wave sk-center">
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
  </div>
</div>

<style>
  :root {
    --sk-color: #1a73e8;
    --header: 80px;
  }

  header {
    padding: 10px 20px;
  }

  .navbar {
    user-select: none;
    height: var(--header);
    justify-content: space-between;
    letter-spacing: 0.3px;
    border-bottom: 5px solid #f2f2f2;
  }

  .navbar-brand {
    font-size: 24px;
    padding-left: 30px;
  }

  .input-group {
    width: 33%;
    max-width: 360px;
  }

  .input-group > :not(:last-child) {
    margin-right: -1px;
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
  }

  .input-group > :not(:first-child) {
    border-top-left-radius: 0;
    border-bottom-left-radius: 0;
    z-index: 2;
  }

  svg {
    vertical-align: -0.125em;
  }

  .content,
  .loading {
    position: fixed;
    top: var(--header);
    width: 100%;
    height: calc(100% - var(--header));
  }

  .content {
    overflow: auto;
  }

  .loading {
    z-index: 2;
    display: flex;
  }

  .anime {
    display: grid;
    margin: 10px;
    text-align: center;
    width: 150px;
    cursor: pointer;
  }

  .playlist {
    height: 230px;
    width: calc(100% - 170px);
    overflow: auto;
    align-self: center;
  }

  li {
    display: inline-block;
    margin: 10px 6px;
    cursor: pointer;
  }

  .play {
    border: 1px solid #6c757d;
    border-radius: 3px;
    padding: 5px;
    color: #343a40;
  }

  @media (max-width: 767px) {
    :root {
      --header: 120px;
    }

    .navbar {
      border-color: transparent;
    }

    .navbar-brand {
      padding-left: 0;
    }

    .input-group {
      width: 66%;
    }
  }

  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
      "Helvetica Neue", Arial, "Noto Sans", "Microsoft YaHei New",
      "Microsoft Yahei", 微软雅黑, 宋体, SimSun, STXihei, 华文细黑, sans-serif;
  }
</style>
