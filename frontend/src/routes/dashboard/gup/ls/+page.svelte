<script lang="ts">
    import { onMount } from 'svelte';
    import {
        ArrowDownRight,
        Banknote,
        CalendarCheck2,
        CalendarDays,
        Filter,
        Info,
        Pencil,
        PiggyBank,
        Plus,
        Save,
        Wallet,
        X
    } from 'lucide-svelte';
    import { api } from '$lib/shared/api';
    import { toast } from '$lib/shared/stores/toast';

    type LsRow = {
        fundingSourceId: string;
        year: number;
        monthNumber: number;
        monthName: string;
        gupLabel: string;
        accountCode: string;
        amount: number;
        realization: number;
        remaining: number;
        transactionCount: number;
    };

    type LsSummary = {
        totalAmount: number;
        totalRealization: number;
        totalRemaining: number;
        filledMonths: number;
        totalMonths: number;
    };

    type LsResponse = {
        year: number;
        rows: LsRow[];
        summary: LsSummary;
    };

    type LsStatus = {
        label: string;
        badge: string;
    };

    const currentYear = new Date().getFullYear();

    let rows: LsRow[] = [];
    let loading = true;
    let saving = false;
    let loadError = '';
    let monthFilter = 'all';
    let modalOpen = false;
    let selectedFundingSourceId = '';
    let amountInput = 0;

    $: filteredRows = monthFilter === 'all'
        ? rows
        : rows.filter((row) => row.fundingSourceId === monthFilter);

    $: totalAmount = filteredRows.reduce(
        (total, row) => total + Number(row.amount || 0),
        0
    );

    $: totalRealization = filteredRows.reduce(
        (total, row) => total + Number(row.realization || 0),
        0
    );

    $: totalRemaining = totalAmount - totalRealization;

    $: filledMonths = filteredRows.filter(
        (row) => Number(row.amount || 0) > 0
    ).length;

    $: selectedRow = rows.find(
        (row) => row.fundingSourceId === selectedFundingSourceId
    ) ?? null;

    $: modalRealization = Number(selectedRow?.realization || 0);
    $: modalRemaining = Number(amountInput || 0) - modalRealization;

    onMount(loadLs);

    async function loadLs() {
        loading = true;
        loadError = '';

        try {
            const result = await (api as any).getGupLs(currentYear) as LsResponse;
            rows = Array.isArray(result?.rows) ? result.rows : [];
        } catch (error) {
            loadError = error instanceof Error
                ? error.message
                : 'Data LS GUP gagal dimuat.';
            toast.error(loadError);
        } finally {
            loading = false;
        }
    }

    function formatRupiah(value: number): string {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            maximumFractionDigits: 0
        }).format(Number(value || 0));
    }

    function formatNumber(value: number): string {
        return new Intl.NumberFormat('id-ID').format(Number(value || 0));
    }

    function parseCurrency(value: string): number {
        return Number(value.replace(/\D/g, '')) || 0;
    }

    function getStatus(row: LsRow): LsStatus {
        if (Number(row.amount || 0) <= 0) {
            return {
                label: 'Belum diinput',
                badge: 'bg-slate-100 text-slate-600'
            };
        }

        if (Number(row.remaining || 0) < 0) {
            return {
                label: 'Melebihi LS',
                badge: 'bg-red-100 text-red-700'
            };
        }

        return {
            label: 'Dana tersedia',
            badge: 'bg-emerald-100 text-emerald-700'
        };
    }

    function openModal(fundingSourceId = '') {
        const currentMonth = new Date().getMonth() + 1;
        const defaultRow = rows.find(
            (row) => row.monthNumber === currentMonth
        ) ?? rows[0] ?? null;

        selectedFundingSourceId = fundingSourceId || defaultRow?.fundingSourceId || '';
        const row = rows.find(
            (item) => item.fundingSourceId === selectedFundingSourceId
        );
        amountInput = Number(row?.amount || 0);
        modalOpen = true;
    }

    function syncAmountInput() {
        amountInput = Number(selectedRow?.amount || 0);
    }

    function closeModal() {
        if (saving) return;
        modalOpen = false;
    }

    async function saveLs() {
        if (!selectedFundingSourceId) {
            toast.error('Bulan LS wajib dipilih.');
            return;
        }

        if (amountInput < 0) {
            toast.error('Nilai LS tidak boleh negatif.');
            return;
        }

        saving = true;
        try {
            const result = await (api as any).updateGupLs({
                year: currentYear,
                fundingSourceId: selectedFundingSourceId,
                amount: amountInput
            }) as LsResponse;

            rows = Array.isArray(result?.rows) ? result.rows : rows;
            modalOpen = false;
            toast.success('LS bulanan berhasil disimpan.');
        } catch (error) {
            toast.error(
                error instanceof Error
                    ? error.message
                    : 'LS bulanan gagal disimpan.'
            );
        } finally {
            saving = false;
        }
    }
