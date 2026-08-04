<script>
    import { cn } from '$lib/shared/utils/utils';
    export let className = '';
</script>

<tr class={cn('border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted', className)} {...$$restProps}>
    <slot />
</tr>
