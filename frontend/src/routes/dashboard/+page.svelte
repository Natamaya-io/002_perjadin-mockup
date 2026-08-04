<script>
    import { userStore } from '$lib/features/auth/store';
    import { onMount } from 'svelte';
    import { api } from '$lib/shared/api';
    import { dashboardSummaryStore } from '$lib/features/dashboard/store';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import { formatCurrency } from '$lib/shared/utils/utils';

    // Granular Components
    import WelcomeBanner from '$lib/features/dashboard/ui/welcome/WelcomeBanner.svelte';
    import WelcomeContent from '$lib/features/dashboard/ui/welcome/WelcomeContent.svelte';
    import WelcomeActions from '$lib/features/dashboard/ui/welcome/WelcomeActions.svelte';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';

    import StatsGrid from '$lib/features/dashboard/ui/stats/StatsGrid.svelte';
    import StatCard from '$lib/features/dashboard/ui/stats/StatCard.svelte';

    import RecentActivityCard from '$lib/features/dashboard/ui/recent/RecentActivityCard.svelte';
    import ActivityItem from '$lib/features/dashboard/ui/recent/ActivityItem.svelte';
    import EmptyActivity from '$lib/features/dashboard/ui/recent/EmptyActivity.svelte';

    import TimelineCalendar from '$lib/features/dashboard/ui/timeline/TimelineCalendar.svelte';
    import PieChart from '$lib/shared/ui/charts/PieChart.svelte';
    
    export let data;

    // Only show loading spinner if we don't have cached data yet
    let isFetchingStats = !$dashboardSummaryStore;

    // Helper for currency
    function formatIDR(amount) {
        return formatCurrency(amount);
    }

    // Compact currency for card display (shorter format for large numbers)
    function formatIDRCompact(amount) {
        if (amount >= 1_000_000_000) {
            const val = (amount / 1_000_000_000).toFixed(2);
            return `Rp ${val.replace('.', ',')} M`;
        }
        if (amount >= 100_000_000) {
            const val = (amount / 1_000_000).toFixed(1);
            return `Rp ${val.replace('.', ',')} Jt`;
        }
        return formatCurrency(amount);
    }

    // Reactive loader visibility based on store state
    $: isFetchingStats = !$dashboardSummaryStore;

    $: stats = $dashboardSummaryStore || {
        totalTrips: 0,
        activeTrips: 0,
        statusCompleted: 0,
        statusInProgress: 0,
        statusAssigned: 0,
        statusRejected: 0,
        reportCompleted: 0,
        reportPending: 0,
        recentRecords: [],
        budgets: []
    };

    onMount(() => {
        // Let the non-blocking +page.ts handle the data fetching (Stale-While-Revalidate).
        // If the store is populated, we hide the loader instantly.
        if ($dashboardSummaryStore) {
            isFetchingStats = false;
        }
    });

    $: totalTrips = stats.totalTrips;
    $: activeTrips = stats.activeTrips;
    
    // Server-Side Chart Data
    $: statusData = [
        { label: 'Completed', value: stats.statusCompleted, color: '#10b981' },
        { label: 'In Progress', value: stats.statusInProgress, color: '#f97316' },
        { label: 'Assigned', value: stats.statusAssigned, color: '#eab308' },
        { label: 'Ditolak', value: stats.statusRejected, color: '#ef4444' }
    ].filter(d => d.value > 0);

    $: reportData = [
         { label: 'Laporan Selesai', value: stats.reportCompleted, color: '#3b82f6' },
         { label: 'Belum Lapor', value: stats.reportPending, color: '#a855f7' }
    ].filter(d => d.value > 0);

    // Recent Records
    $: recentRecords = stats.recentRecords.map((r, i) => ({ ...r, nomorSpdPetugas: String(i+1).padStart(3, '0') }));
    $: newAssignments = stats.statusAssigned;
    $: pendingReports = stats.reportPending;
    $: myUniqueTrips = [...new Map(stats.recentRecords.map(item => [item.spd, item])).values()];

    // --- Budget Filter Logic ---
    const currentDate = new Date();
    let filterMonth = 0; // 0 = Semua Bulan
    let filterYear = currentDate.getFullYear();

    const monthNames = [
        'Semua Bulan', 'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
        'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
    ];

    $: availableYears = (() => {
        const years = new Set();
        years.add(currentDate.getFullYear());
        (stats.budgets || []).forEach(b => years.add(b.year));
        return [...years].sort((a, b) => b - a);
    })();

    $: filteredTotalCost = (stats.budgets || []).reduce((acc, b) => {
        const yearMatch = b.year === filterYear;
        const monthMatch = filterMonth === 0 || b.month === filterMonth;
        
        if (yearMatch && monthMatch) {
            return acc + b.total;
        }
        return acc;
    }, 0);
    
    $: totalCost = (stats.budgets || []).reduce((acc, b) => acc + b.total, 0);

    // Description text based on filter
    $: budgetDescription = filterMonth === 0 
        ? `Realisasi tahun ${filterYear}` 
        : `${monthNames[filterMonth]} ${filterYear}`;

    let showMonthDropdown = false;
    let showYearDropdown = false;

    function selectMonth(m) {
        filterMonth = m;
        showMonthDropdown = false;
    }

    function selectYear(y) {
        filterYear = y;
        showYearDropdown = false;
    }

    function resetFilter() {
        filterMonth = 0;
        filterYear = currentDate.getFullYear();
        showMonthDropdown = false;
        showYearDropdown = false;
    }

    function toggleMonthDropdown() {
        showMonthDropdown = !showMonthDropdown;
        showYearDropdown = false;
    }

    function toggleYearDropdown() {
        showYearDropdown = !showYearDropdown;
        showMonthDropdown = false;
    }

    function handleClickOutside(e) {
        const target = e.target;
        if (!target.closest('.budget-dropdown-wrapper')) {
            showMonthDropdown = false;
            showYearDropdown = false;
        }
    }
