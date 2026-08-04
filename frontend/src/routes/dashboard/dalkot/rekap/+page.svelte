<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import {
        ArrowLeft,
        Calendar,
        CalendarDays,
        FileSpreadsheet,
        Filter,
        LoaderCircle,
        Printer,
        Receipt,
        RefreshCcw,
        Scale,
        TriangleAlert,
        UserRound,
        WalletCards
    } from 'lucide-svelte';
    import { api } from '$lib/shared/api';
    import type {
        DalkotMasterData,
        DalkotRecapResponse,
        DalkotRecapRow,
        DalkotRecapSummary
    } from '$lib/shared/api/types';
    import { toast } from '$lib/shared/stores/toast';

    const monthNames = [
        'Januari',
        'Februari',
        'Maret',
        'April',
        'Mei',
        'Juni',
        'Juli',
        'Agustus',
        'September',
        'Oktober',
        'November',
        'Desember'
    ];

    const now = new Date();

    const emptySummary = (): DalkotRecapSummary => ({
        totalRows: 0,
        totalNominal: 0,
        totalSPJAmount: 0,
        totalRiilAmount: 0,
        balance: 0,
        balanceMinus: false
    });

    const emptyRecap = (): DalkotRecapResponse => ({
        month: now.getMonth() + 1,
        year: now.getFullYear(),
        officerName: '',
        rows: [],
        summary: emptySummary()
    });

    let master: DalkotMasterData | null = null;
    let recap = emptyRecap();
    let selectedMonth = now.getMonth() + 1;
    let selectedYear = now.getFullYear();
    let selectedOfficerId = '';
    let loading = true;
    let loadError = '';

    $: selectedOfficerLabel = recap.officerName || 'Seluruh Petugas Dalkot';
    $: activePeriod = `${monthNames[(recap.month || 1) - 1]} ${recap.year}`;
    $: balanceDisplay = recap.summary.balanceMinus
        ? `- ${formatRupiah(Math.abs(recap.summary.balance))}`
        : formatRupiah(recap.summary.balance);

    onMount(async () => {
        readFiltersFromURL();
        await loadInitialData();
    });

    function readFiltersFromURL() {
        const params = new URLSearchParams(window.location.search);
        const month = Number(params.get('month'));
        const year = Number(params.get('year'));
        if (month >= 1 && month <= 12) selectedMonth = month;
        if (year >= 2020 && year <= 2100) selectedYear = year;
        selectedOfficerId = params.get('officerId') ?? '';
    }

    async function loadInitialData() {
        loading = true;
        loadError = '';
        try {
            const [masterResult, recapResult] = await Promise.all([
                api.getDalkotMaster(),
                api.getDalkotRecap({
                    month: selectedMonth,
                    year: selectedYear,
                    officerId: selectedOfficerId || undefined
                })
            ]);
            master = masterResult;
            recap = recapResult;
        } catch (error) {
            handleLoadError(error);
        } finally {
            loading = false;
        }
    }

    async function applyFilters() {
        if (selectedMonth < 1 || selectedMonth > 12) {
            toast.error('Bulan rekap tidak valid.');
            return;
        }
        if (selectedYear < 2020 || selectedYear > 2100) {
            toast.error('Tahun rekap harus berada antara 2020 dan 2100.');
            return;
        }

        loading = true;
        loadError = '';
        try {
            recap = await api.getDalkotRecap({
                month: selectedMonth,
                year: selectedYear,
                officerId: selectedOfficerId || undefined
            });
            updateURL();
        } catch (error) {
            handleLoadError(error);
        } finally {
            loading = false;
        }
    }

    async function resetFilters() {
        selectedMonth = now.getMonth() + 1;
        selectedYear = now.getFullYear();
        selectedOfficerId = '';
        await applyFilters();
    }

    function updateURL() {
        const params = new URLSearchParams({
            month: String(selectedMonth),
            year: String(selectedYear)
        });
        if (selectedOfficerId) params.set('officerId', selectedOfficerId);
        window.history.replaceState({}, '', `${window.location.pathname}?${params.toString()}`);
    }

    function handleLoadError(error: unknown) {
        loadError =
            error instanceof Error
                ? error.message
                : 'Rekapitulasi Dalkot gagal dimuat.';
        toast.error(loadError);
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

    function formatDalkotType(value: string): string {
        return value === 'SPJ_RILL' ? 'SPJ RILL' : 'Kebijakan Protokol';
    }

    function escapeHtml(value: unknown): string {
        return String(value ?? '')
            .replaceAll('&', '&amp;')
            .replaceAll('<', '&lt;')
            .replaceAll('>', '&gt;')
            .replaceAll('"', '&quot;')
            .replaceAll("'", '&#039;');
    }

    function openMonthlyRecap() {
        goto(`/dashboard/dalkot/rekap/bulanan?month=${recap.month}&year=${recap.year}`);
    }

    function printRecap() {
        const printWindow = window.open('', '_blank', 'width=1180,height=820');
        if (!printWindow) {
            toast.error('Pop-up cetak diblokir oleh browser.');
            return;
        }

        const rowsHtml = recap.rows.length
            ? recap.rows.map(printRow).join('')
            : '<tr><td colspan="6" class="empty">Tidak ada data untuk filter yang dipilih.</td></tr>';

        printWindow.document.open();
        printWindow.document.write(`<!doctype html>
<html lang="id">
<head>
    <meta charset="utf-8">
    <title>Tanda Terima Pembayaran Dalkot</title>
    <style>
        @page { size: A4 landscape; margin: 12mm; }
        * { box-sizing: border-box; }
        body { margin: 0; color: #111; background: #fff; font-family: Arial, Helvetica, sans-serif; font-size: 10px; }
        .paper { width: 100%; margin: 0 auto; }
        .title h1 { margin: 0 0 14px; font-size: 16px; text-align: left; }
        .receipt { width: 48%; margin-bottom: 15px; padding: 9px 12px; border: 1px solid #000; }
        .receipt p { margin: 3px 0; }
        .receipt span { display: inline-block; width: 90px; }
        table { width: 100%; border-collapse: collapse; table-layout: fixed; }
        th, td { border: 1px solid #000; padding: 7px 6px; vertical-align: middle; }
        th { text-align: center; font-size: 9px; }
        td:last-child { text-align: right; white-space: nowrap; }
        .total td { font-weight: 700; }
        .total td:first-child { text-align: right; }
        .empty { padding: 28px; text-align: center !important; }
        .signature { display: flex; justify-content: flex-end; margin-top: 22px; text-align: center; }
        .signature > div { width: 240px; }
        .signature-space { height: 58px; }
        @media print { body { -webkit-print-color-adjust: exact; print-color-adjust: exact; } }
    </style>
</head>
<body onload="window.print()">
    <main class="paper">
        <div class="title"><h1>TANDA TERIMA PEMBAYARAN DALAM KOTA PROTOKOL</h1></div>
        <div class="receipt">
            <p><span>Nama Petugas</span>: ${escapeHtml(selectedOfficerLabel)}</p>
            <p><span>Periode</span>: ${escapeHtml(activePeriod)}</p>
            <p><span>Tanggal Cetak</span>: ${escapeHtml(formatDate(new Date().toISOString().slice(0, 10)))}</p>
        </div>
        <table>
            <thead><tr><th>Tanggal Pelaksanaan</th><th>Pejabat yang Didampingi</th><th>Kegiatan</th><th>Nama Petugas</th><th>Kategori Dalkot</th><th>Nominal</th></tr></thead>
            <tbody>
                ${rowsHtml}
                <tr class="total"><td colspan="5">TOTAL</td><td>${escapeHtml(formatRupiah(recap.summary.totalNominal))}</td></tr>
            </tbody>
        </table>
        <section class="signature"><div><div class="signature-space"></div><strong contenteditable="true">____________________________</strong><br><small contenteditable="true">Petugas Dalkot</small></div></section>
    </main>
</body>
</html>`);
        printWindow.document.close();
        printWindow.focus();
    }

    function printRow(row: DalkotRecapRow): string {
        return `<tr>
            <td>${escapeHtml(formatDate(row.serviceDate))}</td>
            <td>${escapeHtml(row.accompaniedOfficial)}</td>
            <td>${escapeHtml(row.activityName)}</td>
            <td>${escapeHtml(row.officerName)}</td>
            <td>${escapeHtml(row.category)}</td>
            <td>${escapeHtml(formatRupiah(row.nominal))}</td>
        </tr>`;
    }
</script>

<svelte:head>
    <title>Rekapitulasi Dalkot | Perjadin</title>
</svelte:head>

<main class="dalkot-page space-y-6 pb-20">
    <section class="dalkot-rekap-hero">
        <div>
            <span class="dalkot-eyebrow"><Receipt size={14} /> Rekap pembayaran</span>
            <h1>Tanda Terima Pembayaran <span>Dalkot</span></h1>
            <p>Filter periode dan nama petugas, tinjau rincian biaya, lalu cetak rekapitulasi pembayaran.</p>
        </div>
        <button class="dalkot-btn dalkot-btn--outline" type="button" on:click={() => goto('/dashboard/dalkot')}>
            <ArrowLeft size={16} /> Kembali ke Daftar
        </button>
    </section>

    <section class="dalkot-panel">
        <div class="dalkot-panel__header">
            <div>
                <p class="dalkot-panel__kicker">Filter data bulanan</p>
                <h2>Filter Rekapitulasi per Bulan</h2>
                <p>Nilai SPJ, pembayaran riil, balance, dan rincian pembayaran mengikuti bulan serta petugas yang dipilih.</p>
            </div>
        </div>

        <form class="dalkot-filter-grid" on:submit|preventDefault={applyFilters}>
            <div class="dalkot-filter-card">
                <div class="dalkot-filter-card__icon"><CalendarDays size={19} /></div>
                <div class="dalkot-filter-card__field">
                    <label class="dalkot-label" for="rekap-bulan">Bulan</label>
                    <select class="dalkot-input" id="rekap-bulan" bind:value={selectedMonth} disabled={loading}>
                        {#each monthNames as month, index}
                            <option value={index + 1}>{month}</option>
                        {/each}
                    </select>
                </div>
            </div>

            <div class="dalkot-filter-card">
                <div class="dalkot-filter-card__icon"><Calendar size={19} /></div>
                <div class="dalkot-filter-card__field">
                    <label class="dalkot-label" for="rekap-tahun">Tahun</label>
                    <input class="dalkot-input" id="rekap-tahun" type="number" min="2020" max="2100" bind:value={selectedYear} disabled={loading} />
                </div>
            </div>

            <div class="dalkot-filter-card">
                <div class="dalkot-filter-card__icon"><UserRound size={19} /></div>
                <div class="dalkot-filter-card__field">
                    <label class="dalkot-label" for="rekap-petugas">Nama Petugas</label>
                    <select class="dalkot-input" id="rekap-petugas" bind:value={selectedOfficerId} disabled={loading || !master}>
                        <option value="">Seluruh Petugas</option>
                        {#each master?.users ?? [] as officer}
                            <option value={officer.id}>{officer.name}</option>
                        {/each}
                    </select>
                </div>
            </div>

            <div class="dalkot-filter-actions">
                <button class="dalkot-btn dalkot-btn--primary justify-center" type="submit" disabled={loading}>
                    {#if loading}<LoaderCircle class="animate-spin" size={15} />{:else}<Filter size={15} />{/if}
                    Tampilkan
                </button>
                <button class="dalkot-btn dalkot-btn--ghost justify-center" type="button" on:click={resetFilters} disabled={loading}>
                    <RefreshCcw size={15} /> Reset
                </button>
            </div>
        </form>
    </section>

    {#if loadError}
        <section class="dalkot-error" role="alert">{loadError}</section>
    {/if}

    <section class="grid grid-cols-1 gap-5 md:grid-cols-3" aria-label="Ringkasan keuangan Dalkot">
        <div class="dalkot-stat-card">
            <div class="dalkot-stat-card__icon dalkot-stat-card__icon--blue"><FileSpreadsheet size={20} /></div>
            <div>
                <p>Total Biaya SPJ</p>
                <h2>{formatRupiah(recap.summary.totalSPJAmount)}</h2>
                <span>{activePeriod} · {selectedOfficerLabel}</span>
            </div>
        </div>

        <div class="dalkot-stat-card">
            <div class="dalkot-stat-card__icon dalkot-stat-card__icon--indigo"><WalletCards size={20} /></div>
            <div>
                <p>Total Dibayarkan (Riil)</p>
                <h2>{formatRupiah(recap.summary.totalRiilAmount)}</h2>
                <span>{recap.summary.totalRows} rincian pembayaran riil</span>
            </div>
        </div>

        <div class:dalkot-stat-card--danger={recap.summary.balanceMinus} class:dalkot-stat-card--success={!recap.summary.balanceMinus} class="dalkot-stat-card">
            <div class:dalkot-stat-card__icon--red={recap.summary.balanceMinus} class:dalkot-stat-card__icon--green={!recap.summary.balanceMinus} class="dalkot-stat-card__icon">
                {#if recap.summary.balanceMinus}<TriangleAlert size={20} />{:else}<Scale size={20} />{/if}
            </div>
            <div>
                <p>Balance (SPJ - Riil)</p>
                <h2>{balanceDisplay}</h2>
                <span>{recap.summary.balanceMinus ? 'Pengeluaran riil melebihi biaya SPJ' : 'Biaya SPJ masih mencukupi'}</span>
            </div>
        </div>
    </section>

    {#if recap.summary.balanceMinus}
        <section class="dalkot-balance-alert" role="alert" aria-live="polite">
            <div class="dalkot-balance-alert__icon"><TriangleAlert size={19} /></div>
            <div>
                <strong>Peringatan: balance Dalkot minus {formatRupiah(Math.abs(recap.summary.balance))}.</strong>
                <p>Total pembayaran riil pada {activePeriod} lebih besar daripada total biaya SPJ untuk {selectedOfficerLabel}.</p>
            </div>
        </section>
    {/if}

    <section class="dalkot-panel overflow-hidden">
        <div class="dalkot-panel__header">
            <div>
                <p class="dalkot-panel__kicker">Tanda terima pembayaran</p>
                <h2>Rincian Biaya Dalkot</h2>
                <p>Periode {activePeriod} · {selectedOfficerLabel}.</p>
            </div>
            <div class="no-print flex flex-col gap-3 sm:flex-row">
                <button class="dalkot-btn dalkot-btn--ghost" type="button" on:click={openMonthlyRecap}>
                    <CalendarDays size={16} /> Rekap Bulanan SPJ
                </button>
                <button class="dalkot-btn dalkot-btn--gold" type="button" on:click={printRecap}>
                    <Printer size={16} /> Cetak Rekap Riil
                </button>
            </div>
        </div>

        <div class="px-6 pb-4">
            <div class="dalkot-receipt-meta">
                <div><span>Nama Petugas</span><strong>{selectedOfficerLabel}</strong></div>
                <div><span>Periode</span><strong>{activePeriod}</strong></div>
                <div><span>Tanggal Cetak</span><strong>{formatDate(new Date().toISOString().slice(0, 10))}</strong></div>
            </div>
        </div>

        <div class="overflow-x-auto">
            <table class="dalkot-table min-w-[1000px]">
                <thead>
                    <tr>
                        <th>Tanggal Pelaksanaan</th>
                        <th>Pejabat yang Didampingi</th>
                        <th>Kegiatan</th>
                        <th>Nama Petugas</th>
                        <th>Kategori Dalkot</th>
                        <th>Jenis Dalkot</th>
                        <th class="text-right">Nominal</th>
                    </tr>
                </thead>
                <tbody>
                    {#if loading}
                        <tr><td colspan="7" class="dalkot-empty"><LoaderCircle class="mx-auto mb-3 animate-spin" size={25} />Memuat rekapitulasi Dalkot...</td></tr>
                    {:else if recap.rows.length === 0}
                        <tr><td colspan="7" class="dalkot-empty">Belum ada data pembayaran Dalkot untuk filter yang dipilih.</td></tr>
                    {:else}
                        {#each recap.rows as row}
                            <tr>
                                <td class="font-semibold whitespace-nowrap">{formatDate(row.serviceDate)}</td>
                                <td>{row.accompaniedOfficial}</td>
                                <td class="font-semibold text-slate-900">{row.activityName}</td>
                                <td>{row.officerName}</td>
                                <td><span class="dalkot-category-badge">{row.category}</span></td>
                                <td>{formatDalkotType(row.dalkotType)}</td>
                                <td class="text-right font-bold whitespace-nowrap">{formatRupiah(row.nominal)}</td>
                            </tr>
                        {/each}
                        <tr class="dalkot-table__total">
                            <td colspan="6" class="text-right">TOTAL</td>
                            <td class="text-right">{formatRupiah(recap.summary.totalNominal)}</td>
                        </tr>
                    {/if}
                </tbody>
            </table>
        </div>
    </section>
</main>

<style>
    .dalkot-page { --gov-navy: #0d3854; --gov-blue: #1f78a8; --gov-blue-dark: #155d83; --gov-gold: #d99d19; --gov-gold-dark: #b67a07; }
    .dalkot-rekap-hero { position: relative; display: flex; align-items: center; justify-content: space-between; gap: 20px; overflow: hidden; padding: 30px; color: #fff; border-radius: 24px; background: radial-gradient(circle at 86% 6%, rgba(247,200,95,.20), rgba(247,200,95,0) 14rem), linear-gradient(120deg, #0d3854 0%, #164d70 55%, #1f78a8 100%); box-shadow: 0 18px 40px rgba(13,56,84,.20); }
    .dalkot-rekap-hero::before { position: absolute; top: -150px; right: -75px; width: 280px; height: 280px; content: ''; border: 44px solid rgba(255,255,255,.06); border-radius: 50%; }
    .dalkot-rekap-hero > div, .dalkot-rekap-hero > button { position: relative; z-index: 1; }
    .dalkot-eyebrow { display: inline-flex; align-items: center; gap: 8px; color: #dbeafe; font-size: 11px; font-weight: 800; letter-spacing: .1em; text-transform: uppercase; }
    .dalkot-rekap-hero h1 { margin: 10px 0 7px; font-size: clamp(26px, 3vw, 36px); font-weight: 800; line-height: 1.1; }
    .dalkot-rekap-hero h1 span { color: #f6c344; }
    .dalkot-rekap-hero p { max-width: 760px; margin: 0; color: #dceafc; font-size: 14px; line-height: 1.65; }
    .dalkot-panel { border: 1px solid #dbe5ef; border-radius: 20px; background: rgba(255,255,255,.96); box-shadow: 0 8px 22px rgba(15,41,76,.07); }
    .dalkot-panel__header { display: flex; align-items: center; justify-content: space-between; gap: 18px; padding: 24px; border-bottom: 1px solid #e5edf5; }
    .dalkot-panel__kicker { margin: 0 0 6px; color: var(--gov-blue); font-size: 11px; font-weight: 800; letter-spacing: .1em; text-transform: uppercase; }
    .dalkot-panel__header h2 { margin: 0; color: var(--gov-navy); font-size: 20px; font-weight: 800; }
    .dalkot-panel__header p:not(.dalkot-panel__kicker) { margin: 6px 0 0; color: #64748b; font-size: 13px; }
    .dalkot-btn { display: inline-flex; align-items: center; gap: 9px; padding: 11px 16px; border: 0; border-radius: 11px; font-size: 13px; font-weight: 800; line-height: 1; white-space: nowrap; transition: transform .2s ease, box-shadow .2s ease, background .2s ease; }
    .dalkot-btn:not(:disabled):hover { transform: translateY(-1px); }
    .dalkot-btn:disabled { cursor: not-allowed; opacity: .6; }
    .dalkot-btn--primary { color: #fff; background: linear-gradient(135deg, var(--gov-blue), var(--gov-blue-dark)); box-shadow: 0 8px 14px rgba(31,92,168,.18); }
    .dalkot-btn--gold { color: #4a3100; background: linear-gradient(135deg, #f6c344, var(--gov-gold)); box-shadow: 0 8px 14px rgba(233,167,0,.18); }
    .dalkot-btn--outline { color: #fff; border: 1px solid rgba(255,255,255,.45); background: rgba(255,255,255,.09); }
    .dalkot-btn--ghost { color: #475569; background: #f1f5f9; }
    .dalkot-filter-grid { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)) auto; align-items: stretch; gap: 14px; padding: 20px 24px 24px; }
    .dalkot-filter-card { display: flex; align-items: center; gap: 12px; min-width: 0; padding: 14px; border: 1px solid #dbe5ef; border-radius: 14px; background: linear-gradient(180deg,#fbfdff 0%,#f5f9fd 100%); }
    .dalkot-filter-card__icon { display: grid; flex: 0 0 auto; width: 40px; height: 40px; place-items: center; color: var(--gov-blue); border-radius: 11px; background: #e3f0ff; }
    .dalkot-filter-card__field { flex: 1; min-width: 0; }
    .dalkot-label { display: block; margin-bottom: 5px; color: #34465d; font-size: 12px; font-weight: 800; }
    .dalkot-input { width: 100%; height: 40px; padding: 8px 11px; color: #1e293b; border: 1px solid #cbd8e6; border-radius: 10px; outline: none; background: #fff; font-size: 13px; }
    .dalkot-input:focus { border-color: var(--gov-blue); box-shadow: 0 0 0 3px rgba(31,92,168,.12); }
    .dalkot-filter-actions { display: grid; grid-template-columns: 1fr; gap: 8px; min-width: 132px; }
    .dalkot-stat-card { display: flex; align-items: flex-start; gap: 14px; min-height: 132px; padding: 20px; border: 1px solid #dbe5ef; border-radius: 18px; background: #fff; box-shadow: 0 7px 20px rgba(15,41,76,.055); transition: transform .18s ease, box-shadow .18s ease; }
    .dalkot-stat-card:hover { transform: translateY(-2px); box-shadow: 0 15px 30px rgba(15,41,76,.1); }
    .dalkot-stat-card__icon { display: grid; flex: 0 0 auto; width: 45px; height: 45px; place-items: center; border-radius: 14px; }
    .dalkot-stat-card__icon--blue { color: #1f5ca8; background: #e1efff; }
    .dalkot-stat-card__icon--indigo { color: #5348b9; background: #e9e8ff; }
    .dalkot-stat-card__icon--green { color: #0b8a64; background: #dcf6ea; }
    .dalkot-stat-card__icon--red { color: #b91c1c; background: #fee2e2; }
    .dalkot-stat-card p { margin: 0; color: #64748b; font-size: 12px; }
    .dalkot-stat-card h2 { margin: 4px 0; color: var(--gov-navy); font-size: 20px; font-weight: 800; line-height: 1.25; }
    .dalkot-stat-card span { color: #94a3b8; font-size: 11px; }
    .dalkot-stat-card--danger { border-color: #fecaca; background: linear-gradient(180deg,#fff 0%,#fff7f7 100%); box-shadow: 0 8px 22px rgba(185,28,28,.08); }
    .dalkot-stat-card--danger h2 { color: #b91c1c; }
    .dalkot-stat-card--success { border-color: #bbf7d0; background: linear-gradient(180deg,#fff 0%,#f6fff9 100%); }
    .dalkot-balance-alert { display: flex; align-items: flex-start; gap: 14px; padding: 16px 18px; color: #7f1d1d; border: 1px solid #fecaca; border-left: 5px solid #dc2626; border-radius: 14px; background: #fff7f7; box-shadow: 0 7px 18px rgba(185,28,28,.06); }
    .dalkot-balance-alert__icon { display: grid; flex: 0 0 auto; width: 38px; height: 38px; place-items: center; color: #dc2626; border-radius: 11px; background: #fee2e2; }
    .dalkot-balance-alert strong { display: block; font-size: 14px; line-height: 1.45; }
    .dalkot-balance-alert p { margin: 4px 0 0; color: #991b1b; font-size: 12px; line-height: 1.55; }
    .dalkot-receipt-meta { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); padding: 13px 15px; border: 1px solid #dbe5ef; border-radius: 12px; background: #f9fbfd; }
    .dalkot-receipt-meta div { padding: 0 13px; border-right: 1px solid #e2e8f0; }
    .dalkot-receipt-meta div:first-child { padding-left: 0; }
    .dalkot-receipt-meta div:last-child { padding-right: 0; border-right: 0; }
    .dalkot-receipt-meta span { display: block; color: #64748b; font-size: 10px; }
    .dalkot-receipt-meta strong { display: block; margin-top: 3px; color: var(--gov-navy); font-size: 12px; }
    .dalkot-table { width: 100%; border-collapse: collapse; font-size: 12px; }
    .dalkot-table thead th { padding: 14px 15px; color: #fff; background: linear-gradient(100deg,#0d3854,#164d70); font-size: 11px; letter-spacing: .01em; text-align: left; white-space: nowrap; }
    .dalkot-table td { padding: 14px 15px; color: #475569; border-bottom: 1px solid #e6edf4; vertical-align: middle; }
    .dalkot-table tbody tr:not(.dalkot-table__total):hover td { background: #f8fbff; }
    .dalkot-table__total td { color: var(--gov-navy); background: #eaf5f4; font-weight: 800; }
    .dalkot-category-badge { display: inline-flex; padding: 5px 8px; color: #43556b; border: 1px solid #dbe5ef; border-radius: 999px; background: #f3f6fa; font-size: 11px; font-weight: 800; white-space: nowrap; }
    .dalkot-empty { padding: 42px 20px !important; color: #64748b !important; text-align: center; }
    .dalkot-error { padding: 14px 16px; color: #b42318; border: 1px solid #fecdca; border-radius: 12px; background: #fff4f2; font-size: 13px; }
    @media (max-width: 1100px) { .dalkot-filter-grid { grid-template-columns: repeat(2,minmax(0,1fr)); } .dalkot-filter-actions { grid-template-columns: repeat(2,minmax(0,1fr)); } }
    @media (max-width: 768px) { .dalkot-rekap-hero, .dalkot-panel__header { align-items: flex-start; flex-direction: column; } .dalkot-filter-grid { grid-template-columns: 1fr; padding: 18px; } .dalkot-filter-actions { grid-template-columns: 1fr; } .dalkot-receipt-meta { grid-template-columns: 1fr; } .dalkot-receipt-meta div { padding: 10px 0; border-right: 0; border-bottom: 1px solid #e2e8f0; } .dalkot-receipt-meta div:last-child { border-bottom: 0; } }
    @media (max-width: 480px) { .dalkot-btn { width: 100%; justify-content: center; } .dalkot-rekap-hero { padding: 24px 20px; } }
</style>
