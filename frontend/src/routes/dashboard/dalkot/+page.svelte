<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import {
        ChevronDown,
        FileText,
        Pencil,
        Plus,
        Printer,
        RefreshCcw,
        Search,
        Users,
        X
    } from 'lucide-svelte';
    import { api } from '$lib/shared/api';
    import type {
        DalkotCategory,
        DalkotListResponse,
        DalkotOfficer,
        DalkotRecord,
        DalkotStatus,
        DalkotSummary
    } from '$lib/shared/api/types';
    import { toast } from '$lib/shared/stores/toast';

    const emptySummary = (): DalkotSummary => ({
        totalRecords: 0,
        totalOfficers: 0,
        totalSPJAmount: 0,
        totalRiilAmount: 0
    });

    const monthOptions = [
        ['01', 'Januari'],
        ['02', 'Februari'],
        ['03', 'Maret'],
        ['04', 'April'],
        ['05', 'Mei'],
        ['06', 'Juni'],
        ['07', 'Juli'],
        ['08', 'Agustus'],
        ['09', 'September'],
        ['10', 'Oktober'],
        ['11', 'November'],
        ['12', 'Desember']
    ];

    const categoryOptions: DalkotCategory[] = [
        'Jam Kerja',
        'Overtime',
        'Hari Libur'
    ];

    let rows: DalkotRecord[] = [];
    let summary = emptySummary();
    let loading = true;
    let loadError = '';

    let searchQuery = '';
    let categoryFilter = '';
    let monthFilter = '';
    let yearFilter = '';

    let expandedIds = new Set<string>();
    let selectedDetail: DalkotRecord | null = null;

    $: availableYears = Array.from(
        new Set(rows.map((row) => row.serviceDate?.slice(0, 4)).filter(Boolean))
    ).sort((first, second) => Number(second) - Number(first));

    $: filteredRows = [...rows]
        .sort((first, second) => getTimestamp(second) - getTimestamp(first))
        .filter((row) => {
            const keyword = searchQuery.trim().toLowerCase();
            const matchesSearch =
                !keyword ||
                [
                    row.businessId,
                    row.activityName,
                    row.location,
                    row.accompaniedOfficial,
                    row.category,
                    row.assignmentLetterNumber
                ].some((value) => String(value ?? '').toLowerCase().includes(keyword)) ||
                row.officers.some((officer) =>
                    officer.name.toLowerCase().includes(keyword)
                );

            const matchesCategory = !categoryFilter || row.category === categoryFilter;
            const matchesMonth = !monthFilter || row.serviceDate?.slice(5, 7) === monthFilter;
            const matchesYear = !yearFilter || row.serviceDate?.slice(0, 4) === yearFilter;

            return matchesSearch && matchesCategory && matchesMonth && matchesYear;
        });

    onMount(loadDalkot);

    async function loadDalkot() {
        loading = true;
        loadError = '';
        try {
            const result: DalkotListResponse = await api.getDalkotList();
            rows = result.rows ?? [];
            summary = result.summary ?? emptySummary();
        } catch (error) {
            loadError =
                error instanceof Error
                    ? error.message
                    : 'Daftar Dalkot gagal dimuat.';
            toast.error(loadError);
        } finally {
            loading = false;
        }
    }

    function getTimestamp(row: DalkotRecord): number {
        const value = row.createdAt || row.serviceDate || '';
        const normalized = value.length === 10 ? `${value}T00:00:00` : value;
        const timestamp = new Date(normalized).getTime();
        return Number.isNaN(timestamp) ? 0 : timestamp;
    }

    function resetFilters() {
        searchQuery = '';
        categoryFilter = '';
        monthFilter = '';
        yearFilter = '';
    }

    function toggleDetail(id: string) {
        const next = new Set(expandedIds);
        if (next.has(id)) next.delete(id);
        else next.add(id);
        expandedIds = next;
    }

    function openDetailModal(row: DalkotRecord) {
        selectedDetail = row;
    }

    function closeDetailModal() {
        selectedDetail = null;
    }

    function handleWindowKeydown(event: KeyboardEvent) {
        if (event.key === 'Escape' && selectedDetail) {
            closeDetailModal();
        }
    }

    function getSPJOfficers(row: DalkotRecord): DalkotOfficer[] {
        return row.officers.filter((officer) => officer.officerGroup === 'SPJ');
    }

    function getRiilOfficers(row: DalkotRecord): DalkotOfficer[] {
        return row.officers.filter((officer) =>
            ['RIIL', 'RILL'].includes(String(officer.officerGroup).toUpperCase())
        );
    }

    function formatRupiah(value: number): string {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            maximumFractionDigits: 0
        }).format(Number(value || 0));
    }

    function formatDate(value?: string | null): string {
        if (!value) return '-';
        const date = new Date(`${value.slice(0, 10)}T00:00:00`);
        if (Number.isNaN(date.getTime())) return value;
        return new Intl.DateTimeFormat('id-ID', {
            day: '2-digit',
            month: 'long',
            year: 'numeric'
        }).format(date);
    }

    function statusClass(status: DalkotStatus): string {
        switch (status) {
            case 'Sukses':
                return 'border-emerald-200 bg-emerald-50 text-emerald-700';
            case 'Pending':
                return 'border-amber-200 bg-amber-50 text-amber-700';
            default:
                return 'border-slate-200 bg-slate-100 text-slate-700';
        }
    }

    function escapeHtml(value: unknown): string {
        return String(value ?? '')
            .replaceAll('&', '&amp;')
            .replaceAll('<', '&lt;')
            .replaceAll('>', '&gt;')
            .replaceAll('"', '&quot;')
            .replaceAll("'", '&#039;');
    }

    function openPrintWindow(title: string, body: string) {
        const printWindow = window.open('', '_blank', 'width=980,height=760');
        if (!printWindow) {
            toast.error('Pop-up cetak diblokir oleh browser.');
            return;
        }

        printWindow.document.open();
        printWindow.document.write(`
            <!doctype html>
            <html lang="id">
            <head>
                <meta charset="utf-8">
                <title>${escapeHtml(title)}</title>
                <style>
                    body { font-family: Arial, sans-serif; color: #0f172a; margin: 36px; }
                    h1 { font-size: 22px; margin-bottom: 6px; }
                    h2 { font-size: 16px; margin-top: 26px; }
                    p { line-height: 1.6; }
                    table { width: 100%; border-collapse: collapse; margin-top: 18px; }
                    th, td { border: 1px solid #cbd5e1; padding: 9px; text-align: left; }
                    th { background: #f1f5f9; }
                    .actions { margin-bottom: 24px; }
                    .dpr-sheet { break-after: page; page-break-after: always; }
                    .dpr-sheet:last-child { break-after: auto; page-break-after: auto; }
                    @media print { .actions { display: none; } body { margin: 15mm; } }
                </style>
            </head>
            <body>
                <div class="actions"><button onclick="window.print()">Cetak</button></div>
                ${body}
            </body>
            </html>
        `);
        printWindow.document.close();
        printWindow.focus();
    }

    function printTravelReport(row: DalkotRecord) {
        openPrintWindow(
            `Laporan ${row.businessId}`,
            `
                <h1>Laporan Perjalanan Dinas Dalam Kota</h1>
                <p><strong>ID Dalkot:</strong> ${escapeHtml(row.businessId)}</p>
                <table>
                    <tr><th>Tanggal Pelaksanaan</th><td>${escapeHtml(formatDate(row.serviceDate))}</td></tr>
                    <tr><th>Nama Kegiatan</th><td>${escapeHtml(row.activityName)}</td></tr>
                    <tr><th>Pejabat yang Didampingi</th><td>${escapeHtml(row.accompaniedOfficial)}</td></tr>
                    <tr><th>Lokasi</th><td>${escapeHtml(row.location)}</td></tr>
                    <tr><th>Kategori Dalkot</th><td>${escapeHtml(row.category)}</td></tr>
                    <tr><th>Nomor Surat Tugas</th><td>${escapeHtml(row.assignmentLetterNumber || '-')}</td></tr>
                    <tr><th>Tanggal Surat Tugas</th><td>${escapeHtml(formatDate(row.assignmentLetterDate))}</td></tr>
                </table>
                <h2>Isi Laporan</h2>
                <p>${escapeHtml(row.reportContent || '-').replaceAll('\n', '<br>')}</p>
            `
        );
    }

    function buildDprBody(row: DalkotRecord, officer: DalkotOfficer): string {
        const costLabel = officer.officerGroup === 'SPJ' ? 'Biaya SPJ' : 'Biaya Riil';
        return `
            <section class="dpr-sheet">
                <h1>Daftar Pengeluaran Riil</h1>
                <p><strong>ID Dalkot:</strong> ${escapeHtml(row.businessId)}</p>
                <table>
                    <tr><th>Nama Petugas</th><td>${escapeHtml(officer.name)}</td></tr>
                    <tr><th>NIP</th><td>${escapeHtml(officer.nip || '-')}</td></tr>
                    <tr><th>Penugasan</th><td>${escapeHtml(officer.officerGroup)}</td></tr>
                    <tr><th>Kategori Dalkot</th><td>${escapeHtml(row.category)}</td></tr>
                    <tr><th>${escapeHtml(costLabel)}</th><td>${escapeHtml(formatRupiah(officer.amount))}</td></tr>
                    <tr><th>Tanggal Pelaksanaan</th><td>${escapeHtml(formatDate(row.serviceDate))}</td></tr>
                    <tr><th>Kegiatan</th><td>${escapeHtml(row.activityName)}</td></tr>
                </table>
            </section>
        `;
    }

    function printAllSpjDpr(row: DalkotRecord) {
        const officers = getSPJOfficers(row);
        if (officers.length === 0) {
            toast.error('DPR tidak tersedia karena pengajuan ini tidak memiliki Petugas SPJ.');
            return;
        }

        openPrintWindow(
            `DPR SPJ ${row.businessId}`,
            officers.map((officer) => buildDprBody(row, officer)).join('')
        );
    }
