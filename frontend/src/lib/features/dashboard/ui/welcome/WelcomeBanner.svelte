<script>
    export let role = '';
</script>

<div class="relative w-full overflow-hidden rounded-[2rem] bg-gradient-to-br from-blue-700 to-indigo-900 p-6 md:p-10 shadow-lg shadow-indigo-900/10 isolate">
    <!-- Optimized Background: Single Radial Gradient Layer -->
    <div class="absolute inset-0 z-0 pointer-events-none bg-[radial-gradient(circle_at_top_right,_var(--tw-gradient-stops))] from-white/10 via-transparent to-transparent opacity-40"></div>
    
    <!-- Content Wrapper -->
    <div class="relative z-10 flex flex-col gap-6 md:flex-row md:items-start md:justify-between">
        <slot />
    </div>
</div>
