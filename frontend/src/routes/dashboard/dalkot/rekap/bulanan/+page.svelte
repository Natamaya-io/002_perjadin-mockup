<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import {
        ArrowLeft,
        Calendar,
        CalendarCheck,
        CalendarDays,
        Filter,
        LoaderCircle,
        Printer,
        RefreshCcw,
        Users,
        WalletCards
    } from 'lucide-svelte';
    import { api } from '$lib/shared/api';
    import type {
        DalkotMonthlySPJRecapResponse,
        DalkotMonthlySPJRecapRow
    } from '$lib/shared/api/types';
    import { toast } from '$lib/shared/stores/toast';

    const monthNames = [
        'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
        'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
    ];

    const now = new Date();

    const emptyRecap = (): DalkotMonthlySPJRecapResponse => ({
        month: now.getMonth() + 1,
        year: now.getFullYear(),
        account: '026.01.WA.2158.EBA.994.002.S.524113',
        title: 'Penerimaan biaya perjalanan dinas dalam kota protokol yang telah melaksanakan persiapan kegiatan Menteri Ketenagakerjaan dan Wakil Menteri Ketenagakerjaan RI',
        rows: [],
        summary: { totalOfficers: 0, totalDays: 0, totalAmount: 0 }
    });

    let selectedMonth = now.getMonth() + 1;
    let selectedYear = now.getFullYear();
    let recap = emptyRecap();
    let loading = true;
    let loadError = '';

    $: activePeriod = `${monthNames[(recap.month || 1) - 1]} ${recap.year}`;

    onMount(async () => {
        readFiltersFromURL();
        await loadRecap();
    });

    function readFiltersFromURL() {
        const params = new URLSearchParams(window.location.search);
        const month = Number(params.get('month'));
        const year = Number(params.get('year'));
        if (month >= 1 && month <= 12) selectedMonth = month;
        if (year >= 2020 && year <= 2100) selectedYear = year;
    }

    async function loadRecap() {
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
            recap = await api.getDalkotMonthlySPJRecap({
                month: selectedMonth,
                year: selectedYear
            });
            const params = new URLSearchParams({
                month: String(selectedMonth),
                year: String(selectedYear)
            });
            window.history.replaceState({}, '', `${window.location.pathname}?${params.toString()}`);
        } catch (error) {
            loadError = error instanceof Error ? error.message : 'Rekap bulanan SPJ gagal dimuat.';
            toast.error(loadError);
        } finally {
            loading = false;
        }
    }

    async function resetFilters() {
        selectedMonth = now.getMonth() + 1;
        selectedYear = now.getFullYear();
        await loadRecap();
    }

    function formatRupiah(value: number): string {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            maximumFractionDigits: 0
        }).format(Number(value || 0));
    }

    function formatRupiahCompact(value: number): string {
        return `Rp${new Intl.NumberFormat('id-ID', { maximumFractionDigits: 0 }).format(Number(value || 0))}`;
    }

    function formatMonthlyDates(dates: string[]): string {
        const days = Array.from(new Set(dates.map((date) => Number(date.slice(8, 10))).filter(Boolean))).sort((a, b) => a - b);
        return days.length ? `${days.join(', ')} ${activePeriod}` : '-';
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

    function formatShortDate(value?: string | null): string {
        if (!value) return '-';
        const date = new Date(`${value.slice(0, 10)}T00:00:00`);
        if (Number.isNaN(date.getTime())) return value;
        return new Intl.DateTimeFormat('id-ID', {
            day: '2-digit', month: '2-digit', year: 'numeric'
        }).format(date);
    }

    function escapeHtml(value: unknown): string {
        return String(value ?? '')
            .replaceAll('&', '&amp;')
            .replaceAll('<', '&lt;')
            .replaceAll('>', '&gt;')
            .replaceAll('"', '&quot;')
            .replaceAll("'", '&#039;');
    }

    function printOfficialFormat() {
        const printWindow = window.open('', '_blank', 'width=1300,height=850');
        if (!printWindow) {
            toast.error('Pop-up cetak diblokir oleh browser.');
            return;
        }

        const rowsHtml = recap.rows.length
            ? recap.rows.map((row, index) => printRow(row, index)).join('')
            : `<tr class="empty"><td colspan="14">Tidak ada data petugas SPJ untuk periode ${escapeHtml(activePeriod)}.</td></tr>`;

        printWindow.document.open();
        printWindow.document.write(`<!doctype html>
<html lang="id">
<head>
    <meta charset="utf-8">
    <title>Rekap Bulanan SPJ Dalkot - ${escapeHtml(activePeriod)}</title>
    <style>
        @page { size: A4 landscape; margin: 8mm 7mm; }
        * { box-sizing: border-box; }
        body { margin: 0; background: #fff; color: #000; font-family: Arial, Helvetica, sans-serif; font-size: 9px; }
        .sheet { width: 100%; margin: 0 auto; }
        .heading { display: grid; grid-template-columns: 18% 64% 18%; align-items: start; min-height: 42px; margin-bottom: 3px; }
        .period { padding-top: 1px; font-size: 9px; font-weight: 400; text-align: center; }
        .title { padding: 0 6px; font-size: 10px; font-weight: 700; line-height: 1.25; text-align: center; }
        .account { margin: 2px 0 1px; font-size: 10px; font-weight: 700; }
        table { width: 100%; border-collapse: collapse; table-layout: fixed; }
        th, td { padding: 4px 3px; border: 1px solid #000; vertical-align: middle; }
        th { font-weight: 700; line-height: 1.15; text-align: center; }
        tbody td { height: 38px; }
        .c-no { width: 3.2%; } .c-name { width: 18.5%; } .c-gol { width: 5%; } .c-date { width: 17%; }
        .c-tarif { width: 8%; } .c-symbol { width: 3%; } .c-count { width: 3.5%; } .c-unit { width: 4.5%; }
        .c-equal { width: 2.5%; } .c-subtotal { width: 7%; } .c-received { width: 7%; } .c-sign { width: 12%; }
        .c-st { width: 5.5%; } .c-st-date { width: 5.5%; }
        .center { text-align: center; } .right { text-align: right; white-space: nowrap; }
        .name, .dates { line-height: 1.25; text-align: center; } .signature { padding: 4px 6px; }
        .signature span { display: block; } .signature .even { text-align: right; }
        .small { font-size: 8px; line-height: 1.2; word-break: break-word; }
        .total td { height: auto; min-height: 24px; font-weight: 700; }
        .empty td { height: 50px; text-align: center; }
        @media print { body { -webkit-print-color-adjust: exact; print-color-adjust: exact; } }
    </style>
</head>
<body onload="window.print()">
<main class="sheet">
    <div class="heading"><div class="period">${escapeHtml(activePeriod)}</div><div class="title">Daftar: ${escapeHtml(recap.title)}</div><div></div></div>
    <div class="account">Akun: ${escapeHtml(recap.account)}</div>
    <table>
        <colgroup><col class="c-no"><col class="c-name"><col class="c-gol"><col class="c-date"><col class="c-tarif"><col class="c-symbol"><col class="c-count"><col class="c-unit"><col class="c-equal"><col class="c-subtotal"><col class="c-received"><col class="c-sign"><col class="c-st"><col class="c-st-date"></colgroup>
        <thead><tr><th rowspan="2">NO.</th><th rowspan="2">NAMA</th><th rowspan="2">GOL</th><th rowspan="2">TANGGAL</th><th colspan="6">TRANSPORT DALAM KOTA</th><th rowspan="2">JUMLAH YANG<br>DITERIMA</th><th rowspan="2">TANDA TANGAN</th><th rowspan="2">NO ST</th><th rowspan="2">TGL ST</th></tr><tr><th></th><th></th><th></th><th></th><th></th><th></th></tr></thead>
        <tbody>${rowsHtml}${recap.rows.length ? `<tr class="total"><td colspan="6" class="right">TOTAL</td><td class="center">${recap.summary.totalDays}</td><td colspan="2"></td><td class="right">${escapeHtml(formatRupiahCompact(recap.summary.totalAmount))}</td><td class="right">${escapeHtml(formatRupiahCompact(recap.summary.totalAmount))}</td><td colspan="3"></td></tr>` : ''}</tbody>
    </table>
</main>
</body>
</html>`);
        printWindow.document.close();
        printWindow.focus();
    }

    function printRow(row: DalkotMonthlySPJRecapRow, index: number): string {
        const letters = row.assignmentLetterNumbers.length
            ? row.assignmentLetterNumbers.map(escapeHtml).join('<br>')
            : '-';
        const letterDates = row.assignmentLetterDates.length
            ? row.assignmentLetterDates.map((date) => escapeHtml(formatShortDate(date))).join('<br>')
            : '-';
        return `<tr>
            <td class="center">${index + 1}</td>
            <td class="name">${escapeHtml(row.name)}</td>
            <td class="center">${escapeHtml(row.grade || '-')}</td>
            <td class="dates">${escapeHtml(formatMonthlyDates(row.dates))}</td>
            <td class="right">${escapeHtml(formatRupiahCompact(row.rate))}</td>
            <td class="center">x</td>
            <td class="center">${row.totalDays}</td>
            <td class="center">Hari</td>
            <td class="center">=</td>
            <td class="right">${escapeHtml(formatRupiahCompact(row.totalAmount))}</td>
            <td class="right">${escapeHtml(formatRupiahCompact(row.totalAmount))}</td>
            <td class="signature"><span class="${(index + 1) % 2 === 0 ? 'even' : ''}">${index + 1}</span></td>
            <td class="center small">${letters}</td>
            <td class="center small">${letterDates}</td>
        </tr>`;
    }
</script>

<svelte:head>
    <title>Rekap Bulanan SPJ Dalkot | Perjadin</title>
</svelte:head>

<main class="dalkot-page space-y-6 pb-20">
    <section class="dalkot-rekap-hero">
        <div>
            <span class="dalkot-eyebrow"><CalendarDays size={14} /> Rekap SPJ per anggota</span>
            <h1>Rekap Bulanan <span>Dalkot</span></h1>
            <p>Data petugas SPJ dikelompokkan per nama dan per bulan, kemudian jumlah hari dihitung dengan tarif transport dalam kota.</p>
        </div>
        <div class="no-print flex flex-col gap-3 sm:flex-row">
            <button class="dalkot-btn dalkot-btn--outline" type="button" on:click={() => goto(`/dashboard/dalkot/rekap?month=${recap.month}&year=${recap.year}`)}>
                <ArrowLeft size={16} /> Rekap Riil
            </button>
            <button class="dalkot-btn dalkot-btn--gold" type="button" on:click={printOfficialFormat}>
                <Printer size={16} /> Cetak Format Resmi
            </button>
        </div>
    </section>

    <section class="dalkot-panel">
        <div class="dalkot-panel__header">
            <div>
                <p class="dalkot-panel__kicker">Filter periode</p>
                <h2>Pilih Bulan Rekap</h2>
                <p>Satu petugas hanya tampil satu kali; seluruh tanggal penugasannya digabung dalam baris yang sama.</p>
            </div>
        </div>
        <form class="dalkot-filter-grid dalkot-filter-grid--monthly" on:submit|preventDefault={loadRecap}>
            <div class="dalkot-filter-card">
                <div class="dalkot-filter-card__icon"><CalendarDays size={19} /></div>
                <div class="dalkot-filter-card__field">
                    <label class="dalkot-label" for="bulanan-bulan">Bulan</label>
                    <select class="dalkot-input" id="bulanan-bulan" bind:value={selectedMonth} disabled={loading}>
                        {#each monthNames as month, index}<option value={index + 1}>{month}</option>{/each}
                    </select>
                </div>
            </div>
            <div class="dalkot-filter-card">
                <div class="dalkot-filter-card__icon"><Calendar size={19} /></div>
                <div class="dalkot-filter-card__field">
                    <label class="dalkot-label" for="bulanan-tahun">Tahun</label>
                    <input class="dalkot-input" id="bulanan-tahun" type="number" min="2020" max="2100" bind:value={selectedYear} disabled={loading} />
                </div>
            </div>
            <div class="dalkot-filter-actions">
                <button class="dalkot-btn dalkot-btn--primary justify-center" type="submit" disabled={loading}>
                    {#if loading}<LoaderCircle class="animate-spin" size={15} />{:else}<Filter size={15} />{/if} Tampilkan
                </button>
                <button class="dalkot-btn dalkot-btn--ghost justify-center" type="button" on:click={resetFilters} disabled={loading}>
                    <RefreshCcw size={15} /> Reset
                </button>
            </div>
        </form>
    </section>

    {#if loadError}<section class="dalkot-error" role="alert">{loadError}</section>{/if}

    <section class="grid grid-cols-1 gap-5 md:grid-cols-3">
        <div class="dalkot-stat-card"><div class="dalkot-stat-card__icon dalkot-stat-card__icon--blue"><Users size={20} /></div><div><p>Jumlah Petugas</p><h2>{recap.summary.totalOfficers}</h2><span>Petugas SPJ pada periode aktif</span></div></div>
        <div class="dalkot-stat-card"><div class="dalkot-stat-card__icon dalkot-stat-card__icon--indigo"><CalendarCheck size={20} /></div><div><p>Total Hari</p><h2>{recap.summary.totalDays}</h2><span>Akumulasi hari penugasan</span></div></div>
        <div class="dalkot-stat-card"><div class="dalkot-stat-card__icon dalkot-stat-card__icon--green"><WalletCards size={20} /></div><div><p>Total Pembayaran</p><h2>{formatRupiah(recap.summary.totalAmount)}</h2><span>Tarif SPJ {formatRupiah(170000)} per hari</span></div></div>
    </section>

    <section class="dalkot-panel overflow-hidden">
        <div class="dalkot-panel__header">
            <div>
                <p class="dalkot-panel__kicker">Pratinjau format</p>
                <h2>Daftar Penerimaan Biaya Perjalanan Dinas</h2>
                <p>Rekap hanya menggunakan daftar <strong>Petugas SPJ</strong> pada pengajuan berjenis SPJ RILL.</p>
            </div>
        </div>
        <div class="px-5 pb-6 sm:px-6">
            <div class="monthly-report-heading">
                <div class="monthly-report-heading__period">{activePeriod}</div>
                <div class="monthly-report-heading__title">Daftar: {recap.title}</div>
                <div></div>
            </div>
            <div class="monthly-report-account">Akun: {recap.account}</div>
            <div class="overflow-x-auto">
                <table class="monthly-spj-table min-w-[1450px]">
                    <thead>
                        <tr><th rowspan="2">NO.</th><th rowspan="2">NAMA</th><th rowspan="2">GOL</th><th rowspan="2">TANGGAL</th><th colspan="6">TRANSPORT DALAM KOTA</th><th rowspan="2">JUMLAH YANG<br />DITERIMA</th><th rowspan="2">TANDA TANGAN</th><th rowspan="2">NO ST</th><th rowspan="2">TGL ST</th></tr>
                        <tr><th>Tarif</th><th></th><th>Jumlah</th><th>Satuan</th><th></th><th>Total</th></tr>
                    </thead>
                    <tbody>
                        {#if loading}
                            <tr><td colspan="14" class="dalkot-empty"><LoaderCircle class="mx-auto mb-3 animate-spin" size={25} />Memuat rekap bulanan SPJ...</td></tr>
                        {:else if recap.rows.length === 0}
                            <tr><td colspan="14" class="dalkot-empty">Belum ada petugas SPJ untuk periode {activePeriod}.</td></tr>
                        {:else}
                            {#each recap.rows as row, index}
                                <tr>
                                    <td class="num">{index + 1}</td>
                                    <td class="name">{row.name}</td>
                                    <td class="num">{row.grade || '-'}</td>
                                    <td class="dates">{formatMonthlyDates(row.dates)}</td>
                                    <td class="money">{formatRupiahCompact(row.rate)}</td>
                                    <td class="num">x</td>
                                    <td class="num">{row.totalDays}</td>
                                    <td class="num">Hari</td>
                                    <td class="num">=</td>
                                    <td class="money">{formatRupiahCompact(row.totalAmount)}</td>
                                    <td class="money">{formatRupiahCompact(row.totalAmount)}</td>
                                    <td class="signature-cell"><span class:even={(index + 1) % 2 === 0} class="signature-no">{index + 1}</span></td>
                                    <td class="num">{#if row.assignmentLetterNumbers.length}{#each row.assignmentLetterNumbers as letter}{letter}<br />{/each}{:else}-{/if}</td>
                                    <td class="num">{#if row.assignmentLetterDates.length}{#each row.assignmentLetterDates as date}{formatDate(date)}<br />{/each}{:else}-{/if}</td>
                                </tr>
                            {/each}
                            <tr class="total-row"><td colspan="6" class="text-right">TOTAL</td><td class="num">{recap.summary.totalDays}</td><td colspan="2"></td><td class="money">{formatRupiahCompact(recap.summary.totalAmount)}</td><td class="money">{formatRupiahCompact(recap.summary.totalAmount)}</td><td colspan="3"></td></tr>
                        {/if}
                    </tbody>
                </table>
            </div>
        </div>
    </section>
</main>

<style>
    .dalkot-page { --gov-navy: #0d3854; --gov-blue: #1f78a8; --gov-blue-dark: #155d83; --gov-gold: #d99d19; --gov-gold-dark: #b67a07; }
    .dalkot-rekap-hero { position: relative; display: flex; align-items: center; justify-content: space-between; gap: 20px; overflow: hidden; padding: 30px; color: #fff; border-radius: 24px; background: radial-gradient(circle at 86% 6%,rgba(247,200,95,.20),rgba(247,200,95,0) 14rem),linear-gradient(120deg,#0d3854 0%,#164d70 55%,#1f78a8 100%); box-shadow: 0 18px 40px rgba(13,56,84,.20); }
    .dalkot-rekap-hero::before { position: absolute; top: -150px; right: -75px; width: 280px; height: 280px; content: ''; border: 44px solid rgba(255,255,255,.06); border-radius: 50%; }
    .dalkot-rekap-hero > div { position: relative; z-index: 1; }
    .dalkot-eyebrow { display: inline-flex; align-items: center; gap: 8px; color: #dbeafe; font-size: 11px; font-weight: 800; letter-spacing: .1em; text-transform: uppercase; }
    .dalkot-rekap-hero h1 { margin: 10px 0 7px; font-size: clamp(26px,3vw,36px); font-weight: 800; line-height: 1.1; }
    .dalkot-rekap-hero h1 span { color: #f6c344; }
    .dalkot-rekap-hero p { max-width: 760px; margin: 0; color: #dceafc; font-size: 14px; line-height: 1.65; }
    .dalkot-panel { border: 1px solid #dbe5ef; border-radius: 20px; background: rgba(255,255,255,.96); box-shadow: 0 8px 22px rgba(15,41,76,.07); }
    .dalkot-panel__header { display: flex; align-items: center; justify-content: space-between; gap: 18px; padding: 24px; border-bottom: 1px solid #e5edf5; }
    .dalkot-panel__kicker { margin: 0 0 6px; color: var(--gov-blue); font-size: 11px; font-weight: 800; letter-spacing: .1em; text-transform: uppercase; }
    .dalkot-panel__header h2 { margin: 0; color: var(--gov-navy); font-size: 20px; font-weight: 800; }
    .dalkot-panel__header p:not(.dalkot-panel__kicker) { margin: 6px 0 0; color: #64748b; font-size: 13px; }
    .dalkot-btn { display: inline-flex; align-items: center; gap: 9px; padding: 11px 16px; border: 0; border-radius: 11px; font-size: 13px; font-weight: 800; line-height: 1; white-space: nowrap; transition: transform .2s ease,background .2s ease; }
    .dalkot-btn:not(:disabled):hover { transform: translateY(-1px); } .dalkot-btn:disabled { cursor: not-allowed; opacity: .6; }
    .dalkot-btn--primary { color: #fff; background: linear-gradient(135deg,var(--gov-blue),var(--gov-blue-dark)); box-shadow: 0 8px 14px rgba(31,92,168,.18); }
    .dalkot-btn--gold { color: #4a3100; background: linear-gradient(135deg,#f6c344,var(--gov-gold)); box-shadow: 0 8px 14px rgba(233,167,0,.18); }
    .dalkot-btn--outline { color: #fff; border: 1px solid rgba(255,255,255,.45); background: rgba(255,255,255,.09); }
    .dalkot-btn--ghost { color: #475569; background: #f1f5f9; }
    .dalkot-filter-grid { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)) auto; align-items: stretch; gap: 14px; padding: 20px 24px 24px; }
    .dalkot-filter-grid--monthly { grid-template-columns: repeat(2,minmax(0,1fr)) auto; }
    .dalkot-filter-card { display: flex; align-items: center; gap: 12px; min-width: 0; padding: 14px; border: 1px solid #dbe5ef; border-radius: 14px; background: linear-gradient(180deg,#fbfdff 0%,#f5f9fd 100%); }
    .dalkot-filter-card__icon { display: grid; flex: 0 0 auto; width: 40px; height: 40px; place-items: center; color: var(--gov-blue); border-radius: 11px; background: #e3f0ff; }
    .dalkot-filter-card__field { flex: 1; min-width: 0; }
    .dalkot-label { display: block; margin-bottom: 5px; color: #34465d; font-size: 12px; font-weight: 800; }
    .dalkot-input { width: 100%; height: 40px; padding: 8px 11px; color: #1e293b; border: 1px solid #cbd8e6; border-radius: 10px; outline: none; background: #fff; font-size: 13px; }
    .dalkot-input:focus { border-color: var(--gov-blue); box-shadow: 0 0 0 3px rgba(31,92,168,.12); }
    .dalkot-filter-actions { display: grid; grid-template-columns: 1fr; gap: 8px; min-width: 132px; }
    .dalkot-stat-card { display: flex; align-items: flex-start; gap: 14px; min-height: 132px; padding: 20px; border: 1px solid #dbe5ef; border-radius: 18px; background: #fff; box-shadow: 0 7px 20px rgba(15,41,76,.055); }
    .dalkot-stat-card__icon { display: grid; flex: 0 0 auto; width: 45px; height: 45px; place-items: center; border-radius: 14px; }
    .dalkot-stat-card__icon--blue { color: #1f5ca8; background: #e1efff; } .dalkot-stat-card__icon--indigo { color: #5348b9; background: #e9e8ff; } .dalkot-stat-card__icon--green { color: #0b8a64; background: #dcf6ea; }
    .dalkot-stat-card p { margin: 0; color: #64748b; font-size: 12px; } .dalkot-stat-card h2 { margin: 4px 0; color: var(--gov-navy); font-size: 20px; font-weight: 800; line-height: 1.25; } .dalkot-stat-card span { color: #94a3b8; font-size: 11px; }
    .monthly-report-heading { display: grid; grid-template-columns: minmax(130px,.45fr) minmax(420px,1.7fr) minmax(130px,.45fr); align-items: start; gap: 18px; padding: 0 2px 14px; }
    .monthly-report-heading__period { font-weight: 700; text-align: center; } .monthly-report-heading__title { font-weight: 800; line-height: 1.35; text-align: center; } .monthly-report-account { padding: 0 2px 8px; font-size: 13px; font-weight: 800; }
    .monthly-spj-table { width: 100%; border-collapse: collapse; color: #111827; font-size: 12px; }
    .monthly-spj-table th,.monthly-spj-table td { padding: 9px 7px; border: 1px solid #64748b; vertical-align: middle; }
    .monthly-spj-table thead th { color: #0f172a; background: #f8fafc; font-size: 11px; line-height: 1.2; text-align: center; }
    .monthly-spj-table tbody td { background: #fff; } .monthly-spj-table .num { text-align: center; white-space: nowrap; } .monthly-spj-table .money { text-align: right; white-space: nowrap; }
    .monthly-spj-table .dates { min-width: 220px; line-height: 1.45; text-align: center; } .monthly-spj-table .name { min-width: 180px; font-weight: 600; text-align: center; } .monthly-spj-table .signature-cell { min-width: 120px; }
    .signature-no { display: block; font-weight: 600; } .signature-no.even { text-align: right; } .total-row td { font-weight: 800; background: #f1f5f9; }
    .dalkot-empty { padding: 42px 20px !important; color: #64748b !important; text-align: center; }
    .dalkot-error { padding: 14px 16px; color: #b42318; border: 1px solid #fecdca; border-radius: 12px; background: #fff4f2; font-size: 13px; }
    @media (max-width: 1100px) { .dalkot-filter-grid,.dalkot-filter-grid--monthly { grid-template-columns: repeat(2,minmax(0,1fr)); } .dalkot-filter-actions { grid-template-columns: repeat(2,minmax(0,1fr)); } }
    @media (max-width: 900px) { .monthly-report-heading { grid-template-columns: 1fr; } .monthly-report-heading__period { text-align: left; } }
    @media (max-width: 768px) { .dalkot-rekap-hero,.dalkot-panel__header { align-items: flex-start; flex-direction: column; } .dalkot-filter-grid,.dalkot-filter-grid--monthly { grid-template-columns: 1fr; padding: 18px; } .dalkot-filter-actions { grid-template-columns: 1fr; } }
    @media (max-width: 480px) { .dalkot-btn { width: 100%; justify-content: center; } .dalkot-rekap-hero { padding: 24px 20px; } }
</style>
