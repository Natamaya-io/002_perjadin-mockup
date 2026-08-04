<script>
    import { paginatedRecordsStore, paginatedMetadataStore, loadPaginatedRecords, updateRecord, isFetchingRecords } from '$lib/features/pengajuan/store';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';
    import { userStore } from '$lib/features/auth/store';
    import { provincesStore } from '$lib/shared/stores/master-data';
    import { loadingStore, startLoading, stopLoading } from '$lib/shared/stores/loading';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { toast } from '$lib/shared/stores/toast';
    import { getStatusBadge, toTitleCase, formatLocations, formatCurrency } from '$lib/shared/utils/utils';
    import { api } from '$lib/shared/api';
    
    // Components
    import AdminHeader from '$lib/features/admin/ui/AdminHeader.svelte';
    import AdminTableFilters from '$lib/features/admin/ui/AdminTableFilters.svelte';

    import AdminTable from '$lib/features/admin/ui/AdminTable.svelte';
    import AdminTableRow from '$lib/features/admin/ui/AdminTableRow.svelte';
    import CostModal from '$lib/features/admin/ui/CostModal.svelte';
    
    // UI Helpers
    import { ConfirmationModal } from '$lib/shared/ui/confirmation-modal';

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

    // Filter & Sort State
    let searchQuery = '';
    let statusFilter = 'all';
    let sortOption = 'spj-desc';
    let startDate = '';
    let endDate = '';

    let statusOptions = [
        { value: 'all', label: 'Semua Status' },
        { value: 'Submitted', label: 'Menunggu Persetujuan' },
        { value: 'Approved', label: 'Disetujui' },
        { value: 'Rejected', label: 'Ditolak' }
    ];

    /** @type {ReturnType<typeof setTimeout>} */
    let debounceTimer;
    let isInitialMount = true;
    let isModalOpen = false;
    let selectedRecord = null;
    let editingCosts = {};
    let allRecordsInSelectedSpd = [];
    $: if (selectedRecord) {
        allRecordsInSelectedSpd = $paginatedRecordsStore.filter(r => r.spd === selectedRecord.spd);
    } else {
        allRecordsInSelectedSpd = [];
    }

    let pendingGrandTotal = 0; // To store calculation result for confirmation
    let pendingOtherUpdatesToSave = []; // From split hotel feature
    let pendingSpdToSave = '';

    let isConfirmOpen = false;
    let isRejectConfirmOpen = false;
    let isPaidConfirmOpen = false;
    let recordToPay = null;

    let expandedGroups = {};

    function toggleGroup(spd) {
        expandedGroups[spd] = !expandedGroups[spd];
        expandedGroups = expandedGroups; // Trigger reactivity
    }

    let limit = 50;
    let currentCursor = '';

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

    function fetchRecords(append = false) {        let statusParam = statusFilter === 'all' ? undefined : statusFilter;
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

        const params = {
            limit,
            search: searchQuery,
            status: statusParam,
            sort_by: sortByParam,
            start_date: startIso,
            end_date: endIso,
        };

        if (append && currentCursor) {
            params.cursor = currentCursor;
        }

        loadPaginatedRecords(params, append);
    }

    // Grouping is now extremely cheap because the store only contains the current page (~10 SPDs max)
    $: groupedRecordsMap = $paginatedRecordsStore.reduce((acc, record) => {
        if (!acc[record.spd]) {
            acc[record.spd] = { ...record, employeesList: [record] };
        } else {
            acc[record.spd].employeesList.push(record);
            acc[record.spd].employeesList.sort((a,b) => (b.sequenceNumber || 0) - (a.sequenceNumber || 0));
        }
        return acc;
    }, {});

    // Maintain the order of SPDs exactly as returned by Postgres pagination query
    $: uniqueSPDs = [...new Set($paginatedRecordsStore.map(r => r.spd))];
    $: uniqueRecords = uniqueSPDs.map(spd => groupedRecordsMap[spd]).filter(Boolean);

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

    async function openEditModal(record) {
        selectedRecord = record;
        editingCosts = { ...record.costs }; // Clone costs as fallback
        
        startLoading('Memuat detail...');
        try {
            const fullRecord = await api.getRecordById(record.id);
            if (fullRecord) {
                selectedRecord = fullRecord;
                editingCosts = { ...fullRecord.costs };
            }
        } catch (e) {
            console.error("Gagal memuat detail record:", e);
            toast.error("Gagal memuat detail secara penuh. Beberapa foto mungkin tidak tampil.");
        } finally {
            stopLoading();
        }

        pendingOtherUpdatesToSave = []; // Reset
        pendingSpdToSave = selectedRecord.spd || ''; // Reset pending SPD
        isModalOpen = true;
    }

    function handleModalSave(event) {
        const { editingCosts: newCosts, grandTotal, pendingOtherUpdates, spd } = event.detail;

        if (newCosts.localTransport > 500000) {
            toast.warning('Transport Lokal maksimal Rp 500.000');
            return;
        }

        // Store temp state for confirmation
        editingCosts = newCosts;
        pendingGrandTotal = grandTotal;
        pendingOtherUpdatesToSave = pendingOtherUpdates || [];
        pendingSpdToSave = spd || selectedRecord.spd;
        isConfirmOpen = true;
    }

    import { updateMultipleRecords } from '$lib/features/pengajuan/store';

    async function processSave() {
        if (!selectedRecord) return;

        startLoading();
        try {
            const updates = [];
            
            // 1. Primary update
            updates.push({
                id: selectedRecord.id,
                data: {
                    spd: pendingSpdToSave, // Send the edited SPD
                    costs: { ...editingCosts },
                    totalCost: pendingGrandTotal,
                    status: 'Approved'
                }
            });

            // 2. Additional updates from split hotel feature
            if (pendingOtherUpdatesToSave.length > 0) {
                for (const updateInfo of pendingOtherUpdatesToSave) {
                    const targetRecord = $paginatedRecordsStore.find(r => r.id === updateInfo.empId);
                    if (!targetRecord) continue;

                    let fullTargetRecord = null;
                    try {
                        fullTargetRecord = await api.getRecordById(updateInfo.empId);
                    } catch (err) {
                        console.error("Gagal load fullTargetRecord", err);
                    }
                    
                    if (!fullTargetRecord) fullTargetRecord = targetRecord;

                    let newTargetCosts = JSON.parse(JSON.stringify(fullTargetRecord.costs || {}));
                    if (!newTargetCosts.details) newTargetCosts.details = [];
                    
                    // === FIX: inisialisasi detail locations sesuai jumlah lokasi targetRecord
                    // Jika Doni belum pernah buka CostModal, details-nya masih [] (kosong).
                    // Akibatnya index check di bawah selalu false dan hotelRate tidak pernah di-set.
                    const targetLocCount = targetRecord.locations?.length || 1;
                    while (newTargetCosts.details.length < targetLocCount) {
                        newTargetCosts.details.push({
                            transportMode: 'Pesawat/Kendaraan Umum',
                            ticketGo: 0, ticketBack: 0,
                            hotelDays: 0, hotelRate: 0,
                            transportAmount: 0,
                            additionalCosts: [],
                            boardingPassFiles: []
                        });
                    }
                    // === END FIX
                    
                    if (newTargetCosts.details[updateInfo.locationIndex]) {
                        if (updateInfo.isExtend) {
                            if (!newTargetCosts.details[updateInfo.locationIndex].additionalCosts) {
                                newTargetCosts.details[updateInfo.locationIndex].additionalCosts = [];
                            }
                            let targetExtendIdx = newTargetCosts.details[updateInfo.locationIndex].additionalCosts.findIndex(c => c.name === 'Extend Penginapan');
                            if (targetExtendIdx === -1) {
                                newTargetCosts.details[updateInfo.locationIndex].additionalCosts.push({
                                    name: 'Extend Penginapan',
                                    hotelRate: updateInfo.hotelRate,
                                    hotelDays: updateInfo.hotelDays,
                                    amount: updateInfo.hotelRate * updateInfo.hotelDays,
                                    file: updateInfo.hotelFile ? { ...updateInfo.hotelFile } : null,
                                    hotelOriginalRate: updateInfo.hotelOriginalRate || null,
                                    hotelSplitWith: updateInfo.hotelSplitWith || []
                                });
                            } else {
                                const targetCost = newTargetCosts.details[updateInfo.locationIndex].additionalCosts[targetExtendIdx];
                                targetCost.hotelRate = updateInfo.hotelRate;
                                targetCost.hotelDays = updateInfo.hotelDays;
                                targetCost.amount = updateInfo.hotelRate * updateInfo.hotelDays;
                                // Simpan state split agar bisa restore saat modal dibuka lagi
                                if (updateInfo.hotelOriginalRate) targetCost.hotelOriginalRate = updateInfo.hotelOriginalRate;
                                if (updateInfo.hotelSplitWith) targetCost.hotelSplitWith = updateInfo.hotelSplitWith;
                                if (updateInfo.hotelFile) {
                                    targetCost.file = { ...updateInfo.hotelFile };
                                }
                            }
                        } else {
                            newTargetCosts.details[updateInfo.locationIndex].hotelRate = updateInfo.hotelRate;
                            newTargetCosts.details[updateInfo.locationIndex].hotelDays = updateInfo.hotelDays;
                            // Simpan state split agar bisa restore saat modal dibuka lagi
                            if (updateInfo.hotelOriginalRate) {
                                newTargetCosts.details[updateInfo.locationIndex].hotelOriginalRate = updateInfo.hotelOriginalRate;
                            }
                            if (updateInfo.hotelSplitWith) {
                                newTargetCosts.details[updateInfo.locationIndex].hotelSplitWith = updateInfo.hotelSplitWith;
                            }
                            if (updateInfo.hotelFile) {
                                newTargetCosts.details[updateInfo.locationIndex].hotelFile = updateInfo.hotelFile;
                            }
                        }
                    }

                    // Recalculate total for the other person
                    const recalculatedGrandTotal = (newTargetCosts.details || []).reduce((acc, detail, idx) => {
                        const loc = targetRecord.locations && targetRecord.locations[idx] ? targetRecord.locations[idx] : null;
                        let sbmTotal = 0;
                        if (loc) {
                            const provData = $provincesStore.find(p => p.name === loc.province);
                            const rate = provData ? provData.luarKota : (newTargetCosts.dailyAllowanceRate || 0);
                            const start = new Date(loc.startDate);
                            const end = new Date(loc.endDate);
                            if (!isNaN(start.getTime()) && !isNaN(end.getTime())) {
                                const diffDays = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)) + 1;
                                sbmTotal = rate * (diffDays > 0 ? diffDays : 0);
                            }
                        }
                        const hotel = (detail.hotelDays || 0) * (detail.hotelRate || 0);
                        const ticket = Number(detail.ticketGo || 0) + Number(detail.ticketBack || 0);
                        const transport = Number(detail.transportAmount || 0);
                        const addCosts = (detail.additionalCosts || []).reduce((sum, c) => {
                            if (c.name === 'Extend Tiket') {
                                return sum + (Number(c.ticketGo) || 0) + (Number(c.ticketBack) || 0) + (Number(c.amount) || 0);
                            }
                            return sum + (c.amount || 0);
                        }, 0);
                        return acc + sbmTotal + hotel + ticket + transport + addCosts;
                    }, 0);

                    updates.push({
                        id: updateInfo.empId,
                        data: {
                            costs: newTargetCosts,
                            totalCost: recalculatedGrandTotal,
                            status: 'Approved'
                        }
                    });
                }
            }

            await updateMultipleRecords(updates);
            
            toast.success('Rincian biaya berhasil disimpan!');
            fetchRecords();
        } catch (e) {
            toast.error('Gagal menyimpan perubahan.');
        } finally {
            stopLoading();
            isModalOpen = false;
            isConfirmOpen = false; // Close confirmation modal too
        }
    }

    function handleModalReject(event) {
        // Just open confirmation
        isRejectConfirmOpen = true;
    }

    async function processReject() {
        if (!selectedRecord) return;
        startLoading();
        try {
            await updateMultipleRecords([{
                id: selectedRecord.id,
                data: {
                    status: 'Rejected'
                }
            }]);
            toast.success('Pengajuan berhasil ditolak.');
            fetchRecords();
        } catch (e) {
            toast.error('Gagal menolak pengajuan.');
        } finally {
            stopLoading();
            isModalOpen = false;
            isRejectConfirmOpen = false;
        }
    }

    function markAsPaid(record) {
        recordToPay = record;
        isPaidConfirmOpen = true;
    }

    async function processPaid() {
        if (!recordToPay) return;
        startLoading();
        try {
            await updateMultipleRecords([{
                id: recordToPay.id,
                data: {
                    paymentStatus: 'Paid'
                }
            }]);
            toast.success('Dana berhasil dicairkan. Status menjadi Completed.');
            fetchRecords();
        } catch (e) {
            toast.error('Gagal memproses pencairan dana.');
        } finally {
            stopLoading();
            isPaidConfirmOpen = false;
            recordToPay = null;
        }
    }

    function generateSPD(record) {
        // Find the specific travel record ID or use the SPD string?
        // Since SPD is shared, passing spd string will just pick the first one. We should pass ID in the real app, but for now we'll pass SPD and the employee id if possible.
        // Actually, we must change the print route to use record id, but to avoid breaking print, we'll open it with the specific record.
        // For now, keeping the original behavior but warning that it prints the first match.
        // Wait, I can pass &id=record.id to print
        window.open(`/print?type=spd&id=${record.id}&spd=${encodeURIComponent(record.spd)}`, '_blank');
    }

    function generateRincian(record) {
        window.open(`/print?type=rincian&id=${record.id}&spd=${encodeURIComponent(record.spd)}`, '_blank');
    }
    onMount(() => {
        if ($userStore.role !== 'super_admin' && $userStore.role !== 'kasubag') {
            goto('/dashboard');
        }
    });
