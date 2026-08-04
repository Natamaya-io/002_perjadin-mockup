<script>
    import { paginatedRecordsStore, paginatedMetadataStore, loadPaginatedRecords, deleteRecordBySpd, isFetchingRecords, updateRecord } from '$lib/features/pengajuan/store';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';
    import { userStore } from '$lib/features/auth/store';
    import { onMount, onDestroy } from 'svelte';
    import { provincesStore, stakeholdersStore } from '$lib/shared/stores/master-data';
    import { toast } from '$lib/shared/stores/toast';
    import { getStatusBadge, formatLocations, formatCurrency } from '$lib/shared/utils/utils';
    
    import ReviewModal from '$lib/features/pengajuan/ui/ReviewModal.svelte';
    import { ConfirmationModal } from '$lib/shared/ui/confirmation-modal';
    import AdminTableFilters from '$lib/features/admin/ui/AdminTableFilters.svelte';
    import SkeletonTable from '$lib/shared/ui/loader/SkeletonTable.svelte';

    // Filter & Sort State
    let searchQuery = '';
    let statusFilter = 'all';
    let sortOption = 'spj-desc';
    let startDate = '';
    let endDate = '';
    
    let statusOptions = [
        { value: 'all', label: 'Semua Status' },
        { value: 'Draft', label: 'Draft' },
        { value: 'Submitted', label: 'Menunggu Persetujuan' },
        { value: 'Approved', label: 'Disetujui' },
        { value: 'Rejected', label: 'Ditolak' }
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

    // Reactively refetch when filters change (Resetting)
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

        let statusParam = statusFilter === 'all' ? undefined : statusFilter;
        let sortByParam = sortOption;

        let startIso, endIso;
        if (startDate) {
            const start = new Date(startDate);
            start.setHours(0,0,0,0);
            startIso = start.toISOString();
        }
        if (endDate) {
            const end = new Date(endDate);
            end.setHours(23,59,59,999);
            endIso = end.toISOString();
        }

        const targetUserId = ($userStore.role === 'super_admin' || $userStore.role === 'kasubag') ? undefined : $userStore.id;

        const params = {
            limit: 50, // Match Admin Perdin limit
            search: searchQuery,
            status: statusParam,
            sort_by: sortByParam,
            start_date: startIso,
            end_date: endIso,
            user_id: targetUserId
        };

        if (append && currentCursor) {
            params.cursor = currentCursor;
        }

        loadPaginatedRecords(params, append);
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

    let isReviewOpen = false;
    let selectedRecords = [];
    let isDeleteModalOpen = false;
    let spdToDelete = '';

    function handleReview(spd) {
        selectedRecords = $paginatedRecordsStore.filter(r => r.spd === spd);
        isReviewOpen = true;
    }

    function handleDelete(spd) {
        spdToDelete = spd;
        isDeleteModalOpen = true;
    }
    
    async function processDelete() {
        if (spdToDelete) {
            try {
                await deleteRecordBySpd(spdToDelete);
                toast.success(`Pengajuan ${spdToDelete} berhasil dihapus.`);
            } catch (e) {
                toast.error(`Gagal menghapus pengajuan ${spdToDelete}.`);
            } finally {
                spdToDelete = '';
                isDeleteModalOpen = false;
            }
        }
    }

    // Check for unviewed
    $: unviewed = $paginatedRecordsStore.filter(r => r.status === 'Draft' && !r.isViewed && (r.email === $userStore.email || r.employeeId === $userStore.id));
    $: if (unviewed.length > 0) {
        unviewed.forEach(r => updateRecord(r.id, { isViewed: true }));
    }

    onDestroy(() => {
        if (typeof window !== 'undefined') clearTimeout(debounceTimer);
    });

    onMount(() => {
        // Fetch is handled by reactive handleFiltersChanged if store is empty
    });
</script>

<svelte:window on:scroll|passive={handleWindowScroll} />

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Daftar Pengajuan Saya</h1>
            <p class="text-sm text-slate-500 mt-1">Pantau status dan riwayat perjalanan dinas yang telah Anda ajukan.</p>
        </div>
        <div class="flex items-center gap-3">
            <a href="/dashboard/pengajuan/import" class="inline-flex items-center justify-center bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 font-medium px-4 py-2.5 rounded-xl shadow-sm transition-all duration-200">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                Import Excel
            </a>
            <a href="/dashboard/pengajuan/new" class="inline-flex items-center justify-center bg-blue-600 hover:bg-blue-700 text-white font-medium px-5 py-2.5 rounded-xl shadow-sm shadow-blue-500/30 transition-all duration-200">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
                </svg>
                Buat Pengajuan Baru
            </a>
        </div>
    </div>

    <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-100">
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
                        <th class="min-w-[120px] font-semibold text-slate-700 pl-4 py-3 bg-slate-50">ID SPJ</th>
                        <th class="min-w-[250px] font-semibold text-slate-700 py-3 bg-slate-50">Tujuan & Lokasi</th>
                        <th class="min-w-[130px] font-semibold text-slate-700 py-3 bg-slate-50">Tanggal</th>
                        <th class="min-w-[160px] font-semibold text-slate-700 text-center py-3 bg-slate-50">Status</th>
                        <th class="w-[140px] min-w-[140px] font-semibold text-slate-700 text-center pr-4 py-3 bg-slate-50">Aksi</th>
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
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                        </svg>
                                        Belum ada pengajuan.
                                    </div>
                                {/if}
                            </td>
                        </tr>
                    {:else}
                        {#each displayRecords as record (record.spd)}
                                <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                    <td class="pl-4 py-4 align-middle">
                                        <span class="inline-flex items-center font-mono text-[13px] font-bold tracking-widest text-slate-700">{record.spd}</span>
                                        <div class="mt-1.5">
                                            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                                {record.type ? record.type.replace(/_/g, ' ') : 'Dalam Kota'}
                                            </span>
                                        </div>
                                        <div class="mt-2 text-[10px] text-slate-400">
                                            {groupedRecordsMap[record.spd]?.length || 1} Petugas
                                        </div>
                                    </td>
                                    <td class="py-4 align-middle">
                                        <div class="font-medium text-slate-800 text-sm line-clamp-2">
                                            {record.stakeholder ? `${record.purpose} ${record.stakeholder}` : record.purpose}
                                        </div>
                                        <div class="text-xs text-slate-500 mt-1 flex items-center gap-1">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                            </svg>
                                            <span class="line-clamp-2 leading-relaxed">{formatLocations(record)}</span>
                                        </div>
                                    </td>
                                    <td class="py-4 align-middle text-xs text-slate-600">
                                        <div class="whitespace-nowrap">
                                            {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                                        </div>
                                        <div class="text-slate-400 my-0.5 text-[10px]">s/d</div>
                                        <div class="whitespace-nowrap">
                                            {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                                        </div>
                                    </td>
                                    <td class="py-4 align-middle text-center">
                                        <div class="flex flex-col gap-1.5 items-center justify-center">
                                            <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {getStatusBadge(record).class}">
                                                {getStatusBadge(record).label}
                                            </span>
                                        </div>
                                    </td>
                                    <td class="text-center pr-4 py-4 align-middle">
                                        <div class="flex items-center justify-center gap-2">
                                            <button 
                                                class="text-blue-600 hover:text-blue-800 bg-blue-50 hover:bg-blue-100 p-1.5 rounded transition-colors" 
                                                title="Review"
                                                on:click={() => handleReview(record.spd)}
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                                </svg>
                                            </button>
                                            {#if $userStore?.role === 'super_admin' || $userStore?.role === 'kasubag' || (record.creator?.email || record.email) === $userStore?.email}
                                                <button 
                                                    class="text-red-600 hover:text-red-800 bg-red-50 hover:bg-red-100 p-1.5 rounded transition-colors" 
                                                    title="Hapus"
                                                    on:click={() => handleDelete(record.spd)}
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                                    </svg>
                                                </button>
                                            {/if}
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

<ReviewModal
    bind:open={isReviewOpen}
    records={selectedRecords}
    provinces={$provincesStore}
    stakeholders={$stakeholdersStore}
    on:saved={() => fetchRecords(false)}
/>

<ConfirmationModal 
    bind:open={isDeleteModalOpen}
    title="Hapus Pengajuan"
    description="Apakah Anda yakin ingin membatalkan dan menghapus pengajuan SPJ {spdToDelete}? Tindakan ini tidak dapat dibatalkan."
    confirmText="Ya, Hapus"
    cancelText="Batal"
    onConfirm={processDelete}
    variant="destructive"
/>
