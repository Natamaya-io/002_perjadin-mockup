<script>
    import { paginatedRecordsStore, paginatedMetadataStore, loadPaginatedRecords, updateRecord, isFetchingRecords } from '$lib/features/pengajuan/store';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';
    import { onMount, onDestroy } from 'svelte';
    import { userStore } from '$lib/features/auth/store';
    import { getInitials, getStatusBadge, toTitleCase, cn, formatLocations } from '$lib/shared/utils/utils';

    // Components
    import AdminTableFilters from '$lib/features/admin/ui/AdminTableFilters.svelte';
    import EmptyState from '$lib/features/laporan/ui/EmptyState.svelte';
    import Table from '$lib/shared/ui/table/Table.svelte';
    import TableHeader from '$lib/shared/ui/table/TableHeader.svelte';
    import TableRow from '$lib/shared/ui/table/TableRow.svelte';
    import TableHead from '$lib/shared/ui/table/TableHead.svelte';
    import TableBody from '$lib/shared/ui/table/TableBody.svelte';
    import TableCell from '$lib/shared/ui/table/TableCell.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Dialog from '$lib/shared/ui/dialog/Dialog.svelte';
    import DialogHeader from '$lib/shared/ui/dialog/DialogHeader.svelte';
    import DialogTitle from '$lib/shared/ui/dialog/DialogTitle.svelte';
    import TripStepper from '$lib/features/dashboard/ui/roadmap/TripStepper.svelte';
    import DocumentViewer from '$lib/shared/ui/document-viewer/DocumentViewer.svelte';
    import SkeletonTable from '$lib/shared/ui/loader/SkeletonTable.svelte';

    // Filter & Sort State
    let searchQuery = '';
    let statusFilter = 'all'; // 'all', 'Completed', 'Pending'
    let sortOption = 'spj-desc'; // Default to newest SPJ first
    let startDate = '';
    let endDate = '';
    let limit = 50;
    
    let statusOptions = [
        { value: 'all', label: 'Semua Status' },
        { value: 'Pending', label: 'Belum Lapor' },
        { value: 'Completed', label: 'Selesai' }
    ];

    export let data;

    $: if (data?.recordsResponse) {
        paginatedRecordsStore.set(data.recordsResponse.data || []);
        paginatedMetadataStore.set({
            totalItems: data.recordsResponse.totalItems,
            totalRecords: data.recordsResponse.totalRecords,
            nextCursor: data.recordsResponse.nextCursor || null,
            limit: data.recordsResponse.limit
        });
    }

    /** @type {ReturnType<typeof setTimeout>} */
    let debounceTimer;
    let currentCursor = '';
    let isInitialMount = true;

    // Detail Modal State
    let isDetailModalOpen = false;
    let selectedDetailRecord = null;
    let selectedSpd = '';
    
    // Preview Modal State
    let isPreviewOpen = false;
    let previewUrl = '';
    let previewType = '';
    let previewFilename = '';

    function openPreview(url, type, filename) {
        previewUrl = url;
        previewType = type;
        previewFilename = filename;
        isPreviewOpen = true;
    }

    function openDetailModal(spd) {
        selectedSpd = spd;
        // Find the record for the current user, or fallback to the first one
        const spdRecords = groupedRecordsMap[spd] || [];
        selectedDetailRecord = spdRecords.find(r => r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email)) || spdRecords[0];
        isDetailModalOpen = true;
    }

    function getDays(record) {
        if (!record?.startDate || !record?.endDate) return 0;
        const start = Date.parse(record.startDate);
        const end = Date.parse(record.endDate);
        const diffTime = end - start;
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1; 
        return diffDays > 0 ? diffDays : 0;
    }

    // Reactively refetch when filters change
    function handleFiltersChanged() {
        if (typeof window !== 'undefined') {
            if (isInitialMount && data?.recordsResponse) {
                isInitialMount = false;
                return;
            }
            isInitialMount = false;

            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => {
                currentCursor = '';
                fetchRecords(false);
            }, 300);
        }
    }
    $: searchQuery, statusFilter, sortOption, startDate, endDate, handleFiltersChanged();

    function fetchRecords(append = false) {
        if (!$userStore) return;

        const filters = {
            search: searchQuery,
            report_status: statusFilter === 'all' ? undefined : statusFilter,
            sort_by: sortOption,
            start_date: startDate ? new Date(startDate).toISOString() : undefined,
            end_date: endDate ? new Date(endDate).toISOString() : undefined,
            limit
        };

        if (append && currentCursor) {
            filters.cursor = currentCursor;
        }

        loadPaginatedRecords(filters, append);
    }

    let groupedRecordsMap = {};
    let uniqueSPDs = [];
    let displayRecords = [];

    // Single-pass O(N) grouping and deduplication for strict 60fps performance
    $: {
        const userEmail = $userStore?.email;
        const userId = $userStore?.id;
        const newGrouped = {};
        const newUnique = [];
        const newDisplay = [];
        const spdToIndex = {};

        for (const record of $paginatedRecordsStore) {
            const spd = record.spd;
            if (!newGrouped[spd]) {
                newGrouped[spd] = [];
                newUnique.push(spd);
                spdToIndex[spd] = newDisplay.length;
                newDisplay.push(record); // initial representative
            }
            newGrouped[spd].push(record);
            
            // Update representative if this record belongs to the user
            if (record.email === userEmail || (record.employee && record.employee.email === userEmail) || record.employeeId === userId) {
                const idx = spdToIndex[spd];
                newDisplay[idx] = record;
            }
        }
        
        groupedRecordsMap = newGrouped;
        uniqueSPDs = newUnique;
        displayRecords = newDisplay;
    }

    function handleWindowScroll() {
        if (typeof window === 'undefined' || $isFetchingRecords) return;
        const scrollY = window.scrollY;
        const innerHeight = window.innerHeight;
        const scrollHeight = document.body.scrollHeight;
        if (scrollHeight - scrollY - innerHeight < 400) {
            if ($paginatedMetadataStore.nextCursor) {
                currentCursor = $paginatedMetadataStore.nextCursor;
                fetchRecords(true);
            }
        }
    }

    // Check if user has unviewed draft records
    $: unviewed = $paginatedRecordsStore.filter(r => r.status === 'Draft' && !r.isViewed && (r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email)));
    $: {
        if (unviewed.length > 0) {
            unviewed.forEach(r => {
                updateRecord(r.id, { isViewed: true });
            });
        }
    }

    onDestroy(() => {
        if (typeof window !== 'undefined') {
            clearTimeout(debounceTimer);
        }
    });

    onMount(() => {
        // Fetch is handled by reactive handleFiltersChanged if store is empty
    });
