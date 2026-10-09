<script lang="ts">
    import { onMount } from 'svelte';
    import {
        Banknote,
        CircleDollarSign,
        Eye,
        Pencil,
        Plus,
        Save,
        Trash2,
        WalletCards,
        X
    } from 'lucide-svelte';
    import { api } from '$lib/shared/api';
    import { toast } from '$lib/shared/stores/toast';
    import type {
        FinanceAllocation,
        FinanceDashboard,
        FinanceSource,
        SaveFinanceAllocationPayload,
        SaveFinanceSourcePayload
    } from '$lib/shared/api/types';

    const currentYear = new Date().getFullYear();

    let year = currentYear;
    let loading = true;
    let saving = false;
    let dashboard: FinanceDashboard = {
        year,
        sources: [],
        allocations: [],
        summary: {
            totalFunds: 0,
            totalGup: 0,
            totalLs: 0,
            totalUsedGup: 0,
            totalUsedLs: 0,
            totalUsed: 0,
            totalAllocated: 0,
            totalAvailable: 0,
            totalRemaining: 0
        }
    };

    let sourceModalOpen = false;
    let editingSource: FinanceSource | null = null;

    let viewModalOpen = false;
    let viewingSource: FinanceSource | null = null;
    let sourceForm: SaveFinanceSourcePayload = emptySourceForm();

    let allocationModalOpen = false;
    let allocationType: 'GUP' | 'LS' = 'GUP';
    let allocationForm: SaveFinanceAllocationPayload = emptyAllocationForm('GUP');

    $: availableSources = dashboard.sources.filter(
        (source) => source.isActive && Number(source.available || 0) > 0
    );

    onMount(loadData);

    function emptySourceForm(): SaveFinanceSourcePayload {
        return {
            year,
            code: '',
            name: '',
            amount: 0,
            notes: '',
            isActive: true
        };
    }

    function emptyAllocationForm(type: 'GUP' | 'LS'): SaveFinanceAllocationPayload {
        return {
            financeSourceId: '',
            allocationType: type,
            amount: 0,
            notes: ''
        };
    }

    async function loadData() {
        loading = true;
        try {
            dashboard = await api.getFinanceSources(year);
        } catch (error) {
            toast.error(error instanceof Error ? error.message : 'Sumber dana gagal dimuat.');
        } finally {
            loading = false;
        }
    }

    async function changeYear() {
        await loadData();
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
        return new Intl.DateTimeFormat('id-ID', {
            day: '2-digit',
            month: 'short',
            year: 'numeric'
        }).format(new Date(value));
    }

    function openViewSource(source: FinanceSource) {
        viewingSource = source;
        viewModalOpen = true;
    }

    function closeViewSource() {
        viewModalOpen = false;
        viewingSource = null;
    }

    function editViewedSource() {
        if (!viewingSource) return;
        const source = viewingSource;
        closeViewSource();
        openEditSource(source);
    }

    function openCreateSource() {
        editingSource = null;
        sourceForm = emptySourceForm();
        sourceModalOpen = true;
    }

    function openEditSource(source: FinanceSource) {
        editingSource = source;
        sourceForm = {
            year: source.year,
            code: source.code,
            name: source.name,
            amount: source.amount,
            notes: source.notes || '',
            isActive: source.isActive
        };
        sourceModalOpen = true;
    }

    function closeSourceModal() {
        if (saving) return;
        sourceModalOpen = false;
        editingSource = null;
    }

    async function saveSource() {
        if (!sourceForm.code.trim() || !sourceForm.name.trim()) {
            toast.error('Kode dan nama sumber dana wajib diisi.');
            return;
        }
        if (Number(sourceForm.amount) < 0) {
            toast.error('Nominal sumber dana tidak boleh negatif.');
            return;
        }

        saving = true;
        try {
            if (editingSource) {
                await api.updateFinanceSource(editingSource.id, {
                    ...sourceForm,
                    amount: Number(sourceForm.amount)
                });
                toast.success('Sumber dana berhasil diperbarui.');
            } else {
                await api.createFinanceSource({
                    ...sourceForm,
                    amount: Number(sourceForm.amount)
                });
                toast.success('Sumber dana berhasil ditambahkan.');
            }
            sourceModalOpen = false;
            editingSource = null;
            await loadData();
        } catch (error) {
            toast.error(error instanceof Error ? error.message : 'Sumber dana gagal disimpan.');
        } finally {
            saving = false;
        }
    }

    async function deleteSource(source: FinanceSource) {
        const ok = window.confirm(`Hapus sumber dana ${source.code} - ${source.name}?`);
        if (!ok) return;

        try {
            await api.deleteFinanceSource(source.id);
            toast.success('Sumber dana berhasil dihapus.');
            if (viewingSource?.id === source.id) {
                closeViewSource();
            }
            await loadData();
        } catch (error) {
            toast.error(error instanceof Error ? error.message : 'Sumber dana gagal dihapus.');
        }
    }

    function openAllocation(type: 'GUP' | 'LS') {
        allocationType = type;
        allocationForm = emptyAllocationForm(type);
        allocationForm.financeSourceId = availableSources[0]?.id || '';
        allocationModalOpen = true;
    }

    function closeAllocationModal() {
        if (saving) return;
        allocationModalOpen = false;
    }

    $: selectedAllocationSource = dashboard.sources.find(
        (source) => source.id === allocationForm.financeSourceId
    ) ?? null;

    async function saveAllocation() {
        if (!allocationForm.financeSourceId) {
            toast.error('Pilih sumber dana terlebih dahulu.');
            return;
        }
        if (Number(allocationForm.amount) <= 0) {
            toast.error('Nominal alokasi harus lebih dari 0.');
            return;
        }
        if (
            selectedAllocationSource &&
            Number(allocationForm.amount) > Number(selectedAllocationSource.available)
        ) {
            toast.error('Nominal melebihi saldo sumber dana yang tersedia.');
            return;
        }

        saving = true;
        try {
            await api.createFinanceAllocation({
                ...allocationForm,
                allocationType,
                amount: Number(allocationForm.amount)
            });
            toast.success(`Dana berhasil dialokasikan ke ${allocationType}.`);
            allocationModalOpen = false;
            await loadData();
        } catch (error) {
            toast.error(error instanceof Error ? error.message : 'Alokasi dana gagal disimpan.');
        } finally {
            saving = false;
        }
    }

    async function deleteAllocation(allocation: FinanceAllocation) {
        const ok = window.confirm(
            `Hapus alokasi ${allocation.allocationType} sebesar ${formatRupiah(allocation.amount)}?`
        );
        if (!ok) return;

        try {
            await api.deleteFinanceAllocation(allocation.id);
            toast.success('Alokasi berhasil dihapus dan saldo dikembalikan.');
            await loadData();
        } catch (error) {
            toast.error(error instanceof Error ? error.message : 'Alokasi gagal dihapus.');
        }
    }
