<script>
    import { cn } from '$lib/shared/utils/utils';
    export let className = '';
</script>

<tbody class={cn('[&_tr:last-child]:border-0', className)} {...$$restProps}>
    <slot />
</tbody>
