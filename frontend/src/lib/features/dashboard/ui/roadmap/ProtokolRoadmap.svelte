<script>
    import { userStore } from '$lib/features/auth/store';
    import { toTitleCase } from '$lib/shared/utils/utils';
    import { fade, fly } from 'svelte/transition';
    export let records = [];

    // Tampilkan hanya yang belum sepenuhnya selesai (belum dibayar) atau batasi 5 terbaru jika sudah banyak
    $: activeRecords = [...records].sort((a, b) => new Date(b.startDate) - new Date(a.startDate));

    let expandedCards = {};

    // Auto-expand the first record by default if it exists
    $: if (activeRecords.length > 0 && Object.keys(expandedCards).length === 0) {
        expandedCards[activeRecords[0].id] = true;
    }

    function toggleCard(id) {
        expandedCards[id] = !expandedCards[id];
        expandedCards = expandedCards;
    }

    function getSteps(record) {
        const step1Done = true; // Selalu true jika record ada
        const step2Done = record.reportStatus === 'Completed' || record.status === 'Approved' || record.paymentStatus === 'Paid';
        const step3Done = record.status === 'Approved' || record.paymentStatus === 'Paid';
        const step4Done = record.paymentStatus === 'Paid';

        return [
            { 
                id: 1, 
                title: 'Perjalanan Dinas', 
                description: 'Surat tugas diterbitkan', 
                isCompleted: step1Done, 
                isCurrent: !step2Done,
                icon: 'M0 0' // Placeholder, real SVG below
            },
            {
                id: 2,
                title: 'Laporan',
                description: 'Input rincian biaya & laporan',
                isCompleted: step2Done,
                isCurrent: step1Done && !step2Done,
                actionLabel: (!step2Done && $userStore.role !== 'kasubag') ? 'Input Laporan' : null,
                actionLink: '/dashboard/laporan'
            },
            { 
                id: 3, 
                title: 'Review Keuangan', 
                description: 'Menunggu approval', 
                isCompleted: step3Done, 
                isCurrent: step2Done && !step3Done
            },
            { 
                id: 4, 
                title: 'Billing Cair', 
                description: 'Pembayaran & kwitansi dicetak', 
                isCompleted: step4Done, 
                isCurrent: step3Done && !step4Done 
            }
        ];
    }
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between">
        <div>
            <h2 class="text-xl font-bold text-slate-900 tracking-tight">Roadmap Perjalanan Aktif</h2>
            <p class="text-sm text-slate-500 mt-1">Pantau progres pengajuan, pelaporan, hingga pencairan billing Anda.</p>
        </div>
    </div>

    {#if activeRecords.length === 0}
        <div class="bg-white rounded-xl border border-slate-200 border-dashed p-12 text-center shadow-sm">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-slate-300 mx-auto mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
            <h3 class="text-base font-semibold text-slate-800">Belum ada perjalanan</h3>
            <p class="text-sm text-slate-500 mt-1">Saat ada perjalanan dinas baru, progresnya akan muncul di sini.</p>
        </div>
    {:else}
        <div class="grid gap-4">
            {#each activeRecords as record (record.id)}
                {@const steps = getSteps(record)}
                {@const isFullyDone = record.paymentStatus === 'Paid'}
                {@const isExpanded = expandedCards[record.id]}
                <div class="bg-white rounded-2xl border {isFullyDone ? 'border-emerald-200/60 bg-emerald-50/10' : 'border-slate-200'} shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
                    {#if isFullyDone}
                        <div class="absolute top-0 right-0 bg-emerald-500 text-white text-[10px] font-bold px-3 py-1 rounded-bl-xl uppercase tracking-wider z-10">
                            Selesai
                        </div>
                    {/if}
                    
                    <!-- Card Header (Always Visible) -->
                    <div class="p-5 md:p-6 cursor-pointer select-none" on:click={() => toggleCard(record.id)}>
                        <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
                            <div class="flex-1 pr-8 md:pr-0">
                                <div class="flex flex-wrap items-center gap-2 mb-1.5">
                                    <span class="text-xs font-mono font-bold text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded-full border border-indigo-100">{record.nomorSpdPetugas || record.spd}</span>
                                    <span class="text-xs text-slate-400 font-medium">
                                        {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short'})} - {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                                    </span>
                                </div>
                                <h3 class="font-bold text-lg text-slate-900 leading-tight pr-6">{record.purpose}</h3>
                                <p class="text-sm text-slate-500 mt-1 flex items-center gap-1.5">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                    </svg>
                                    {toTitleCase(record.location)}, {toTitleCase(record.province)}
                                </p>
                            </div>
                            
                            <div class="flex items-center gap-4 shrink-0">
                                {#if record.paymentStatus !== 'Paid'}
                                    <div class="text-right">
                                        <div class="text-[10px] uppercase font-bold text-slate-400 tracking-wider mb-1 hidden md:block">Status Saat Ini</div>
                                        <div class="inline-flex items-center gap-2 bg-blue-50 border border-blue-100 px-3 py-1.5 rounded-lg text-blue-700 font-semibold text-xs md:text-sm">
                                            <span class="relative flex h-2.5 w-2.5">
                                                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                                                <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-blue-500"></span>
                                            </span>
                                            {steps.find(s => s.isCurrent)?.title || 'Menunggu'}
                                        </div>
                                    </div>
                                {/if}
                                <div class="p-1 rounded-md hover:bg-slate-100 transition-colors text-slate-400">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 transition-transform duration-300 {isExpanded ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                    </svg>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Roadmap Content (Collapsible) -->
                    {#if isExpanded}
                        <div class="px-5 md:px-6 pb-6 pt-2 border-t border-slate-100 bg-slate-50/30">
                            <!-- Stepper Desktop (Horizontal) -->
                            <div class="hidden md:block relative mt-6">
                                <div class="absolute left-6 top-5 w-[calc(100%-3rem)] h-1 bg-slate-100 rounded-full"></div>
                                <div class="absolute left-6 top-5 h-1 bg-indigo-500 rounded-full transition-all duration-700" style="width: {(steps.filter(s => s.isCompleted).length - 1) / (steps.length - 1) * 100}%"></div>
                                
                                <div class="flex justify-between relative z-10">
                                    {#each steps as step, i}
                                        <div class="flex flex-col items-center w-1/4 relative group">
                                            <div class="h-10 w-10 rounded-full flex items-center justify-center transition-all duration-500 {step.isCompleted ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/30' : (step.isCurrent ? 'bg-white border-2 border-indigo-500 text-indigo-600 ring-4 ring-indigo-50' : 'bg-white border-2 border-slate-200 text-slate-400')}">
                                                {#if step.isCompleted}
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                                        <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                                                    </svg>
                                                {:else}
                                                    <span class="text-sm font-bold">{i + 1}</span>
                                                {/if}
                                            </div>
                                            <div class="text-center mt-3">
                                                <div class="text-[11px] font-bold uppercase tracking-wider {step.isCompleted || step.isCurrent ? 'text-slate-800' : 'text-slate-400'}">{step.title}</div>
                                                <div class="text-[10px] text-slate-500 mt-0.5">{step.description}</div>
                                                
                                                {#if step.actionLabel}
                                                    <a href={step.actionLink} class="mt-2 inline-block px-3 py-1 text-[10px] font-bold uppercase tracking-wider text-white bg-indigo-600 hover:bg-indigo-700 rounded-md shadow-sm transition-colors cursor-pointer">
                                                        {step.actionLabel}
                                                    </a>
                                                {/if}
                                            </div>
                                        </div>
                                    {/each}
                                </div>
                            </div>

                            <!-- Stepper Mobile (Vertical) -->
                            <div class="md:hidden flex flex-col space-y-6 relative pl-4 mt-6">
                                <div class="absolute left-6 top-2 bottom-2 w-0.5 bg-slate-100"></div>
                                <div class="absolute left-6 top-2 w-0.5 bg-indigo-500 transition-all duration-700" style="height: {(steps.filter(s => s.isCompleted).length - 1) / (steps.length - 1) * 100}%"></div>
                                
                                {#each steps as step, i}
                                    <div class="flex items-start gap-4 relative z-10">
                                        <div class="h-5 w-5 mt-0.5 rounded-full flex shrink-0 items-center justify-center transition-all duration-500 {step.isCompleted ? 'bg-indigo-600 text-white shadow-md' : (step.isCurrent ? 'bg-white border-2 border-indigo-500 text-indigo-600 ring-2 ring-indigo-50' : 'bg-white border-2 border-slate-200 text-slate-400')}">
                                            {#if step.isCompleted}
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
                                                    <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                                                </svg>
                                            {:else}
                                                <div class="h-1.5 w-1.5 rounded-full {step.isCurrent ? 'bg-indigo-500' : 'bg-transparent'}"></div>
                                            {/if}
                                        </div>
                                        <div class="flex-1 pb-1">
                                            <div class="text-xs font-bold uppercase tracking-wider {step.isCompleted || step.isCurrent ? 'text-slate-800' : 'text-slate-400'}">{step.title}</div>
                                            <div class="text-[11px] text-slate-500 mt-0.5">{step.description}</div>
                                            
                                            {#if step.actionLabel}
                                                <a href={step.actionLink} class="mt-2 inline-block px-3 py-1.5 text-[10px] font-bold uppercase tracking-wider text-white bg-indigo-600 hover:bg-indigo-700 rounded-md shadow-sm transition-colors cursor-pointer">
                                                    {step.actionLabel}
                                                </a>
                                            {/if}
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        </div>
                    {/if}
                </div>
            {/each}
        </div>
    {/if}
</div>