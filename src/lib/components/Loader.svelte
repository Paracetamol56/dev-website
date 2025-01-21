<script lang="ts">
  import { Loader2 } from 'lucide-svelte';

  export let size: number = 32;
</script>

<Loader2 class="animate-spin text-ctp-mauve" {size} />