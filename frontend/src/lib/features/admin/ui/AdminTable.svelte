<script>
    import Table from '$lib/shared/ui/table/Table.svelte';
    import TableHeader from '$lib/shared/ui/table/TableHeader.svelte';
    import TableRow from '$lib/shared/ui/table/TableRow.svelte';
    import TableHead from '$lib/shared/ui/table/TableHead.svelte';
    import TableBody from '$lib/shared/ui/table/TableBody.svelte';
</script>

<div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden">
    <div class="overflow-x-auto w-full">
        <Table class="w-full text-sm text-left">
            <TableHeader class="bg-slate-50 border-b border-slate-200">
                <TableRow class="hover:bg-slate-50/50">
                    <TableHead class="w-[100px] min-w-[100px] font-semibold text-slate-700 pl-4 py-3">ID / SPD</TableHead>
                    <TableHead class="min-w-[200px] font-semibold text-slate-700 py-3">Pegawai</TableHead>
                    <TableHead class="min-w-[250px] font-semibold text-slate-700 py-3">Detail Perjalanan</TableHead>
                    <TableHead class="w-[120px] min-w-[120px] font-semibold text-slate-700 py-3">Status</TableHead>
                    <TableHead class="min-w-[160px] font-semibold text-slate-700 text-right py-3">Total Biaya</TableHead>
                    <TableHead class="w-[180px] min-w-[180px] font-semibold text-slate-700 text-right pr-4 py-3">Aksi</TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                <slot />
            </TableBody>
        </Table>
    </div>
</div>
