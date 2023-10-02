<script lang="ts">
	import { page } from "$app/stores";
	import axios from "axios";
	import { onMount } from "svelte";
	import { addToast } from "../+layout.svelte";
	import { goto } from "$app/navigation";
	import { user } from "$lib/stores";

  onMount(() => {
    const token: string = $page.url.searchParams.get('token') ?? '';
    const redirect: string = $page.url.searchParams.get('redirect') ?? '/';
    
    // Send the token to the server
    axios.get('/api/auth/verify', { params: { token: token } })
      .then((reponse) => {
        if (reponse.status !== 200) {
          console.error(reponse);
          addToast({
            data: {
              title: 'Invalid token',
              description: 'The token you provided is invalid or has expired, please try again',
              color: 'bg-ctp-red'
            }
          });
          goto(redirect);
          return;
        }
        // Store the user in a cookie
        document.cookie = `user=${JSON.stringify(reponse.data)}; path=/; max-age=2592000; samesite=strict`;
        user.set(reponse.data);

        addToast({
          data: {
            title: 'Successfully logged in',
            description: 'You are now logged in',
            color: 'bg-ctp-green'
          }
        });
        goto(redirect);
      })
      .catch(() => {
        addToast({
          data: {
            title: 'Failed to verify token',
            description: 'An error occured while verifying your token',
            color: 'bg-ctp-red'
          }
        });
        goto(redirect);
      });
  });
</script>