</script>

<svelte:window on:scroll|passive={handleWindowScroll} />

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Laporan Pasca Dinas</h1>
            <p class="text-sm text-slate-500 mt-1">Upload dan kelola dokumen laporan hasil perjalanan dinas Anda.</p>
        </div>
    </div>

    <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-100 mb-6">
        <AdminTableFilters
            bind:searchQuery
            bind:statusFilter
            bind:sortOption
            bind:startDate
            bind:endDate
            statusOptions={statusOptions}
        />
    </div>

    <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden relative flex flex-col">
        <!-- NATIVE SCROLL CONTAINER -->
        <div class="overflow-x-auto w-full relative min-h-[400px]">
            <table class="w-full text-sm text-left relative border-collapse">
                <thead class="bg-slate-50 sticky top-0 z-20 shadow-sm border-b border-slate-200">
                    <tr>
                        <th class="min-w-[120px] font-semibold text-slate-700 pl-4 py-3 bg-slate-50">{$userStore.role !== 'protokol' ? 'ID SPJ' : 'No. SPJ'}</th>
                        <th class="min-w-[250px] font-semibold text-slate-700 py-3 bg-slate-50">Tujuan & Lokasi</th>
                        <th class="min-w-[160px] font-semibold text-slate-700 py-3 bg-slate-50">Tanggal</th>
                        <th class="w-[120px] min-w-[120px] font-semibold text-slate-700 py-3 bg-slate-50">Status Laporan</th>
                        <th class="w-[150px] min-w-[150px] font-semibold text-slate-700 text-center pr-4 py-3 bg-slate-50">Aksi</th>
                    </tr>
                </thead>
                <tbody>
                    {#if displayRecords.length === 0}
                        <tr>
                            <td colspan="5" class="p-0">
                                {#if $isFetchingRecords}
                                    <SkeletonTable rows={8} />
                                {:else}
                                    <div class="p-12 text-center text-slate-500 w-full">
                                        <EmptyState />
                                    </div>
                                {/if}
                            </td>
                        </tr>
                    {:else}
                        {#each displayRecords as record (record.spd)}
                                <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                    <td class="pl-4 py-4 align-middle">
                                        <span class="inline-flex items-center font-mono text-[13px] font-bold tracking-widest text-slate-700">{record.spd}</span>
                                        <div class="mt-1">
                                            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                                {record.type ? record.type.replace(/_/g, ' ') : 'Dalam Kota'}
                                            </span>
                                        </div>
                                        <div class="mt-2 text-[10px] text-slate-400">
                                            {groupedRecordsMap[record.spd]?.length || 1} Petugas
                                        </div>
                                    </td>
                                    <td class="py-4 align-top">
                                        <div class="font-medium text-slate-800 text-sm line-clamp-2">
                                            {record.stakeholder ? `${record.purpose} ${record.stakeholder}` : record.purpose}
                                        </div>
                                        {#if record.agenda}
                                        <div class="text-[11px] text-slate-600 mt-1.5 font-medium">
                                            <span class="text-slate-400 font-normal mr-1">Agenda:</span>{record.agenda}
                                        </div>
                                        {/if}
                                        <div class="text-xs text-slate-500 mt-1 flex items-center gap-1">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                            </svg>
                                            <span class="line-clamp-2 leading-relaxed">{formatLocations(record)}</span>
                                        </div>
                                    </td>
                                    <td class="py-4 align-top text-xs text-slate-600">
                                        <div class="whitespace-nowrap">
                                            {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                                        </div>
                                        <div class="text-slate-400 my-0.5 text-[10px]">s/d</div>
                                        <div class="whitespace-nowrap">
                                            {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                                        </div>
                                    </td>
                                    <td class="py-4 align-top">
                                        <div class="flex flex-col gap-1.5 items-start">
                                            <span class={cn("inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border", 
                                                record.reportStatus === 'Completed' ? "bg-emerald-50 text-emerald-700 border-emerald-100" : 
                                                record.reportStatus === 'Draft' ? "bg-sky-50 text-sky-700 border-sky-100" : 
                                                "bg-amber-50 text-amber-700 border-amber-100")}>
                                                {record.reportStatus === 'Completed' ? 'Selesai' : 
                                                 record.reportStatus === 'Draft' ? 'Draft' : 'Pending'}
                                            </span>
                                        </div>
                                    </td>
                                    <td class="text-center pr-4 py-4 align-top">
                                        <div class="flex flex-col gap-2">
                                            <div class="flex items-center justify-center gap-2">
                                                <a href={`/dashboard/laporan/${encodeURIComponent(record.spd)}`} class="flex-1">
                                                    <button 
                                                        class={cn("w-full px-2 py-1.5 rounded-lg text-[11px] font-medium shadow-sm transition-all border flex items-center justify-center gap-1.5 whitespace-nowrap", 
                                                            record.reportStatus === 'Completed' || $userStore.role === 'kasubag'
                                                            ? "bg-white text-slate-700 border-slate-200 hover:bg-slate-50 hover:text-blue-600" 
                                                            : "bg-blue-600 border-blue-600 hover:bg-blue-700 text-white shadow-blue-500/20")}
                                                    >
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                                        </svg>
                                                        {$userStore.role === 'kasubag' ? 'Lihat Laporan' : (record.reportStatus === 'Completed' ? 'Edit Laporan' : 'Input Laporan')}
                                                    </button>
                                                </a>
                                            </div>
                                        </div>
                                    </td>
                                </tr>
                            {/each}
                        
                        {#if $isFetchingRecords && currentCursor}
                            <tr>
                                <td colspan="5" class="p-6 text-center">
                                    <div class="flex items-center justify-center gap-1">
                                        <LottieLoader size="32px" />
                                        <span class="text-sm text-slate-500">Memuat data selanjutnya...</span>
                                    </div>
                                </td>
                            </tr>
                        {/if}
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>

<!-- Detail Modal (Surat Tugas & Roadmap) -->
<Dialog bind:open={isDetailModalOpen} class="!w-[95vw] !max-w-5xl !p-0 overflow-hidden flex flex-col rounded-2xl shadow-2xl !max-h-[85vh]">
    <div class="border-b border-slate-100 p-6 bg-white sticky top-0 z-20 flex items-center shrink-0">
        <DialogTitle class="text-xl font-bold text-slate-800 m-0">Detail Perjalanan Dinas</DialogTitle>
    </div>
    
    <div class="p-4 md:p-8 bg-slate-50 space-y-6 md:space-y-8 w-full">
        {#if selectedDetailRecord}
            <div class="grid grid-cols-1 md:grid-cols-3 gap-5 md:gap-6">
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-blue-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col h-full relative overflow-hidden group">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-blue-500 to-cyan-400"></div>
                    <div class="flex items-center gap-3 mb-4 md:mb-5 pb-3 border-b border-slate-100/80 relative z-10">
                        <div class="p-2.5 bg-gradient-to-br from-blue-50 to-blue-100/50 text-blue-600 rounded-xl shadow-sm border border-blue-100">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Informasi Dasar</h4>
                    </div>
                    <div class="space-y-4.5 flex-1 relative z-10">
                        <div>
                            <span class="text-[11px] font-bold text-slate-400 uppercase tracking-widest block mb-1.5">Nomor SPJ</span>
                            <span class="font-mono text-slate-800 font-bold bg-slate-50 px-3 py-1.5 rounded-lg border border-slate-200 inline-block text-sm shadow-sm">{selectedDetailRecord.spd}</span>
                        </div>
                        <div>
                            <span class="text-[11px] font-bold text-slate-400 uppercase tracking-widest block mb-1.5">Periode</span>
                            <span class="font-semibold text-slate-700 block text-[13px] md:text-sm bg-slate-50/50 p-2 rounded-lg border border-slate-100">
                                {new Date(selectedDetailRecord.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short'})} &rarr; {new Date(selectedDetailRecord.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                            </span>
                        </div>
                    </div>
                </div>

                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-emerald-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col h-full relative overflow-hidden group">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-emerald-500 to-teal-400"></div>
                    <div class="flex items-center gap-3 mb-4 md:mb-5 pb-3 border-b border-slate-100/80 relative z-10">
                        <div class="p-2.5 bg-gradient-to-br from-emerald-50 to-emerald-100/50 text-emerald-600 rounded-xl shadow-sm border border-emerald-100">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" /></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Lokasi</h4>
                    </div>
                    <div class="space-y-4.5 flex-1 relative z-10">
                        <div class="bg-slate-50 p-3 rounded-lg border border-slate-100 text-sm font-bold">
                            {formatLocations(selectedDetailRecord)}
                        </div>
                    </div>
                </div>

                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-indigo-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col h-full relative overflow-hidden group">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-indigo-500 to-purple-400"></div>
                    <div class="flex items-center gap-3 mb-4 md:mb-5 pb-3 border-b border-slate-100/80 relative z-10">
                        <div class="p-2.5 bg-gradient-to-br from-indigo-50 to-indigo-100/50 text-indigo-600 rounded-xl shadow-sm border border-indigo-100">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Tim</h4>
                    </div>
                    <div class="text-sm font-bold text-slate-600">
                        {selectedDetailRecord.employeesList.length} Petugas terlibat dalam SPJ ini.
                    </div>
                </div>
            </div>

            <div class="space-y-6">
                <div class="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
                    <div class="p-4 bg-slate-50 border-b border-slate-100 font-bold text-xs uppercase tracking-widest text-slate-600">Roadmap Perjalanan</div>
                    <div class="p-6 overflow-x-auto">
                        <TripStepper record={selectedDetailRecord} />
                    </div>
                </div>
            </div>
        {/if}
    </div>
    
    <div class="p-6 border-t border-slate-200 bg-white flex justify-end">
        <Button class="min-w-[120px]" variant="outline" on:click={() => isDetailModalOpen = false}>Tutup</Button>
    </div>
</Dialog>

<!-- Document Preview Modal -->
<Dialog open={isPreviewOpen} hideCloseButton={true} on:close={() => isPreviewOpen = false} class="!w-[90vw] !max-w-6xl !h-[85vh] !p-0 overflow-hidden rounded-xl">
    <div class="h-full flex flex-col">
        <div class="flex items-center justify-between px-6 py-4 border-b border-slate-100 bg-white">
            <h3 class="font-bold text-slate-800">Pratinjau Dokumen</h3>
            <button type="button" class="p-2 text-slate-400 hover:text-red-500 transition-colors" aria-label="Tutup pratinjau" on:click={() => isPreviewOpen = false}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
        </div>
        <div class="flex-1 bg-slate-100">
            {#if previewUrl}
                <DocumentViewer url={previewUrl} type={previewType} filename={previewFilename} />
            {/if}
        </div>
    </div>
</Dialog>
