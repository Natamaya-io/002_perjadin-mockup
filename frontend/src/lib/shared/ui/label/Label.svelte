<script>
    import { cn } from '$lib/shared/utils/utils';
    export let className = '';
    
    $: classes = cn(
        'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
        className
    );
</script>

<label class={classes} {...$$restProps}>
    <slot />
</label>
