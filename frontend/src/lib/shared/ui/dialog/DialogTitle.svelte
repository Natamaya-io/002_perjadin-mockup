<script>
  import { cn } from "$lib/shared/utils/utils";
  let className = undefined;
  export { className as class };
</script>

<h2 class={cn("text-lg font-semibold leading-none tracking-tight", className)}>
  <slot />
</h2>