</script>


<div class="space-y-8 pb-20 relative">
    {#if isFetchingStats}
        <div class="absolute inset-0 z-50 bg-slate-100/80 flex items-start justify-center pt-32 rounded-2xl">
            <div class="flex flex-col items-center justify-center gap-1 bg-white p-4 rounded-xl shadow-lg border border-slate-200">
                <LottieLoader size="80px" />
                <span class="text-sm font-semibold text-slate-600 animate-pulse -mt-2">Menghitung Statistik...</span>
            </div>
        </div>
    {/if}
    <!-- 1. Welcome Section (Organism) -->
    <WelcomeBanner role={$userStore.role}>
        <WelcomeContent role={$userStore.role} />
        
        <WelcomeActions>
            {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}
            <a href="/dashboard/pengajuan/new" class="w-full md:w-auto">
                <Button class="w-full md:w-auto bg-white text-blue-600 hover:bg-blue-50 border-0 shadow-lg font-semibold h-12 px-6 rounded-xl transition-transform active:scale-95 whitespace-nowrap">
                    + Pengajuan Perjalanan
                </Button>
            </a>
            {/if}
            {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}
                 <a href="/dashboard/admin/perdin" class="w-full md:w-auto">
                    <Button variant="outline" class="w-full md:w-auto bg-blue-600/30 border-white/20 text-white hover:bg-blue-600/50 hover:text-white h-12 px-6 rounded-xl transition-transform active:scale-95 whitespace-nowrap">
                       Kelola Biaya
                    </Button>
                 </a>            {/if}
        </WelcomeActions>
    </WelcomeBanner>

    <!-- 1.5 Alert Penugasan Baru (Khusus Protokol) -->
    {#if $userStore.role !== 'super_admin' && $userStore.role !== 'kasubag'}
        {#if newAssignments > 0}
        <div class="bg-gradient-to-r from-rose-500 to-pink-600 rounded-2xl p-6 shadow-lg shadow-rose-500/20 text-white flex flex-col md:flex-row items-center justify-between gap-6 relative overflow-hidden mb-8 animate-in fade-in slide-in-from-bottom-2">
            <div class="absolute -right-6 -top-6 h-32 w-32 rounded-full bg-white/10 blur-2xl pointer-events-none"></div>
            <div class="absolute -left-6 -bottom-6 h-24 w-24 rounded-full bg-white/10 blur-xl pointer-events-none"></div>
            
            <div class="relative z-10 flex flex-col sm:flex-row items-center text-center sm:text-left gap-5 w-full md:w-auto">
                <div class="h-14 w-14 rounded-full bg-white/30 flex items-center justify-center shrink-0 border border-white/20 shadow-inner">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-7 w-7 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                    </svg>
                </div>
                <div>
                    <h3 class="text-xl sm:text-2xl font-bold tracking-tight mb-1">Anda memiliki {newAssignments} Penugasan Baru!</h3>
                    <p class="text-rose-100 text-sm">Ada perjalanan dinas baru yang menunggu konfirmasi dan tindakan dari Anda.</p>
                </div>
            </div>
            
            <div class="relative z-10 w-full md:w-auto shrink-0 mt-2 md:mt-0">
                <a href="/dashboard/laporan?from=notif" class="inline-flex w-full md:w-auto items-center justify-center px-6 py-3 text-sm font-bold tracking-wide text-rose-600 bg-white hover:bg-rose-50 rounded-xl shadow-md shadow-black/10 transition-all hover:scale-105 active:scale-95">
                    Lihat Detail Penugasan
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M14 5l7 7m0 0l-7 7m7-7H3" />
                    </svg>
                </a>
            </div>
        </div>
        {:else}
        <div class="bg-white border border-slate-200 rounded-2xl p-5 flex flex-col sm:flex-row items-center justify-between gap-4 mb-8 shadow-sm">
            <div class="flex items-center gap-4 text-center sm:text-left">
                <div class="h-12 w-12 rounded-full bg-slate-50 flex items-center justify-center border border-slate-100 shrink-0">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                </div>
                <div>
                    <h3 class="text-base font-bold text-slate-700">Tidak ada penugasan baru</h3>
                    <p class="text-slate-500 text-sm mt-0.5">Anda sudah menyelesaikan semua konfirmasi penugasan saat ini.</p>
                </div>
            </div>
            <div class="w-full sm:w-auto">
                <a href="/dashboard/laporan" class="inline-flex w-full sm:w-auto items-center justify-center px-4 py-2 text-xs font-semibold text-slate-600 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-lg transition-colors">
                    Lihat Riwayat Perjalanan
                </a>
            </div>
        </div>
        {/if}
    {/if}

    <!-- 2. Stats Section (Organism) -->
    <StatsGrid>
            <StatCard 
                title="Total Perjalanan" 
                value={totalTrips} 
                description="Kegiatan tercatat tahun ini" 
                iconColor="blue"
                bgClass="bg-gradient-to-br from-blue-500 to-indigo-600 border-transparent shadow-lg shadow-blue-500/20"
                textColorClass="text-white"
                titleColorClass="text-blue-100"
                descColorClass="text-blue-50"
                blobClass="bg-white/10 group-hover:bg-white/20"
                iconContainerClass="bg-white/20 text-white"
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                    </svg>
                </div>
            </StatCard>

            {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}
            <StatCard 
                title="Total Anggaran" 
                value={formatIDR(filteredTotalCost)} 
                description={budgetDescription} 
                iconColor="emerald"
                bgClass="bg-gradient-to-br from-emerald-500 to-teal-600 border-transparent shadow-lg shadow-emerald-500/20"
                textColorClass="text-white"
                titleColorClass="text-emerald-100"
                descColorClass="text-emerald-50"
                blobClass="bg-white/10 group-hover:bg-white/20"
                iconContainerClass="bg-white/20 text-white"
                isSecret={true}
                autoShrink={true}
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </div>

                <!-- svelte-ignore a11y-click-events-have-key-events -->
                <!-- svelte-ignore a11y-no-static-element-interactions -->
                <div slot="filter" class="flex items-center gap-1.5 flex-wrap -mt-1" on:click|stopPropagation>
                    <!-- Month Dropdown -->
                    <div class="budget-dropdown-wrapper">
                        <button class="budget-pill" on:click={toggleMonthDropdown}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 opacity-70" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
                            <span>{filterMonth === 0 ? 'Semua' : monthNames[filterMonth].substring(0, 3)}</span>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-2.5 w-2.5 opacity-60 transition-transform" class:rotate-180={showMonthDropdown} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7" /></svg>
                        </button>
                        {#if showMonthDropdown}
                            <div class="budget-dropdown-panel">
                                <div class="budget-dropdown-grid">
                                    {#each monthNames as name, i}
                                        <button 
                                            class="budget-dropdown-item" 
                                            class:active={filterMonth === i}
                                            on:click={() => selectMonth(i)}
                                        >
                                            {i === 0 ? 'Semua' : name.substring(0, 3)}
                                        </button>
                                    {/each}
                                </div>
                            </div>
                        {/if}
                    </div>

                    <!-- Year Dropdown -->
                    <div class="budget-dropdown-wrapper">
                        <button class="budget-pill" on:click={toggleYearDropdown}>
                            <span>{filterYear}</span>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-2.5 w-2.5 opacity-60 transition-transform" class:rotate-180={showYearDropdown} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7" /></svg>
                        </button>
                        {#if showYearDropdown}
                            <div class="budget-dropdown-panel budget-dropdown-panel-sm">
                                {#each availableYears as year}
                                    <button 
                                        class="budget-dropdown-item-row" 
                                        class:active={filterYear === year}
                                        on:click={() => selectYear(year)}
                                    >
                                        {year}
                                    </button>
                                {/each}
                            </div>
                        {/if}
                    </div>

                    <!-- Reset Button -->
                    {#if filterMonth !== 0 || filterYear !== currentDate.getFullYear()}
                        <button 
                            class="budget-filter-reset"
                            on:click={resetFilter}
                            title="Reset filter"
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </button>
                    {/if}
                </div>
            </StatCard>
            {/if}

            <StatCard 
                title="Sedang Berjalan" 
                value={activeTrips} 
                description="Tim aktif di lapangan" 
                iconColor="amber"
                bgClass="bg-gradient-to-br from-amber-500 to-orange-600 border-transparent shadow-lg shadow-amber-500/20"
                textColorClass="text-white"
                titleColorClass="text-amber-100"
                descColorClass="text-amber-50"
                blobClass="bg-white/10 group-hover:bg-white/20"
                iconContainerClass="bg-white/20 text-white"
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </div>
            </StatCard>

            <StatCard
                title="Laporan Pending"
                value={pendingReports}
                description="Menunggu kelengkapan dokumen"
                iconColor="purple"
                bgClass="bg-gradient-to-br from-purple-500 to-fuchsia-600 border-transparent shadow-lg shadow-purple-500/20"
                textColorClass="text-white"
                titleColorClass="text-purple-100"
                descColorClass="text-purple-50"
                blobClass="bg-white/10 group-hover:bg-white/20"
                iconContainerClass="bg-white/20 text-white"
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                </div>
            
                <div slot="action">
                    {#if pendingReports > 0 && $userStore.role !== 'super_admin' && $userStore.role !== 'kasubag'}
                    <a href="/dashboard/laporan?from=notif" class="inline-flex items-center justify-center px-3 py-1.5 text-[10px] font-bold tracking-wide text-white bg-purple-500 hover:bg-purple-600 rounded-lg shadow-sm transition-colors shadow-purple-500/20 hover:shadow-purple-500/40">
                        LIHAT <span class="sr-only">Laporan Pending</span>
                    </a>
                    {/if}
                </div>
            </StatCard>
        </StatsGrid>
        
        <!-- 2.5 Charts Section (New) -->
        {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6 animate-in fade-in slide-in-from-bottom-4 duration-700 delay-150">
            <PieChart title="Status Pengajuan" data={statusData} />
            <PieChart title="Status Laporan" data={reportData} />
        </div>
        {/if}

        <!-- 3. Recent Activity Section (Organism) -->
        <RecentActivityCard 
            title={$userStore.role !== 'super_admin' && $userStore.role !== 'kasubag' ? "Riwayat Terbaru" : "Aktivitas Terbaru"}
            viewAllLink={$userStore.role === 'super_admin' || $userStore.role === 'kasubag' ? "/dashboard/admin/perdin" : ""}
        >
            {#each recentRecords as record (record.id)}
                <ActivityItem {record} {formatIDR} />
            {/each}
            
            {#if recentRecords.length === 0}
                <EmptyActivity />
            {/if}
        </RecentActivityCard>

        <!-- 4. Timeline Calendar (Protokol User) -->
        {#if $userStore.role !== 'super_admin' && $userStore.role !== 'kasubag'}
            <div class="pt-4">
                <TimelineCalendar records={myUniqueTrips} />
            </div>
        {/if}
</div>

<svelte:window on:click={handleClickOutside} />

<style>
    :global(.budget-dropdown-wrapper) {
        position: relative;
    }

    :global(.budget-pill) {
        display: inline-flex;
        align-items: center;
        gap: 0.3rem;
        background: rgba(255, 255, 255, 0.25);
        border: 1px solid rgba(255, 255, 255, 0.25);
        border-radius: 2rem;
        padding: 0.25rem 0.6rem;
        font-size: 0.65rem;
        font-weight: 600;
        color: white;
        cursor: pointer;
        outline: none;
        transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
        white-space: nowrap;
        letter-spacing: 0.01em;
    }

    :global(.budget-pill:hover) {
        background: rgba(255, 255, 255, 0.28);
        border-color: rgba(255, 255, 255, 0.4);
        transform: translateY(-1px);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    }

    :global(.budget-pill:active) {
        transform: scale(0.96);
    }

    :global(.budget-dropdown-panel) {
        position: absolute;
        top: calc(100% + 6px);
        left: 0;
        z-index: 50;
        min-width: 190px;
        background: rgba(15, 23, 42, 0.98);
        border: 1px solid rgba(255, 255, 255, 0.12);
        border-radius: 0.75rem;
        padding: 0.4rem;
        box-shadow: 
            0 10px 40px rgba(0, 0, 0, 0.35),
            0 0 0 1px rgba(255, 255, 255, 0.05) inset;
        animation: dropdownSlideIn 0.15s cubic-bezier(0.16, 1, 0.3, 1);
    }

    :global(.budget-dropdown-panel-sm) {
        min-width: 80px;
    }

    :global(.budget-dropdown-grid) {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 2px;
    }

    :global(.budget-dropdown-item) {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0.35rem 0.25rem;
        font-size: 0.65rem;
        font-weight: 500;
        color: rgba(255, 255, 255, 0.7);
        border-radius: 0.4rem;
        cursor: pointer;
        transition: all 0.15s ease;
        border: none;
        background: transparent;
        white-space: nowrap;
    }

    :global(.budget-dropdown-item:hover) {
        background: rgba(255, 255, 255, 0.1);
        color: white;
    }

    :global(.budget-dropdown-item.active) {
        background: rgba(16, 185, 129, 0.5);
        color: white;
        font-weight: 700;
        box-shadow: 0 0 8px rgba(16, 185, 129, 0.3);
    }

    :global(.budget-dropdown-item-row) {
        display: flex;
        align-items: center;
        width: 100%;
        padding: 0.4rem 0.75rem;
        font-size: 0.7rem;
        font-weight: 500;
        color: rgba(255, 255, 255, 0.7);
        border-radius: 0.4rem;
        cursor: pointer;
        transition: all 0.15s ease;
        border: none;
        background: transparent;
    }

    :global(.budget-dropdown-item-row:hover) {
        background: rgba(255, 255, 255, 0.1);
        color: white;
    }

    :global(.budget-dropdown-item-row.active) {
        background: rgba(16, 185, 129, 0.5);
        color: white;
        font-weight: 700;
    }

    :global(.budget-filter-reset) {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 1.35rem;
        height: 1.35rem;
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.15);
        border: 1px solid rgba(255, 255, 255, 0.2);
        color: white;
        cursor: pointer;
        transition: all 0.2s ease;
        padding: 0;
    }

    :global(.budget-filter-reset:hover) {
        background: rgba(255, 80, 80, 0.45);
        border-color: rgba(255, 80, 80, 0.6);
        transform: scale(1.15) rotate(90deg);
    }

    :global(.rotate-180) {
        transform: rotate(180deg);
    }

    @keyframes -global-dropdownSlideIn {
        from {
            opacity: 0;
            transform: translateY(-4px) scale(0.96);
        }
        to {
            opacity: 1;
            transform: translateY(0) scale(1);
        }
    }
</style>
