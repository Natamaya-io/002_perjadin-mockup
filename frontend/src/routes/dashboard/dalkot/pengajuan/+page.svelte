<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import {
        ArrowLeft,
        CalendarDays,
        Check,
        ChevronDown,
        FileText,
        Info,
        MapPin,
        Save,
        Search,
        Upload,
        Users,
        X
    } from 'lucide-svelte';
    import { api } from '$lib/shared/api';
    import type {
        DalkotCategory,
        DalkotMasterData,
        DalkotRecord,
        DalkotUserOption
    } from '$lib/shared/api/types';
    import { toast } from '$lib/shared/stores/toast';
    import { userStore } from '$lib/features/auth/store';

    type FormState = {
        serviceDate: string;
        category: DalkotCategory | '';
        accompaniedOfficial: string;
        assignmentLetterNumber: string;
        assignmentLetterDate: string;
        activityName: string;
        location: string;
        reportContent: string;
        status: 'Draft' | 'Pending' | 'Sukses';
        document: File | null;
    };

    const emptyMaster = (): DalkotMasterData => ({
        users: [],
        categories: [
            { value: 'Jam Kerja', label: 'Jam Kerja', amount: 100000 },
            { value: 'Overtime', label: 'Overtime', amount: 150000 },
            { value: 'Hari Libur', label: 'Hari Libur', amount: 250000 }
        ],
        officials: [],
        locations: [],
        spjAmount: 170000,
        jamKerjaAmount: 100000,
        overtimeAmount: 150000,
        hariLiburAmount: 250000
    });

    const emptyForm = (): FormState => ({
        serviceDate: '',
        category: 'Jam Kerja',
        accompaniedOfficial: '',
        assignmentLetterNumber: '',
        assignmentLetterDate: '',
        activityName: '',
        location: '',
        reportContent: '',
        status: 'Draft',
        document: null
    });

    let master = emptyMaster();
    let form = emptyForm();
    let editingId = '';
    let existingDetail: DalkotRecord | null = null;

    let selectedSPJUserIds: string[] = [];
    let selectedRiilUserIds: string[] = [];
    let showSPJDropdown = false;
    let showRiilDropdown = false;
    let searchSPJ = '';
    let searchRiil = '';

    let loading = true;
    let saving = false;
    let loadError = '';

    $: protocolUsers = master.users.filter(
        (user) =>
            user.role.toLowerCase() === 'protokol' &&
            user.isActive !== false
    );

    $: filteredSPJUsers = filterUsers(protocolUsers, searchSPJ);
    $: filteredRiilUsers = filterUsers(protocolUsers, searchRiil);

    $: selectedSPJUsers = protocolUsers.filter((user) =>
        selectedSPJUserIds.includes(user.id)
    );
    $: selectedRiilUsers = protocolUsers.filter((user) =>
        selectedRiilUserIds.includes(user.id)
    );

    $: hasSPJ = selectedSPJUserIds.length > 0;
    $: riilAmountPerOfficer = getRiilAmount(form.category);
    $: totalSPJAmount = selectedSPJUserIds.length * master.spjAmount;
    $: totalRiilAmount = selectedRiilUserIds.length * riilAmountPerOfficer;

    onMount(loadPage);

    async function loadPage() {
        loading = true;
        loadError = '';

        try {
            master = await api.getDalkotMaster();

            const id = new URLSearchParams(window.location.search).get('id');
            if (id) {
                editingId = id;
                existingDetail = await api.getDalkotById(id);
                fillEditForm(existingDetail);
            }
        } catch (error) {
            loadError =
                error instanceof Error
                    ? error.message
                    : 'Form Pengajuan Dalkot gagal dimuat.';
            toast.error(loadError);
        } finally {
            loading = false;
        }
    }

    function fillEditForm(detail: DalkotRecord) {
        form = {
            serviceDate: detail.serviceDate?.slice(0, 10) ?? '',
            category: detail.category ?? 'Jam Kerja',
            accompaniedOfficial: detail.accompaniedOfficial ?? '',
            assignmentLetterNumber: detail.assignmentLetterNumber ?? '',
            assignmentLetterDate: detail.assignmentLetterDate?.slice(0, 10) ?? '',
            activityName: detail.activityName ?? '',
            location: detail.location ?? '',
            reportContent: detail.reportContent ?? '',
            status: detail.status ?? 'Draft',
            document: null
        };

        selectedSPJUserIds = detail.officers
            .filter((officer) => officer.officerGroup === 'SPJ')
            .map((officer) => officer.employeeId ?? '')
            .filter(Boolean);

        selectedRiilUserIds = detail.officers
            .filter((officer) =>
                ['RIIL', 'RILL'].includes(String(officer.officerGroup).toUpperCase())
            )
            .map((officer) => officer.employeeId ?? '')
            .filter(Boolean);
    }

    function filterUsers(users: DalkotUserOption[], keywordValue: string) {
        const keyword = keywordValue.trim().toLowerCase();
        if (!keyword) return users;

        return users.filter((user) =>
            [user.name, user.email, user.nip, user.position, user.grade].some(
                (value) => String(value ?? '').toLowerCase().includes(keyword)
            )
        );
    }

    function toggleSPJUser(userId: string) {
        selectedSPJUserIds = selectedSPJUserIds.includes(userId)
            ? selectedSPJUserIds.filter((id) => id !== userId)
            : [...selectedSPJUserIds, userId];
    }

    function toggleRiilUser(userId: string) {
        // Daftar RIIL berdiri sendiri. User yang sama tetap dapat dipilih di SPJ.
        selectedRiilUserIds = selectedRiilUserIds.includes(userId)
            ? selectedRiilUserIds.filter((id) => id !== userId)
            : [...selectedRiilUserIds, userId];
    }

    function selectCategory(value: string) {
        if (['Jam Kerja', 'Overtime', 'Hari Libur'].includes(value)) {
            form.category = value as DalkotCategory;
        }
    }

    function getRiilAmount(category: DalkotCategory | ''): number {
        switch (category) {
            case 'Jam Kerja':
                return master.jamKerjaAmount;
            case 'Overtime':
                return master.overtimeAmount;
            case 'Hari Libur':
                return master.hariLiburAmount;
            default:
                return 0;
        }
    }

    function handleDocumentChange(event: Event) {
        const input = event.currentTarget as HTMLInputElement;
        const file = input.files?.[0] ?? null;

        if (!file) {
            form.document = null;
            return;
        }

        const allowedTypes = ['application/pdf', 'image/jpeg', 'image/png'];
        if (!allowedTypes.includes(file.type)) {
            toast.error('Dokumentasi hanya boleh PDF, JPG, JPEG, atau PNG.');
            input.value = '';
            form.document = null;
            return;
        }

        if (file.size > 5 * 1024 * 1024) {
            toast.error('Ukuran dokumentasi maksimal 5 MB.');
            input.value = '';
            form.document = null;
            return;
        }

        form.document = file;
    }

    function validateForm(): boolean {
        if (!form.serviceDate) {
            toast.error('Tanggal Pelaksanaan Dinas wajib diisi.');
            return false;
        }
        if (!form.category) {
            toast.error('Kategori Dalkot wajib dipilih.');
            return false;
        }
        if (!form.accompaniedOfficial) {
            toast.error('Pejabat yang Didampingi wajib dipilih.');
            return false;
        }
        if (
            selectedSPJUserIds.length === 0 &&
            selectedRiilUserIds.length === 0
        ) {
            toast.error('Pilih minimal satu Petugas SPJ atau Petugas Rill.');
            return false;
        }

        if ($userStore.role === 'protokol') {
    const currentUserId = String(
        $userStore.id ?? ''
    );

    const currentUserSelected =
        selectedSPJUserIds.includes(
            currentUserId
        ) ||
        selectedRiilUserIds.includes(
            currentUserId
        );

    if (!currentUserSelected) {
        toast.error(
            'Nama Anda wajib dipilih sebagai Petugas SPJ atau Petugas Riil.'
        );

        return false;
    }
}
        if (!form.activityName.trim()) {
            toast.error('Nama Kegiatan wajib diisi.');
            return false;
        }
        if (!form.location) {
            toast.error('Lokasi wajib dipilih.');
            return false;
        }

        // Laporan dan dokumentasi hanya wajib ketika Petugas SPJ dipilih.
        if (hasSPJ && !form.reportContent.trim()) {
            toast.error('Isi Laporan wajib diisi apabila terdapat Petugas SPJ.');
            return false;
        }
        if (
            hasSPJ &&
            !form.document &&
            !existingDetail?.documentFilename &&
            !existingDetail?.documentUrl
        ) {
            toast.error('Dokumentasi wajib diunggah apabila terdapat Petugas SPJ.');
            return false;
        }

        return true;
    }

    function buildFormData(): FormData {
        const data = new FormData();
        data.append('serviceDate', form.serviceDate);
        data.append('category', form.category);
        data.append('accompaniedOfficial', form.accompaniedOfficial);
        data.append('assignmentLetterNumber', form.assignmentLetterNumber.trim());
        data.append('assignmentLetterDate', form.assignmentLetterDate);
        data.append('activityName', form.activityName.trim());
        data.append('location', form.location);
        data.append('reportContent', form.reportContent.trim());
        data.append('status', form.status);
        data.append('spjOfficerIds', JSON.stringify(selectedSPJUserIds));
        data.append('riilOfficerIds', JSON.stringify(selectedRiilUserIds));

        if (form.document) {
            data.append('document', form.document, form.document.name);
        }
        return data;
    }

    async function saveDalkot() {
        if (!validateForm()) return;

        saving = true;
        try {
            if (editingId) {
                await api.updateDalkot(editingId, buildFormData());
                toast.success('Pengajuan Dalkot berhasil diperbarui.');
            } else {
                await api.createDalkot(buildFormData());
                toast.success('Pengajuan Dalkot berhasil disimpan.');
            }
            await goto('/dashboard/dalkot');
        } catch (error) {
            toast.error(
                error instanceof Error
                    ? error.message
                    : 'Pengajuan Dalkot gagal disimpan.'
            );
        } finally {
            saving = false;
        }
    }

    function selectedNames(users: DalkotUserOption[]): string {
        if (users.length === 0) return 'Belum ada petugas dipilih';
        if (users.length <= 2) return users.map((user) => user.name).join(', ');
        return `${users.slice(0, 2).map((user) => user.name).join(', ')} dan ${users.length - 2} lainnya`;
    }

    function formatRupiah(value: number): string {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            maximumFractionDigits: 0
        }).format(Number(value || 0));
    }

    function closeDropdowns() {
        showSPJDropdown = false;
        showRiilDropdown = false;
    }
