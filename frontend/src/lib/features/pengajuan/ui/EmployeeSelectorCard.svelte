<script>
    import { createEventDispatcher } from 'svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';

    /** @type {any[]} */
    export let employees = [];
    /** @type {any[]} */
    export let selectedEmployees = [];
    /** @type {string[]} */
    export let disabledIds = [];
    /** @type {boolean} */
    export let readonly = false;
    /** @type {boolean} */
    export let isLoading = false;

    const dispatch = createEventDispatcher();
    
    let searchQuery = '';

    $: filteredEmployees = employees.filter(e => 
        (e.name || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
        (e.nip || '').includes(searchQuery) ||
        (e.jabatan || '').toLowerCase().includes(searchQuery.toLowerCase())
    );

    /** @param {any} id */
    function toggleEmployee(id) {
        if (readonly) return;
        dispatch('toggle', id);
    }
</script>

<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden flex flex-col h-[500px]">
    <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center">
        <h3 class="font-semibold text-slate-800">Pilih Petugas</h3>
        <span class="text-xs font-medium px-2 py-1 bg-blue-100 text-blue-700 rounded-full">
            {selectedEmployees.length}/{employees.length}
        </span>
    </div>
    <div class="p-3 border-b border-slate-100">
        <div class="relative">
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
            </div>
            <Input type="text" placeholder="Cari nama, NIP, atau jabatan..." bind:value={searchQuery} class="pl-9 bg-slate-50/50 border-slate-200 text-sm h-9" />
        </div>
    </div>
    <div class="p-4 overflow-y-auto flex-1 space-y-2">
        {#if isLoading}
            <div class="space-y-3 animate-pulse">
                {#each Array(5) as _}
                <div class="flex items-center space-x-3 p-3 rounded-lg border border-slate-100 bg-slate-50/50">
                    <div class="h-5 w-5 bg-slate-200 rounded shrink-0"></div>
                    <div class="grid gap-2 w-full">
                        <div class="h-4 bg-slate-200 rounded w-1/2"></div>
                        <div class="h-3 bg-slate-200 rounded w-3/4"></div>
                        <div class="h-2 bg-slate-200 rounded w-1/3 mt-1"></div>
                    </div>
                </div>
                {/each}
            </div>
        {:else if filteredEmployees.length === 0}
            <div class="text-center py-6 text-sm text-slate-500 italic">Tidak ada petugas yang cocok.</div>
        {:else}
            {#each filteredEmployees as employee (employee.id)}
                {@const isDisabled = disabledIds.includes(employee.id)}
                <label class="flex items-center space-x-3 p-3 rounded-lg border border-transparent {readonly || isDisabled ? 'cursor-default opacity-50 bg-slate-50' : 'hover:border-slate-200 hover:bg-slate-50 cursor-pointer'} transition-all has-[:checked]:bg-blue-50/30 has-[:checked]:border-blue-200">
                    <input type="checkbox"
                        checked={selectedEmployees.includes(employee.id)}
                        on:change={() => !isDisabled && toggleEmployee(employee.id)}
                        disabled={readonly || isDisabled}
                        class="accent-blue-600 h-5 w-5 rounded border-slate-300 shrink-0 {readonly || isDisabled ? 'opacity-50' : ''}"
                    />
                    <div class="grid gap-0.5">
                        <span class="text-sm font-semibold text-slate-800 flex items-center gap-2">
                            {employee.name}
                            {#if isDisabled}
                                <span class="text-[10px] text-red-500 bg-red-50 px-1.5 py-0.5 rounded border border-red-100 font-medium">Sibuk</span>
                            {/if}
                        </span>
                        {#if employee.jabatan}
                            <span class="text-xs font-medium text-slate-700">{employee.jabatan}</span>
                        {/if}
                        {#if (employee.pangkat && employee.pangkat !== '-') || (employee.golongan && employee.golongan !== '-')}
                            <span class="text-xs text-slate-500">
                                {#if (employee.pangkat && employee.pangkat !== '-') && (employee.golongan && employee.golongan !== '-')}
                                    {employee.pangkat} ({employee.golongan})
                                {:else}
                                    {(employee.pangkat !== '-' ? employee.pangkat : '') || (employee.golongan !== '-' ? employee.golongan : '')}
                                {/if}
                            </span>
                        {/if}
                        <span class="text-[10px] text-slate-400 font-mono">NIP. {employee.nip || '-'}</span>
                    </div>
                </label>
            {/each}
        {/if}
    </div>
    {#if !readonly}
    <div class="p-4 border-t border-slate-100 bg-slate-50/50">
        <p class="text-xs text-slate-500 text-center mb-3">Pastikan data sudah benar sebelum menyimpan.</p>
        <Button class="w-full bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20" on:click={() => dispatch('submit')}>
            Simpan Pengajuan
        </Button>
    </div>
    {/if}
</div>
