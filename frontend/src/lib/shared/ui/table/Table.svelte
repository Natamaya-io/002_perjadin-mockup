<script>
    import { cn } from '$lib/shared/utils/utils';
    export let className = '';
</script>

<div class="relative w-full overflow-auto table-scrollbar table-scroll-shadows">
    <table class={cn('w-full caption-bottom text-sm', className)} {...$$restProps}>
        <slot />
    </table>
</div>