</script>

<svelte:window on:click={closeDropdowns} />

<svelte:head>
    <title>{editingId ? 'Edit Pengajuan Dalkot' : 'Pengajuan Dalkot'} | Perjadin</title>
</svelte:head>

<main class="space-y-6 pb-20">
    <section class="rounded-3xl bg-gradient-to-br from-slate-950 via-slate-800 to-sky-900 p-6 text-white shadow-xl sm:p-8">
        <div class="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
            <div class="max-w-3xl">
                <p class="text-xs font-bold uppercase tracking-[0.2em] text-sky-200">Perjalanan Dinas Dalam Kota</p>
                <h1 class="mt-3 text-3xl font-black sm:text-4xl">
                    {editingId ? 'Edit Pengajuan Dalkot' : 'Form Pengajuan Dalkot'}
                </h1>
                <p class="mt-3 text-sm leading-6 text-slate-200">
                    Pilih Petugas SPJ dan Petugas Rill secara terpisah. User Protokol yang sama boleh dipilih pada kedua dropdown.
                </p>
            </div>

            <button
                type="button"
                class="inline-flex items-center justify-center gap-2 rounded-2xl border border-white/20 bg-white/10 px-5 py-3 text-sm font-bold hover:bg-white/20"
                on:click={() => goto('/dashboard/dalkot')}
            >
                <ArrowLeft size={18} /> Kembali ke Daftar
            </button>
        </div>
    </section>

    {#if loadError}
        <section class="rounded-2xl border border-red-200 bg-red-50 p-5 text-sm text-red-700">{loadError}</section>
    {/if}

    {#if loading}
        <section class="flex min-h-96 items-center justify-center rounded-3xl border border-slate-200 bg-white text-sm text-slate-500">
            Memuat Form Pengajuan Dalkot...
        </section>
    {:else}
        <form class="space-y-6" on:submit|preventDefault={saveDalkot}>
            <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm sm:p-7">
                <div class="flex items-start gap-3">
                    <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-blue-50 text-blue-700">
                        <CalendarDays size={21} />
                    </div>
                    <div>
                        <h2 class="text-xl font-black text-slate-900">Data Pelaksanaan</h2>
                        <p class="mt-1 text-sm text-slate-500">Lengkapi informasi kegiatan dan Surat Tugas.</p>
                    </div>
                </div>

                <div class="mt-6 grid grid-cols-1 gap-5 md:grid-cols-2">
                    <div>
                        <label for="serviceDate" class="mb-2 block text-sm font-bold text-slate-700">Tanggal Pelaksanaan Dinas</label>
                        <input id="serviceDate" bind:value={form.serviceDate} type="date" required class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm" disabled={saving} />
                    </div>

                    <div>
                        <label for="official" class="mb-2 block text-sm font-bold text-slate-700">Pejabat yang Didampingi</label>
                        <select id="official" bind:value={form.accompaniedOfficial} required class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm" disabled={saving}>
                            <option value="">Pilih Pejabat</option>
                            {#each master.officials as item}
                                <option value={item.value}>{item.label}</option>
                            {/each}
                        </select>
                    </div>

                    <div>
                        <label for="assignmentLetterNumber" class="mb-2 block text-sm font-bold text-slate-700">Nomor Surat Tugas</label>
                        <input id="assignmentLetterNumber" bind:value={form.assignmentLetterNumber} type="text" placeholder="Contoh: ST/123/VIII/2026" class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm" disabled={saving} />
                        <p class="mt-2 text-xs text-slate-500">Opsional. Digunakan pada rekap bulanan SPJ.</p>
                    </div>

                    <div>
                        <label for="assignmentLetterDate" class="mb-2 block text-sm font-bold text-slate-700">Tanggal Surat Tugas</label>
                        <input id="assignmentLetterDate" bind:value={form.assignmentLetterDate} type="date" class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm" disabled={saving} />
                        <p class="mt-2 text-xs text-slate-500">Boleh dikosongkan apabila Surat Tugas belum diterbitkan.</p>
                    </div>
                </div>
            </section>

            <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm sm:p-7">
                <div>
                    <p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-400">Kategori Dalkot</p>
                    <h2 class="mt-2 text-xl font-black text-slate-900">Pilih kategori biaya Rill</h2>
                    <p class="mt-1 text-sm text-slate-500">Kategori ditampilkan sebagai card seperti form awal.</p>
                </div>

                <div class="mt-6 grid grid-cols-1 gap-4 md:grid-cols-3">
                    {#each master.categories as category}
                        <button
                            type="button"
                            aria-pressed={form.category === category.value}
                            class={`relative rounded-2xl border-2 p-5 text-left transition ${
                                form.category === category.value
                                    ? 'border-blue-600 bg-blue-50 shadow-sm'
                                    : 'border-slate-200 bg-white hover:border-blue-300 hover:bg-slate-50'
                            }`}
                            on:click={() => selectCategory(category.value)}
                            disabled={saving}
                        >
                            {#if form.category === category.value}
                                <span class="absolute right-4 top-4 flex h-7 w-7 items-center justify-center rounded-full bg-blue-600 text-white">
                                    <Check size={16} />
                                </span>
                            {/if}
                            <p class="font-black text-slate-900">{category.label}</p>
                            <p class="mt-2 text-sm font-bold text-blue-700">{formatRupiah(category.amount ?? 0)} / Petugas Rill</p>
                        </button>
                    {/each}
                </div>
            </section>

            <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm sm:p-7">
                <div class="flex items-start gap-3">
                    <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-emerald-50 text-emerald-700">
                        <Users size={21} />
                    </div>
                    <div>
                        <h2 class="text-xl font-black text-slate-900">Petugas User Protokol</h2>
                        <p class="mt-1 text-sm text-slate-500">User yang sama dapat dicentang pada Petugas SPJ dan Petugas Rill sekaligus.</p>
                    </div>
                </div>

                {#if protocolUsers.length === 0}
                    <div class="mt-6 rounded-2xl border border-amber-200 bg-amber-50 p-5 text-sm text-amber-800">
                        Tidak ada User aktif dengan role Protokol. Tambahkan atau ubah role User melalui menu User.
                    </div>
                {:else}
                    <div class="mt-6 grid grid-cols-1 gap-5 lg:grid-cols-2">
                        <div class="relative" on:click|stopPropagation>
                            <label class="mb-2 block text-sm font-bold text-slate-700">Dropdown Petugas SPJ</label>
                            <button
                                type="button"
                                class="flex w-full items-center justify-between rounded-xl border border-slate-300 bg-white px-4 py-3 text-left"
                                on:click={() => {
                                    showSPJDropdown = !showSPJDropdown;
                                    showRiilDropdown = false;
                                }}
                                disabled={saving}
                            >
                                <span class="min-w-0">
                                    <span class="block font-bold text-slate-800">{selectedSPJUserIds.length} petugas dipilih</span>
                                    <span class="mt-1 block truncate text-xs text-slate-500">{selectedNames(selectedSPJUsers)}</span>
                                </span>
                                <ChevronDown size={18} />
                            </button>

                            {#if showSPJDropdown}
                                <div class="absolute z-40 mt-2 w-full rounded-2xl border border-slate-200 bg-white p-3 shadow-2xl">
                                    <div class="relative">
                                        <Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                                        <input bind:value={searchSPJ} type="search" placeholder="Cari Petugas SPJ..." class="w-full rounded-xl border border-slate-300 py-2.5 pl-10 pr-3 text-sm" />
                                    </div>
                                    <div class="mt-3 max-h-72 space-y-1 overflow-y-auto">
                                        {#each filteredSPJUsers as user (user.id)}
                                            <label class="flex cursor-pointer items-start gap-3 rounded-xl px-3 py-2 hover:bg-slate-50">
                                                <input type="checkbox" checked={selectedSPJUserIds.includes(user.id)} class="mt-1 h-4 w-4" on:change={() => toggleSPJUser(user.id)} />
                                                <span class="min-w-0">
                                                    <span class="block font-bold text-slate-800">{user.name}</span>
                                                    <span class="block text-xs text-slate-500">{user.position || user.nip || user.email}</span>
                                                </span>
                                            </label>
                                        {/each}
                                    </div>
                                </div>
                            {/if}
                        </div>

                        <div class="relative" on:click|stopPropagation>
                            <label class="mb-2 block text-sm font-bold text-slate-700">Dropdown Petugas Rill</label>
                            <button
                                type="button"
                                class="flex w-full items-center justify-between rounded-xl border border-slate-300 bg-white px-4 py-3 text-left"
                                on:click={() => {
                                    showRiilDropdown = !showRiilDropdown;
                                    showSPJDropdown = false;
                                }}
                                disabled={saving}
                            >
                                <span class="min-w-0">
                                    <span class="block font-bold text-slate-800">{selectedRiilUserIds.length} petugas dipilih</span>
                                    <span class="mt-1 block truncate text-xs text-slate-500">{selectedNames(selectedRiilUsers)}</span>
                                </span>
                                <ChevronDown size={18} />
                            </button>

                            {#if showRiilDropdown}
                                <div class="absolute z-40 mt-2 w-full rounded-2xl border border-slate-200 bg-white p-3 shadow-2xl">
                                    <div class="relative">
                                        <Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                                        <input bind:value={searchRiil} type="search" placeholder="Cari Petugas Rill..." class="w-full rounded-xl border border-slate-300 py-2.5 pl-10 pr-3 text-sm" />
                                    </div>
                                    <div class="mt-3 max-h-72 space-y-1 overflow-y-auto">
                                        {#each filteredRiilUsers as user (user.id)}
                                            <label class="flex cursor-pointer items-start gap-3 rounded-xl px-3 py-2 hover:bg-slate-50">
                                                <input type="checkbox" checked={selectedRiilUserIds.includes(user.id)} class="mt-1 h-4 w-4" on:change={() => toggleRiilUser(user.id)} />
                                                <span class="min-w-0">
                                                    <span class="block font-bold text-slate-800">{user.name}</span>
                                                    <span class="block text-xs text-slate-500">{user.position || user.nip || user.email}</span>
                                                </span>
                                            </label>
                                        {/each}
                                    </div>
                                </div>
                            {/if}
                        </div>
                    </div>
                {/if}
            </section>

            <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm sm:p-7">
                <div class="flex items-start gap-3">
                    <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-50 text-amber-700">
                        <FileText size={21} />
                    </div>
                    <div>
                        <h2 class="text-xl font-black text-slate-900">Rincian Kegiatan</h2>
                        <p class="mt-1 text-sm text-slate-500">Laporan dan dokumentasi hanya wajib apabila Petugas SPJ dipilih.</p>
                    </div>
                </div>

                <div class="mt-6 space-y-5">
                    <div>
                        <label for="activityName" class="mb-2 block text-sm font-bold text-slate-700">Nama Kegiatan</label>
                        <input id="activityName" bind:value={form.activityName} type="text" required class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm" disabled={saving} />
                    </div>

                    <div>
                        <label for="location" class="mb-2 block text-sm font-bold text-slate-700">Lokasi</label>
                        <div class="relative">
                            <MapPin size={17} class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400" />
                            <select id="location" bind:value={form.location} required class="w-full rounded-xl border border-slate-300 bg-white py-3 pl-11 pr-4 text-sm" disabled={saving}>
                                <option value="">Pilih Lokasi</option>
                                {#each master.locations as item}
                                    <option value={item.value}>{item.label}</option>
                                {/each}
                            </select>
                        </div>
                    </div>

                    <div>
                        <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
                            <label for="reportContent" class="text-sm font-bold text-slate-700">
                                Isi Laporan {hasSPJ ? '(wajib)' : '(opsional)'}
                            </label>
                            {#if !hasSPJ}
                                <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-bold text-slate-600">Tidak wajib untuk Rill saja</span>
                            {/if}
                        </div>
                        <textarea id="reportContent" bind:value={form.reportContent} rows="6" class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm" disabled={saving}></textarea>
                    </div>

                    <div>
                        <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
                            <label class="text-sm font-bold text-slate-700">Dokumentasi {hasSPJ ? '(wajib)' : '(opsional)'}</label>
                            {#if !hasSPJ}
                                <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-bold text-slate-600">Tidak wajib untuk Rill saja</span>
                            {/if}
                        </div>

                        <label class="flex cursor-pointer flex-col items-center justify-center gap-3 rounded-2xl border-2 border-dashed border-slate-300 bg-slate-50 px-5 py-8 text-center hover:border-blue-400">
                            <Upload size={25} />
                            <span class="text-sm font-bold text-slate-700">
                                {#if form.document}
                                    {form.document.name}
                                {:else if existingDetail?.documentOriginalName}
                                    Dokumen tersimpan: {existingDetail.documentOriginalName}
                                {:else}
                                    Pilih JPG, PNG, atau PDF
                                {/if}
                            </span>
                            <span class="text-xs text-slate-500">Maksimal 5 MB</span>
                            <input type="file" accept=".pdf,.jpg,.jpeg,.png,application/pdf,image/jpeg,image/png" class="hidden" on:change={handleDocumentChange} disabled={saving} />
                        </label>

                        {#if form.document}
                            <button type="button" class="mt-3 inline-flex items-center gap-2 text-sm font-bold text-red-600" on:click={() => (form.document = null)}>
                                <X size={16} /> Hapus file pilihan
                            </button>
                        {/if}
                    </div>
                </div>
            </section>

            <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm sm:p-7">
                <h2 class="text-xl font-black text-slate-900">Ringkasan Biaya Terpisah</h2>
                <p class="mt-1 text-sm text-slate-500">Biaya SPJ dan Biaya Rill tidak dijumlahkan menjadi satu total.</p>

                <div class="mt-5 grid grid-cols-1 gap-4 md:grid-cols-2">
                    <div class="rounded-2xl border border-blue-100 bg-blue-50 p-5">
                        <p class="text-xs font-bold uppercase tracking-wide text-blue-700">Total Biaya SPJ</p>
                        <p class="mt-3 text-sm font-bold text-blue-950">{selectedSPJUserIds.length} orang × {formatRupiah(master.spjAmount)}</p>
                        <p class="mt-2 text-2xl font-black text-blue-950">{formatRupiah(totalSPJAmount)}</p>
                    </div>

                    <div class="rounded-2xl border border-emerald-100 bg-emerald-50 p-5">
                        <p class="text-xs font-bold uppercase tracking-wide text-emerald-700">Total Biaya Rill</p>
                        <p class="mt-3 text-sm font-bold text-emerald-950">{selectedRiilUserIds.length} orang × {formatRupiah(riilAmountPerOfficer)}</p>
                        <p class="mt-2 text-2xl font-black text-emerald-950">{formatRupiah(totalRiilAmount)}</p>
                    </div>
                </div>

                <div class="mt-5 flex gap-3 rounded-2xl border border-blue-100 bg-blue-50 p-4 text-sm text-blue-800">
                    <Info size={18} class="shrink-0" />
                    <p>User yang dipilih pada kedua dropdown dihitung satu kali sebagai SPJ dan satu kali sebagai Rill, tetapi kedua total tetap ditampilkan terpisah.</p>
                </div>
            </section>

            <div class="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
                <button type="button" class="rounded-2xl border border-slate-300 bg-white px-6 py-3 text-sm font-bold" on:click={() => goto('/dashboard/dalkot')} disabled={saving}>Batal</button>
                <button type="submit" class="inline-flex items-center justify-center gap-2 rounded-2xl bg-slate-900 px-6 py-3 text-sm font-bold text-white disabled:opacity-50" disabled={saving}>
                    <Save size={18} />
                    {saving ? 'Menyimpan...' : editingId ? 'Simpan Perubahan' : 'Simpan Data'}
                </button>
            </div>
        </form>
    {/if}
</main>