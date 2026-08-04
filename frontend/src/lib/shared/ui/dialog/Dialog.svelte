<script>
  import { createEventDispatcher, onDestroy } from 'svelte';
  import { portal } from '$lib/shared/actions/portal';
  export let open = false;
  export let hideCloseButton = false;
  let className = '';
  export { className as class };

  const dispatch = createEventDispatcher();
  
  function close() {
    dispatch('close');
    open = false;
  }

  $: if (typeof document !== 'undefined') {
    if (open) {
      document.body.style.overflow = 'hidden';
      document.body.style.touchAction = 'none'; // Further prevent scrolling on touch devices
    } else {
      document.body.style.overflow = '';
      document.body.style.touchAction = '';
    }
  }

  onDestroy(() => {
    if (typeof document !== 'undefined') {
      document.body.style.overflow = '';
      document.body.style.touchAction = '';
    }
  });
</script>

{#if open}
  <div use:portal>
    <div class="fixed inset-0 z-[100] bg-slate-900/90" role="button" tabindex="0" on:click={close} on:keydown={(e) => e.key === 'Escape' && close()}></div>
    <div class="fixed left-[50%] top-[50%] z-[100] grid w-[calc(100%-1rem)] sm:w-[calc(100%-2rem)] md:w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-4 md:p-6 shadow-lg duration-200 rounded-xl md:rounded-2xl bg-white max-h-[calc(100dvh-1rem)] sm:max-h-[calc(100dvh-2rem)] md:max-h-[90dvh] overflow-y-auto overflow-x-hidden {className}">
      <slot />
      {#if !hideCloseButton}
      <button class="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground bg-slate-100/90 hover:bg-slate-200" on:click={close}>
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4"><line x1="18" x2="6" y1="6" y2="18"></line><line x1="6" x2="18" y1="6" y2="18"></line></svg>
        <span class="sr-only">Close</span>
      </button>
      {/if}
    </div>
  </div>
{/if}
