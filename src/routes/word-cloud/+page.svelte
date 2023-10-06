<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/stores";
	import JoinForm from "./JoinForm.svelte";
	import WordForm from "./WordForm.svelte";
	import axios from "axios";
	import { addToast } from "../+layout.svelte";
	import { browser } from "$app/environment";
	import { onMount } from "svelte";
	import type { WordCloudSession } from "./utils";

  let session: WordCloudSession | null = null;

  onMount(() => {
    if ($page.url.searchParams.has("session")) {
      axios
        .get(`/api/word-cloud/${$page.url.searchParams.get("session")}`)
        .then((res) => {
          if (res.data) {
            session = res.data;
          } else {
            $page.url.searchParams.delete("session");
            if (browser) {
              goto($page.url.toString());
            }
          }
        })
        .catch((error) => {
          console.error(error);
          $page.url.searchParams.delete("session");
          goto($page.url.toString());
          addToast({data:{
            title: "Warning",
            description: "The session you are trying to join does not exist or has been closed.",
            color: "bg-ctp-orange",
          }});
        });
    }
  });

  const joinSession = (s: WordCloudSession) => {
    session = s;
    $page.url.searchParams.set("session", s.id);
    if (browser) {
      goto($page.url.toString());
    }
  };
</script>

{#if !session}
  <JoinForm joinSession={joinSession} />
{:else}
  <WordForm session={session} />
{/if}
