<script>
    import { userStore } from '$lib/features/auth/store';
    import { createEventDispatcher } from 'svelte';

    export let record;

    const dispatch = createEventDispatcher();

    $: steps = getSteps(record);

    function getSteps(record) {
        if (!record) return [];
        const step1Done = true; // Selalu true jika record ada
        const step2Done = record.reportStatus === 'Completed' || record.status === 'Approved' || record.paymentStatus === 'Paid';
        const step3Done = record.status === 'Approved' || record.paymentStatus === 'Paid';
        const step4Done = record.paymentStatus === 'Paid';

        return [
            { 
                id: 1, 
                title: 'Perjalanan Dinas', 
                description: 'Surat tugas & petugas ditugaskan', 
                isCompleted: step1Done, 
                isCurrent: !step2Done
            },
            {
                id: 2,
                title: 'Laporan',
                description: 'Input rincian biaya & laporan',
                isCompleted: step2Done,
                isCurrent: step1Done && !step2Done,
                actionLabel: (!step2Done && $userStore.role !== 'kasubag') ? 'Input Laporan' : null,
                actionLink: !step2Done ? `/dashboard/laporan/${record.spd}` : null
            },
            { 
                id: 3, 
                title: 'Review Keuangan', 
                description: 'Review & approve rincian biaya', 
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

<div class="w-full">
    <!-- Stepper Desktop (Horizontal) -->
    <div class="hidden md:block relative mt-8 mb-8 px-4 w-full">
        <!-- Base line -->
        <div class="absolute left-10 right-10 top-5 h-1 bg-slate-200 rounded-full" style="width: calc(100% - 5rem);"></div>
        <!-- Active line -->
        <div class="absolute left-10 top-5 h-1 bg-indigo-500 rounded-full transition-all duration-700" style="width: calc({(steps.filter(s => s.isCompleted).length - 1) / (steps.length - 1) * 100}% - {steps.filter(s => s.isCompleted).length > 1 ? '5rem' : '0rem'} * {(steps.filter(s => s.isCompleted).length - 1) / (steps.length - 1)}); min-width: {(steps.filter(s => s.isCompleted).length - 1) / (steps.length - 1) * 100 > 0 ? 'calc(' + ((steps.filter(s => s.isCompleted).length - 1) / (steps.length - 1) * 100) + '% - 5rem)' : '0px'}"></div>
        
        <div class="flex justify-between relative z-10 w-full">
            {#each steps as step, i}
                <div class="flex flex-col items-center w-1/4 relative group">
                    <div class="h-10 w-10 bg-white rounded-full p-1 z-10">
                        <div class="h-full w-full rounded-full flex items-center justify-center transition-all duration-500 {step.isCompleted ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/30' : (step.isCurrent ? 'bg-white border-2 border-indigo-500 text-indigo-600 ring-4 ring-indigo-50' : 'bg-slate-100 border-2 border-slate-200 text-slate-400')}">
                            {#if step.isCompleted}
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                                    <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                                </svg>
                            {:else}
                                <span class="text-sm font-bold">{i + 1}</span>
                            {/if}
                        </div>
                    </div>
                    <div class="text-center mt-3">
                        <div class="text-[11px] font-bold uppercase tracking-wider {step.isCompleted || step.isCurrent ? 'text-slate-800' : 'text-slate-400'} leading-tight">{step.title}</div>
                        <div class="text-[10px] text-slate-500 mt-1 leading-tight px-2">{step.description}</div>
                        
                        {#if step.actionLabel && !step.isCompleted}
                            {#if step.actionEvent}
                                <button type="button" on:click={(e) => { e.stopPropagation(); dispatch(step.actionEvent); }} class="mt-2.5 inline-flex items-center gap-1 px-3 py-1.5 text-[10px] font-bold uppercase tracking-wider text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg shadow-sm transition-colors cursor-pointer">
                                    {step.actionLabel}
                                </button>
                            {:else}
                                <a href={step.actionLink} on:click|stopPropagation class="mt-2.5 inline-flex items-center gap-1 px-3 py-1.5 text-[10px] font-bold uppercase tracking-wider text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg shadow-sm transition-colors cursor-pointer">
                                    {step.actionLabel}
                                </a>
                            {/if}
                        {/if}
                    </div>
                </div>
            {/each}
        </div>
    </div>

    <!-- Stepper Mobile (Vertical) -->
    <div class="md:hidden flex flex-col space-y-6 relative pl-4 mt-6 mb-6">
        <div class="absolute left-6 top-2 bottom-2 w-0.5 bg-slate-200"></div>
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
                    
                    {#if step.actionLabel && !step.isCompleted}
                        {#if step.actionEvent}
                            <button type="button" on:click={(e) => { e.stopPropagation(); dispatch(step.actionEvent); }} class="mt-2 inline-block px-3 py-1.5 text-[10px] font-bold uppercase tracking-wider text-white bg-indigo-600 hover:bg-indigo-700 rounded-md shadow-sm transition-colors cursor-pointer">
                                {step.actionLabel}
                            </button>
                        {:else}
                            <a href={step.actionLink} on:click|stopPropagation class="mt-2 inline-block px-3 py-1.5 text-[10px] font-bold uppercase tracking-wider text-white bg-indigo-600 hover:bg-indigo-700 rounded-md shadow-sm transition-colors cursor-pointer">
                                {step.actionLabel}
                            </a>
                        {/if}
                    {/if}
                </div>
            </div>
        {/each}
    </div>
</div>