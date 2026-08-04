<script>
    import { fly } from 'svelte/transition';
    import { userStore } from '$lib/features/auth/store';
    export let title = 'Aktivitas Terbaru';
    export let viewAllLink = '';
    
    $: isProtokol = $userStore.role !== 'super_admin' && $userStore.role !== 'kasubag';
</script>

<div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden flex flex-col h-full">
    <!-- Header -->
    <div class="px-6 py-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/95">
        <div class="flex items-center gap-2">
            <div class="h-4 w-1 bg-blue-500 rounded-full"></div>
            <h3 class="font-bold text-slate-800 text-lg tracking-tight">{title}</h3>
        </div>
        {#if viewAllLink}
            <a href={viewAllLink} class="group flex items-center gap-1 text-sm font-semibold text-blue-600 hover:text-blue-700 transition-colors">
                <span>Lihat Semua</span>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 transition-transform group-hover:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3" />
                </svg>
            </a>
        {/if}
    </div>

    <!-- List Container -->
    <div class="flex-1 overflow-x-auto">
        <table class="w-full text-sm text-left border-collapse">
            <thead class="bg-slate-50/80 text-slate-500 font-semibold uppercase text-xs tracking-wider border-b border-slate-100">
                <tr>
                    {#if isProtokol}
                        <th class="px-6 py-4 font-medium">Nomor SPJ</th>
                        <th class="px-6 py-4 font-medium">Tujuan & Lokasi</th>
                        <th class="px-6 py-4 font-medium">Tanggal</th>
                        <th class="px-6 py-4 font-medium text-center">Status</th>
                    {:else}
                        <th class="px-6 py-4 font-medium">Pegawai</th>
                        <th class="px-6 py-4 font-medium">Tujuan & Lokasi</th>
                        <th class="px-6 py-4 font-medium">Tanggal</th>
                        <th class="px-6 py-4 font-medium text-center">Status</th>
                        <th class="px-6 py-4 font-medium text-right">Biaya</th>
                    {/if}
                </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 bg-white">
                <slot />
            </tbody>
        </table>
    </div>
</div>
