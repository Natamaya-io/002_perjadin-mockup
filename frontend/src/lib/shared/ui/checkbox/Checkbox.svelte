<script>
  import { cn } from "$lib/shared/utils/utils";
  
  export let checked = false;
  export let id = '';
  export let className = '';
</script>

<div class="flex items-center space-x-2">
  <input 
    type="checkbox" 
    {id} 
    bind:checked 
    class={cn(
      "h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary accent-black", 
      className
    )}
    {...$$restProps}
    on:change
  />
  <slot />
</div>
