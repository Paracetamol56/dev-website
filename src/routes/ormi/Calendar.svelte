<script lang="ts">

  const weekDays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

  // Create an array of 52 weeks before the current date
  const weeks = Array.from({ length: 52 }, (_, i) => {
    const days = Array.from({ length: 7 }, (_, j) => {
      const date = new Date();
      date.setDate(date.getDate() - date.getDay() + i * 7 + j + 1);
      return date;
    });
    return days;
  });

  console.log(weeks);

  let hovering: Date | null = null;
  $: console.log(hovering);
</script>

<div class="w-full max-w-4xl p-1 mx-auto bg-ctp-mantle rounded-md">
  <table class="w-full">
    <tbody>
      {#each weekDays as day, i}
        <tr>
          {#if i % 2 == 0}
            <td class="w-8 p-px">
              <p class="text-xs text-right">{day}</p>
            </td>
          {:else}
            <td></td>
          {/if}
          {#each weeks as week}
            <td class="p-0.5">
              <!-- svelte-ignore a11y-no-static-element-interactions -->
              <div
                class="relative aspect-square bg-ctp-crust rounded-sm"
                on:mouseenter={(e) => hovering = week[i]}
                on:mouseleave={() => hovering = null}
              >
                {#if hovering === week[i]}
                  <div class="absolute bottom-[calc(100%+0.5rem)] left-1/2 transform -translate-x-1/2">
                    <div class="relative z-10 rounded-lg bg-ctp-text shadow">
                      <div class="absolute square-2 bg-ctp-text left-1/2 -bottom-1 transform -translate-x-1/2 rotate-45"></div>
                      <p class="px-4 py-1 text-ctp-mantle whitespace-nowrap">0 activity on {hovering.toLocaleDateString()}</p>
                    </div>
                  </div>
                {/if}
              </div>
            </td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
</div>
