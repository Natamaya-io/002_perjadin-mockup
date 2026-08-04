<script>
    import ReportTitle from './header/ReportTitle.svelte';
    import ReportFilters from './header/ReportFilters.svelte';

    export let searchQuery = '';
    export let statusFilter = 'all';
    export let sortOption = 'date-desc';
</script>

<div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
    <ReportTitle />
    <ReportFilters 
        bind:searchQuery 
        bind:statusFilter 
        bind:sortOption 
    />
</div>
