<script>
    import { cn } from '$lib/shared/utils/utils';
    export let title = '';
    export let value = '';
    export let description = '';
    export let iconColor = 'blue'; // blue, emerald, amber, purple, rose
    export let bgClass = 'bg-white border-slate-100';
    export let textColorClass = 'text-slate-900';
    export let titleColorClass = 'text-slate-400';
    export let descColorClass = 'text-slate-500';
    export let blobClass = 'bg-slate-50 group-hover:bg-slate-100/50';
    export let iconContainerClass = '';
    export let isSecret = false;
    export let autoShrink = false;

    let isRevealed = false;

    const colorClasses = {
        blue: 'bg-blue-50 text-blue-600 group-hover:bg-blue-600 group-hover:text-white',
        emerald: 'bg-emerald-50 text-emerald-600 group-hover:bg-emerald-600 group-hover:text-white',
        amber: 'bg-amber-50 text-amber-600 group-hover:bg-amber-600 group-hover:text-white',
        purple: 'bg-purple-50 text-purple-600 group-hover:bg-purple-600 group-hover:text-white',
        rose: 'bg-rose-50 text-rose-600 group-hover:bg-rose-600 group-hover:text-white'
    };
</script>

<div class={cn("group relative rounded-2xl p-6 border shadow-sm hover:shadow-lg transition-all duration-300 ease-out hover:-translate-y-1", bgClass)}>
    <!-- Decorative background blob (clipped) -->
    <div class="absolute inset-0 overflow-hidden rounded-2xl pointer-events-none">
        <div class={cn("absolute -right-6 -top-6 h-24 w-24 rounded-full transition-all group-hover:scale-150", blobClass)}></div>
    </div>

    <div class="relative z-10 flex flex-col h-full justify-between gap-4">
        <div class="flex justify-between items-start">
            <div class={cn("p-3.5 rounded-xl transition-colors duration-300", iconContainerClass || colorClasses[iconColor] || colorClasses.blue)}>
                <slot name="icon" />
            </div>

            <!-- Optional Action Slot (e.g. for buttons) -->
            <div class="ml-auto flex items-center gap-2">
                {#if isSecret}
                    <button 
                        class={cn("p-1.5 rounded-full hover:bg-white/20 transition-colors", textColorClass)} 
                        on:click={() => isRevealed = !isRevealed}
                        title={isRevealed ? "Sembunyikan Nilai" : "Tampilkan Nilai"}
                    >
                        {#if isRevealed}
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 opacity-80" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                            </svg>
                        {:else}
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 opacity-80" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                            </svg>
                        {/if}
                    </button>
                {/if}
                <slot name="action" />
            </div>
        </div>

        <!-- Filter Slot (for inline filters like month/year selectors) -->
        <slot name="filter" />

        <div class="space-y-1 overflow-hidden min-w-0">
            <div 
                class={cn(
                    "font-bold tracking-tight font-feature-settings-tnum",
                    autoShrink ? "text-xl sm:text-2xl lg:text-[1.6rem]" : "text-2xl lg:text-3xl",
                    textColorClass
                )} 
                title={isSecret && !isRevealed ? "Tersembunyi" : value}
            >
                {#if isSecret && !isRevealed}
                    <span class="tracking-widest">••••••••</span>
                {:else}
                    <span class="stat-value-text">{value}</span>
                {/if}
            </div>
            <div class="flex flex-col">
                <span class={cn("text-xs font-semibold uppercase tracking-wider truncate", titleColorClass)}>{title}</span>
                <span class={cn("text-[11px] font-medium mt-0.5 truncate", descColorClass)}>{description}</span>
            </div>
        </div>
    </div>
</div>
<style>
    .font-feature-settings-tnum {
        font-feature-settings: "tnum";
        font-variant-numeric: tabular-nums;
    }
    .stat-value-text {
        display: block;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    /* Auto-shrink for very long currency values on small containers */
    @container (max-width: 220px) {
        .stat-value-text {
            font-size: 1.1rem;
        }
    }
</style>
