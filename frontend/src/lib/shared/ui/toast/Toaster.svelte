<script>
    import { toast } from '$lib/shared/stores/toast';
    import ToastItem from './Toast.svelte';
</script>

<div
    class="pointer-events-none fixed bottom-4 right-4 flex flex-col items-end gap-2"
    style="z-index: 9999;"
    aria-live="polite"
    aria-atomic="true"
>
    <div class="pointer-events-auto flex flex-col items-end">
        {#each $toast as item (item.id)}
            <ToastItem {item} />
        {/each}
    </div>
</div>