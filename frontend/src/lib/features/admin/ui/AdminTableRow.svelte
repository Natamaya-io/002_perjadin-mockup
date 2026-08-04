<script>
    import { createEventDispatcher } from 'svelte';
    import { cn, getStatusBadge, formatCurrency } from '$lib/shared/utils/utils';
    import { userStore } from '$lib/features/auth/store';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import TableRow from '$lib/shared/ui/table/TableRow.svelte';
    import TableCell from '$lib/shared/ui/table/TableCell.svelte';

    export let record;

    const dispatch = createEventDispatcher();

</script>

<TableRow class="hover:bg-slate-50 transition-colors border-b border-slate-100 last:border-0">
    <TableCell class="pl-4 py-3 align-top">
        <div class="flex flex-col gap-0.5">
            <span class="font-mono text-[10px] text-slate-400 leading-none">#{record.id}</span>
            <span class="inline-flex items-center font-mono text-[13px] font-bold tracking-widest text-slate-700">{record.spd}</span>
        </div>
    </TableCell>
    <TableCell class="py-3 align-top">
        <div class="flex flex-col gap-0.5">
            <span class="font-semibold text-slate-900 text-sm">{record.employee?.name || '-'}</span>
            <span class="text-xs text-slate-500">{record.employee?.rank || '-'}</span>
        </div>
    </TableCell>
    <TableCell class="py-3 align-top">
        <div class="flex flex-col gap-0.5 max-w-[250px]">
            <span class="font-medium text-slate-700 text-sm truncate" title="{record.location}, {record.province}">
                {record.location}, {record.province}
                {#if record.locations && record.locations.length > 1}
                    <span class="ml-1 text-[10px] px-1.5 py-0.5 bg-blue-50 text-blue-600 rounded-full border border-blue-100">
                        +{record.locations.length - 1} Lokasi
                    </span>
                {/if}
            </span>
            <span class="text-xs text-slate-500 flex items-center gap-1">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})} s/d {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
            </span>
        </div>
    </TableCell>
    <TableCell class="py-3 align-top">
        <span class={cn("inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border", getStatusBadge(record).class)}>
            {getStatusBadge(record).label}
        </span>
    </TableCell>
    <TableCell class="text-right font-mono font-medium text-slate-700 py-3 align-top">{formatCurrency(record.totalCost)}</TableCell>
    <TableCell class="text-right pr-4 py-3 align-top">
        <div class="flex justify-end items-center gap-1">
            {#if $userStore.role !== 'kasubag'}
                <Button variant="outline" size="sm" class="h-8 text-xs border-slate-200 bg-white hover:bg-blue-50 hover:text-blue-600 hover:border-blue-200 transition-all shadow-sm" on:click={() => dispatch('edit', record)}>
                    {$userStore.role === 'protokol' ? (record.status === 'Draft' ? 'Input Dokumen & Biaya' : 'Edit Dokumen & Biaya') : 'Review Biaya'}
                </Button>
                <div class="h-4 w-px bg-slate-200 mx-1"></div>
            {/if}
            <div class="flex items-center gap-1">
                {#if $userStore.role !== 'protokol'}
                    <button class="p-1.5 hover:bg-slate-100 rounded-md text-slate-400 hover:text-slate-700 transition-colors" title="Cetak SPD" on:click={() => dispatch('printSPD', record)}>
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><line x1="16" x2="8" y1="13" y2="13"/><line x1="16" x2="8" y1="17" y2="17"/><line x1="10" x2="8" y1="9" y2="9"/></svg>
                    </button>
                    <button class="p-1.5 hover:bg-slate-100 rounded-md text-slate-400 hover:text-slate-700 transition-colors" title="Cetak Rincian" on:click={() => dispatch('printRincian', record)}>
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="16" height="20" x="4" y="2" rx="2"/><line x1="8" x2="16" y1="6" y2="6"/><line x1="16" x2="16" y1="14" y2="18"/><path d="M16 10h.01"/><path d="M12 10h.01"/><path d="M8 10h.01"/><path d="M12 14h.01"/><path d="M8 14h.01"/><path d="M12 18h.01"/><path d="M8 18h.01"/></svg>
                    </button>
                {/if}
            </div>
        </div>
    </TableCell>
</TableRow>