</script>

<svelte:head>
    <title>Dalkot | Perjadin</title>
</svelte:head>

<svelte:window on:keydown={handleWindowKeydown} />

<main class="space-y-6 pb-20">
    <section class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-slate-950 via-slate-800 to-blue-900 p-6 text-white shadow-xl sm:p-8">
        <div class="relative z-10 flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
            <div class="max-w-3xl">
                <p class="text-xs font-bold uppercase tracking-[0.2em] text-blue-200">Modul Perjalanan Dinas</p>
                <h1 class="mt-3 text-4xl font-black">Dalam Kota <span class="text-blue-300">Dalkot</span></h1>
                <p class="mt-3 max-w-2xl text-sm leading-6 text-slate-200">
                    Kelola pengajuan dan rincian Petugas SPJ serta Petugas Riil dalam satu tabel yang rapi.
                </p>
            </div>

            <button type="button" class="inline-flex items-center justify-center gap-2 rounded-2xl bg-white px-5 py-3 text-sm font-black text-slate-900 shadow-lg" on:click={() => goto('/dashboard/dalkot/pengajuan')}>
                <Plus size={18} /> Tambah Pengajuan
            </button>
        </div>
    </section>

    <section class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
            <p class="text-sm font-bold text-slate-500">Total Pengajuan</p>
            <p class="mt-3 text-3xl font-black text-slate-900">{summary.totalRecords}</p>
            <p class="mt-1 text-xs text-slate-500">Data Dalkot tercatat</p>
        </article>
        <article class="rounded-2xl border border-indigo-100 bg-indigo-50 p-5 shadow-sm">
            <div class="flex items-center gap-2 text-indigo-700"><Users size={18} /><p class="text-sm font-bold">Petugas Ditugaskan</p></div>
            <p class="mt-3 text-3xl font-black text-indigo-950">{summary.totalOfficers}</p>
            <p class="mt-1 text-xs text-indigo-700">Petugas unik per pengajuan</p>
        </article>
        <article class="rounded-2xl border border-amber-100 bg-amber-50 p-5 shadow-sm">
            <p class="text-sm font-bold text-amber-700">Total Biaya SPJ</p>
            <p class="mt-3 text-2xl font-black text-amber-950">{formatRupiah(summary.totalSPJAmount)}</p>
            <p class="mt-1 text-xs text-amber-700">Dihitung terpisah</p>
        </article>
        <article class="rounded-2xl border border-emerald-100 bg-emerald-50 p-5 shadow-sm">
            <p class="text-sm font-bold text-emerald-700">Total Biaya Riil</p>
            <p class="mt-3 text-2xl font-black text-emerald-950">{formatRupiah(summary.totalRiilAmount)}</p>
            <p class="mt-1 text-xs text-emerald-700">Jam kerja, overtime, hari libur</p>
        </article>
    </section>

    <section class="rounded-[20px] border border-[#dbe5ef] bg-white shadow-[0_8px_22px_rgba(15,41,76,0.07)]">
        <div class="flex flex-col gap-5 border-b border-[#e5edf5] p-6 xl:flex-row xl:items-end xl:justify-between">
            <div>
                <p class="mb-1.5 text-[11px] font-extrabold uppercase tracking-[0.1em] text-[#1f72d4]">Pencarian data</p>
                <h2 class="text-xl font-extrabold text-[#0b1f3a]">Filter Daftar Dalkot</h2>
                <p class="mt-1.5 text-[13px] text-slate-500">Cari pengajuan berdasarkan kegiatan, lokasi, kategori, bulan, atau tahun pelaksanaan.</p>
            </div>

            <div class="grid w-full grid-cols-1 gap-3 sm:grid-cols-2 xl:w-auto xl:grid-cols-5">
                <div class="relative sm:col-span-2 xl:col-span-1">
                    <Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                    <input bind:value={searchQuery} type="search" placeholder="Cari data..." class="w-full rounded-[10px] border border-slate-300 py-2.5 pl-10 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100" />
                </div>
                <select bind:value={categoryFilter} class="rounded-[10px] border border-slate-300 bg-white px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100">
                    <option value="">Semua Kategori</option>
                    {#each categoryOptions as category}
                        <option value={category}>{category}</option>
                    {/each}
                </select>
                <select bind:value={monthFilter} class="rounded-[10px] border border-slate-300 bg-white px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100">
                    <option value="">Semua Bulan</option>
                    {#each monthOptions as month}
                        <option value={month[0]}>{month[1]}</option>
                    {/each}
                </select>
                <select bind:value={yearFilter} class="rounded-[10px] border border-slate-300 bg-white px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100">
                    <option value="">Semua Tahun</option>
                    {#each availableYears as year}
                        <option value={year}>{year}</option>
                    {/each}
                </select>
                <button type="button" class="inline-flex items-center justify-center gap-2 rounded-[10px] border border-slate-300 bg-white px-4 py-2.5 text-sm font-bold text-slate-700 transition hover:bg-slate-50" on:click={resetFilters}>
                    <RefreshCcw size={16} /> Reset
                </button>
            </div>
        </div>
    </section>

    {#if loadError}
        <section class="rounded-2xl border border-red-200 bg-red-50 p-5 text-sm text-red-700">{loadError}</section>
    {/if}

    <section class="overflow-hidden rounded-[20px] border border-[#dbe5ef] bg-white shadow-[0_8px_22px_rgba(15,41,76,0.07)]">
        <div class="flex flex-col gap-4 border-b border-[#e5edf5] p-6 lg:flex-row lg:items-center lg:justify-between">
            <div>
                <p class="mb-1.5 text-[11px] font-extrabold uppercase tracking-[0.1em] text-[#1f72d4]">Data operasional</p>
                <h2 class="text-xl font-extrabold text-[#0b1f3a]">Tabel Dalam Kota</h2>
                <p class="mt-1.5 text-[13px] text-slate-500">Klik ID Dalkot untuk membuka rincian Petugas SPJ dan Petugas Riil. Tombol View & Edit membuka detail pengajuan.</p>
            </div>
            <button type="button" class="inline-flex items-center justify-center gap-2 rounded-[10px] bg-[#1f72d4] px-4 py-2.5 text-sm font-extrabold text-white shadow-sm transition hover:bg-[#185eaf]" on:click={() => goto('/dashboard/dalkot/pengajuan')}>
                <Plus size={16} /> Tambah Pengajuan
            </button>
        </div>

        <div class="overflow-x-auto">
            <table class="w-full min-w-[1100px] border-collapse text-left text-xs">
                <thead>
                    <tr class="bg-gradient-to-r from-[#071a31] to-[#0b1f3a] text-white">
                        <th class="whitespace-nowrap px-[15px] py-3.5 text-[11px] font-extrabold tracking-[0.01em]">ID Dalkot</th>
                        <th class="whitespace-nowrap px-[15px] py-3.5 text-[11px] font-extrabold tracking-[0.01em]">Tanggal</th>
                        <th class="min-w-[250px] px-[15px] py-3.5 text-[11px] font-extrabold tracking-[0.01em]">Nama Kegiatan</th>
                        <th class="whitespace-nowrap px-[15px] py-3.5 text-[11px] font-extrabold tracking-[0.01em]">Kategori Dalkot</th>
                        <th class="whitespace-nowrap px-[15px] py-3.5 text-right text-[11px] font-extrabold tracking-[0.01em]">Total Biaya SPJ</th>
                        <th class="whitespace-nowrap px-[15px] py-3.5 text-right text-[11px] font-extrabold tracking-[0.01em]">Total Biaya Riil</th>
                        <th class="whitespace-nowrap px-[15px] py-3.5 text-[11px] font-extrabold tracking-[0.01em]">Status</th>
                        <th class="whitespace-nowrap px-[15px] py-3.5 text-center text-[11px] font-extrabold tracking-[0.01em]">Aksi</th>
                    </tr>
                </thead>

                <tbody class="bg-white">
                    {#if loading}
                        <tr>
                            <td colspan="8" class="border-b border-[#e6edf4] px-6 py-16 text-center text-slate-500">Memuat daftar Dalkot...</td>
                        </tr>
                    {:else if filteredRows.length === 0}
                        <tr>
                            <td colspan="8" class="border-b border-[#e6edf4] bg-[#fbfdff] px-6 py-14 text-center text-[#7890a3]">
                                <FileText size={30} class="mx-auto text-[#8bb5d4]" />
                                <strong class="mt-3 block text-sm text-[#274d69]">Belum ada pengajuan Dalkot</strong>
                                <span class="mt-1 block text-xs">Tambahkan pengajuan baru atau ubah filter pencarian.</span>
                            </td>
                        </tr>
                    {:else}
                        {#each filteredRows as row (row.id)}
                            <tr class={`transition-colors ${expandedIds.has(row.id) ? 'bg-[#f4f9fd]' : 'hover:bg-[#f8fbff]'}`}>
                                <td class="border-b border-[#e6edf4] px-[15px] py-3.5 align-middle">
                                    <button
                                        type="button"
                                        class={`inline-flex min-w-[86px] items-center justify-between gap-2.5 rounded-[8px] border px-[9px] py-[8px] text-[11px] font-extrabold leading-none shadow-[0_2px_7px_rgba(31,114,212,0.07)] transition ${expandedIds.has(row.id) ? 'border-[#1f78a8] bg-gradient-to-br from-[#1f78a8] to-[#155d83] text-white shadow-[0_6px_14px_rgba(31,120,168,0.18)]' : 'border-[#bdd9f1] bg-gradient-to-b from-[#f4f9ff] to-[#e8f4ff] text-[#15508a] hover:-translate-y-px hover:border-[#78afe0] hover:bg-[#dceeff] hover:shadow-[0_5px_12px_rgba(31,114,212,0.12)]'}`}
                                        title={`${expandedIds.has(row.id) ? 'Tutup' : 'Buka'} rincian petugas ${row.businessId}`}
                                        aria-label={`${expandedIds.has(row.id) ? 'Tutup' : 'Buka'} rincian petugas ${row.businessId}`}
                                        aria-expanded={expandedIds.has(row.id)}
                                        on:click={() => toggleDetail(row.id)}
                                    >
                                        <span>{row.businessId}</span>
                                        <ChevronDown size={12} class={`shrink-0 transition-transform duration-200 ${expandedIds.has(row.id) ? 'rotate-180' : ''}`} />
                                    </button>
                                </td>
                                <td class="whitespace-nowrap border-b border-[#e6edf4] px-[15px] py-3.5 align-middle font-semibold text-slate-700">{formatDate(row.serviceDate)}</td>
                                <td class="border-b border-[#e6edf4] px-[15px] py-3.5 align-middle">
                                    <div class="max-w-[280px] font-semibold leading-snug text-slate-900">{row.activityName}</div>
                                    <div class="mt-1 text-[11px] text-slate-500">{row.dalkotType.replaceAll('_', ' ')}</div>
                                </td>
                                <td class="whitespace-nowrap border-b border-[#e6edf4] px-[15px] py-3.5 align-middle">
                                    <span class="inline-flex rounded-full border border-[#c8ddf1] bg-[#edf5fc] px-2.5 py-1 text-[11px] font-extrabold text-[#1a4e88]">{row.category}</span>
                                </td>
                                <td class="whitespace-nowrap border-b border-[#e6edf4] px-[15px] py-3.5 text-right align-middle font-bold text-slate-900">{formatRupiah(row.totalSPJAmount)}</td>
                                <td class="whitespace-nowrap border-b border-[#e6edf4] px-[15px] py-3.5 text-right align-middle font-bold text-slate-900">{formatRupiah(row.totalRiilAmount)}</td>
                                <td class="whitespace-nowrap border-b border-[#e6edf4] px-[15px] py-3.5 align-middle">
                                    <span class={`inline-flex rounded-full border px-2.5 py-1 text-[11px] font-bold ${statusClass(row.status)}`}>{row.status}</span>
                                </td>
                                <td class="whitespace-nowrap border-b border-[#e6edf4] px-[15px] py-3.5 align-middle">
                                    <div class="flex items-center justify-center gap-2">
                                        <button
                                            type="button"
                                            class="inline-grid h-[31px] w-[31px] place-items-center rounded-lg bg-[#1f72d4] text-white transition hover:-translate-y-px hover:brightness-95"
                                            title="Cetak Laporan Perjadin"
                                            aria-label={`Cetak laporan ${row.businessId}`}
                                            on:click={() => printTravelReport(row)}
                                        >
                                            <FileText size={14} />
                                        </button>
                                        <button
                                            type="button"
                                            class={`inline-grid h-[31px] w-[31px] place-items-center rounded-lg text-white transition ${getSPJOfficers(row).length > 0 ? 'bg-[#0a9d6e] hover:-translate-y-px hover:brightness-95' : 'cursor-not-allowed bg-slate-300'}`}
                                            title={getSPJOfficers(row).length > 0 ? 'Cetak semua DPR Petugas SPJ' : 'DPR tidak tersedia karena tidak ada Petugas SPJ'}
                                            aria-label={`Cetak semua DPR SPJ ${row.businessId}`}
                                            disabled={getSPJOfficers(row).length === 0}
                                            on:click={() => printAllSpjDpr(row)}
                                        >
                                            <Printer size={14} />
                                        </button>
                                        <button
                                            type="button"
                                            class="inline-grid h-[31px] w-[31px] place-items-center rounded-lg bg-[#0b1f3a] text-white transition hover:-translate-y-px hover:brightness-95"
                                            title="View dan Edit"
                                            aria-label={`View dan Edit ${row.businessId}`}
                                            on:click={() => openDetailModal(row)}
                                        >
                                            <Pencil size={14} />
                                        </button>
                                    </div>
                                </td>
                            </tr>

                            {#if expandedIds.has(row.id)}
                                <tr>
                                    <td colspan="8" class="border-b border-[#dbe8f1] bg-[#f4f9fd] px-3 pb-4 pt-1.5">
                                        <div class="grid gap-3 pt-1 xl:grid-cols-2">
                                            <section class="rounded-[12px] border border-[#d3e5f1] bg-[#eaf5fd] px-2 pb-2 pt-1.5 shadow-[inset_0_1px_0_rgba(255,255,255,0.85)]" aria-label="Rincian Petugas SPJ">
                                                <div class="relative z-10 mb-[-1px] ml-0.5 inline-flex min-w-[172px] items-center gap-2 rounded-t-[7px] border border-[#b9cad6] border-b-white bg-white px-3 py-2 text-[11px] font-extrabold leading-none text-[#153e64]">
                                                    <Printer size={13} class="text-[#1f72d4]" />
                                                    <span>Rincian Petugas SPJ</span>
                                                </div>
                                                <div class="overflow-hidden rounded-b-[9px] rounded-tr-[9px] border border-[#b9cad6] bg-white">
                                                    <div class="overflow-x-auto">
                                                        <table class="w-full min-w-[480px] table-fixed border-collapse text-xs">
                                                            <thead>
                                                                <tr class="bg-gradient-to-r from-[#071a31] to-[#0b1f3a] text-white">
                                                                    <th class="w-[44%] border-r border-white/10 px-3 py-2.5 text-left text-[10px] font-extrabold uppercase tracking-[0.025em]">Nama Petugas</th>
                                                                    <th class="w-[24%] border-r border-white/10 px-3 py-2.5 text-left text-[10px] font-extrabold uppercase tracking-[0.025em]">Kategori Dalkot</th>
                                                                    <th class="w-[32%] px-3 py-2.5 text-right text-[10px] font-extrabold uppercase tracking-[0.025em]">Biaya SPJ</th>
                                                                </tr>
                                                            </thead>
                                                            <tbody>
                                                                {#if getSPJOfficers(row).length === 0}
                                                                    <tr><td colspan="3" class="px-3 py-8 text-center text-slate-500">Belum ada Petugas SPJ.</td></tr>
                                                                {:else}
                                                                    {#each getSPJOfficers(row) as officer (officer.id)}
                                                                        <tr class="hover:bg-[#f5faff]">
                                                                            <td class="border-t border-[#e9f0f5] px-3 py-2.5 font-semibold text-slate-800">
                                                                                {officer.name}
                                                                                <div class="mt-0.5 text-[10px] font-normal text-slate-500">{officer.position || officer.nip || '-'}</div>
                                                                            </td>
                                                                            <td class="border-t border-[#e9f0f5] px-3 py-2.5 text-slate-600">SPJ</td>
                                                                            <td class="whitespace-nowrap border-t border-[#e9f0f5] px-3 py-2.5 text-right font-semibold text-slate-800">{formatRupiah(officer.amount)}</td>
                                                                        </tr>
                                                                    {/each}
                                                                {/if}
                                                            </tbody>
                                                        </table>
                                                    </div>
                                                </div>
                                            </section>

                                            <section class="rounded-[12px] border border-[#d3e5f1] bg-[#eaf5fd] px-2 pb-2 pt-1.5 shadow-[inset_0_1px_0_rgba(255,255,255,0.85)]" aria-label="Rincian Petugas Riil">
                                                <div class="relative z-10 mb-[-1px] ml-0.5 inline-flex min-w-[172px] items-center gap-2 rounded-t-[7px] border border-[#b9cad6] border-b-white bg-white px-3 py-2 text-[11px] font-extrabold leading-none text-[#153e64]">
                                                    <Users size={13} class="text-[#1f72d4]" />
                                                    <span>Rincian Petugas Riil</span>
                                                </div>
                                                <div class="overflow-hidden rounded-b-[9px] rounded-tr-[9px] border border-[#b9cad6] bg-white">
                                                    <div class="overflow-x-auto">
                                                        <table class="w-full min-w-[480px] table-fixed border-collapse text-xs">
                                                            <thead>
                                                                <tr class="bg-gradient-to-r from-[#071a31] to-[#0b1f3a] text-white">
                                                                    <th class="w-[44%] border-r border-white/10 px-3 py-2.5 text-left text-[10px] font-extrabold uppercase tracking-[0.025em]">Nama Petugas</th>
                                                                    <th class="w-[24%] border-r border-white/10 px-3 py-2.5 text-left text-[10px] font-extrabold uppercase tracking-[0.025em]">Kategori Dalkot</th>
                                                                    <th class="w-[32%] px-3 py-2.5 text-right text-[10px] font-extrabold uppercase tracking-[0.025em]">Biaya Riil</th>
                                                                </tr>
                                                            </thead>
                                                            <tbody>
                                                                {#if getRiilOfficers(row).length === 0}
                                                                    <tr><td colspan="3" class="px-3 py-8 text-center text-slate-500">Belum ada Petugas Riil.</td></tr>
                                                                {:else}
                                                                    {#each getRiilOfficers(row) as officer (officer.id)}
                                                                        <tr class="hover:bg-[#f5faff]">
                                                                            <td class="border-t border-[#e9f0f5] px-3 py-2.5 font-semibold text-slate-800">
                                                                                {officer.name}
                                                                                <div class="mt-0.5 text-[10px] font-normal text-slate-500">{officer.position || officer.nip || '-'}</div>
                                                                            </td>
                                                                            <td class="border-t border-[#e9f0f5] px-3 py-2.5 text-slate-600">{row.category}</td>
                                                                            <td class="whitespace-nowrap border-t border-[#e9f0f5] px-3 py-2.5 text-right font-semibold text-slate-800">{formatRupiah(officer.amount)}</td>
                                                                        </tr>
                                                                    {/each}
                                                                {/if}
                                                            </tbody>
                                                        </table>
                                                    </div>
                                                </div>
                                            </section>
                                        </div>
                                    </td>
                                </tr>
                            {/if}
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>
    </section>
</main>

{#if selectedDetail}
    <div class="fixed inset-0 z-[120] flex items-center justify-center p-4 sm:p-6">
        <button type="button" class="absolute inset-0 bg-slate-950/65 backdrop-blur-[2px]" aria-label="Tutup detail Dalkot" on:click={closeDetailModal}></button>

        <div class="relative z-10 max-h-[92vh] w-full max-w-4xl overflow-hidden rounded-[20px] border border-[#dbe5ef] bg-white shadow-[0_28px_70px_rgba(3,20,38,0.35)]" role="dialog" aria-modal="true" aria-labelledby="dalkot-detail-title" tabindex="-1">
            <header class="flex items-start justify-between gap-5 bg-gradient-to-r from-[#071a31] via-[#0b1f3a] to-[#155d83] px-6 py-5 text-white">
                <div>
                    <p class="text-[10px] font-extrabold uppercase tracking-[0.16em] text-blue-200">Detail pengajuan Dalkot</p>
                    <div class="mt-2 flex flex-wrap items-center gap-2">
                        <h2 id="dalkot-detail-title" class="text-xl font-extrabold">{selectedDetail.activityName}</h2>
                        <span class="rounded-md border border-white/20 bg-white/10 px-2 py-1 font-mono text-[11px] font-extrabold tracking-wider">{selectedDetail.businessId}</span>
                    </div>
                    <p class="mt-1.5 text-xs text-slate-300">Lihat data lengkap sebelum membuka form perubahan.</p>
                </div>
                <button type="button" class="inline-grid h-9 w-9 shrink-0 place-items-center rounded-lg border border-white/15 bg-white/10 text-white transition hover:bg-white/20" aria-label="Tutup modal" on:click={closeDetailModal}>
                    <X size={17} />
                </button>
            </header>

            <div class="max-h-[calc(92vh-154px)] overflow-y-auto bg-[#f7fafc] p-5 sm:p-6">
                <div class="grid gap-4 lg:grid-cols-[1fr_280px]">
                    <section class="rounded-[14px] border border-[#dbe5ef] bg-white p-5 shadow-[0_6px_16px_rgba(15,41,76,0.05)]">
                        <div class="mb-4 flex flex-wrap items-center justify-between gap-3 border-b border-[#e5edf5] pb-4">
                            <div>
                                <p class="text-[10px] font-extrabold uppercase tracking-[0.1em] text-[#1f72d4]">Informasi kegiatan</p>
                                <h3 class="mt-1 text-base font-extrabold text-[#0b1f3a]">Data Utama Dalkot</h3>
                            </div>
                            <span class={`inline-flex rounded-full border px-2.5 py-1 text-[11px] font-bold ${statusClass(selectedDetail.status)}`}>{selectedDetail.status}</span>
                        </div>

                        <dl class="grid gap-x-5 gap-y-4 sm:grid-cols-2">
                            <div>
                                <dt class="text-[10px] font-extrabold uppercase tracking-wide text-slate-400">Tanggal Pelaksanaan</dt>
                                <dd class="mt-1 text-sm font-semibold text-slate-800">{formatDate(selectedDetail.serviceDate)}</dd>
                            </div>
                            <div>
                                <dt class="text-[10px] font-extrabold uppercase tracking-wide text-slate-400">Kategori Dalkot</dt>
                                <dd class="mt-1 text-sm font-semibold text-slate-800">{selectedDetail.category}</dd>
                            </div>
                            <div>
                                <dt class="text-[10px] font-extrabold uppercase tracking-wide text-slate-400">Jenis Dalkot</dt>
                                <dd class="mt-1 text-sm font-semibold text-slate-800">{selectedDetail.dalkotType.replaceAll('_', ' ')}</dd>
                            </div>
                            <div>
                                <dt class="text-[10px] font-extrabold uppercase tracking-wide text-slate-400">Lokasi</dt>
                                <dd class="mt-1 text-sm font-semibold text-slate-800">{selectedDetail.location || '-'}</dd>
                            </div>
                            <div class="sm:col-span-2">
                                <dt class="text-[10px] font-extrabold uppercase tracking-wide text-slate-400">Pimpinan yang Didampingi</dt>
                                <dd class="mt-1 text-sm font-semibold text-slate-800">{selectedDetail.accompaniedOfficial || '-'}</dd>
                            </div>
                            <div>
                                <dt class="text-[10px] font-extrabold uppercase tracking-wide text-slate-400">Nomor Surat Tugas</dt>
                                <dd class="mt-1 break-words text-sm font-semibold text-slate-800">{selectedDetail.assignmentLetterNumber || '-'}</dd>
                            </div>
                            <div>
                                <dt class="text-[10px] font-extrabold uppercase tracking-wide text-slate-400">Tanggal Surat Tugas</dt>
                                <dd class="mt-1 text-sm font-semibold text-slate-800">{formatDate(selectedDetail.assignmentLetterDate)}</dd>
                            </div>
                        </dl>

                        <div class="mt-5 border-t border-[#e5edf5] pt-4">
                            <p class="text-[10px] font-extrabold uppercase tracking-wide text-slate-400">Isi Laporan</p>
                            <p class="mt-2 whitespace-pre-line text-sm leading-6 text-slate-700">{selectedDetail.reportContent || '-'}</p>
                        </div>
                    </section>

                    <aside class="space-y-4">
                        <section class="rounded-[14px] border border-[#f1d9a5] bg-[#fffaf0] p-4 shadow-[0_6px_16px_rgba(15,41,76,0.04)]">
                            <p class="text-[10px] font-extrabold uppercase tracking-wide text-[#9a6508]">Total Biaya SPJ</p>
                            <p class="mt-2 text-xl font-extrabold text-[#684306]">{formatRupiah(selectedDetail.totalSPJAmount)}</p>
                            <p class="mt-1 text-[11px] text-[#9a6508]">{getSPJOfficers(selectedDetail).length} Petugas SPJ</p>
                        </section>
                        <section class="rounded-[14px] border border-[#bfe4d6] bg-[#f0fbf7] p-4 shadow-[0_6px_16px_rgba(15,41,76,0.04)]">
                            <p class="text-[10px] font-extrabold uppercase tracking-wide text-[#087852]">Total Biaya Riil</p>
                            <p class="mt-2 text-xl font-extrabold text-[#07583e]">{formatRupiah(selectedDetail.totalRiilAmount)}</p>
                            <p class="mt-1 text-[11px] text-[#087852]">{getRiilOfficers(selectedDetail).length} Petugas Riil</p>
                        </section>
                        <div class="rounded-[14px] border border-blue-100 bg-blue-50 p-4 text-[11px] font-semibold leading-5 text-blue-800">
                            Biaya SPJ dan biaya Riil ditampilkan terpisah dan tidak dijumlahkan.
                        </div>
                    </aside>
                </div>
            </div>

            <footer class="flex flex-col-reverse gap-3 border-t border-[#e5edf5] bg-white px-6 py-4 sm:flex-row sm:justify-end">
                <button type="button" class="rounded-[9px] border border-slate-300 bg-white px-4 py-2.5 text-sm font-bold text-slate-700 transition hover:bg-slate-50" on:click={closeDetailModal}>Tutup</button>
                <button type="button" class="inline-flex items-center justify-center gap-2 rounded-[9px] bg-[#0b1f3a] px-4 py-2.5 text-sm font-bold text-white transition hover:bg-[#132f52]" on:click={() => goto(`/dashboard/dalkot/pengajuan?id=${encodeURIComponent(selectedDetail.id)}`)}>
                    <Pencil size={15} /> Edit Pengajuan
                </button>
            </footer>
        </div>
    </div>
{/if}
