<script lang="ts">
    import { onMount } from 'svelte';
    import {
        Info,
        Eye,
        FileDown,
        FileText,
        Pencil,
        Plus,
        Printer,
        ReceiptText,
        RefreshCcw,
        Search,
        Trash2,
        Upload,
        X
    } from 'lucide-svelte';
    import { api } from '$lib/shared/api';
    import { toast } from '$lib/shared/stores/toast';

    type ProcurementType = {
        id: string;
        name: string;
        accountCode: string;
        mak: string;
    };

    type FundingSource = {
        id: string;
        year: number;
        monthNumber: number;
        monthName: string;
        gupLabel: string;
    };

    type GupRow = {
        id: string;
        businessId: string;
        paymentDescription: string;
        procurementTypeId: string;
        procurementTypeName: string;
        accountCode: string;
        mak: string;
        fundingSourceId?: string | null;
        fundingSourceLabel: string;
        valueAmount: number;
        paidAmount: number;
        taxAmount: number;
        differenceAmount: number;
        receiptDate: string;
        recipient: string;
        pum: string;
        documentOriginalName?: string;
        documentMime?: string;
        documentSize?: number;
        documentUrl?: string;
        createdAt?: string;
        updatedAt?: string;
    };

    type GupForm = {
        paymentDescription: string;
        procurementTypeId: string;
        fundingSourceId: string;
        valueAmount: number;
        paidAmount: number;
        taxAmount: number;
        receiptDate: string;
        recipient: string;
        pum: string;
        document: File | null;
    };

    const months = [
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

    const emptyForm = (): GupForm => ({
        paymentDescription: '',
        procurementTypeId: '',
        fundingSourceId: '',
        valueAmount: 0,
        paidAmount: 0,
        taxAmount: 0,
        receiptDate: new Date().toISOString().slice(0, 10),
        recipient: '',
        pum: '',
        document: null
    });

    let rows: GupRow[] = [];
    let procurementTypes: ProcurementType[] = [];
    let fundingSources: FundingSource[] = [];
    let loading = true;
    let saving = false;
    let loadError = '';


    let searchQuery = '';
    let procurementFilter = '';
    let monthFilter = '';

    let formOpen = false;
    let detailOpen = false;
    let deleteOpen = false;
    let editingId: string | null = null;
    let selectedRow: GupRow | null = null;
    let form: GupForm = emptyForm();
    let fileInput: HTMLInputElement;

    $: selectedProcurement = procurementTypes.find(
        (item) => item.id === form.procurementTypeId
    );

    $: formDifference = form.valueAmount - form.paidAmount - form.taxAmount;

    function getRowTimestamp(row: GupRow): number {
        const value = row.createdAt || row.receiptDate || '';
        const timestamp = new Date(value).getTime();
        return Number.isNaN(timestamp) ? 0 : timestamp;
    }

    $: sortedRows = [...rows].sort((firstRow, secondRow) => {
        const timeDifference =
            getRowTimestamp(secondRow) - getRowTimestamp(firstRow);

        if (timeDifference !== 0) {
            return timeDifference;
        }

        return (firstRow.businessId || '').localeCompare(
            secondRow.businessId || ''
        );
    });

    $: filteredRows = sortedRows.filter((row) => {
        const normalizedSearch = searchQuery.trim().toLowerCase();
        const matchSearch =
            normalizedSearch === '' ||
            row.paymentDescription?.toLowerCase().includes(normalizedSearch) ||
            row.recipient?.toLowerCase().includes(normalizedSearch) ||
            row.pum?.toLowerCase().includes(normalizedSearch) ||
            row.accountCode?.toLowerCase().includes(normalizedSearch) ||
            row.mak?.toLowerCase().includes(normalizedSearch);

        const matchProcurement =
            procurementFilter === '' || row.procurementTypeId === procurementFilter;

        const matchMonth =
            monthFilter === '' || row.receiptDate?.slice(5, 7) === monthFilter;

        return matchSearch && matchProcurement && matchMonth;
    });

    onMount(() => {
        loadPage();
    });

    async function loadPage() {
        loading = true;
        loadError = '';

        try {
            const client = api as any;
            const [transactionResult, masterResult] = await Promise.all([
                client.getGupTransactions(),
                client.getGupMaster(new Date().getFullYear())
            ]);

            rows = Array.isArray(transactionResult)
                ? transactionResult
                : transactionResult?.data ?? [];

            procurementTypes = masterResult?.procurementTypes ?? [];
            fundingSources = masterResult?.fundingSources ?? [];
        } catch (error) {
            loadError = error instanceof Error ? error.message : 'Data GUP gagal dimuat.';
            toast.error(loadError);
        } finally {
            loading = false;
        }
    }

    function resetFilters() {
        searchQuery = '';
        procurementFilter = '';
        monthFilter = '';
    }



    function openAddForm() {
        editingId = null;
        selectedRow = null;
        form = emptyForm();

        if (fileInput) {
            fileInput.value = '';
        }

        formOpen = true;
    }

    function openEditForm(row: GupRow) {
        editingId = row.id;
        selectedRow = row;

        form = {
            paymentDescription: row.paymentDescription ?? '',
            procurementTypeId: row.procurementTypeId ?? '',
            fundingSourceId: row.fundingSourceId ?? '',
            valueAmount: Number(row.valueAmount ?? 0),
            paidAmount: Number(row.paidAmount ?? 0),
            taxAmount: Number(row.taxAmount ?? 0),
            receiptDate: row.receiptDate?.slice(0, 10) ?? '',
            recipient: row.recipient ?? '',
            pum: row.pum ?? '',
            document: null
        };

        detailOpen = false;
        formOpen = true;
    }

    function closeForm(force = false) {
        if (saving && !force) {
            return;
        }

        formOpen = false;
        editingId = null;
        selectedRow = null;
        form = emptyForm();

        if (fileInput) {
            fileInput.value = '';
        }
    }

    function openDetail(row: GupRow) {
        selectedRow = row;
        detailOpen = true;
    }

    function confirmDelete(row: GupRow) {
        selectedRow = row;
        deleteOpen = true;
    }

    function parseCurrency(value: string): number {
        return Number(value.replace(/\D/g, '')) || 0;
    }

    function formatNumber(value: number): string {
        return new Intl.NumberFormat('id-ID').format(Number(value || 0));
    }

    function formatRupiah(value: number): string {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            maximumFractionDigits: 0
        }).format(Number(value || 0));
    }

    function formatDate(value: string): string {
        if (!value) return '-';
        const date = new Date(`${value.slice(0, 10)}T00:00:00`);
        if (Number.isNaN(date.getTime())) return value;
        return new Intl.DateTimeFormat('id-ID', {
            day: '2-digit',
            month: 'long',
            year: 'numeric'
        }).format(date);
    }

    function formatFileSize(size = 0): string {
        if (!size) return '';
        if (size < 1024) return `${size} B`;
        if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
        return `${(size / (1024 * 1024)).toFixed(2)} MB`;
    }

    function handleFileChange(event: Event) {
        const input = event.currentTarget as HTMLInputElement;
        const file = input.files?.[0] ?? null;

        if (!file) {
            form.document = null;
            return;
        }

        const validMime = ['application/pdf', 'image/jpeg'].includes(file.type);
        const validExtension = /\.(pdf|jpe?g)$/i.test(file.name);

        if (!validMime || !validExtension) {
            toast.error('Dokumen hanya boleh berupa PDF, JPG, atau JPEG.');
            input.value = '';
            form.document = null;
            return;
        }

        if (file.size > 5 * 1024 * 1024) {
            toast.error('Ukuran dokumen maksimal 5 MB.');
            input.value = '';
            form.document = null;
            return;
        }

        form.document = file;
    }

    function validateForm(): boolean {
        if (!form.paymentDescription.trim()) {
            toast.error('Pembayaran wajib diisi.');
            return false;
        }

        if (!form.procurementTypeId) {
            toast.error('Jenis pengadaan wajib dipilih.');
            return false;
        }

        if (!form.fundingSourceId) {
            toast.error('Sumber dana wajib dipilih.');
            return false;
        }

        if (!form.receiptDate) {
            toast.error('Tanggal kwitansi wajib diisi.');
            return false;
        }

        if (!form.recipient.trim()) {
            toast.error('Penerima wajib diisi.');
            return false;
        }

        if (!form.pum.trim()) {
            toast.error('PUM wajib diisi.');
            return false;
        }

        return true;
    }

    function buildFormData(): FormData {
        const data = new FormData();
        data.append('paymentDescription', form.paymentDescription.trim());
        data.append('procurementTypeId', form.procurementTypeId);
        data.append('fundingSourceId', form.fundingSourceId);
        data.append('valueAmount', String(form.valueAmount));
        data.append('paidAmount', String(form.paidAmount));
        data.append('taxAmount', String(form.taxAmount));
        data.append('receiptDate', form.receiptDate);
        data.append('recipient', form.recipient.trim());
        data.append('pum', form.pum.trim());

        if (form.document) {
            data.append('document', form.document, form.document.name);
        }

        return data;
    }

    async function saveGup() {
        if (!validateForm()) {
            return;
        }

        const currentEditingId = editingId;
        const isEditing = Boolean(currentEditingId);

        saving = true;

        try {
            const client = api as any;

            if (isEditing && currentEditingId) {
                await client.updateGup(
                    currentEditingId,
                    buildFormData()
                );
            } else {
                await client.createGup(
                    buildFormData()
                );
            }

            const successMessage = isEditing
                ? 'Data GUP berhasil diperbarui.'
                : 'Data GUP berhasil ditambahkan.';

            await loadPage();

            saving = false;
            closeForm(true);
            toast.success(successMessage);
        } catch (error) {
            saving = false;

            const message =
                error instanceof Error
                    ? error.message
                    : 'Data GUP gagal disimpan.';

            toast.error(message);
        }
    }

    async function deleteGup() {
        if (!selectedRow) return;
        saving = true;

        try {
            await (api as any).deleteGup(selectedRow.id);
            toast.success('Data GUP berhasil dihapus.');
            deleteOpen = false;
            selectedRow = null;
            await loadPage();
        } catch (error) {
            toast.error(error instanceof Error ? error.message : 'Data GUP gagal dihapus.');
        } finally {
            saving = false;
        }
    }

    function printPage() {
        window.print();
    }