</script>

<svelte:window on:scroll|passive={handleWindowScroll} />

<div class="space-y-6 pb-20">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
        <div>
            <h1 class="text-2xl font-bold text-slate-900 tracking-tight">Rekap & Kalkulasi</h1>
            <p class="text-sm text-slate-500 mt-1">Review dan verifikasi rincian biaya perjalanan dinas.</p>
        </div>
    </div>

    <div class="bg-white rounded-xl shadow-sm border border-slate-200 p-4 mb-6">
        <AdminTableFilters
            bind:searchQuery
            bind:statusFilter
            bind:sortOption
            bind:startDate
            bind:endDate
            statusOptions={statusOptions}
        />
    </div>
    {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}

        <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
            <div class="hidden lg:block overflow-x-auto w-full relative">
                <table class="w-full text-left text-sm border-collapse min-w-[900px] relative">
                    <thead class="bg-slate-50 border-b border-slate-200 text-xs uppercase font-semibold text-slate-500 sticky top-0 z-10 shadow-sm">
                        <tr>
                            <th class="px-6 py-4 whitespace-nowrap w-[12%]">ID SPJ</th>
                            <th class="px-6 py-4 whitespace-nowrap w-[28%]">Lokasi</th>
                            <th class="px-6 py-4 whitespace-nowrap w-[20%]">Tanggal</th>
                            <th class="px-6 py-4 whitespace-nowrap text-right w-[25%]">Total Biaya Akhir</th>
                            <th class="px-6 py-4 whitespace-nowrap text-center w-[15%]">Status</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100">
                        {#if uniqueRecords.length === 0}
                            {#if $isFetchingRecords}
                                {#each Array(5) as _}
                                    <tr class="bg-white border-b border-slate-100 animate-pulse">
                                        <td class="px-6 py-4 whitespace-nowrap"><div class="h-4 bg-slate-200 rounded w-20"></div></td>
                                        <td class="px-6 py-4">
                                            <div class="h-4 bg-slate-200 rounded w-48 mb-2"></div>
                                            <div class="h-3 bg-slate-100 rounded w-32"></div>
                                        </td>
                                        <td class="px-6 py-4 whitespace-nowrap"><div class="h-4 bg-slate-200 rounded w-32"></div></td>
                                        <td class="px-6 py-4 whitespace-nowrap text-right"><div class="h-6 bg-slate-200 rounded-lg w-24 ml-auto"></div></td>
                                        <td class="px-6 py-4 whitespace-nowrap text-center"><div class="h-5 bg-slate-200 rounded-full w-20 mx-auto"></div></td>
                                    </tr>
                                {/each}
                            {:else}
                                <tr>
                                    <td colspan="5" class="p-12 text-center text-slate-500 bg-slate-50/50">
                                        Belum ada pengajuan yang masuk.
                                    </td>
                                </tr>
                            {/if}
                        {:else}
                            {#each uniqueRecords as record (record.spd)}
                                <!-- Group Header Row -->
                            <tr class="bg-slate-50/80 border-b border-slate-200 cursor-pointer hover:bg-slate-100 transition-colors select-none" on:click={() => toggleGroup(record.spd)}>
                                <td class="px-6 py-3 whitespace-nowrap">
                                    <div class="flex items-center gap-3">
                                        <button class="p-1 rounded-md hover:bg-slate-200 transition-colors focus:outline-none shrink-0" aria-label="Toggle details">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-500 transition-transform duration-200 {expandedGroups[record.spd] ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                            </svg>
                                        </button>
                                        <span class="inline-flex items-center font-mono text-[13px] font-bold tracking-widest text-slate-700">{record.spd}</span>
                                    </div>
                                </td>
                                <td class="px-6 py-3">
                                    <div class="text-sm text-slate-800 font-medium line-clamp-1" title={formatLocations(record)}>{formatLocations(record)}</div>
                                </td>
                                <td class="px-6 py-3 whitespace-nowrap">
                                    <div class="text-xs text-slate-500 flex items-center gap-1.5">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                        </svg>
                                        <span>{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})} &mdash; {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</span>
                                    </div>
                                </td>
                                <td class="px-6 py-3 whitespace-nowrap text-right">
                                    <div class="px-3 py-1 inline-flex bg-blue-50 text-blue-700 font-mono font-bold text-xs rounded-lg border border-blue-100">
                                        {formatCurrency(record.employeesList.reduce((sum, e) => sum + (e.totalCost || 0), 0))}
                                    </div>
                                </td>
                                <td class="px-6 py-3 whitespace-nowrap text-center">
                                    <span class="inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide border {getStatusBadge(record).class}">
                                        {getStatusBadge(record).label}
                                    </span>
                                </td>
                            </tr>
                            <!-- Employee Rows -->
                            {#if expandedGroups[record.spd]}
                                <!-- Nested Table Head for Employee Data inside a single spanning cell to break column dependency -->
                                <tr>
                                    <td colspan="5" class="p-0 border-b border-slate-200">
                                        <table class="w-full text-left text-sm border-collapse bg-white">
                                            <thead class="bg-slate-100/50 border-y border-slate-200 text-[11px] uppercase tracking-wider font-semibold text-slate-500 shadow-inner">
                                                <tr>
                                                    <th class="px-6 py-3 text-left font-semibold w-[12%]">NO. SPD</th>
                                                    <th class="px-6 py-3 text-left font-semibold w-[28%]">Nama Pegawai & Kontak</th>
                                                    <th class="px-6 py-3 text-left font-semibold w-[20%]">Kelengkapan</th>
                                                    <th class="px-6 py-3 text-right font-semibold w-[25%]">Total Biaya</th>
                                                    <th class="px-6 py-3 text-center font-semibold w-[15%]">Aksi</th>
                                                </tr>
                                            </thead>
                                            <tbody class="divide-y divide-slate-100">
                                                {#each record.employeesList as empRecord}
                                                    <tr class="hover:bg-slate-50/50 transition-colors">
                                                        <td class="px-6 py-4 align-middle">
                                                            <div class="flex items-center gap-3 pl-2">
                                                                <div class="w-2 h-2 rounded-full bg-slate-300 shrink-0"></div>
                                                                <div class="font-mono font-bold text-slate-700">{String(empRecord.sequenceNumber || 0).padStart(3, '0')}</div>
                                                            </div>
                                                        </td>
                                                        <td class="px-6 py-4 align-middle">
                                                            <div class="font-medium text-slate-900">{empRecord.employee?.name || '-'}</div>
                                                            <div class="text-xs text-slate-500 mb-1">
                                                                {#if empRecord.employee?.jabatan && empRecord.employee.jabatan !== '-'}
                                                                    <div class="font-medium text-slate-600">{empRecord.employee.jabatan}</div>
                                                                {/if}
                                                                {#if empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-' && empRecord.employee?.golongan && empRecord.employee.golongan !== '-'}
                                                                    <div class="mt-0.5">{empRecord.employee.pangkat} ({empRecord.employee.golongan})</div>
                                                                {:else if empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-'}
                                                                    <div class="mt-0.5">{empRecord.employee.pangkat}</div>
                                                                {:else if empRecord.employee?.golongan && empRecord.employee.golongan !== '-'}
                                                                    <div class="mt-0.5">{empRecord.employee.golongan}</div>
                                                                {:else if empRecord.employee?.rank && empRecord.employee.rank !== '-'}
                                                                    <div class="mt-0.5">{empRecord.employee.rank}</div>
                                                                {/if}
                                                            </div>
                                                            <div class="flex items-center gap-1.5 text-sm text-slate-600">
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" /></svg>
                                                                <span class="text-xs">{empRecord.employee?.nomorHp || '-'}</span>
                                                            </div>
                                                        </td>
                                                        <td class="px-6 py-4 align-middle text-left">
                                                            <span class="inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide border {empRecord.status === 'Rejected' ? 'bg-red-50 text-red-700 border-red-200' : empRecord.status === 'Draft' ? 'bg-yellow-50 text-yellow-700 border-yellow-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200'}">
                                                                {empRecord.status === 'Rejected' ? 'Rejected' : empRecord.status === 'Draft' ? 'Belum Lengkap' : 'Lengkap'}
                                                            </span>
                                                        </td>
                                                        <td class="px-6 py-4 align-middle text-right font-mono font-medium text-blue-600">
                                                            {formatCurrency(empRecord.totalCost || 0)}
                                                        </td>
                                                        <td class="px-6 py-4 align-middle pr-6">
                                                            <div class="flex items-center justify-center gap-2">
                                                                {#if empRecord.status === 'Approved' && empRecord.paymentStatus !== 'Paid'}
                                                                    <button class="bg-emerald-600 text-white hover:bg-emerald-700 px-3 py-1.5 rounded-md text-xs font-bold transition-all shadow-md shadow-emerald-500/20 whitespace-nowrap" on:click={(e) => { e.stopPropagation(); markAsPaid(empRecord); }}>
                                                                        Cairkan Dana
                                                                    </button>
                                                                {/if}
                                                                <button class="bg-white border {empRecord.status === 'Approved' ? 'border-emerald-200 text-emerald-600 hover:bg-emerald-50' : 'border-slate-200 text-blue-600 hover:bg-blue-50'} px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm whitespace-nowrap" on:click={(e) => { e.stopPropagation(); openEditModal(empRecord); }}>
                                                                    {empRecord.status === 'Approved' ? 'Detail & Edit' : 'Review'}
                                                                </button>
                                                            </div>
                                                        </td>
                                                    </tr>
                                                {/each}
                                            </tbody>
                                        </table>
                                    </td>
                                </tr>
                            {/if}
                            {/each}
                        
                        {#if $isFetchingRecords && currentCursor}
                            <tr>
                                <td colspan="5" class="p-4 text-center">
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

            <!-- Mobile Stacked/Card View -->
            <div class="lg:hidden flex flex-col divide-y divide-slate-100 bg-slate-50">
                {#each uniqueRecords as record (record.spd)}
                    <div class="flex flex-col">
                        <!-- Group Header (Mobile) -->
                        <div role="button" tabindex="0" class="p-4 bg-slate-100/80 border-b border-slate-200 cursor-pointer hover:bg-slate-200 transition-colors select-none" on:click={() => toggleGroup(record.spd)} on:keydown={(e) => e.key === 'Enter' && toggleGroup(record.spd)}>
                            <div class="flex justify-between items-start mb-2 gap-2">
                                <div class="flex items-center flex-wrap gap-2">
                                    <span class="inline-flex items-center font-mono text-[13px] font-bold tracking-widest text-slate-700">{record.spd}</span>
                                    <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                        {record.type ? record.type.replace(/_/g, ' ') : 'Dalam Kota'}
                                    </span>
                                </div>
                                <button class="p-1 rounded-md bg-white border border-slate-200 shadow-sm hover:bg-slate-50 transition-colors focus:outline-none" aria-label="Toggle details">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-500 transition-transform duration-200 {expandedGroups[record.spd] ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                    </svg>
                                </button>
                            </div>
                            <div class="text-sm font-semibold text-slate-800 leading-tight mb-2 pr-6">{record.purpose}</div>
                            <div class="flex flex-col gap-1 text-xs text-slate-600">
                                <div class="flex items-center gap-1.5">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                    </svg>
                                    <span class="truncate">{formatLocations(record)}</span>
                                </div>
                                <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1.5 sm:gap-0 mt-1">
                                    <div class="flex items-center gap-1.5 text-slate-500 min-w-0">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                        </svg>
                                        <span class="truncate">{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})} &mdash; {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</span>
                                    </div>
                                    <span class="font-semibold text-[10px] text-slate-500 uppercase tracking-wider shrink-0">{record.employeesList.length} Petugas</span>
                                </div>
                            </div>
                        </div>

                        <!-- Employees (Mobile) -->
                        {#if expandedGroups[record.spd]}
                            <div class="flex flex-col divide-y divide-slate-100 bg-white">
                                {#each record.employeesList as empRecord}
                                    <div class="p-4 flex flex-col gap-3">
                                        <div class="flex justify-between items-center text-[10px] font-semibold text-slate-400 uppercase tracking-wider">
                                            <span>NO. SPD: {String(empRecord.sequenceNumber || 0).padStart(3, '0')}</span>
                                            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide border {getStatusBadge(empRecord).class}">
                                                {getStatusBadge(empRecord).label}
                                            </span>
                                        </div>
                                        <div class="flex justify-between items-end gap-2">
                                            <div class="flex flex-col min-w-0">
                                                <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-0.5">Pegawai</span>
                                                <div class="font-medium text-sm text-slate-900 leading-tight truncate">{empRecord.employee?.name || '-'}</div>
                                            </div>
                                            <div class="flex gap-1.5">
                                                {#if empRecord.status === 'Approved' && empRecord.paymentStatus !== 'Paid'}
                                                    <button class="bg-emerald-600 text-white hover:bg-emerald-700 px-3 py-1.5 rounded-md text-[10px] font-bold uppercase transition-all shadow-md shadow-emerald-500/20 whitespace-nowrap" on:click={(e) => { e.stopPropagation(); markAsPaid(empRecord); }}>
                                                        Cairkan Dana
                                                    </button>
                                                {/if}
                                                <button 
                                                    class="bg-white border {empRecord.status === 'Approved' ? 'border-emerald-200 text-emerald-600 hover:bg-emerald-50' : 'border-slate-200 text-blue-600 hover:bg-blue-50'} px-3 py-1.5 rounded-md text-[10px] font-bold uppercase transition-all shadow-sm whitespace-nowrap" 
                                                    on:click={(e) => { e.stopPropagation(); openEditModal(empRecord); }}
                                                >
                                                    {empRecord.status === 'Approved' ? 'Edit' : 'Review'}
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </div>
                {/each}
                {#if uniqueRecords.length === 0}
                    {#if $isFetchingRecords}
                        {#each Array(4) as _}
                            <div class="p-4 bg-white border-b border-slate-200 animate-pulse flex flex-col gap-3">
                                <div class="flex justify-between items-start">
                                    <div class="flex items-center gap-2">
                                        <div class="h-5 bg-slate-200 rounded w-24"></div>
                                        <div class="h-4 bg-slate-200 rounded-full w-16"></div>
                                    </div>
                                    <div class="h-6 bg-slate-200 rounded w-6"></div>
                                </div>
                                <div class="h-4 bg-slate-200 rounded w-3/4"></div>
                                <div class="flex flex-col gap-2 mt-1">
                                    <div class="h-3 bg-slate-200 rounded w-full"></div>
                                    <div class="flex justify-between items-center">
                                        <div class="h-3 bg-slate-200 rounded w-1/2"></div>
                                        <div class="h-3 bg-slate-200 rounded w-16"></div>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    {:else}
                        <div class="p-12 text-center text-slate-500 bg-white">
                            Belum ada pengajuan yang masuk.
                        </div>
                    {/if}
                {/if}
            </div>

            <!-- Removed Pagination Footer -->
        </div>
    {:else}
        <div class="flex flex-col items-center justify-center p-12 text-center border-2 border-dashed border-slate-200 rounded-xl bg-slate-50">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            <h3 class="text-lg font-medium text-slate-900">Akses Dibatasi</h3>
            <p class="text-slate-500 max-w-sm mt-1">Halaman ini khusus untuk Admin.</p>
        </div>
    {/if}

    <CostModal
        bind:open={isModalOpen}
        record={selectedRecord}
        allRecordsInSpd={allRecordsInSelectedSpd}
        bind:editingCosts={editingCosts}
        on:close={() => isModalOpen = false}
        on:save={handleModalSave}
        on:reject={handleModalReject}
    />
    <ConfirmationModal
        bind:open={isConfirmOpen}
        title="Simpan Rincian Biaya"
        description="Apakah Anda yakin data rincian biaya ini sudah sesuai? Status akan diubah menjadi Approved."
        confirmText="Ya, Simpan & Approve"
        onConfirm={processSave}
    />
    <ConfirmationModal
        bind:open={isRejectConfirmOpen}
        title="Tolak Pengajuan"
        description="Apakah Anda yakin ingin menolak pengajuan ini? Status akan dikembalikan dan ditandai sebagai Rejected."
        confirmText="Ya, Tolak Pengajuan"
        confirmButtonClass="bg-red-600 hover:bg-red-700 focus:ring-red-500"
        onConfirm={processReject}
    />
    <ConfirmationModal
        bind:open={isPaidConfirmOpen}
        title="Cairkan Dana"
        description="Apakah Anda yakin dana untuk pegawai ini sudah dicairkan? Status akan berubah menjadi Completed."
        confirmText="Ya, Tandai Lunas"
        confirmButtonClass="bg-emerald-600 hover:bg-emerald-700 focus:ring-emerald-500"
        onConfirm={processPaid}
    />
</div>