</script>

<svelte:head>
    <title>LS Bulanan | Perjadin</title>
</svelte:head>

<main class="space-y-6 pb-20">
    <section class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-slate-950 via-sky-950 to-sky-800 p-6 text-white shadow-xl sm:p-8">
        <div class="absolute -right-24 -top-24 h-64 w-64 rounded-full bg-sky-300/20 blur-3xl"></div>
        <div class="absolute -bottom-28 left-20 h-72 w-72 rounded-full bg-amber-300/10 blur-3xl"></div>

        <div class="relative z-10 flex flex-col gap-6 xl:flex-row xl:items-end xl:justify-between">
            <div class="max-w-3xl">
                <div class="mb-4 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/10 px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-sky-100">
                    <Wallet size={16} />
                    Sumber Dana GUP
                </div>
                <h1 class="text-3xl font-black tracking-tight text-white sm:text-4xl">
                    LS Bulanan
                </h1>
                <p class="mt-3 max-w-2xl text-sm leading-6 text-slate-200 sm:text-base">
                    LS merupakan pemasukan yang menjadi sumber dana GUP setiap bulan. Januari mendanai GUP 1, Februari mendanai GUP 2, dan seterusnya hingga GUP 12.
                </p>
            </div>

            <button
                type="button"
                on:click={() => openModal()}
                class="inline-flex items-center justify-center gap-2 rounded-2xl bg-amber-400 px-5 py-3 text-sm font-black text-slate-950 shadow-lg transition hover:bg-amber-300"
            >
                <Plus size={17} />
                Input LS Bulanan
            </button>
        </div>
    </section>

    {#if loadError}
        <section class="rounded-2xl border border-red-200 bg-red-50 p-5 text-sm text-red-700">
            {loadError}
            <button type="button" on:click={loadLs} class="ml-2 font-bold underline">
                Muat ulang
            </button>
        </section>
    {/if}

    <section class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-3xl border border-sky-100 bg-white p-5 shadow-sm">
            <div class="flex items-center justify-between gap-3">
                <div>
                    <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Total LS</p>
                    <h2 class="mt-2 text-2xl font-black text-slate-900">{formatRupiah(totalAmount)}</h2>
                </div>
                <span class="grid h-12 w-12 place-items-center rounded-2xl bg-sky-50 text-sky-700">
                    <Banknote size={22} />
                </span>
            </div>
        </article>

        <article class="rounded-3xl border border-emerald-100 bg-white p-5 shadow-sm">
            <div class="flex items-center justify-between gap-3">
                <div>
                    <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Realisasi GUP</p>
                    <h2 class="mt-2 text-2xl font-black text-emerald-700">{formatRupiah(totalRealization)}</h2>
                </div>
                <span class="grid h-12 w-12 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
                    <ArrowDownRight size={22} />
                </span>
            </div>
        </article>

        <article class="rounded-3xl border border-amber-100 bg-white p-5 shadow-sm">
            <div class="flex items-center justify-between gap-3">
                <div>
                    <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Sisa Dana</p>
                    <h2
                        class:text-red-700={totalRemaining < 0}
                        class="mt-2 text-2xl font-black text-amber-700"
                    >
                        {formatRupiah(totalRemaining)}
                    </h2>
                </div>
                <span class="grid h-12 w-12 place-items-center rounded-2xl bg-amber-50 text-amber-700">
                    <PiggyBank size={22} />
                </span>
            </div>
        </article>

        <article class="rounded-3xl border border-violet-100 bg-white p-5 shadow-sm">
            <div class="flex items-center justify-between gap-3">
                <div>
                    <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Bulan Terisi</p>
                    <h2 class="mt-2 text-2xl font-black text-violet-700">
                        {filledMonths} / {filteredRows.length}
                    </h2>
                </div>
                <span class="grid h-12 w-12 place-items-center rounded-2xl bg-violet-50 text-violet-700">
                    <CalendarCheck2 size={22} />
                </span>
            </div>
        </article>
    </section>

    <section class="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div class="border-b border-slate-200 p-5 sm:p-6">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
                <div>
                    <p class="text-xs font-bold uppercase tracking-[0.2em] text-sky-600">Pemetaan Bulanan</p>
                    <h2 class="mt-2 text-xl font-black text-slate-900">LS → GUP</h2>
                    <p class="mt-1 text-sm text-slate-500">
                        Kode akun LS: <strong>S.522141</strong>. Realisasi dihitung dari transaksi Pengajuan GUP berdasarkan sumber dana GUP yang sesuai.
                    </p>
                </div>

                <div class="w-full lg:w-72">
                    <label for="lsMonthFilter" class="mb-2 block text-xs font-bold uppercase tracking-wider text-slate-500">
                        Filter LS per Bulan
                    </label>
                    <div class="relative">
                        <CalendarDays size={17} class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-sky-700" />
                        <select
                            id="lsMonthFilter"
                            bind:value={monthFilter}
                            class="w-full appearance-none rounded-2xl border border-slate-300 bg-white py-3 pl-11 pr-10 text-sm font-bold text-slate-800 outline-none transition focus:border-sky-600 focus:ring-2 focus:ring-sky-100"
                        >
                            <option value="all">Semua Bulan</option>
                            {#each rows as row}
                                <option value={row.fundingSourceId}>
                                    {row.monthName} — {row.gupLabel}
                                </option>
                            {/each}
                        </select>
                    </div>
                </div>
            </div>

            <div class="mt-4 inline-flex items-center gap-2 rounded-full bg-sky-50 px-3 py-1.5 text-xs font-bold text-sky-700">
                <Filter size={14} />
                {#if monthFilter === 'all'}
                    Menampilkan semua bulan
                {:else}
                    Menampilkan LS {filteredRows[0]?.monthName || '-'} — {filteredRows[0]?.gupLabel || '-'}
                {/if}
            </div>
        </div>

        {#if loading}
            <div class="flex min-h-72 items-center justify-center p-6 text-sm text-slate-500">
                Memuat data LS bulanan...
            </div>
        {:else if filteredRows.length === 0}
            <div class="flex min-h-72 items-center justify-center p-6 text-sm text-slate-500">
                Data LS tidak ditemukan.
            </div>
        {:else}
            <div class="grid grid-cols-1 gap-4 p-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 sm:p-6">
                {#each filteredRows as row (row.fundingSourceId)}
                    {@const status = getStatus(row)}
                    <article class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
                        <div class="flex items-start justify-between gap-3">
                            <div>
                                <p class="text-xs font-bold uppercase tracking-wide text-slate-400">{row.monthName}</p>
                                <h3 class="mt-1 text-lg font-black text-slate-900">{row.gupLabel}</h3>
                            </div>
                            <button
                                type="button"
                                on:click={() => openModal(row.fundingSourceId)}
                                class="grid h-9 w-9 place-items-center rounded-xl bg-sky-50 text-sky-700 hover:bg-sky-100"
                                title={`Input LS ${row.monthName}`}
                                aria-label={`Input LS ${row.monthName}`}
                            >
                                <Pencil size={16} />
                            </button>
                        </div>

                        <div class="mt-4 space-y-2 text-sm">
                            <div class="flex justify-between gap-3">
                                <span class="text-slate-500">LS / Pemasukan</span>
                                <strong>{formatRupiah(row.amount)}</strong>
                            </div>
                            <div class="flex justify-between gap-3">
                                <span class="text-slate-500">Realisasi</span>
                                <strong class="text-emerald-700">{formatRupiah(row.realization)}</strong>
                            </div>
                            <div class="flex justify-between gap-3">
                                <span class="text-slate-500">Sisa Dana</span>
                                <strong class:text-red-700={row.remaining < 0} class="text-amber-700">
                                    {formatRupiah(row.remaining)}
                                </strong>
                            </div>
                        </div>

                        <div class="mt-4 flex items-center justify-between gap-2">
                            <span class={`rounded-full px-2.5 py-1 text-[11px] font-bold ${status.badge}`}>
                                {status.label}
                            </span>
                            <span class="text-[11px] font-semibold text-slate-400">{row.accountCode}</span>
                        </div>
                    </article>
                {/each}
            </div>
        {/if}
    </section>
</main>

{#if modalOpen}
    <div
        class="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 p-4"
        role="presentation"
        on:click={(event) => {
            if (event.currentTarget === event.target) closeModal();
        }}
    >
        <div class="max-h-[92vh] w-full max-w-2xl overflow-y-auto rounded-3xl bg-white shadow-2xl" role="dialog" aria-modal="true" aria-labelledby="lsModalTitle">
            <div class="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white px-6 py-5">
                <div>
                    <h3 id="lsModalTitle" class="text-xl font-black text-slate-900">Input LS Bulanan</h3>
                    <p class="mt-1 text-sm text-slate-500">Input pemasukan LS yang menjadi sumber dana GUP bulan terkait.</p>
                </div>
                <button
                    type="button"
                    on:click={closeModal}
                    disabled={saving}
                    class="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-slate-100 text-slate-600 hover:bg-slate-200 disabled:opacity-50"
                    aria-label="Tutup form LS"
                >
                    <X size={19} />
                </button>
            </div>

            <form class="space-y-5 p-6" on:submit|preventDefault={saveLs}>
                <div class="grid grid-cols-1 gap-5 md:grid-cols-2">
                    <div>
                        <label class="mb-2 block text-sm font-bold text-slate-700" for="modalLsBulan">Bulan</label>
                        <select
                            id="modalLsBulan"
                            bind:value={selectedFundingSourceId}
                            on:change={syncAmountInput}
                            class="w-full rounded-2xl border border-slate-300 px-4 py-3 font-semibold text-slate-900 focus:outline-none focus:ring-2 focus:ring-sky-700"
                        >
                            {#each rows as row}
                                <option value={row.fundingSourceId}>{row.monthName}</option>
                            {/each}
                        </select>
                    </div>

                    <div>
                        <label class="mb-2 block text-sm font-bold text-slate-700" for="modalLsGup">Sumber Dana GUP</label>
                        <input
                            id="modalLsGup"
                            readonly
                            value={selectedRow?.gupLabel || '-'}
                            class="w-full rounded-2xl border border-sky-200 bg-sky-50 px-4 py-3 font-black text-sky-800"
                        />
                    </div>

                    <div>
                        <label class="mb-2 block text-sm font-bold text-slate-700" for="modalLsAccount">Kode Akun LS</label>
                        <input
                            id="modalLsAccount"
                            value="S.522141"
                            readonly
                            class="w-full rounded-2xl border border-slate-200 bg-slate-100 px-4 py-3 font-semibold text-slate-600"
                        />
                    </div>

                    <div>
                        <label class="mb-2 block text-sm font-bold text-slate-700" for="modalLsNilai">Nilai LS / Pemasukan</label>
                        <input
                            id="modalLsNilai"
                            required
                            inputmode="numeric"
                            value={formatNumber(amountInput)}
                            on:input={(event) => {
                                amountInput = parseCurrency(event.currentTarget.value);
                            }}
                            class="w-full rounded-2xl border border-slate-300 px-4 py-3 text-lg font-black text-slate-900 focus:outline-none focus:ring-2 focus:ring-sky-700"
                            placeholder="Contoh: 50000000"
                        />
                    </div>

                    <div>
                        <label class="mb-2 block text-sm font-bold text-slate-700" for="modalLsRealisasi">Realisasi GUP</label>
                        <input
                            id="modalLsRealisasi"
                            readonly
                            value={formatRupiah(modalRealization)}
                            class="w-full rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 font-bold text-emerald-700"
                        />
                    </div>

                    <div>
                        <label class="mb-2 block text-sm font-bold text-slate-700" for="modalLsSisa">Sisa Sumber Dana</label>
                        <input
                            id="modalLsSisa"
                            readonly
                            value={formatRupiah(modalRemaining)}
                            class:text-red-700={modalRemaining < 0}
                            class="w-full rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 font-bold text-amber-700"
                        />
                    </div>
                </div>

                <div class="rounded-2xl border border-sky-100 bg-sky-50 p-4 text-sm leading-6 text-sky-800">
                    <Info size={17} class="mr-2 inline" />
                    Contoh: LS Januari otomatis menjadi sumber dana <strong>GUP 1</strong>. Realisasi transaksi dengan sumber dana GUP 1 akan mengurangi saldo LS Januari.
                </div>

                <div class="flex flex-col-reverse gap-3 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
                    <button
                        type="button"
                        on:click={closeModal}
                        disabled={saving}
                        class="rounded-2xl bg-slate-100 px-5 py-3 text-sm font-bold text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                    >
                        Batal
                    </button>
                    <button
                        type="submit"
                        disabled={saving}
                        class="inline-flex items-center justify-center gap-2 rounded-2xl bg-sky-700 px-5 py-3 text-sm font-bold text-white hover:bg-sky-800 disabled:opacity-50"
                    >
                        <Save size={17} />
                        {saving ? 'Menyimpan...' : 'Simpan LS'}
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}
