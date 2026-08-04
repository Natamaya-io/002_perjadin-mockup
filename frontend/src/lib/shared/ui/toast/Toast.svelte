<script>
    import { toast } from '$lib/shared/stores/toast';
    import { fly, fade } from 'svelte/transition';
    import { onMount } from 'svelte';

    export let item;

    const icons = {
        success: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>`,
        error: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>`,
        info: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>`,
        warning: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>`
    };

    const colors = {
        success: 'bg-emerald-50 text-emerald-800 border-emerald-200',
        error: 'bg-red-50 text-red-800 border-red-200',
        info: 'bg-blue-50 text-blue-800 border-blue-200',
        warning: 'bg-amber-50 text-amber-800 border-amber-200'
    };
    
    const iconColors = {
        success: 'text-emerald-500',
        error: 'text-red-500',
        info: 'text-blue-500',
        warning: 'text-amber-500'
    };
</script>

<div 
    in:fly={{ y: 20, duration: 300 }} 
    out:fade={{ duration: 200 }}
    class="flex items-center w-[calc(100vw-2rem)] sm:w-[380px] p-4 mb-3 rounded-xl border shadow-lg shadow-black/5 {colors[item.type]} backdrop-blur-sm relative overflow-hidden group"
    role="alert"
>
    <!-- Icon -->
    <div class="inline-flex items-center justify-center flex-shrink-0 w-8 h-8 rounded-lg bg-white {iconColors[item.type]} shadow-sm">
        {@html icons[item.type]}
    </div>
    
    <!-- Content -->
    <div class="ml-3 text-sm font-medium pr-6">{item.message}</div>
    
    <!-- Close Button -->
    <button 
        type="button" 
        class="ml-auto -mx-1.5 -my-1.5 rounded-lg focus:ring-2 focus:ring-black/5 p-1.5 inline-flex items-center justify-center h-8 w-8 opacity-60 hover:opacity-100 transition-opacity" 
        on:click={() => toast.remove(item.id)} 
        aria-label="Close"
    >
        <span class="sr-only">Close</span>
        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path></svg>
    </button>
    
    <!-- Progress Bar (Optional visual flair) -->
    <div class="absolute bottom-0 left-0 h-0.5 bg-current opacity-20 animate-shrink origin-left w-full" style="animation-duration: {item.duration}ms;"></div>
</div>

<style>
    @keyframes shrink {
        from { transform: scaleX(1); }
        to { transform: scaleX(0); }
    }
    .animate-shrink {
        animation-name: shrink;
        animation-timing-function: linear;
        animation-fill-mode: forwards;
    }
</style>