</script>

<svelte:head>
    <title>Pengajuan GUP | Perjadin</title>
</svelte:head>

<main class="space-y-6 pb-20">
    <section class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
            <div>
                <p class="mb-1 text-sm text-slate-500">Modul Pengajuan Ganti Uang Persediaan</p>
                <h1 class="text-2xl font-bold text-slate-900 sm:text-3xl">Pengajuan GUP</h1>
                <p class="mt-2 max-w-3xl text-sm text-slate-500">
                    Tabel menampilkan ringkasan pengajuan. Gunakan tombol View untuk rincian,
                    Cetak Kwitansi untuk bukti pembayaran, dan Cetak SPBY untuk Surat Perintah Bayar.
                </p>
            </div>

            <div class="flex flex-col gap-3 sm:flex-row print:hidden">
                <button
                    type="button"
                    class="inline-flex items-center justify-center gap-2 rounded-xl bg-amber-500 px-5 py-3 text-sm font-semibold text-slate-950 shadow-sm transition hover:bg-amber-600"
                    on:click={printPage}
                >
                    <Printer size={17} />
                    Cetak Rekapitulasi
                </button>

                <button
                    type="button"
                    class="inline-flex items-center justify-center gap-2 rounded-xl bg-blue-200 px-5 py-3 text-sm font-semibold text-slate-900 shadow-sm transition hover:bg-blue-300"
                    on:click={openAddForm}
                >
                    <Plus size={17} />
                    Tambah GUP
                </button>
            </div>
        </div>
    </section>

    <section class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
        <div class="flex flex-col gap-3 border-b border-slate-200 px-6 py-5 sm:flex-row sm:items-center sm:justify-between">
            <div>
                <h2 class="text-lg font-bold text-slate-900">Tabel Pengajuan GUP</h2>
                <p class="text-sm text-slate-500">Kode Akun dan MAK dibuat otomatis berdasarkan Jenis Pengadaan.</p>
            </div>
            <div class="flex items-center gap-2 text-sm text-slate-500 print:hidden">
                <Info size={17} class="text-blue-500" />
                <span>Klik View untuk melihat rincian dan mengubah data.</span>
            </div>
        </div>

        <div class="grid grid-cols-1 gap-3 border-b border-slate-200 bg-slate-50/70 px-6 py-4 md:grid-cols-2 xl:grid-cols-4 print:hidden">
            <div>
                <label for="filterJenis" class="mb-2 block text-xs font-bold uppercase tracking-wide text-slate-500">Jenis Pengadaan</label>
                <select id="filterJenis" bind:value={procurementFilter} class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-slate-800">
                    <option value="">Semua Jenis Pengadaan</option>
                    {#each procurementTypes as item}
                        <option value={item.id}>{item.name}</option>
                    {/each}
                </select>
            </div>

            <div>
                <label for="filterBulan" class="mb-2 block text-xs font-bold uppercase tracking-wide text-slate-500">Bulan Transaksi</label>
                <select id="filterBulan" bind:value={monthFilter} class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-slate-800">
                    <option value="">Semua Bulan</option>
                    {#each months as month}
                        <option value={month[0]}>{month[1]}</option>
                    {/each}
                </select>
            </div>

            <div>
                <label for="searchGup" class="mb-2 block text-xs font-bold uppercase tracking-wide text-slate-500">Search</label>
                <div class="relative">
                    <Search size={17} class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400" />
                    <input id="searchGup" bind:value={searchQuery} type="search" placeholder="Cari pembayaran, penerima, PUM, akun..." class="w-full rounded-xl border border-slate-300 bg-white py-3 pl-11 pr-4 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-slate-800" />
                </div>
            </div>

            <div class="flex items-end gap-3">
                <button type="button" on:click={resetFilters} class="inline-flex flex-1 items-center justify-center gap-2 rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm font-semibold text-slate-700 transition hover:bg-slate-100">
                    <RefreshCcw size={16} />
                    Reset Filter
                </button>
                <div class="min-w-24 rounded-xl bg-slate-900 px-3 py-3 text-center text-xs font-bold text-white">
                    {filteredRows.length} data
                </div>
            </div>
        </div>

        {#if loadError}
            <div class="m-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
                {loadError}
                <button type="button" class="ml-2 font-bold underline" on:click={loadPage}>Muat ulang</button>
            </div>
        {/if}

        <div class="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden relative flex flex-col">
    <div class="overflow-x-auto w-full relative min-h-[400px]">
        <table class="w-full text-sm text-left relative border-collapse">
            <thead
                class="bg-slate-50 sticky top-0 z-20 shadow-sm border-b border-slate-200"
            >
                <tr>
                    <th
                        class="w-[60px] min-w-[60px] bg-slate-50 py-3 pl-4 font-semibold text-slate-700"
                    >
                        No
                    </th>

                    <th
                        class="min-w-[240px] bg-slate-50 py-3 font-semibold text-slate-700"
                    >
                        Pembayaran
                    </th>

                    <th
                        class="min-w-[135px] bg-slate-50 py-3 text-right font-semibold text-slate-700"
                    >
                        Nilai
                    </th>

                    <th
                        class="min-w-[145px] bg-slate-50 py-3 text-right font-semibold text-slate-700"
                    >
                        Dibayarkan
                    </th>

                    <th
                        class="min-w-[110px] bg-slate-50 py-3 text-right font-semibold text-slate-700"
                    >
                        Pajak
                    </th>

                    <th
                        class="min-w-[110px] bg-slate-50 py-3 text-right font-semibold text-slate-700"
                    >
                        Selisih
                    </th>

                    <th
                        class="min-w-[145px] bg-slate-50 py-3 pl-5 font-semibold text-slate-700"
                    >
                        Tanggal
                    </th>

                    <th
                        class="w-[190px] min-w-[190px] bg-slate-50 py-3 pr-4 text-center font-semibold text-slate-700 print:hidden"
                    >
                        Aksi
                    </th>
                </tr>
            </thead>

            <tbody>
                {#if loading}
                    <tr>
                        <td colspan="8" class="p-0">
                            <div
                                class="flex min-h-[320px] flex-col items-center justify-center text-slate-500"
                            >
                                <div
                                    class="mb-3 h-8 w-8 animate-spin rounded-full border-2 border-slate-200 border-t-blue-600"
                                ></div>

                                <p class="text-sm">
                                    Memuat data GUP...
                                </p>
                            </div>
                        </td>
                    </tr>
                {:else if filteredRows.length === 0}
                    <tr>
                        <td colspan="8" class="p-0">
                            <div
                                class="flex min-h-[320px] flex-col items-center justify-center px-6 text-center text-slate-500"
                            >
                                <FileText
                                    size={48}
                                    strokeWidth={1.5}
                                    class="mb-4 text-slate-300"
                                />

                                <p class="font-semibold text-slate-700">
                                    Belum ada pengajuan GUP
                                </p>

                                <p class="mt-1 text-sm text-slate-400">
                                    Tambahkan data baru atau ubah filter pencarian.
                                </p>
                            </div>
                        </td>
                    </tr>
                {:else}
                    {#each filteredRows as row, index (row.id)}
                        <tr
                            class="border-b border-slate-100 bg-white transition-colors hover:bg-slate-50/50"
                        >
                            <td
                                class="py-4 pl-4 align-middle font-semibold text-slate-700"
                            >
                                {filteredRows.length - index}
                            </td>

                            <td class="py-4 align-middle">
                                <div
                                    class="max-w-[280px] font-medium text-slate-800"
                                >
                                    {row.paymentDescription}
                                </div>

                                <div
                                    class="mt-1.5 flex flex-wrap items-center gap-2"
                                >
                                    <span
                                        class="inline-flex items-center rounded-full border border-indigo-200 bg-indigo-50 px-2 py-0.5 text-[9px] font-bold tracking-wide text-indigo-700"
                                    >
                                        {row.procurementTypeName || 'Tanpa jenis'}
                                    </span>
                                </div>

                                <div
                                    class="mt-1.5 text-[10px] text-slate-400"
                                >
                                    {row.accountCode || '-'}
                                </div>
                            </td>

                            <td
                                class="whitespace-nowrap py-4 text-right align-middle text-sm font-semibold text-slate-800"
                            >
                                {formatRupiah(row.valueAmount)}
                            </td>

                            <td
                                class="whitespace-nowrap py-4 text-right align-middle text-sm font-semibold text-slate-800"
                            >
                                {formatRupiah(row.paidAmount)}
                            </td>

                            <td
                                class="whitespace-nowrap py-4 text-right align-middle text-sm text-slate-600"
                            >
                                {formatRupiah(row.taxAmount)}
                            </td>

                            <td
                                class="whitespace-nowrap py-4 text-right align-middle"
                            >
                                <span
                                    class:bg-emerald-50={Number(row.differenceAmount) === 0}
                                    class:text-emerald-700={Number(row.differenceAmount) === 0}
                                    class:border-emerald-200={Number(row.differenceAmount) === 0}
                                    class:bg-amber-50={Number(row.differenceAmount) !== 0}
                                    class:text-amber-700={Number(row.differenceAmount) !== 0}
                                    class:border-amber-200={Number(row.differenceAmount) !== 0}
                                    class="inline-flex rounded-full border px-2.5 py-1 text-[10px] font-bold"
                                >
                                    {formatRupiah(row.differenceAmount)}
                                </span>
                            </td>

                            <td
                                class="py-4 pl-5 align-middle text-xs text-slate-600"
                            >
                                <div class="whitespace-nowrap">
                                    {formatDate(row.receiptDate)}
                                </div>

                                <div class="mt-1 text-[10px] text-slate-400">
                                    {row.fundingSourceLabel || '-'}
                                </div>
                            </td>

                            <td
                                class="py-4 pr-4 text-center align-middle print:hidden"
                            >
                                <div
                                    class="flex items-center justify-center gap-2"
                                >
                                    <button
                                        type="button"
                                        class="rounded bg-blue-50 p-1.5 text-blue-600 transition-colors hover:bg-blue-100 hover:text-blue-800"
                                        title="Lihat detail"
                                        aria-label="Lihat detail GUP"
                                        on:click={() => openDetail(row)}
                                    >
                                        <Eye size={16} />
                                    </button>

                                    <button
                                        type="button"
                                        class="rounded bg-amber-50 p-1.5 text-amber-600 transition-colors hover:bg-amber-100 hover:text-amber-800"
                                        title="Cetak Kwitansi"
                                        aria-label="Cetak Kwitansi"
                                        on:click={printPage}
                                    >
                                        <ReceiptText size={16} />
                                    </button>

                                    <button
                                        type="button"
                                        class="rounded bg-sky-50 p-1.5 text-sky-600 transition-colors hover:bg-sky-100 hover:text-sky-800"
                                        title="Cetak SPBY"
                                        aria-label="Cetak SPBY"
                                        on:click={printPage}
                                    >
                                        <FileDown size={16} />
                                    </button>

                                    <button
                                        type="button"
                                        class="rounded bg-red-50 p-1.5 text-red-600 transition-colors hover:bg-red-100 hover:text-red-800"
                                        title="Hapus"
                                        aria-label="Hapus data GUP"
                                        on:click={() => confirmDelete(row)}
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </div>
                            </td>
                        </tr>
                    {/each}
                {/if}
            </tbody>
        </table>
    </div>
</div>
    </section>

    <section class="grid grid-cols-1 gap-6 xl:grid-cols-3 print:hidden">
        <article class="rounded-2xl border border-blue-100 bg-blue-50 p-5">
            <h3 class="font-bold text-blue-900">Kode Akun Otomatis</h3>
            <p class="mt-2 text-sm leading-6 text-blue-800">Kode akun dan MAK mengikuti jenis pengadaan yang dipilih pada formulir.</p>
        </article>
        <article class="rounded-2xl border border-amber-100 bg-amber-50 p-5">
            <h3 class="font-bold text-amber-900">Perhitungan Otomatis</h3>
            <p class="mt-2 text-sm leading-6 text-amber-800">Selisih dihitung dari Nilai dikurangi Jumlah Dibayarkan dan Pajak.</p>
        </article>
        <article class="rounded-2xl border border-emerald-100 bg-emerald-50 p-5">
            <h3 class="font-bold text-emerald-900">Dokumen Pendukung</h3>
            <p class="mt-2 text-sm leading-6 text-emerald-800">PDF atau JPEG maksimal 5 MB hanya ditampilkan melalui menu View.</p>
        </article>
    </section>
</main>

{#if formOpen}
    <div
        class="fixed inset-0 z-[100] flex items-center justify-center bg-slate-950/60 p-4"
        role="presentation"
        on:click={(event) => {
            if (event.currentTarget === event.target) {
                closeForm();
            }
        }}
    >
        <div
            class="max-h-[92vh] w-full max-w-5xl overflow-y-auto rounded-2xl bg-white shadow-2xl"
            role="dialog"
            aria-modal="true"
            aria-labelledby="gupFormTitle"
        >
            <div
                class="sticky top-0 z-10 flex items-center justify-between border-b border-slate-200 bg-white px-6 py-5"
            >
                <div>
                    <p class="text-sm font-semibold text-blue-600">
                        Form Pengajuan GUP
                    </p>
                    <h2
                        id="gupFormTitle"
                        class="text-xl font-bold text-slate-900"
                    >
                        {editingId ? 'Edit GUP' : 'Tambah GUP'}
                    </h2>
                </div>

                <button
                    type="button"
                    on:click={() => closeForm()}
                    disabled={saving}
                    class="rounded-lg p-2 text-slate-500 hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
                    aria-label="Tutup form pengajuan GUP"
                    title="Tutup"
                >
                    <X size={21} />
                </button>
            </div>

            <form
                class="space-y-6 p-6"
                on:submit|preventDefault={saveGup}
            >
                <div class="grid grid-cols-1 gap-5 md:grid-cols-2">
                    <div class="md:col-span-2">
                        <label
                            for="paymentDescription"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Pembayaran
                        </label>
                        <textarea
                            id="paymentDescription"
                            bind:value={form.paymentDescription}
                            rows="3"
                            class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-slate-800"
                            placeholder="Masukkan uraian pembayaran"
                            disabled={saving}
                        ></textarea>
                    </div>

                    <div>
                        <label
                            for="procurementType"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Jenis Pengadaan
                        </label>
                        <select
                            id="procurementType"
                            bind:value={form.procurementTypeId}
                            class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-slate-800"
                            disabled={saving}
                        >
                            <option value="">Pilih Jenis Pengadaan</option>
                            {#each procurementTypes as item}
                                <option value={item.id}>{item.name}</option>
                            {/each}
                        </select>
                    </div>

                    <div>
                        <label
                            for="fundingSource"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Sumber Dana
                        </label>
                        <select
                            id="fundingSource"
                            bind:value={form.fundingSourceId}
                            class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-slate-800"
                            disabled={saving}
                        >
                            <option value="">Pilih Sumber Dana</option>
                            {#each fundingSources as item}
                                <option value={item.id}>
                                    {item.monthName} · {item.gupLabel}
                                </option>
                            {/each}
                        </select>
                    </div>

                    <div>
                        <label
                            for="accountCode"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Kode Akun
                        </label>
                        <input
                            id="accountCode"
                            value={selectedProcurement?.accountCode ?? ''}
                            readonly
                            class="w-full rounded-xl border border-slate-200 bg-slate-100 px-4 py-3 text-sm text-slate-600"
                        />
                    </div>

                    <div>
                        <label
                            for="mak"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            MAK
                        </label>
                        <input
                            id="mak"
                            value={selectedProcurement?.mak ?? ''}
                            readonly
                            class="w-full rounded-xl border border-slate-200 bg-slate-100 px-4 py-3 text-sm text-slate-600"
                        />
                    </div>

                    <div>
                        <label
                            for="valueAmount"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Nilai
                        </label>
                        <div class="relative">
                            <span
                                class="absolute left-4 top-1/2 -translate-y-1/2 text-sm font-semibold text-slate-500"
                            >
                                Rp
                            </span>
                            <input
                                id="valueAmount"
                                value={formatNumber(form.valueAmount)}
                                on:input={(event) => {
                                    form.valueAmount = parseCurrency(
                                        event.currentTarget.value
                                    );
                                }}
                                inputmode="numeric"
                                class="w-full rounded-xl border border-slate-300 py-3 pl-11 pr-4 text-right text-sm focus:outline-none focus:ring-2 focus:ring-slate-800"
                                disabled={saving}
                            />
                        </div>
                    </div>

                    <div>
                        <label
                            for="paidAmount"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Jumlah Dibayarkan
                        </label>
                        <div class="relative">
                            <span
                                class="absolute left-4 top-1/2 -translate-y-1/2 text-sm font-semibold text-slate-500"
                            >
                                Rp
                            </span>
                            <input
                                id="paidAmount"
                                value={formatNumber(form.paidAmount)}
                                on:input={(event) => {
                                    form.paidAmount = parseCurrency(
                                        event.currentTarget.value
                                    );
                                }}
                                inputmode="numeric"
                                class="w-full rounded-xl border border-slate-300 py-3 pl-11 pr-4 text-right text-sm focus:outline-none focus:ring-2 focus:ring-slate-800"
                                disabled={saving}
                            />
                        </div>
                    </div>

                    <div>
                        <label
                            for="taxAmount"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Pajak
                        </label>
                        <div class="relative">
                            <span
                                class="absolute left-4 top-1/2 -translate-y-1/2 text-sm font-semibold text-slate-500"
                            >
                                Rp
                            </span>
                            <input
                                id="taxAmount"
                                value={formatNumber(form.taxAmount)}
                                on:input={(event) => {
                                    form.taxAmount = parseCurrency(
                                        event.currentTarget.value
                                    );
                                }}
                                inputmode="numeric"
                                class="w-full rounded-xl border border-slate-300 py-3 pl-11 pr-4 text-right text-sm focus:outline-none focus:ring-2 focus:ring-slate-800"
                                disabled={saving}
                            />
                        </div>
                    </div>

                    <div>
                        <label
                            for="differenceAmount"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Selisih
                        </label>
                        <input
                            id="differenceAmount"
                            value={formatRupiah(formDifference)}
                            readonly
                            class="w-full rounded-xl border border-slate-200 bg-slate-100 px-4 py-3 text-right text-sm font-bold text-slate-700"
                        />
                    </div>

                    <div>
                        <label
                            for="receiptDate"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Tanggal Kwitansi
                        </label>
                        <input
                            id="receiptDate"
                            type="date"
                            bind:value={form.receiptDate}
                            class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-slate-800"
                            disabled={saving}
                        />
                    </div>

                    <div>
                        <label
                            for="recipient"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Penerima
                        </label>
                        <input
                            id="recipient"
                            bind:value={form.recipient}
                            class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-slate-800"
                            placeholder="Nama penerima"
                            disabled={saving}
                        />
                    </div>

                    <div>
                        <label
                            for="pum"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            PUM
                        </label>
                        <input
                            id="pum"
                            bind:value={form.pum}
                            class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-slate-800"
                            placeholder="Nama PUM"
                            disabled={saving}
                        />
                    </div>

                    <div class="md:col-span-2">
                        <label
                            for="document"
                            class="mb-2 block text-sm font-semibold text-slate-700"
                        >
                            Dokumen Pendukung
                        </label>
                        <label
                            class="flex cursor-pointer items-center justify-center gap-3 rounded-xl border-2 border-dashed border-slate-300 bg-slate-50 px-5 py-6 text-sm text-slate-600 transition hover:border-blue-400 hover:bg-blue-50"
                        >
                            <Upload size={20} />
                            <span>
                                {#if form.document}
                                    {form.document.name}
                                {:else if editingId &&
                                selectedRow?.documentOriginalName}
                                    Dokumen lama:
                                    {selectedRow.documentOriginalName}
                                {:else}
                                    Pilih PDF, JPG, atau JPEG maksimal 5 MB
                                {/if}
                            </span>
                            <input
                                bind:this={fileInput}
                                id="document"
                                type="file"
                                accept=".pdf,.jpg,.jpeg,application/pdf,image/jpeg"
                                class="hidden"
                                on:change={handleFileChange}
                                disabled={saving}
                            />
                        </label>
                    </div>
                </div>

                <div
                    class="flex flex-col-reverse gap-3 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end"
                >
                    <button
                        type="button"
                        on:click={() => closeForm()}
                        disabled={saving}
                        class="rounded-xl border border-slate-300 bg-white px-5 py-3 text-sm font-semibold text-slate-700 hover:bg-slate-50 disabled:opacity-50"
                    >
                        Batal
                    </button>
                    <button
                        type="submit"
                        disabled={saving}
                        class="rounded-xl bg-slate-900 px-5 py-3 text-sm font-semibold text-white hover:bg-slate-800 disabled:opacity-50"
                    >
                        {saving
                            ? 'Menyimpan...'
                            : editingId
                              ? 'Simpan Perubahan'
                              : 'Simpan GUP'}
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}

{#if detailOpen && selectedRow}
    <div class="fixed inset-0 z-[100] flex items-center justify-center bg-slate-950/60 p-4" role="presentation" on:click={(event) => event.currentTarget === event.target && (detailOpen = false)}>
        <div class="max-h-[92vh] w-full max-w-4xl overflow-y-auto rounded-2xl bg-white shadow-2xl" role="dialog" aria-modal="true">
            <div class="sticky top-0 z-10 flex items-center justify-between border-b border-slate-200 bg-white px-6 py-5">
                <div>
                    <p class="text-sm font-semibold text-blue-600">Detail Pengajuan</p>
                    <h2 class="text-xl font-bold text-slate-900">{selectedRow.businessId || 'Data GUP'}</h2>
                </div>
                <button type="button" on:click={() => (detailOpen = false)} class="rounded-lg p-2 text-slate-500 hover:bg-slate-100"><X size={21} /></button>
            </div>

            <div class="space-y-6 p-6">
                <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                    {#each [
                        ['Pembayaran', selectedRow.paymentDescription],
                        ['Jenis Pengadaan', selectedRow.procurementTypeName],
                        ['Kode Akun', selectedRow.accountCode],
                        ['MAK', selectedRow.mak],
                        ['Sumber Dana', selectedRow.fundingSourceLabel],
                        ['Tanggal Kwitansi', formatDate(selectedRow.receiptDate)],
                        ['Penerima', selectedRow.recipient],
                        ['PUM', selectedRow.pum]
                    ] as item}
                        <div class="rounded-xl border border-slate-200 bg-slate-50 p-4">
                            <p class="text-xs font-bold uppercase tracking-wide text-slate-500">{item[0]}</p>
                            <p class="mt-2 text-sm font-semibold text-slate-900">{item[1] || '-'}</p>
                        </div>
                    {/each}
                </div>

                <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
                    <div class="rounded-xl bg-slate-900 p-4 text-white"><p class="text-xs text-slate-300">Nilai</p><p class="mt-2 font-bold">{formatRupiah(selectedRow.valueAmount)}</p></div>
                    <div class="rounded-xl bg-blue-600 p-4 text-white"><p class="text-xs text-blue-100">Dibayarkan</p><p class="mt-2 font-bold">{formatRupiah(selectedRow.paidAmount)}</p></div>
                    <div class="rounded-xl bg-amber-500 p-4 text-slate-950"><p class="text-xs text-amber-900">Pajak</p><p class="mt-2 font-bold">{formatRupiah(selectedRow.taxAmount)}</p></div>
                    <div class="rounded-xl bg-emerald-600 p-4 text-white"><p class="text-xs text-emerald-100">Selisih</p><p class="mt-2 font-bold">{formatRupiah(selectedRow.differenceAmount)}</p></div>
                </div>

                <div>
                    <h3 class="mb-3 flex items-center gap-2 font-bold text-slate-900"><FileText size={18} /> Dokumen Pendukung</h3>
                    {#if selectedRow.documentUrl}
                        <div class="overflow-hidden rounded-xl border border-slate-200 bg-slate-50">
                            <div class="flex items-center justify-between border-b border-slate-200 px-4 py-3 text-sm">
                                <span class="font-semibold text-slate-700">{selectedRow.documentOriginalName || 'Dokumen GUP'}</span>
                                <span class="text-xs text-slate-500">{formatFileSize(selectedRow.documentSize)}</span>
                            </div>
                            {#if selectedRow.documentMime === 'application/pdf'}
                                <iframe title="Dokumen GUP" src={selectedRow.documentUrl} class="h-[520px] w-full"></iframe>
                            {:else}
                                <img src={selectedRow.documentUrl} alt="Dokumen GUP" class="mx-auto max-h-[520px] object-contain p-4" />
                            {/if}
                        </div>
                    {:else}
                        <div class="rounded-xl border border-dashed border-slate-300 bg-slate-50 p-8 text-center text-sm text-slate-500">Tidak ada dokumen pendukung.</div>
                    {/if}
                </div>

                <div class="flex flex-col-reverse gap-3 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
                    <button type="button" on:click={() => (detailOpen = false)} class="rounded-xl border border-slate-300 bg-white px-5 py-3 text-sm font-semibold text-slate-700 hover:bg-slate-50">Tutup</button>
                    <button type="button" on:click={() => openEditForm(selectedRow)} class="inline-flex items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 py-3 text-sm font-semibold text-white hover:bg-blue-700"><Pencil size={16} /> Edit Data</button>
                </div>
            </div>
        </div>
    </div>
{/if}

{#if deleteOpen && selectedRow}
    <div class="fixed inset-0 z-[110] flex items-center justify-center bg-slate-950/60 p-4">
        <div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
            <div class="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-100 text-red-700"><Trash2 size={22} /></div>
            <h2 class="text-xl font-bold text-slate-900">Hapus data GUP?</h2>
            <p class="mt-2 text-sm leading-6 text-slate-500">Data “{selectedRow.paymentDescription}” dan dokumen pendukungnya akan dihapus permanen.</p>
            <div class="mt-6 flex justify-end gap-3">
                <button type="button" on:click={() => (deleteOpen = false)} disabled={saving} class="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-semibold text-slate-700 hover:bg-slate-50">Batal</button>
                <button type="button" on:click={deleteGup} disabled={saving} class="rounded-xl bg-red-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-red-700 disabled:opacity-50">{saving ? 'Menghapus...' : 'Hapus'}</button>
            </div>
        </div>
    </div>
{/if}

<style>
    @media print {
        :global(body) {
            background: white !important;
        }

        :global(aside),
        :global(header) {
            display: none !important;
        }

        main {
            padding: 0;
        }
    }
</style>
