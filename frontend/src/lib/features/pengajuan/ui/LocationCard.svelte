<script>
    import { createEventDispatcher } from 'svelte';
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import Textarea from '$lib/shared/ui/textarea/Textarea.svelte';
    import { toTitleCase } from '$lib/shared/utils/utils';
    import regenciesData from '$lib/shared/assets/regencies.json';

    export let locations = [];
    export let purpose = '';
    export let agenda = '';
    export let provinces = [];
    export let readonly = false;

    const dispatch = createEventDispatcher();



    function addLocation() {
        dispatch('add');
    }

    /** @param {number} index */
    function removeLocation(index) {
        dispatch('remove', index);
    }

    /** @param {string} provinceName */
    function getRegencies(provinceName) {
        const found = regenciesData.find(p => p.province === provinceName.toUpperCase());
        return found ? found.regencies : [];
    }
</script>

<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
        <h3 class="font-semibold text-slate-800 flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Lokasi & Tujuan
        </h3>
    </div>
    
    <div class="p-6 space-y-6">
        <!-- Dynamic Location Rows -->
        <div class="space-y-6">
            {#each locations as loc, i}
                <div class="space-y-4 p-4 rounded-xl border border-slate-100 bg-slate-50/30 relative group">
                    {#if locations.length > 1 && !readonly}
                        <button 
                            type="button"
                            on:click={() => removeLocation(i)}
                            class="absolute -top-2 -right-2 p-1.5 bg-white border border-red-100 text-red-400 rounded-full opacity-0 group-hover:opacity-100 transition-opacity shadow-sm hover:text-red-600 z-10"
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                                <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                            </svg>
                        </button>
                    {/if}

                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        <div class="space-y-2">
                            <Label class="text-slate-600 text-sm">Provinsi *</Label>
                            <Select bind:value={loc.province} class="h-10 text-sm {readonly ? 'opacity-70 cursor-not-allowed' : ''}" disabled={readonly} on:change={() => loc.location = ''}>
                                <option value="" disabled selected>Pilih Provinsi</option>
                                {#each provinces as prov}
                                    <option value={toTitleCase(prov.name)}>{toTitleCase(prov.name)}</option>
                                {/each}
                            </Select>
                        </div>
                        <div class="space-y-2">
                            <Label class="text-slate-600 text-sm">Lokasi Dinas (Kab/Kota)</Label>
                            {#if loc.province && getRegencies(loc.province).length > 0}
                                <Select bind:value={loc.location} class="h-10 text-sm {readonly ? 'opacity-70 cursor-not-allowed' : ''}" disabled={readonly}>
                                    <option value="" disabled selected>Pilih Kab/Kota</option>
                                    {#each getRegencies(loc.province) as regency}
                                        <option value={toTitleCase(regency)}>{toTitleCase(regency)}</option>
                                    {/each}
                                </Select>
                            {:else}
                                <Input placeholder="Pilih provinsi terlebih dahulu" bind:value={loc.location} class="h-10 text-sm {readonly ? 'opacity-70 cursor-not-allowed' : ''}" disabled={readonly || !loc.province} />
                            {/if}
                        </div>
                    </div>

                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        <div class="space-y-1">
                            <Label class="text-slate-500 text-[10px] uppercase tracking-wider font-bold">Tanggal Mulai</Label>
                            <Input type="date" bind:value={loc.startDate} min={i > 0 ? (locations[i-1].endDate || locations[i-1].startDate || '') : ''} class="h-9 text-xs {readonly ? 'opacity-70' : ''}" disabled={readonly} />
                        </div>
                        <div class="space-y-1">
                            <Label class="text-slate-500 text-[10px] uppercase tracking-wider font-bold">Tanggal Selesai</Label>
                            <Input type="date" bind:value={loc.endDate} min={loc.startDate || (i > 0 ? (locations[i-1].endDate || locations[i-1].startDate || '') : '')} class="h-9 text-xs {readonly ? 'opacity-70' : ''}" disabled={readonly} />
                        </div>
                    </div>
                </div>
            {/each}
        </div>

        <!-- Extend Button -->
        {#if !readonly}
            <div class="space-y-2">
                <Label class="text-slate-600 font-medium text-sm">Extend</Label>
                <button 
                    type="button" 
                    on:click={addLocation}
                    class="h-10 w-10 flex items-center justify-center bg-blue-600 text-white rounded-full shadow-md hover:bg-blue-700 transition-all active:scale-95"
                    title="Tambah Lokasi"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                </button>
            </div>
        {/if}

        <hr class="border-slate-100" />

        <!-- Purpose Section -->
        <div class="space-y-3">
            <Label class="text-slate-600 font-medium">Tujuan Perjalanan</Label>
            <div class="grid grid-cols-1 gap-2">
                <label class="relative flex items-center p-3 rounded-xl border border-slate-200 {readonly ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-slate-50'} transition-colors has-[:checked]:border-blue-500 has-[:checked]:bg-blue-50/30">
                    <input type="radio" name="purpose" value="persiapan" bind:group={purpose} disabled={readonly} class="accent-blue-600 h-4 w-4 mr-3" />
                    <span class="text-sm font-medium text-slate-700">Persiapan dan Pendampingan Kunjungan Kerja</span>
                </label>
                <label class="relative flex items-center p-3 rounded-xl border border-slate-200 {readonly ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-slate-50'} transition-colors has-[:checked]:border-blue-500 has-[:checked]:bg-blue-50/30">
                    <input type="radio" name="purpose" value="koordinasi" bind:group={purpose} disabled={readonly} class="accent-blue-600 h-4 w-4 mr-3" />
                    <span class="text-sm font-medium text-slate-700">Koordinasi dan Konsultasi Kunjungan Kerja</span>
                </label>
            </div>
        </div>

        <!-- Agenda Section -->
        <div class="space-y-2">
            <Label class="text-slate-600 font-medium">Detail Agenda</Label>
            <Textarea placeholder="Jelaskan detail agenda kegiatan secara singkat..." bind:value={agenda} class="min-h-[100px] text-sm resize-y {readonly ? 'opacity-70 cursor-not-allowed' : ''}" disabled={readonly} />
        </div>
    </div>
</div>
