<script lang="ts">
    import { onMount } from 'svelte';
    import {
        FileText,
        Info,
        Pencil,
        Plus,
        Printer,
        Search,
        X
    } from 'lucide-svelte';
    import { api } from '$lib/shared/api';
    import { toast } from '$lib/shared/stores/toast';

    type ReportRow = {
        procurementTypeId: string;
        procurementTypeName: string;
        accountCode: string;
        mak: string;
        transactionCount: number;
        realization: number;
        submissionValue: number;
        budget: number;
        remaining: number;
        absorption: number;
    };

    type ReportSummary = {
        totalBudget: number;
        totalRealization: number;
        totalRemaining: number;
        totalAbsorption: number;
        totalTransactions: number;
        totalSubmissionValue: number;
        totalProcurementType: number;
    };

    type ReportResponse = {
        year: number;
        rows: ReportRow[];
        summary: ReportSummary;
    };

    type PieSlice = {
        label: string;
        value: number;
        percentage: number;
        displayPercentage: number;
        offset: number;
        color: string;
    };

    const palette = [
        '#0ea5e9',
        '#10b981',
        '#f59e0b',
        '#8b5cf6',
        '#ef4444',
        '#14b8a6',
        '#6366f1',
        '#f97316',
        '#22c55e',
        '#ec4899',
        '#64748b',
        '#84cc16'
    ];

    const emptySummary = (): ReportSummary => ({
        totalBudget: 0,
        totalRealization: 0,
        totalRemaining: 0,
        totalAbsorption: 0,
        totalTransactions: 0,
        totalSubmissionValue: 0,
        totalProcurementType: 0
    });

    const currentYear = new Date().getFullYear();

    let reportRows: ReportRow[] = [];
    let summary: ReportSummary = emptySummary();
    let loading = true;
    let saving = false;
    let loadError = '';
    let searchQuery = '';
    let budgetModalOpen = false;
    let selectedProcurementTypeId = '';
    let budgetInput = 0;
    let hoveredOverviewSlice: PieSlice | null = null;
    let hoveredCompositionSlice: PieSlice | null = null;
    let hoveredCardSlice: {
        chartId: string;
        slice: PieSlice;
    } | null = null;

    $: filteredRows = reportRows.filter((row) => {
        const keyword = searchQuery.trim().toLowerCase();
        if (!keyword) return true;
        return [
            row.procurementTypeName,
            row.accountCode,
            row.mak
        ].some((value) => value.toLowerCase().includes(keyword));
    });

    $: activeCompositionRows = reportRows.filter(
        (row) => Number(row.realization || 0) > 0
    );

    $: totalBudgetSlices = getBudgetSlices(
        summary.totalBudget,
        summary.totalRealization
    );

    $: compositionSlices = getCompositionSlices(
        activeCompositionRows
    );

    $: selectedBudgetRow = reportRows.find(
        (row) => row.procurementTypeId === selectedProcurementTypeId
    ) ?? null;

    $: modalRealization = Number(selectedBudgetRow?.realization || 0);
    $: modalRemaining = Number(budgetInput || 0) - modalRealization;
    $: modalAbsorption = Number(budgetInput || 0) > 0
        ? modalRealization / Number(budgetInput || 0) * 100
        : 0;

    onMount(loadReport);

    async function loadReport() {
        loading = true;
        loadError = '';

        try {
            const result = await (api as any).getGupReport(currentYear) as ReportResponse;
            reportRows = Array.isArray(result?.rows) ? result.rows : [];
            summary = result?.summary ?? emptySummary();
        } catch (error) {
            loadError = error instanceof Error
                ? error.message
                : 'Laporan GUP gagal dimuat.';
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

    function formatPercent(value: number): string {
        return `${new Intl.NumberFormat('id-ID', {
            minimumFractionDigits: 0,
            maximumFractionDigits: 2
        }).format(Number(value || 0))}%`;
    }

    function clamp(value: number, minimum: number, maximum: number): number {
        return Math.min(Math.max(value, minimum), maximum);
    }

    function createPieSlices(
    items: Array<{
        label: string;
        value: number;
        color: string;
    }>
): PieSlice[] {
    const validItems = items
        .map((item) => ({
            ...item,
            value: Math.max(Number(item.value || 0), 0)
        }))
        .filter((item) => item.value > 0);

    const total = validItems.reduce(
        (sum, item) => sum + item.value,
        0
    );

    if (total <= 0) {
        return [];
    }

    let offset = 0;

    return validItems.map((item) => {
        const percentage =
            (item.value / total) * 100;

        const slice: PieSlice = {
            label: item.label,
            value: item.value,
            percentage,
            displayPercentage: percentage,
            offset,
            color: item.color
        };

        offset += percentage;

        return slice;
    });
}

    type PiePoint = {
        x: number;
        y: number;
    };

function polarToCartesian(
    centerX: number,
    centerY: number,
    radius: number,
    angleInDegrees: number
): PiePoint {
    const angleInRadians =
        (angleInDegrees * Math.PI) / 180;

    return {
        x:
            centerX +
            radius * Math.cos(angleInRadians),
        y:
            centerY +
            radius * Math.sin(angleInRadians)
    };
}

function createPieSectorPath(
    offset: number,
    percentage: number,
    radius = 48
): string {
    const centerX = 50;
    const centerY = 50;

    const safePercentage = Math.min(
        Math.max(percentage, 0),
        100
    );

    if (safePercentage <= 0) {
        return '';
    }

    const startAngle =
        offset * 3.6 - 90;

    const endAngle =
        (offset + safePercentage) * 3.6 - 90;

    const startPoint = polarToCartesian(
        centerX,
        centerY,
        radius,
        startAngle
    );

    const endPoint = polarToCartesian(
        centerX,
        centerY,
        radius,
        endAngle
    );

    const largeArcFlag =
        safePercentage > 50 ? 1 : 0;

    return [
        `M ${centerX} ${centerY}`,
        `L ${startPoint.x} ${startPoint.y}`,
        `A ${radius} ${radius} 0 ${largeArcFlag} 1 ${endPoint.x} ${endPoint.y}`,
        'Z'
    ].join(' ');
}

    function getBudgetSlices(
        budgetValue: number,
        realizationValue: number
    ): PieSlice[] {
        const budget = Math.max(Number(budgetValue || 0), 0);
        const realization = Math.max(Number(realizationValue || 0), 0);

        if (budget <= 0 && realization <= 0) {
            return [{
                label: 'Anggaran belum diinput',
                value: 0,
                percentage: 100,
                displayPercentage: 0,
                offset: 0,
                color: '#cbd5e1'
            }];
        }

        if (budget <= 0) {
            return [{
                label: 'Realisasi tanpa anggaran',
                value: realization,
                percentage: 100,
                displayPercentage: 100,
                offset: 0,
                color: '#ef4444'
            }];
        }

        if (realization > budget) {
            return createPieSlices([
                {
                    label: 'Realisasi sesuai anggaran',
                    value: budget,
                    color: '#f59e0b'
                },
                {
                    label: 'Realisasi melebihi anggaran',
                    value: realization - budget,
                    color: '#ef4444'
                }
            ]);
        }

        return createPieSlices([
            {
                label: 'Realisasi',
                value: realization,
                color: '#10b981'
            },
            {
                label: 'Sisa anggaran',
                value: budget - realization,
                color: '#e2e8f0'
            }
        ]);
    }

    function getCompositionSlices(rows: ReportRow[]): PieSlice[] {
        return createPieSlices(
            rows.map((row, index) => ({
                label: row.procurementTypeName,
                value: Number(row.realization || 0),
                color: palette[index % palette.length]
            }))
        );
    }

    function buildBudgetGradient(budget: number, realization: number): string {
        const safeBudget = Number(budget || 0);
        const safeRealization = Number(realization || 0);

        if (safeBudget <= 0) {
            return 'conic-gradient(#94a3b8 0deg 360deg)';
        }

        if (safeRealization > safeBudget) {
            const budgetShare = safeBudget / safeRealization * 100;
            return `conic-gradient(#f59e0b 0% ${budgetShare}%, #ef4444 ${budgetShare}% 100%)`;
        }

        const realizationShare = clamp(safeRealization / safeBudget * 100, 0, 100);
        return `conic-gradient(#10b981 0% ${realizationShare}%, #e2e8f0 ${realizationShare}% 100%)`;
    }

    function buildCompositionGradient(rows: ReportRow[]): string {
        const total = rows.reduce(
            (sum, row) => sum + Number(row.realization || 0),
            0
        );

        if (total <= 0) {
            return 'conic-gradient(#cbd5e1 0deg 360deg)';
        }

        let cursor = 0;
        const parts = rows.map((row, index) => {
            const start = cursor;
            cursor += Number(row.realization || 0) / total * 100;
            return `${palette[index % palette.length]} ${start}% ${cursor}%`;
        });

        return `conic-gradient(${parts.join(', ')})`;
    }

    function getStatus(row: ReportRow) {
        if (row.budget <= 0) {
            return {
                label: 'Belum Diinput',
                badge: 'border-slate-200 bg-slate-100 text-slate-600',
                progress: 'bg-slate-400'
            };
        }

        if (row.realization > row.budget) {
            return {
                label: 'Melebihi Anggaran',
                badge: 'border-red-200 bg-red-100 text-red-700',
                progress: 'bg-red-600'
            };
        }

        if (row.absorption >= 90) {
            return {
                label: 'Hampir Habis',
                badge: 'border-amber-200 bg-amber-100 text-amber-700',
                progress: 'bg-amber-500'
            };
        }

        return {
            label: 'Terkendali',
            badge: 'border-emerald-200 bg-emerald-100 text-emerald-700',
            progress: 'bg-emerald-600'
        };
    }

    function openBudgetModal(procurementTypeId = '') {
        const selectedId = procurementTypeId || reportRows[0]?.procurementTypeId || '';
        selectedProcurementTypeId = selectedId;
        const row = reportRows.find((item) => item.procurementTypeId === selectedId);
        budgetInput = Number(row?.budget || 0);
        budgetModalOpen = true;
    }

    function syncBudgetInput() {
        const row = reportRows.find(
            (item) => item.procurementTypeId === selectedProcurementTypeId
        );
        budgetInput = Number(row?.budget || 0);
    }

    function closeBudgetModal() {
        if (saving) return;
        budgetModalOpen = false;
    }

    async function saveBudget() {
        if (!selectedProcurementTypeId) {
            toast.error('Jenis Pengadaan wajib dipilih.');
            return;
        }

        if (budgetInput < 0) {
            toast.error('Anggaran tidak boleh negatif.');
            return;
        }

        saving = true;
        try {
            const result = await (api as any).updateGupReportBudget({
                year: currentYear,
                procurementTypeId: selectedProcurementTypeId,
                amount: budgetInput
            }) as ReportResponse;

            reportRows = Array.isArray(result?.rows) ? result.rows : [];
            summary = result?.summary ?? emptySummary();
            budgetModalOpen = false;
            toast.success('Anggaran berhasil disimpan.');
        } catch (error) {
            toast.error(
                error instanceof Error
                    ? error.message
                    : 'Anggaran gagal disimpan.'
            );
        } finally {
            saving = false;
        }
    }

    function printReport() {
        window.print();
    }
</script>

<svelte:head>
    <title>Laporan dan Rekapitulasi GUP | Perjadin</title>
</svelte:head>

<main class="space-y-6 pb-20">
    <section class="no-print relative overflow-hidden rounded-3xl bg-gradient-to-br from-slate-950 via-slate-800 to-sky-900 p-6 text-white shadow-xl sm:p-8">
        <div class="absolute -right-24 -top-24 h-64 w-64 rounded-full bg-sky-400/20 blur-3xl"></div>
        <div class="absolute -bottom-28 left-20 h-72 w-72 rounded-full bg-amber-300/10 blur-3xl"></div>

        <div class="relative z-10 flex flex-col gap-6 xl:flex-row xl:items-end xl:justify-between">
            <div class="max-w-3xl">
                <div class="mb-4 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/10 px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-sky-100">
                    <FileText size={16} />
                    Visualisasi Serapan Anggaran GUP
                </div>
                <h1 class="text-3xl font-black tracking-tight sm:text-4xl">
                    Laporan dan Rekapitulasi
                </h1>
                <p class="mt-3 max-w-2xl text-sm leading-6 text-slate-200 sm:text-base">
                    Pantau Anggaran, Realisasi, Sisa Anggaran, dan Persentase Serapan dalam bentuk pie chart per Jenis Pengadaan.
                </p>
            </div>

            <div class="flex flex-col gap-3 sm:flex-row">
                <button
                    type="button"
                    on:click={() => openBudgetModal()}
                    class="inline-flex items-center justify-center gap-2 rounded-2xl bg-amber-400 px-5 py-3 text-sm font-bold text-slate-950 shadow-lg shadow-amber-500/20 transition hover:bg-amber-300"
                >
                    <Pencil size={17} />
                    Input Anggaran
                </button>
                <button
                    type="button"
                    on:click={printReport}
                    class="inline-flex items-center justify-center gap-2 rounded-2xl border border-white/20 bg-white/10 px-5 py-3 text-sm font-bold text-white transition hover:bg-white/20"
                >
                    <Printer size={17} />
                    Cetak Laporan
                </button>
            </div>
        </div>
    </section>

    {#if loadError}
        <section class="rounded-2xl border border-red-200 bg-red-50 p-5 text-sm text-red-700">
            {loadError}
            <button type="button" on:click={loadReport} class="ml-2 font-bold underline">
                Muat ulang
            </button>
        </section>
    {/if}

    <section class="grid grid-cols-1 gap-6 xl:grid-cols-5">
        <article class="overflow-hidden rounded-3xl border border-slate-200 bg-white p-5 shadow-sm xl:col-span-2">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div>
                    <p class="text-xs font-bold uppercase tracking-[0.2em] text-slate-400">Overview</p>
                    <h2 class="mt-2 text-xl font-black text-slate-900">Total Serapan Anggaran</h2>
                    <p class="mt-1 text-sm text-slate-500">Komposisi total Realisasi dan Sisa Anggaran.</p>
                </div>
                <div class="rounded-2xl bg-slate-900 px-4 py-3 text-right text-white">
                    <p class="text-xs text-slate-300">Serapan</p>
                    <p class="text-2xl font-black">{formatPercent(summary.totalAbsorption)}</p>
                </div>
            </div>

            <div class="mt-5 grid grid-cols-1 gap-4 md:grid-cols-2">
                <div class="flex h-64 items-center justify-center rounded-3xl bg-slate-50 p-4">
                    <div class="relative h-48 w-48">
                        <svg
                            viewBox="0 0 100 100"
                            class="h-full w-full drop-shadow-sm"
                            role="img"
                            aria-label="Komposisi total realisasi dan sisa anggaran"
                        >
                            {#each totalBudgetSlices as slice}
                                <g
                                    role="img"
                                    aria-label={`${slice.label}: ${formatRupiah(slice.value)}`}
                                    class="cursor-pointer transition-opacity duration-150"
                                    class:opacity-30={hoveredOverviewSlice !== null &&
                                        hoveredOverviewSlice.label !== slice.label}
                                    on:mouseenter={() => {
                                        hoveredOverviewSlice = slice;
                                    }}
                                    on:mouseleave={() => {
                                        hoveredOverviewSlice = null;
                                    }}
                                >
                                    {#if slice.percentage >= 99.999}
                                        <circle
                                            cx="50"
                                            cy="50"
                                            r="48"
                                            fill={slice.color}
                                        />
                                    {:else}
                                        <path
                                            d={createPieSectorPath(
                                                slice.offset,
                                                slice.percentage
                                            )}
                                            fill={slice.color}
                                        />
                                    {/if}

                                    <title>
                                        {slice.label}: {formatRupiah(slice.value)}
                                    </title>
                                </g>
                            {/each}
                        </svg>

                        <div class="pointer-events-none absolute inset-[28%] flex items-center justify-center rounded-full bg-white text-center shadow-sm">
                            {#if hoveredOverviewSlice}
                                <div class="px-1">
                                    <p class="text-[9px] font-bold uppercase leading-tight text-slate-500">
                                        {hoveredOverviewSlice.label}
                                    </p>
                                    <p class="mt-1 text-xs font-black text-slate-900">
                                        {formatRupiah(hoveredOverviewSlice.value)}
                                    </p>
                                    <p class="mt-1 text-[10px] text-slate-500">
                                        {formatPercent(hoveredOverviewSlice.displayPercentage)}
                                    </p>
                                </div>
                            {:else}
                                <div>
                                    <p class="text-[10px] font-bold uppercase tracking-wide text-slate-400">
                                        Serapan
                                    </p>
                                    <p class="mt-1 text-lg font-black text-slate-900">
                                        {formatPercent(summary.totalAbsorption)}
                                    </p>
                                </div>
                            {/if}
                        </div>
                    </div>
                </div>

                <div class="grid content-start grid-cols-1 gap-3">
                    <div class="rounded-2xl border border-slate-100 bg-slate-50 p-4">
                        <p class="text-xs font-semibold text-slate-500">Total Anggaran</p>
                        <h3 class="mt-1 text-xl font-black text-slate-900">
                            {formatRupiah(summary.totalBudget)}
                        </h3>
                    </div>
                    <div class="rounded-2xl border border-emerald-100 bg-emerald-50 p-4">
                        <p class="text-xs font-semibold text-emerald-700">Total Realisasi</p>
                        <h3 class="mt-1 text-xl font-black text-emerald-800">
                            {formatRupiah(summary.totalRealization)}
                        </h3>
                    </div>
                    <div class="rounded-2xl border border-amber-100 bg-amber-50 p-4">
                        <p class="text-xs font-semibold text-amber-700">Sisa Anggaran</p>
                        <h3
                            class:text-red-700={summary.totalRemaining < 0}
                            class="mt-1 text-xl font-black text-amber-800"
                        >
                            {formatRupiah(summary.totalRemaining)}
                        </h3>
                    </div>
                </div>
            </div>
        </article>

        <article class="overflow-hidden rounded-3xl border border-slate-200 bg-white p-5 shadow-sm xl:col-span-3">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div>
                    <p class="text-xs font-bold uppercase tracking-[0.2em] text-slate-400">Komposisi</p>
                    <h2 class="mt-2 text-xl font-black text-slate-900">Realisasi per Jenis Pengadaan</h2>
                    <p class="mt-1 text-sm text-slate-500">Pie chart ini ditarik dari data Realisasi pada menu Pengajuan GUP.</p>
                </div>
                <div class="inline-flex items-center gap-2 rounded-2xl bg-sky-50 px-4 py-3 text-sm font-bold text-sky-700">
                    <FileText size={16} />
                    {summary.totalProcurementType} Jenis
                </div>
            </div>

            <div class="mt-5 grid grid-cols-1 gap-5 lg:grid-cols-5">
                <div class="flex min-h-72 flex-col items-center justify-center gap-5 rounded-3xl bg-slate-50 p-5 lg:col-span-3 sm:flex-row">
                    <div class="relative h-52 w-52 shrink-0">
                        <svg
                            viewBox="0 0 100 100"
                            class="h-full w-full drop-shadow-sm"
                            role="img"
                            aria-label="Realisasi per Jenis Pengadaan"
                        >
                            {#if compositionSlices.length === 0}
                                <circle
                                    cx="50"
                                    cy="50"
                                    r="48"
                                    fill="#cbd5e1"
                                />
                            {:else}
                                {#each compositionSlices as slice}
                                    <g
                                        role="img"
                                        aria-label={`${slice.label}: ${formatRupiah(slice.value)}`}
                                        class="cursor-pointer transition-opacity duration-150"
                                        class:opacity-30={hoveredCompositionSlice !== null &&
                                            hoveredCompositionSlice.label !== slice.label}
                                        on:mouseenter={() => {
                                            hoveredCompositionSlice = slice;
                                        }}
                                        on:mouseleave={() => {
                                            hoveredCompositionSlice = null;
                                        }}
                                    >
                                        {#if slice.percentage >= 99.999}
                                            <circle
                                                cx="50"
                                                cy="50"
                                                r="48"
                                                fill={slice.color}
                                            />
                                        {:else}
                                            <path
                                                d={createPieSectorPath(
                                                    slice.offset,
                                                    slice.percentage
                                                )}
                                                fill={slice.color}
                                            />
                                        {/if}

                                        <title>
                                            {slice.label}: {formatRupiah(slice.value)}
                                        </title>
                                    </g>
                                {/each}
                            {/if}
                        </svg>

                        {#if hoveredCompositionSlice}
                            <div class="pointer-events-none absolute left-1/2 top-1/2 z-20 w-44 -translate-x-1/2 -translate-y-1/2 rounded-xl border border-slate-200 bg-white/95 p-3 text-center shadow-xl backdrop-blur">
                                <p class="break-words text-xs font-bold leading-tight text-slate-900">
                                    {hoveredCompositionSlice.label}
                                </p>
                                <p class="mt-1 text-sm font-black text-slate-900">
                                    {formatRupiah(hoveredCompositionSlice.value)}
                                </p>
                                <p class="mt-1 text-xs text-slate-500">
                                    {formatPercent(hoveredCompositionSlice.displayPercentage)}
                                    dari total realisasi
                                </p>
                            </div>
                        {/if}
                    </div>

                    <div class="max-h-60 w-full space-y-2 overflow-y-auto pr-1 sm:max-w-64">
                        {#if activeCompositionRows.length === 0}
                            <p class="text-center text-sm text-slate-500">Belum ada realisasi.</p>
                        {:else}
                            {#each activeCompositionRows as row, index}
                                <div class="flex items-center justify-between gap-3 text-xs">
                                    <div class="flex min-w-0 items-center gap-2">
                                        <span
                                            class="h-2.5 w-2.5 shrink-0 rounded-full"
                                            style:background={palette[index % palette.length]}
                                        ></span>
                                        <span class="truncate text-slate-600">
                                            {row.procurementTypeName}
                                        </span>
                                    </div>
                                    <strong class="shrink-0 text-slate-800">
                                        {formatRupiah(row.realization)}
                                    </strong>
                                </div>
                            {/each}
                        {/if}
                    </div>
                </div>

                <div class="space-y-3 lg:col-span-2">
                    <div class="rounded-2xl border border-slate-100 bg-slate-50 p-4">
                        <p class="text-xs font-semibold text-slate-500">Jumlah Transaksi GUP</p>
                        <h3 class="mt-1 text-2xl font-black text-slate-900">{summary.totalTransactions}</h3>
                    </div>
                    <div class="rounded-2xl border border-slate-100 bg-slate-50 p-4">
                        <p class="text-xs font-semibold text-slate-500">Nilai Pengajuan</p>
                        <h3 class="mt-1 text-2xl font-black text-slate-900">{formatRupiah(summary.totalSubmissionValue)}</h3>
                    </div>
                    <div class="rounded-2xl border border-slate-100 bg-slate-50 p-4 text-sm text-slate-500">
                        <Info size={17} class="mr-2 inline text-sky-600" />
                        Jenis Pengadaan dengan realisasi nol tetap muncul pada kartu detail, tetapi tidak mendominasi pie komposisi.
                    </div>
                </div>
            </div>
        </article>
    </section>

    <section class="rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div class="flex flex-col gap-4 border-b border-slate-200 p-5 sm:flex-row sm:items-center sm:justify-between sm:p-6">
            <div>
                <p class="text-xs font-bold uppercase tracking-[0.2em] text-slate-400">Detail Serapan</p>
                <h2 class="mt-2 text-xl font-black text-slate-900">Pie Chart Tiap Jenis Pengadaan</h2>
                <p class="mt-1 text-sm text-slate-500">Setiap kartu menampilkan komposisi Realisasi dan Sisa Anggaran berdasarkan anggaran yang diinput.</p>
            </div>

            <div class="no-print flex w-full flex-col gap-3 sm:w-auto sm:flex-row">
                <div class="relative w-full sm:w-80">
                    <Search size={17} class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400" />
                    <input
                        bind:value={searchQuery}
                        type="search"
                        placeholder="Cari jenis pengadaan / kode akun..."
                        class="w-full rounded-2xl border border-slate-300 bg-slate-50 py-3 pl-11 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-slate-800"
                    />
                </div>
                <button
                    type="button"
                    on:click={() => openBudgetModal()}
                    class="inline-flex items-center justify-center gap-2 rounded-2xl bg-slate-900 px-5 py-3 text-sm font-bold text-white transition hover:bg-slate-800"
                >
                    <Plus size={17} />
                    Input
                </button>
            </div>
        </div>

        {#if loading}
            <div class="flex min-h-80 items-center justify-center p-6 text-sm text-slate-500">
                Memuat laporan GUP...
            </div>
        {:else if filteredRows.length === 0}
            <div class="flex min-h-80 items-center justify-center p-6 text-sm text-slate-500">
                Data tidak ditemukan.
            </div>
        {:else}
            <div class="grid grid-cols-1 gap-5 p-5 sm:grid-cols-2 sm:p-6 xl:grid-cols-3 2xl:grid-cols-4">
                {#each filteredRows as row}
                    {@const status = getStatus(row)}
                    {@const budgetSlices = getBudgetSlices(row.budget, row.realization)}
                    <article class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm transition hover:-translate-y-1 hover:shadow-lg">
                        <div class="flex items-start justify-between gap-4">
                            <div>
                                <p class="text-xs font-bold uppercase tracking-[0.16em] text-slate-400">{row.accountCode}</p>
                                <h3 class="mt-2 min-h-[44px] break-words text-base font-black leading-snug text-slate-900" title={row.procurementTypeName}>{row.procurementTypeName}</h3>
                                <p class="mt-1 text-xs text-slate-400">{row.transactionCount} transaksi pengajuan</p>
                            </div>
                            <button
                                type="button"
                                on:click={() => openBudgetModal(row.procurementTypeId)}
                                class="no-print flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-blue-50 text-blue-700 transition hover:bg-blue-100"
                                title="Input Anggaran"
                                aria-label={`Input anggaran ${row.procurementTypeName}`}
                            >
                                <Pencil size={17} />
                            </button>
                        </div>

                        <div
    class="mt-5 flex h-52 items-center justify-center rounded-3xl bg-slate-50 p-4"
>
    <div class="relative h-40 w-40">
        <svg
            viewBox="0 0 100 100"
            class="h-full w-full drop-shadow-sm"
            role="img"
            aria-label={`Chart anggaran ${row.procurementTypeName}`}
        >
            {#each budgetSlices as slice}
                <g
                    role="img"
                    aria-label={`${slice.label}: ${formatRupiah(
                        slice.value
                    )}`}
                    class="cursor-pointer transition-opacity duration-150"
                    class:opacity-30={hoveredCardSlice?.chartId ===
                        row.procurementTypeId &&
                        hoveredCardSlice.slice.label !==
                            slice.label}
                    on:mouseenter={() => {
                        hoveredCardSlice = {
                            chartId: row.procurementTypeId,
                            slice
                        };
                    }}
                    on:mouseleave={() => {
                        hoveredCardSlice = null;
                    }}
                >
                    {#if slice.percentage >= 99.999}
                        <circle
                            cx="50"
                            cy="50"
                            r="48"
                            fill={slice.color}
                        />
                    {:else}
                        <path
                            d={createPieSectorPath(
                                slice.offset,
                                slice.percentage
                            )}
                            fill={slice.color}
                        />
                    {/if}

                    <title>
                        {slice.label}: {formatRupiah(
                            slice.value
                        )}
                    </title>
                </g>
            {/each}
        </svg>

        {#if hoveredCardSlice?.chartId === row.procurementTypeId}
            <div
                class="pointer-events-none absolute left-1/2 top-1/2 z-20 w-40 -translate-x-1/2 -translate-y-1/2 rounded-xl border border-slate-200 bg-white/95 p-3 text-center shadow-xl backdrop-blur"
            >
                <p
                    class="text-[11px] font-bold leading-tight text-slate-900"
                >
                    {hoveredCardSlice.slice.label}
                </p>

                <p
                    class="mt-1 text-xs font-black text-slate-900"
                >
                    {formatRupiah(
                        hoveredCardSlice.slice.value
                    )}
                </p>

                <p class="mt-1 text-[10px] text-slate-500">
                    {formatPercent(
                        hoveredCardSlice.slice.percentage
                    )}
                </p>
            </div>
        {/if}
    </div>
</div>

                        <div class="mt-5 grid grid-cols-2 gap-3 text-sm">
                            <div class="rounded-2xl bg-slate-50 p-3">
                                <p class="text-xs font-semibold text-slate-500">Anggaran</p>
                                <p class="mt-1 font-black text-slate-900">{formatRupiah(row.budget)}</p>
                            </div>
                            <div class="rounded-2xl bg-emerald-50 p-3">
                                <p class="text-xs font-semibold text-emerald-700">Realisasi</p>
                                <p class="mt-1 font-black text-emerald-800">{formatRupiah(row.realization)}</p>
                            </div>
                            <div class="rounded-2xl bg-amber-50 p-3">
                                <p class="text-xs font-semibold text-amber-700">Sisa</p>
                                <p class:text-red-700={row.remaining < 0} class="mt-1 font-black text-amber-800">{formatRupiah(row.remaining)}</p>
                            </div>
                            <div class="rounded-2xl bg-violet-50 p-3">
                                <p class="text-xs font-semibold text-violet-700">Serapan</p>
                                <p class="mt-1 font-black text-violet-800">{formatPercent(row.absorption)}</p>
                            </div>
                        </div>

                        <div class="mt-4">
                            <div class="h-2.5 overflow-hidden rounded-full bg-slate-100">
                                <div
                                    class={`h-full rounded-full transition-all duration-500 ${status.progress}`}
                                    style:width={`${clamp(row.absorption, 0, 100)}%`}
                                ></div>
                            </div>
                            <div class="mt-3 flex min-w-0 items-start justify-between gap-3">
                                <span class={`inline-flex shrink-0 rounded-full border px-3 py-1 text-xs font-bold ${status.badge}`}>
                                    {status.label}
                                </span>
                                <div class="min-w-0 flex-1 text-right" title={row.mak}>
                                    <p class="text-[10px] font-semibold uppercase tracking-wide text-slate-400">MAK</p>
                                    <p class="mt-0.5 break-all text-[10px] font-semibold leading-4 text-slate-500">{row.mak || '-'}</p>
                                </div>
                            </div>
                        </div>
                    </article>
                {/each}
            </div>
        {/if}
    </section>

    <section class="no-print grid grid-cols-1 gap-5 lg:grid-cols-3">
        <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm lg:col-span-2">
            <h3 class="text-lg font-black text-slate-900">Cara baca pie chart</h3>
            <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
                <div class="rounded-2xl bg-slate-50 p-4">
                    <div class="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-600 font-black text-white">✓</div>
                    <p class="font-bold text-slate-900">Realisasi</p>
                    <p class="mt-1 text-sm text-slate-500">Bagian anggaran yang sudah digunakan berdasarkan Pengajuan GUP.</p>
                </div>
                <div class="rounded-2xl bg-slate-50 p-4">
                    <div class="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-slate-500 font-black text-white">Rp</div>
                    <p class="font-bold text-slate-900">Sisa Anggaran</p>
                    <p class="mt-1 text-sm text-slate-500">Selisih Anggaran dikurangi Realisasi untuk jenis pengadaan tersebut.</p>
                </div>
                <div class="rounded-2xl bg-slate-50 p-4">
                    <div class="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-red-600 font-black text-white">!</div>
                    <p class="font-bold text-slate-900">Melebihi Anggaran</p>
                    <p class="mt-1 text-sm text-slate-500">Muncul jika Realisasi lebih besar daripada Anggaran yang diinput.</p>
                </div>
            </div>
        </div>

        <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
            <h3 class="text-lg font-black text-slate-900">Data Sumber</h3>
            <dl class="mt-4 space-y-3 text-sm">
                <div class="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3">
                    <dt class="text-slate-500">Jenis Pengadaan</dt>
                    <dd class="font-black text-slate-900">{summary.totalProcurementType}</dd>
                </div>
                <div class="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3">
                    <dt class="text-slate-500">Transaksi GUP</dt>
                    <dd class="font-black text-slate-900">{summary.totalTransactions}</dd>
                </div>
                <div class="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3">
                    <dt class="text-slate-500">Realisasi</dt>
                    <dd class="font-black text-slate-900">{formatRupiah(summary.totalRealization)}</dd>
                </div>
            </dl>
        </div>
    </section>
</main>

{#if budgetModalOpen}
    <div
        class="no-print fixed inset-0 z-[100] flex items-center justify-center bg-black/50 p-4"
        role="presentation"
        on:click={(event) => {
            if (event.currentTarget === event.target) closeBudgetModal();
        }}
    >
        <div class="max-h-[92vh] w-full max-w-3xl overflow-y-auto rounded-3xl bg-white shadow-2xl" role="dialog" aria-modal="true">
            <div class="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white px-6 py-5">
                <div>
                    <h3 class="text-xl font-black text-slate-900">Input Anggaran Jenis Pengadaan</h3>
                    <p class="mt-1 text-sm text-slate-500">Input ini khusus alokasi anggaran per Jenis Pengadaan. Sumber dana LS dikelola melalui menu GUP &gt; LS.</p>
                </div>
                <button
                    type="button"
                    on:click={closeBudgetModal}
                    disabled={saving}
                    class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-slate-100 hover:bg-slate-200 disabled:opacity-50"
                    aria-label="Tutup modal"
                >
                    <X size={19} class="text-slate-600" />
                </button>
            </div>

            <form class="space-y-5 p-6" on:submit|preventDefault={saveBudget}>
                <div class="grid grid-cols-1 gap-5 md:grid-cols-2">
                    <div>
                        <label for="budgetProcurement" class="mb-2 block text-sm font-bold text-slate-700">Jenis Pengadaan</label>
                        <select
                            id="budgetProcurement"
                            bind:value={selectedProcurementTypeId}
                            on:change={syncBudgetInput}
                            required
                            class="w-full rounded-2xl border border-slate-300 px-4 py-3 focus:outline-none focus:ring-2 focus:ring-slate-900"
                        >
                            {#each reportRows as row}
                                <option value={row.procurementTypeId}>{row.procurementTypeName}</option>
                            {/each}
                        </select>
                    </div>

                    <div>
                        <label for="budgetAccount" class="mb-2 block text-sm font-bold text-slate-700">Kode Akun</label>
                        <input
                            id="budgetAccount"
                            readonly
                            value={selectedBudgetRow?.accountCode || '-'}
                            class="w-full rounded-2xl border border-slate-200 bg-slate-100 px-4 py-3 font-semibold text-slate-600 focus:outline-none"
                        />
                    </div>

                    <div>
                        <label for="budgetAmount" class="mb-2 block text-sm font-bold text-slate-700">Input Anggaran</label>
                        <input
                            id="budgetAmount"
                            required
                            min="0"
                            inputmode="numeric"
                            value={formatNumber(budgetInput)}
                            on:input={(event) => {
                                budgetInput = parseCurrency(event.currentTarget.value);
                            }}
                            class="w-full rounded-2xl border border-slate-300 px-4 py-3 text-lg font-black text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-900"
                            placeholder="Contoh: 10000000"
                        />
                    </div>

                    <div>
                        <label for="budgetRealization" class="mb-2 block text-sm font-bold text-slate-700">Realisasi</label>
                        <input id="budgetRealization" readonly value={formatRupiah(modalRealization)} class="w-full rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 font-bold text-emerald-700 focus:outline-none" />
                    </div>

                    <div>
                        <label for="budgetRemaining" class="mb-2 block text-sm font-bold text-slate-700">Sisa Anggaran</label>
                        <input id="budgetRemaining" readonly value={formatRupiah(modalRemaining)} class:text-red-700={modalRemaining < 0} class="w-full rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 font-bold text-amber-700 focus:outline-none" />
                    </div>

                    <div>
                        <label for="budgetPercentage" class="mb-2 block text-sm font-bold text-slate-700">Persentase</label>
                        <input id="budgetPercentage" readonly value={formatPercent(modalAbsorption)} class="w-full rounded-2xl border border-violet-200 bg-violet-50 px-4 py-3 font-bold text-violet-700 focus:outline-none" />
                    </div>
                </div>

                <div class="rounded-2xl border border-blue-100 bg-blue-50 p-4 text-sm text-blue-700">
                    <Info size={17} class="mr-2 inline" />
                    Setelah disimpan, nilai anggaran akan memperbarui pie chart, sisa anggaran, dan persentase serapan tiap jenis pengadaan.
                </div>

                <div class="flex flex-col-reverse gap-3 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
                    <button type="button" on:click={closeBudgetModal} disabled={saving} class="rounded-2xl bg-slate-100 px-5 py-3 text-sm font-bold text-slate-700 hover:bg-slate-200 disabled:opacity-50">
                        Batal
                    </button>
                    <button type="submit" disabled={saving} class="rounded-2xl bg-slate-900 px-5 py-3 text-sm font-bold text-white hover:bg-slate-800 disabled:opacity-50">
                        {saving ? 'Menyimpan...' : 'Simpan Anggaran'}
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}

<style>
    @media print {
        :global(body) {
            background: white !important;
        }

        :global(aside),
        :global(header),
        .no-print {
            display: none !important;
        }

        main {
            padding: 0 !important;
        }

        article,
        section {
            break-inside: avoid;
        }
    }
</style>
