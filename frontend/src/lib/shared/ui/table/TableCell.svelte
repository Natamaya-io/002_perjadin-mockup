<script>
    import { cn } from '$lib/shared/utils/utils';
    export let className = '';
</script>

<td class={cn('p-4 align-middle [&:has([role=checkbox])]:pr-0', className)} {...$$restProps}>
    <slot />
</td>
