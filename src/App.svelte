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

  let list: anime[] = [];

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
    const resp = await fetch("/list");
    if (resp.ok) {
      const json = await resp.json();
      if (Array.isArray(json)) {
        list = json as anime[];
        return;
      }
    }
    alert("Failed to get list");
  });
</script>

<header class="navbar navbar-expand flex-column flex-md-row">
  <a class="navbar-brand text-primary m-0 mr-md-3" href="/">My Anime</a>
</header>
<div class="content">
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

<style>
  header {
    padding: 10px 20px;
  }

  .navbar {
    user-select: none;
    height: 80px;
    justify-content: space-between;
    letter-spacing: 0.3px;
    border-bottom: 5px solid #f2f2f2;
  }

  .navbar-brand {
    font-size: 24px;
    padding-left: 30px;
  }

  .content {
    position: fixed;
    top: 0;
    margin-top: 80px;
    width: 100%;
    height: calc(100% - 80px);
    overflow: auto;
  }

  .anime {
    display: grid;
    margin: 10px;
    text-align: center;
    width: 150px;
    cursor: pointer;
  }

  .playlist {
    height: 208px;
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
    .navbar {
      border-color: transparent;
    }

    .navbar-brand {
      padding-left: 0;
    }
  }

  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
      "Helvetica Neue", Arial, "Noto Sans", "Microsoft YaHei New",
      "Microsoft Yahei", 微软雅黑, 宋体, SimSun, STXihei, 华文细黑, sans-serif;
  }
</style>