</script>

<svelte:head>
    <title>Sumber Dana | Keuangan Perjadin</title>
</svelte:head>

<main class="space-y-6 pb-20">
    <section class="overflow-hidden rounded-3xl bg-gradient-to-br from-slate-950 via-blue-950 to-indigo-800 p-6 text-white shadow-xl sm:p-8">
        <div class="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
            <div>
                <div class="mb-3 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/10 px-4 py-2 text-xs font-bold uppercase tracking-[0.18em] text-blue-100">
                    <WalletCards size={16} /> Keuangan
                </div>
                <h1 class="text-3xl font-black sm:text-4xl">Sumber Dana</h1>
                <p class="mt-3 max-w-3xl text-sm leading-6 text-slate-200 sm:text-base">
                    Kelola dana tersedia dan alokasikan saldo ke kebutuhan GUP atau LS. Setiap alokasi otomatis mengurangi saldo tersedia pada sumber dana yang dipilih.
                </p>
            </div>

            <div class="flex flex-col gap-3 sm:flex-row">
                <select
                    bind:value={year}
                    on:change={changeYear}
                    class="rounded-2xl border border-white/20 bg-white/10 px-4 py-3 text-sm font-bold text-white outline-none"
                >
                    {#each [currentYear - 1, currentYear, currentYear + 1] as optionYear}
                        <option class="text-slate-900" value={optionYear}>{optionYear}</option>
                    {/each}
                </select>
                <button
                    type="button"
                    on:click={openCreateSource}
                    class="inline-flex items-center justify-center gap-2 rounded-2xl bg-amber-400 px-5 py-3 text-sm font-black text-slate-950 hover:bg-amber-300"
                >
                    <Plus size={17} /> Tambah Sumber Dana
                </button>
            </div>
        </div>
    </section>

    <section class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
            <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Total Keuangan</p>
            <h2 class="mt-2 text-2xl font-black text-slate-900">{formatRupiah(dashboard.summary.totalFunds)}</h2>
        </article>
        <article class="rounded-3xl border border-blue-100 bg-white p-5 shadow-sm">
            <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Alokasi GUP</p>
            <h2 class="mt-2 text-2xl font-black text-blue-700">{formatRupiah(dashboard.summary.totalGup)}</h2>
        </article>
        <article class="rounded-3xl border border-violet-100 bg-white p-5 shadow-sm">
            <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Alokasi LS</p>
            <h2 class="mt-2 text-2xl font-black text-violet-700">{formatRupiah(dashboard.summary.totalLs)}</h2>
        </article>
        <article class="rounded-3xl border border-emerald-100 bg-white p-5 shadow-sm">
            <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Sisa Dana Aktual</p>
            <h2 class="mt-2 text-2xl font-black text-emerald-700">{formatRupiah(dashboard.summary.totalRemaining)}</h2>
        </article>
    </section>

    <section class="grid grid-cols-1 gap-5 lg:grid-cols-2">
        <article class="rounded-3xl border border-blue-100 bg-white p-6 shadow-sm">
            <div class="flex items-start justify-between gap-4">
                <div>
                    <span class="grid h-12 w-12 place-items-center rounded-2xl bg-blue-50 text-blue-700">
                        <Banknote size={23} />
                    </span>
                    <h2 class="mt-4 text-2xl font-black text-slate-900">GUP</h2>
                    <p class="mt-1 text-sm text-slate-500">Input dana GUP dari saldo Sumber Dana yang masih tersedia.</p>
                </div>
                <button
                    type="button"
                    on:click={() => openAllocation('GUP')}
                    disabled={availableSources.length === 0}
                    class="rounded-2xl bg-blue-600 px-4 py-2.5 text-sm font-bold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
                >
                    Input GUP
                </button>
            </div>
            <div class="mt-6 grid grid-cols-3 gap-3 rounded-2xl bg-blue-50 p-4">
                <div>
                    <p class="text-[10px] font-bold uppercase text-blue-500">Pagu</p>
                    <p class="mt-1 font-black text-blue-900">{formatRupiah(dashboard.summary.totalGup)}</p>
                </div>
                <div>
                    <p class="text-[10px] font-bold uppercase text-amber-500">Realisasi</p>
                    <p class="mt-1 font-black text-amber-700">{formatRupiah(dashboard.summary.totalUsedGup)}</p>
                </div>
                <div>
                    <p class="text-[10px] font-bold uppercase text-emerald-500">Sisa</p>
                    <p class="mt-1 font-black text-emerald-700">{formatRupiah(dashboard.summary.totalGup - dashboard.summary.totalUsedGup)}</p>
                </div>
            </div>
        </article>

        <article class="rounded-3xl border border-violet-100 bg-white p-6 shadow-sm">
            <div class="flex items-start justify-between gap-4">
                <div>
                    <span class="grid h-12 w-12 place-items-center rounded-2xl bg-violet-50 text-violet-700">
                        <CircleDollarSign size={23} />
                    </span>
                    <h2 class="mt-4 text-2xl font-black text-slate-900">LS</h2>
                    <p class="mt-1 text-sm text-slate-500">Input dana LS dari saldo Sumber Dana yang masih tersedia.</p>
                </div>
                <button
                    type="button"
                    on:click={() => openAllocation('LS')}
                    disabled={availableSources.length === 0}
                    class="rounded-2xl bg-violet-600 px-4 py-2.5 text-sm font-bold text-white hover:bg-violet-700 disabled:cursor-not-allowed disabled:opacity-50"
                >
                    Input LS
                </button>
            </div>
            <div class="mt-6 grid grid-cols-3 gap-3 rounded-2xl bg-violet-50 p-4">
                <div>
                    <p class="text-[10px] font-bold uppercase text-violet-500">Pagu</p>
                    <p class="mt-1 font-black text-violet-900">{formatRupiah(dashboard.summary.totalLs)}</p>
                </div>
                <div>
                    <p class="text-[10px] font-bold uppercase text-amber-500">Realisasi</p>
                    <p class="mt-1 font-black text-amber-700">{formatRupiah(dashboard.summary.totalUsedLs)}</p>
                </div>
                <div>
                    <p class="text-[10px] font-bold uppercase text-emerald-500">Sisa</p>
                    <p class="mt-1 font-black text-emerald-700">{formatRupiah(dashboard.summary.totalLs - dashboard.summary.totalUsedLs)}</p>
                </div>
            </div>
        </article>
    </section>

    <section class="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div class="flex items-center justify-between border-b border-slate-200 p-5 sm:p-6">
            <div>
                <h2 class="text-xl font-black text-slate-900">Sumber Dana</h2>
                <p class="mt-1 text-sm text-slate-500">Ringkasan sumber dana. Klik View untuk melihat rincian keuangan, GUP, LS, dan pengelolaan data.</p>
            </div>
        </div>

        {#if loading}
            <div class="p-10 text-center text-sm text-slate-500">Memuat sumber dana...</div>
        {:else if dashboard.sources.length === 0}
            <div class="p-10 text-center text-sm text-slate-500">Belum ada sumber dana pada tahun {year}.</div>
        {:else}
            <div class="overflow-x-auto">
                <table class="min-w-full text-sm">
                    <thead class="bg-slate-50 text-left text-xs font-bold uppercase tracking-wider text-slate-500">
                        <tr>
                            <th class="px-5 py-4">Kode</th>
                            <th class="px-5 py-4">Nama Sumber Dana</th>
                            <th class="px-5 py-4 text-right">Keuangan</th>
                            <th class="px-5 py-4">Status</th>
                            <th class="px-5 py-4 text-right">Aksi</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100">
                        {#each dashboard.sources as source (source.id)}
                            <tr class="hover:bg-slate-50/70">
                                <td class="px-5 py-4 font-black text-slate-800">{source.code}</td>
                                <td class="px-5 py-4">
                                    <div class="font-bold text-slate-800">{source.name}</div>
                                    <div class="mt-1 text-xs text-slate-400">Tahun {source.year}</div>
                                </td>
                                <td class="px-5 py-4 text-right font-semibold text-slate-800">{formatRupiah(source.amount)}</td>
                                <td class="px-5 py-4">
                                    <span class={`rounded-full px-2.5 py-1 text-xs font-bold ${source.isActive ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-500'}`}>
                                        {source.isActive ? 'Aktif' : 'Nonaktif'}
                                    </span>
                                </td>
                                <td class="px-5 py-4 text-right">
                                    <button
                                        type="button"
                                        on:click={() => openViewSource(source)}
                                        class="inline-flex items-center gap-2 rounded-xl bg-blue-50 px-3.5 py-2 text-xs font-black text-blue-700 transition hover:bg-blue-100"
                                        title="Lihat rincian sumber dana"
                                    >
                                        <Eye size={15} /> View
                                    </button>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </section>

    <section class="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div class="border-b border-slate-200 p-5 sm:p-6">
            <h2 class="text-xl font-black text-slate-900">Riwayat Alokasi GUP & LS</h2>
            <p class="mt-1 text-sm text-slate-500">Setiap input pada Card GUP atau LS tercatat di sini.</p>
        </div>

        {#if dashboard.allocations.length === 0}
            <div class="p-10 text-center text-sm text-slate-500">Belum ada alokasi dana.</div>
        {:else}
            <div class="overflow-x-auto">
                <table class="min-w-full text-sm">
                    <thead class="bg-slate-50 text-left text-xs font-bold uppercase tracking-wider text-slate-500">
                        <tr>
                            <th class="px-5 py-4">Tanggal</th>
                            <th class="px-5 py-4">Jenis</th>
                            <th class="px-5 py-4">Sumber Dana</th>
                            <th class="px-5 py-4 text-right">Nominal</th>
                            <th class="px-5 py-4">Keterangan</th>
                            <th class="px-5 py-4 text-right">Aksi</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100">
                        {#each dashboard.allocations as allocation (allocation.id)}
                            <tr>
                                <td class="px-5 py-4 text-slate-500">{formatDate(allocation.createdAt)}</td>
                                <td class="px-5 py-4">
                                    <span class={`rounded-full px-2.5 py-1 text-xs font-black ${allocation.allocationType === 'GUP' ? 'bg-blue-100 text-blue-700' : 'bg-violet-100 text-violet-700'}`}>
                                        {allocation.allocationType}
                                    </span>
                                </td>
                                <td class="px-5 py-4">
                                    <div class="font-bold text-slate-800">{allocation.sourceCode}</div>
                                    <div class="text-xs text-slate-400">{allocation.sourceName}</div>
                                </td>
                                <td class="px-5 py-4 text-right font-black text-slate-800">{formatRupiah(allocation.amount)}</td>
                                <td class="px-5 py-4 text-slate-500">{allocation.notes || '-'}</td>
                                <td class="px-5 py-4 text-right">
                                    <button type="button" on:click={() => deleteAllocation(allocation)} class="grid h-9 w-9 place-items-center rounded-xl bg-red-50 text-red-700 hover:bg-red-100" title="Hapus alokasi">
                                        <Trash2 size={16} />
                                    </button>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </section>
</main>


{#if viewModalOpen && viewingSource}
    <div class="fixed inset-0 z-[110] flex items-center justify-center bg-slate-950/60 p-4 backdrop-blur-sm" role="presentation">
        <div class="max-h-[92vh] w-full max-w-4xl overflow-y-auto rounded-3xl bg-white shadow-2xl" role="dialog" aria-modal="true" aria-label="Detail Sumber Dana">
            <div class="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white px-6 py-5">
                <div>
                    <div class="mb-2 flex flex-wrap items-center gap-2">
                        <span class="rounded-full bg-blue-50 px-3 py-1 text-xs font-black text-blue-700">{viewingSource.code}</span>
                        <span class={`rounded-full px-3 py-1 text-xs font-black ${viewingSource.isActive ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-500'}`}>
                            {viewingSource.isActive ? 'Aktif' : 'Nonaktif'}
                        </span>
                    </div>
                    <h3 class="text-2xl font-black text-slate-900">{viewingSource.name}</h3>
                    <p class="mt-1 text-sm text-slate-500">Detail lengkap Sumber Dana tahun {viewingSource.year}.</p>
                </div>
                <button type="button" on:click={closeViewSource} class="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-slate-100 text-slate-600 hover:bg-slate-200" aria-label="Tutup detail">
                    <X size={18} />
                </button>
            </div>

            <div class="space-y-6 p-6">
                <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
                    <div class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                        <p class="text-[10px] font-black uppercase tracking-wider text-slate-400">Total Keuangan</p>
                        <p class="mt-2 text-lg font-black text-slate-900">{formatRupiah(viewingSource.amount)}</p>
                    </div>
                    <div class="rounded-2xl border border-emerald-100 bg-emerald-50 p-4">
                        <p class="text-[10px] font-black uppercase tracking-wider text-emerald-500">Sisa Aktual</p>
                        <p class="mt-2 text-lg font-black text-emerald-800">{formatRupiah(viewingSource.remainingTotal)}</p>
                    </div>
                    <div class="rounded-2xl border border-cyan-100 bg-cyan-50 p-4">
                        <p class="text-[10px] font-black uppercase tracking-wider text-cyan-600">Belum Dialokasikan</p>
                        <p class="mt-2 text-lg font-black text-cyan-800">{formatRupiah(viewingSource.available)}</p>
                    </div>
                    <div class="rounded-2xl border border-slate-200 bg-white p-4">
                        <p class="text-[10px] font-black uppercase tracking-wider text-slate-400">Total Realisasi</p>
                        <p class="mt-2 text-lg font-black text-slate-800">{formatRupiah(viewingSource.usedTotal)}</p>
                    </div>
                </div>

                <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
                    <section class="rounded-2xl border border-blue-100 bg-blue-50/60 p-5">
                        <div class="mb-4 flex items-center gap-3">
                            <span class="grid h-10 w-10 place-items-center rounded-xl bg-blue-100 text-blue-700"><Banknote size={19} /></span>
                            <div>
                                <h4 class="font-black text-slate-900">Anggaran GUP</h4>
                                <p class="text-xs text-slate-500">Pagu, realisasi, dan sisa GUP.</p>
                            </div>
                        </div>
                        <dl class="space-y-3 text-sm">
                            <div class="flex items-center justify-between gap-4"><dt class="text-slate-500">Pagu GUP</dt><dd class="font-black text-blue-700">{formatRupiah(viewingSource.allocatedGup)}</dd></div>
                            <div class="flex items-center justify-between gap-4"><dt class="text-slate-500">Realisasi GUP</dt><dd class="font-black text-amber-700">{formatRupiah(viewingSource.usedGup)}</dd></div>
                            <div class="border-t border-blue-100 pt-3 flex items-center justify-between gap-4"><dt class="font-bold text-slate-700">Sisa GUP</dt><dd class="font-black text-emerald-700">{formatRupiah(viewingSource.remainingGup)}</dd></div>
                        </dl>
                    </section>

                    <section class="rounded-2xl border border-violet-100 bg-violet-50/60 p-5">
                        <div class="mb-4 flex items-center gap-3">
                            <span class="grid h-10 w-10 place-items-center rounded-xl bg-violet-100 text-violet-700"><CircleDollarSign size={19} /></span>
                            <div>
                                <h4 class="font-black text-slate-900">Anggaran LS</h4>
                                <p class="text-xs text-slate-500">Pagu, realisasi, dan sisa LS.</p>
                            </div>
                        </div>
                        <dl class="space-y-3 text-sm">
                            <div class="flex items-center justify-between gap-4"><dt class="text-slate-500">Pagu LS</dt><dd class="font-black text-violet-700">{formatRupiah(viewingSource.allocatedLs)}</dd></div>
                            <div class="flex items-center justify-between gap-4"><dt class="text-slate-500">Realisasi LS</dt><dd class="font-black text-amber-700">{formatRupiah(viewingSource.usedLs)}</dd></div>
                            <div class="border-t border-violet-100 pt-3 flex items-center justify-between gap-4"><dt class="font-bold text-slate-700">Sisa LS</dt><dd class="font-black text-emerald-700">{formatRupiah(viewingSource.remainingLs)}</dd></div>
                        </dl>
                    </section>
                </div>

                <section class="rounded-2xl border border-slate-200 p-5">
                    <h4 class="font-black text-slate-900">Informasi Sumber Dana</h4>
                    <div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
                        <div><p class="text-xs font-bold uppercase tracking-wider text-slate-400">Tahun</p><p class="mt-1 font-bold text-slate-800">{viewingSource.year}</p></div>
                        <div><p class="text-xs font-bold uppercase tracking-wider text-slate-400">Kode</p><p class="mt-1 font-bold text-slate-800">{viewingSource.code}</p></div>
                        <div class="sm:col-span-2"><p class="text-xs font-bold uppercase tracking-wider text-slate-400">Nama Sumber Dana</p><p class="mt-1 font-bold text-slate-800">{viewingSource.name}</p></div>
                        <div class="sm:col-span-2"><p class="text-xs font-bold uppercase tracking-wider text-slate-400">Keterangan</p><p class="mt-1 whitespace-pre-wrap text-sm leading-6 text-slate-600">{viewingSource.notes || '-'}</p></div>
                        <div><p class="text-xs font-bold uppercase tracking-wider text-slate-400">Dibuat</p><p class="mt-1 text-sm font-semibold text-slate-700">{formatDate(viewingSource.createdAt)}</p></div>
                        <div><p class="text-xs font-bold uppercase tracking-wider text-slate-400">Diperbarui</p><p class="mt-1 text-sm font-semibold text-slate-700">{formatDate(viewingSource.updatedAt)}</p></div>
                    </div>
                </section>
            </div>

            <div class="sticky bottom-0 flex flex-col-reverse gap-3 border-t border-slate-200 bg-white px-6 py-4 sm:flex-row sm:items-center sm:justify-between">
                <button type="button" on:click={() => deleteSource(viewingSource)} class="inline-flex items-center justify-center gap-2 rounded-2xl bg-red-50 px-4 py-2.5 text-sm font-black text-red-700 hover:bg-red-100">
                    <Trash2 size={16} /> Hapus
                </button>
                <div class="flex flex-col-reverse gap-3 sm:flex-row">
                    <button type="button" on:click={closeViewSource} class="rounded-2xl border border-slate-300 px-5 py-2.5 text-sm font-bold text-slate-700 hover:bg-slate-50">Tutup</button>
                    <button type="button" on:click={editViewedSource} class="inline-flex items-center justify-center gap-2 rounded-2xl bg-amber-400 px-5 py-2.5 text-sm font-black text-slate-950 hover:bg-amber-300">
                        <Pencil size={16} /> Edit Data
                    </button>
                </div>
            </div>
        </div>
    </div>
{/if}

{#if sourceModalOpen}
    <div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 p-4" role="presentation">
        <div class="w-full max-w-xl rounded-3xl bg-white shadow-2xl" role="dialog" aria-modal="true">
            <div class="flex items-center justify-between border-b border-slate-200 px-6 py-5">
                <div>
                    <h3 class="text-xl font-black text-slate-900">{editingSource ? 'Edit Sumber Dana' : 'Tambah Sumber Dana'}</h3>
                    <p class="mt-1 text-sm text-slate-500">Keuangan yang tersedia dapat dialokasikan ke GUP atau LS.</p>
                </div>
                <button type="button" on:click={closeSourceModal} class="grid h-10 w-10 place-items-center rounded-full bg-slate-100 text-slate-600"><X size={18} /></button>
            </div>

            <form class="space-y-4 p-6" on:submit|preventDefault={saveSource}>
                <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <label class="block">
                        <span class="mb-2 block text-sm font-bold text-slate-700">Tahun</span>
                        <input type="number" bind:value={sourceForm.year} min="2000" max="2100" class="w-full rounded-2xl border border-slate-300 px-4 py-3 outline-none focus:border-blue-500" />
                    </label>
                    <label class="block">
                        <span class="mb-2 block text-sm font-bold text-slate-700">Kode Sumber Dana</span>
                        <input bind:value={sourceForm.code} placeholder="Contoh: DIPA-01" class="w-full rounded-2xl border border-slate-300 px-4 py-3 outline-none focus:border-blue-500" />
                    </label>
                </div>

                <label class="block">
                    <span class="mb-2 block text-sm font-bold text-slate-700">Nama Sumber Dana</span>
                    <input bind:value={sourceForm.name} placeholder="Contoh: DIPA Sekretariat Jenderal" class="w-full rounded-2xl border border-slate-300 px-4 py-3 outline-none focus:border-blue-500" />
                </label>

                <label class="block">
                    <span class="mb-2 block text-sm font-bold text-slate-700">Jumlah Keuangan</span>
                    <input type="number" bind:value={sourceForm.amount} min="0" step="1" class="w-full rounded-2xl border border-slate-300 px-4 py-3 outline-none focus:border-blue-500" />
                    {#if editingSource}
                        <span class="mt-1 block text-xs text-slate-400">Minimal {formatRupiah(editingSource.allocatedTotal)} karena dana tersebut sudah dialokasikan.</span>
                    {/if}
                </label>

                <label class="block">
                    <span class="mb-2 block text-sm font-bold text-slate-700">Keterangan</span>
                    <textarea bind:value={sourceForm.notes} rows="3" class="w-full rounded-2xl border border-slate-300 px-4 py-3 outline-none focus:border-blue-500"></textarea>
                </label>

                <label class="flex items-center gap-3 rounded-2xl bg-slate-50 p-4">
                    <input type="checkbox" bind:checked={sourceForm.isActive} class="h-4 w-4" />
                    <span class="text-sm font-bold text-slate-700">Sumber dana aktif dan dapat digunakan</span>
                </label>

                <div class="flex justify-end gap-3 pt-2">
                    <button type="button" on:click={closeSourceModal} class="rounded-2xl border border-slate-300 px-5 py-3 text-sm font-bold text-slate-700">Batal</button>
                    <button type="submit" disabled={saving} class="inline-flex items-center gap-2 rounded-2xl bg-blue-600 px-5 py-3 text-sm font-bold text-white disabled:opacity-50">
                        <Save size={16} /> {saving ? 'Menyimpan...' : 'Simpan'}
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}

{#if allocationModalOpen}
    <div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 p-4" role="presentation">
        <div class="w-full max-w-xl rounded-3xl bg-white shadow-2xl" role="dialog" aria-modal="true">
            <div class="flex items-center justify-between border-b border-slate-200 px-6 py-5">
                <div>
                    <h3 class="text-xl font-black text-slate-900">Input Keuangan {allocationType}</h3>
                    <p class="mt-1 text-sm text-slate-500">Dana diambil dari saldo tersedia pada Sumber Dana.</p>
                </div>
                <button type="button" on:click={closeAllocationModal} class="grid h-10 w-10 place-items-center rounded-full bg-slate-100 text-slate-600"><X size={18} /></button>
            </div>

            <form class="space-y-4 p-6" on:submit|preventDefault={saveAllocation}>
                <label class="block">
                    <span class="mb-2 block text-sm font-bold text-slate-700">Pilih Sumber Dana</span>
                    <select bind:value={allocationForm.financeSourceId} class="w-full rounded-2xl border border-slate-300 bg-white px-4 py-3 outline-none focus:border-blue-500">
                        {#each availableSources as source}
                            <option value={source.id}>{source.code} — {source.name} ({formatRupiah(source.available)} tersedia)</option>
                        {/each}
                    </select>
                </label>

                {#if selectedAllocationSource}
                    <div class="rounded-2xl bg-emerald-50 p-4">
                        <p class="text-xs font-bold uppercase tracking-wider text-emerald-600">Saldo Tersedia</p>
                        <p class="mt-1 text-xl font-black text-emerald-800">{formatRupiah(selectedAllocationSource.available)}</p>
                    </div>
                {/if}

                <label class="block">
                    <span class="mb-2 block text-sm font-bold text-slate-700">Nominal untuk {allocationType}</span>
                    <input type="number" bind:value={allocationForm.amount} min="1" step="1" class="w-full rounded-2xl border border-slate-300 px-4 py-3 outline-none focus:border-blue-500" />
                </label>

                <label class="block">
                    <span class="mb-2 block text-sm font-bold text-slate-700">Keterangan</span>
                    <textarea bind:value={allocationForm.notes} rows="3" placeholder={`Contoh: Alokasi dana ${allocationType} tahap 1`} class="w-full rounded-2xl border border-slate-300 px-4 py-3 outline-none focus:border-blue-500"></textarea>
                </label>

                <div class="flex justify-end gap-3 pt-2">
                    <button type="button" on:click={closeAllocationModal} class="rounded-2xl border border-slate-300 px-5 py-3 text-sm font-bold text-slate-700">Batal</button>
                    <button type="submit" disabled={saving || !allocationForm.financeSourceId} class="inline-flex items-center gap-2 rounded-2xl bg-blue-600 px-5 py-3 text-sm font-bold text-white disabled:opacity-50">
                        <Save size={16} /> {saving ? 'Menyimpan...' : `Simpan ${allocationType}`}
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}
