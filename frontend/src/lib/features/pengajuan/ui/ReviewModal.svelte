<script>
    import { createEventDispatcher } from 'svelte';
    import { portal } from '$lib/shared/actions/portal';
    import { fade, fly } from 'svelte/transition';
    
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import Textarea from '$lib/shared/ui/textarea/Textarea.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';
    import CostEstimateCard from '$lib/features/pengajuan/ui/CostEstimateCard.svelte';
    import ConfirmationModal from '$lib/shared/ui/confirmation-modal/ConfirmationModal.svelte';
    
    import { api } from '$lib/shared/api';
    import { updateRecord, deleteRecord, addRecord } from '$lib/features/pengajuan/store';
    import { userStore } from '$lib/features/auth/store';
    import { toast } from '$lib/shared/stores/toast';
    import { formatCurrency, getStatusBadge, toTitleCase } from '$lib/shared/utils/utils';
    import regenciesData from '$lib/shared/assets/regencies.json';
    
    export let open = false;
    export let records = []; // All records for the selected SPD
    export let provinces = [];
    export let stakeholders = [];
    
    const dispatch = createEventDispatcher();
    
    let isEditing = false;
    let isLoading = false;
    
    /** @type {any[]} */
    let protokolOfficers = [];
    let selectedEmployeeIds = [];
    let initialEmployeeIds = [];
    let prevOpen = false;
    
    let showSpdConfirmModal = false;
    
    let searchQuery = '';
    $: filteredOfficers = protokolOfficers.filter(officer => 
        (officer.name || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
        (officer.nip || '').includes(searchQuery) ||
        (officer.jabatan || '').toLowerCase().includes(searchQuery.toLowerCase())
    );
    
    // Form state initialized when records change
    let formData = {
        locations: [],
        purpose: '',
        stakeholder: '',
        agenda: '',
        suratTugasNumber: ''
    };

    /** @param {string} provinceName */
    function getRegencies(provinceName) {
        if (!provinceName) return [];
        const found = regenciesData.find(p => p.province === provinceName.toUpperCase());
        return found ? found.regencies : [];
    }

    function addLocation() {
        formData.locations = [
            ...formData.locations,
            { startDate: '', endDate: '', location: '', province: '' }
        ];
    }

    function removeLocation(index) {
        if (formData.locations.length <= 1) return;
        formData.locations = formData.locations.filter((_, i) => i !== index);
    }
    
    $: if (open && !prevOpen) {
        prevOpen = true;
        // Fetch officers dynamically when modal opens
        api.getUsers({ role: 'protokol' }).then(users => protokolOfficers = users).catch(console.error);
        
        if (records.length > 0) {
            const firstRecord = records[0];
            // Format dates to YYYY-MM-DD for input type="date"
            const formatForInput = (isoString) => {
                if (!isoString) return '';
                const timestamp = Date.parse(isoString);
                return isNaN(timestamp) ? '' : new Date(timestamp).toISOString().split('T')[0];
            };
            
            formData = {
                locations: firstRecord.locations && firstRecord.locations.length > 0
                    ? firstRecord.locations.map(l => ({
                        startDate: formatForInput(l.startDate),
                        endDate: formatForInput(l.endDate),
                        location: l.location || '',
                        province: l.province || ''
                    }))
                    : [{
                        startDate: formatForInput(firstRecord.startDate),
                        endDate: formatForInput(firstRecord.endDate),
                        location: firstRecord.location || '',
                        province: firstRecord.province || ''
                    }],
                purpose: firstRecord.purpose || '',
                stakeholder: firstRecord.stakeholder || '',
                agenda: firstRecord.agenda || '',
                suratTugasNumber: firstRecord.suratTugasNumber || ''
            };
            
            selectedEmployeeIds = records.map(r => r.employee?.id).filter(Boolean);
            initialEmployeeIds = [...selectedEmployeeIds];
            isEditing = false;
        }
    } else if (!open && prevOpen) {
        prevOpen = false;
    }
    
    $: baseRecord = records[0] || {};
    $: isEditable = ((baseRecord.status === 'Draft' || baseRecord.status === 'Submitted' || baseRecord.status === 'In Progress') && $userStore?.role !== 'kasubag') || $userStore?.role === 'super_admin';
    
    // Calculate SBM Cost
	$: costBreakdown = Object.values((formData.locations && formData.locations.length > 0
		? formData.locations
		: []
		).reduce((acc, loc) => {
			const provData = provinces.find(p => p.name === loc.province);
			const rate = provData ? provData.luarKota : 0;
			const start = Date.parse(loc.startDate);
			const end = Date.parse(loc.endDate);
			let locDays = 0;
			if (!isNaN(start) && !isNaN(end)) {
				const diffTime = end - start;
				const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
				locDays = diffDays > 0 ? diffDays : 0;
			}
            
            if (locDays > 0 && loc.province) {
                if (!acc[loc.province]) {
                    acc[loc.province] = { id: loc.province, province: loc.province, days: 0, rate: rate };
                }
                acc[loc.province].days += locDays;
            }
			return acc;
		}, {}));

	$: days = costBreakdown.reduce((sum, item) => sum + item.days, 0);
	$: totalCostPerPerson = costBreakdown.reduce((sum, item) => sum + (item.rate * item.days), 0);
	$: sbmRateAvg = days > 0 ? totalCostPerPerson / days : 0;

    // Overall start/end for the record
    $: minStartDate = formData.locations.reduce((min, loc) => {
        if (!loc.startDate) return min;
        if (!min || loc.startDate < min) return loc.startDate;
        return min;
    }, '');

    $: maxEndDate = formData.locations.reduce((max, loc) => {
        if (!loc.endDate) return max;
        if (!max || loc.endDate > max) return loc.endDate;
        return max;
    }, '');

	// When editing, cost is based on new selection. Otherwise based on records length.
	$: currentEmployeeCount = isEditing ? selectedEmployeeIds.length : records.length;
	$: totalSBMCost = totalCostPerPerson * currentEmployeeCount;

    let newSuratTugasFile = null;

    function handleFileChange(event) {
        const file = event.target.files[0];
        if (file) {
            newSuratTugasFile = file;
        }
    }

    async function saveSuratTugasFileOnly() {
        if (!newSuratTugasFile) return;

        isLoading = true;
        try {
            const res = await api.uploadFile(newSuratTugasFile);
            const uploadedPath = res.path;

            // Update all records in the SPD group
            for (const record of records) {
                await updateRecord(record.id, {
                    suratTugasPath: uploadedPath
                });
            }

            toast.success('Surat Tugas berhasil diperbarui.');
            newSuratTugasFile = null;
            dispatch('saved');
        } catch (e) {
            toast.error('Gagal memperbarui Surat Tugas.');
        } finally {
            isLoading = false;
        }
    }

	async function saveSuratTugasNumberOnly() {
        if (!formData.suratTugasNumber) {
            toast.error('Nomor Surat Tugas tidak boleh kosong.');
            return;
        }

        isLoading = true;
        try {
            // Update all records in the SPD group
            for (const record of records) {
                await updateRecord(record.id, {
                    suratTugasNumber: formData.suratTugasNumber
                });
            }

            toast.success('Nomor Surat Tugas berhasil diperbarui.');
            dispatch('saved');
            // We don't close the modal, just let them see the updated state
        } catch (e) {
            toast.error('Gagal memperbarui Nomor Surat Tugas.');
        } finally {
            isLoading = false;
        }
    }

	async function saveSpdNumberOnly() {
        if (!formData.spdNumberInput) {
            toast.error('Nomor SPD tidak boleh kosong jika ingin diubah.');
            return;
        }

        isLoading = true;
        try {
            const finalSpd = `ID-SPJ-${String(formData.spdNumberInput).padStart(3, '0')}`;
            
            // Update all records in the SPD group
            for (const record of records) {
                await updateRecord(record.id, {
                    spd: finalSpd
                });
            }

            toast.success('Nomor SPJ berhasil diperbarui.');
            
            // Beri waktu backend untuk menyinkronkan sequence number di background 
            // sembari menahan state loading di tombol agar UX lebih mulus
            await new Promise(resolve => setTimeout(resolve, 600));
            
            dispatch('saved');
            close();
        } catch (e) {
            toast.error('Gagal memperbarui Nomor SPD.');
        } finally {
            isLoading = false;
        }
    }

	async function handleSave() {
        if (formData.locations.some(loc => !loc.startDate || !loc.endDate || !loc.province)) {
            toast.error('Harap lengkapi field wajib di setiap lokasi.');
            return;
        }
        
        if (days <= 0) {
            toast.error('Tanggal selesai tidak boleh mendahului tanggal mulai.');
            return;
        }

        if (selectedEmployeeIds.length === 0) {
            toast.error('Pilih minimal satu petugas.');
            return;
        }

        isLoading = true;
        try {
            let finalSpd = baseRecord.spd;
            if (formData.spdNumberInput && formData.spdNumberInput !== (baseRecord.spd ? baseRecord.spd.replace('ID-SPJ-', '') : '')) {
                finalSpd = `ID-SPJ-${String(formData.spdNumberInput).padStart(3, '0')}`;
            }

            const commonData = {
                ...formData,
                spd: finalSpd, // Include potentially edited SPD
                startDate: minStartDate ? new Date(minStartDate).toISOString() : null,
                endDate: maxEndDate ? new Date(maxEndDate).toISOString() : null,
                // Backward compatibility for summary
                location: formData.locations[0]?.location || '',
                province: formData.locations[0]?.province || '',
                totalCost: totalCostPerPerson // Store per-person cost
            };

            // 1. Delete removed employees
            const toDeleteIds = initialEmployeeIds.filter(id => !selectedEmployeeIds.includes(id));
            for (const id of toDeleteIds) {
                const recordToDelete = records.find(r => r.employee.id === id);
                if (recordToDelete) {
                    await deleteRecord(recordToDelete.id);
                }
            }

            // 2. Update existing employees
            const toUpdateIds = initialEmployeeIds.filter(id => selectedEmployeeIds.includes(id));
            for (const id of toUpdateIds) {
                const recordToUpdate = records.find(r => r.employee.id === id);
                if (recordToUpdate) {
                    await updateRecord(recordToUpdate.id, commonData);
                }
            }

            // 3. Add new employees
            const toAddIds = selectedEmployeeIds.filter(id => !initialEmployeeIds.includes(id));
            if (toAddIds.length > 0) {
                const newEmployees = protokolOfficers.filter(u => toAddIds.includes(u.id));
                const tripData = {
                    spd: baseRecord.spd,
                    email: $userStore.email,
                    type: baseRecord.type,
                    ...commonData,
                    employees: newEmployees
                };
                await addRecord(tripData);
            }

            toast.success('Pengajuan berhasil diperbarui.');
            
            // Beri waktu backend untuk menyinkronkan sequence number di background
            // sembari menahan state loading di tombol
            await new Promise(resolve => setTimeout(resolve, 600));
            
            isEditing = false;
            dispatch('saved');
            close();
        } catch (e) {
            toast.error('Gagal memperbarui pengajuan.');
        } finally {
            isLoading = false;
        }
    }
    
    function close() {
        open = false;
        isEditing = false;
        dispatch('close');
    }
</script>

{#if open && records.length > 0}
  <div use:portal>
    <!-- Background Overlay -->
    <div class="fixed inset-0 z-[100] bg-slate-900/90" role="button" tabindex="0" aria-label="Close modal" on:click={close} on:keydown={(e) => e.key === 'Escape' && close()}></div>
    
    <!-- Modal Dialog -->
    <div class="fixed left-[50%] top-[50%] z-[100] w-full max-w-4xl translate-x-[-50%] translate-y-[-50%] border border-slate-200 bg-white shadow-2xl sm:rounded-2xl overflow-hidden max-h-[90vh] flex flex-col">
        <!-- Header -->
        <div class="px-6 py-5 border-b border-slate-100 flex justify-between items-start bg-white">
            <div class="flex-1 min-w-0 pr-4">
                <div class="flex flex-wrap items-center gap-2 sm:gap-3 mb-1.5">
                    <h2 class="text-lg sm:text-xl font-bold text-slate-800 leading-tight">Detail Pengajuan</h2>
                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {getStatusBadge(baseRecord).class}">
                        {getStatusBadge(baseRecord).label}
                    </span>
                </div>
                <div class="flex flex-wrap items-center gap-2 text-sm text-slate-500">
                    <span class="font-mono text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded border border-indigo-100 text-xs sm:text-sm">{baseRecord.spd}</span>
                    <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold capitalize tracking-wide bg-slate-100 text-slate-600 border border-slate-200">
                        {baseRecord.type ? baseRecord.type.replace(/_/g, ' ') : 'Dalam Kota'}
                    </span>
                </div>
            </div>
            <button class="rounded-full p-2 hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-colors" on:click={close} aria-label="Close">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-5 w-5"><line x1="18" x2="6" y1="6" y2="18"></line><line x1="6" x2="18" y1="6" y2="18"></line></svg>
            </button>
        </div>
        
        <!-- Body -->
        <div class="flex-1 overflow-y-auto p-6 bg-slate-50/50">
            <div class="grid grid-cols-1 lg:grid-cols-5 gap-8">
                <!-- Left Column: Details -->
                <div class="lg:col-span-3 space-y-6">
                    <div class="bg-white rounded-xl border border-slate-200 p-5 shadow-sm">
                        <h3 class="text-sm font-bold text-slate-800 flex items-center gap-2 mb-5">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            Informasi Dasar
                        </h3>
                        
                        <div class="space-y-6">
                            {#each formData.locations as loc, i}
                                <div class="space-y-4 p-4 rounded-xl border border-slate-100 bg-slate-50/30 relative group">
                                    {#if formData.locations.length > 1 && isEditing}
                                        <button 
                                            type="button"
                                            on:click={() => removeLocation(i)}
                                            class="absolute -top-2 -right-2 p-1 bg-white border border-red-100 text-red-400 rounded-full shadow-sm hover:text-red-600 z-10"
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                                                <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                            </svg>
                                        </button>
                                    {/if}

                                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                        <div class="space-y-1.5">
                                            <Label class="text-slate-600 text-xs">Provinsi *</Label>
                                            <Select bind:value={loc.province} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-blue-200 focus:border-blue-500'} on:change={() => loc.location = ''}>
                                                <option value="" disabled>Pilih Provinsi</option>
                                                {#each provinces as prov}
                                                    <option value={toTitleCase(prov.name)}>{toTitleCase(prov.name)}</option>
                                                {/each}
                                            </Select>
                                        </div>
                                        <div class="space-y-1.5">
                                            <Label class="text-slate-600 text-xs">Lokasi Dinas (Kab/Kota)</Label>
                                            {#if isEditing && loc.province && getRegencies(loc.province).length > 0}
                                                <Select bind:value={loc.location} class="h-10 text-sm bg-white border-blue-200 focus:border-blue-500" disabled={!isEditing}>
                                                    <option value="" disabled selected>Pilih Kab/Kota</option>
                                                    {#each getRegencies(loc.province) as regency}
                                                        <option value={toTitleCase(regency)}>{toTitleCase(regency)}</option>
                                                    {/each}
                                                </Select>
                                            {:else}
                                                <Input value={toTitleCase(loc.location)} on:input={(e) => loc.location = e.target.value} disabled={!isEditing || !loc.province} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-blue-200 focus:border-blue-500'} placeholder={!loc.province ? 'Pilih provinsi terlebih dahulu' : 'Contoh: Surabaya'} />
                                            {/if}
                                        </div>
                                    </div>

                                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                        <div class="space-y-1.5">
                                            <Label class="text-slate-600 text-xs">Tanggal Mulai *</Label>
                                            <Input type="date" bind:value={loc.startDate} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-blue-200 focus:border-blue-500'} />
                                        </div>
                                        <div class="space-y-1.5">
                                            <Label class="text-slate-600 text-xs">Tanggal Selesai *</Label>
                                            <Input type="date" bind:value={loc.endDate} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-blue-200 focus:border-blue-500'} />
                                        </div>
                                    </div>
                                </div>
                            {/each}

                            {#if isEditing}
                                <div class="flex justify-center pt-2">
                                    <Button variant="outline" size="sm" class="text-indigo-600 border-indigo-200 hover:bg-indigo-50 flex items-center gap-1.5 h-8" on:click={addLocation}>
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                                        </svg>
                                        Tambah Lokasi
                                    </Button>
                                </div>
                            {/if}

                            <div class="space-y-1.5 pt-2">
                                <Label class="text-slate-600 text-xs flex justify-between items-center">
                                    <span>Nomor SPJ</span>
                                    {#if !isEditing && isEditable && formData.spdNumberInput !== (baseRecord.spd ? baseRecord.spd.replace('ID-SPJ-', '') : '')}
                                        <button 
                                            on:click={() => showSpdConfirmModal = true}
                                            class="text-[10px] text-emerald-600 hover:text-emerald-700 font-bold uppercase tracking-tight flex items-center gap-1"
                                            disabled={isLoading}
                                        >
                                            {#if isLoading}
                                                <svg class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                                </svg>
                                            {:else}
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" /></svg>
                                            {/if}
                                            Simpan Nomor
                                        </button>
                                    {/if}
                                </Label>
                                <div class="flex items-center">
                                    <span class="inline-flex items-center px-3 border border-r-0 border-slate-300 bg-slate-100 text-slate-500 text-sm rounded-l-md h-10">
                                        ID-SPJ-
                                    </span>
                                    <Input 
                                        type="number"
                                        min="1"
                                        bind:value={formData.spdNumberInput} 
                                        disabled={!isEditing && !isEditable} 
                                        class="flex-1 rounded-none rounded-r-md h-10 border-slate-300 focus:ring-emerald-500 focus:border-emerald-500 {(!isEditing && !isEditable) ? 'bg-slate-50 opacity-70 cursor-default' : (isEditing ? 'bg-white' : 'bg-emerald-50/30')}" 
                                        placeholder="Cth: 005" 
                                    />
                                </div>
                                {#if !isEditing && isEditable}
                                    <p class="text-[10px] text-slate-400 italic">Isi untuk mengganti nomor SPJ.</p>
                                {/if}
                            </div>

                            <div class="space-y-1.5 pt-2">
                                <Label class="text-slate-600 text-xs flex justify-between items-center">
                                    <span>Surat Tugas</span>
                                    {#if isEditable && newSuratTugasFile}
                                        <button 
                                            on:click={saveSuratTugasFileOnly}
                                            class="text-[10px] text-emerald-600 hover:text-emerald-700 font-bold uppercase tracking-tight flex items-center gap-1"
                                            disabled={isLoading}
                                        >
                                            {#if isLoading}
                                                <LottieLoader size="32px" className="brightness-0 invert" />
                                            {:else}
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" /></svg>
                                            {/if}
                                            Simpan File
                                        </button>
                                    {/if}
                                </Label>
                                
                                {#if baseRecord.suratTugasPath || newSuratTugasFile}
                                    <div class="flex items-center justify-between p-2.5 rounded-lg border border-slate-200 bg-slate-50">
                                        <div class="flex items-center gap-2 overflow-hidden">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-rose-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                            </svg>
                                            <span class="text-sm font-medium text-slate-700 truncate" title={newSuratTugasFile ? newSuratTugasFile.name : baseRecord.suratTugasPath}>
                                                {newSuratTugasFile ? newSuratTugasFile.name : baseRecord.suratTugasPath}
                                            </span>
                                        </div>
                                        <div class="flex items-center gap-2 shrink-0">
                                            {#if baseRecord.suratTugasPath && !newSuratTugasFile}
                                                <a href={`/uploads/${baseRecord.suratTugasPath}`} target="_blank" rel="noopener noreferrer" class="inline-flex items-center justify-center px-3 py-1.5 text-xs font-medium text-blue-700 bg-blue-100 hover:bg-blue-200 rounded-md transition-colors">
                                                    Lihat
                                                </a>
                                            {/if}
                                            {#if isEditable}
                                                <label class="cursor-pointer inline-flex items-center justify-center px-3 py-1.5 text-xs font-medium text-emerald-700 bg-emerald-100 hover:bg-emerald-200 rounded-md transition-colors">
                                                    {baseRecord.suratTugasPath ? 'Ganti' : 'Pilih'}
                                                    <input type="file" class="hidden" accept="image/*,application/pdf" on:change={handleFileChange} />
                                                </label>
                                            {/if}
                                        </div>
                                    </div>
                                {:else}
                                    <div class="flex flex-col items-center justify-center gap-2 p-4 border border-dashed border-slate-200 rounded-lg bg-slate-50 text-center">
                                        <p class="text-xs text-slate-500 italic">Tidak ada Surat Tugas yang dilampirkan.</p>
                                        {#if isEditable}
                                            <label class="cursor-pointer inline-flex items-center justify-center px-4 py-2 text-xs font-bold text-blue-600 bg-white border border-blue-200 hover:bg-blue-50 rounded-lg shadow-sm transition-all active:scale-95 uppercase tracking-wide">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
                                                Upload Surat Tugas
                                                <input type="file" class="hidden" accept="image/*,application/pdf" on:change={handleFileChange} />
                                            </label>
                                        {/if}
                                    </div>
                                {/if}
                            </div>

                            <div class="space-y-1.5 pt-2">
                                <Label class="text-slate-600 text-xs flex justify-between items-center">
                                    <span>Nomor Surat Tugas</span>
                                    {#if !isEditing && isEditable && formData.suratTugasNumber !== (records[0]?.suratTugasNumber || '')}
                                        <button 
                                            on:click={saveSuratTugasNumberOnly}
                                            class="text-[10px] text-emerald-600 hover:text-emerald-700 font-bold uppercase tracking-tight flex items-center gap-1"
                                            disabled={isLoading}
                                        >
                                            {#if isLoading}
                                                <LottieLoader size="32px" className="brightness-0 invert" />
                                            {:else}
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" /></svg>
                                            {/if}
                                            Simpan Nomor
                                        </button>
                                    {/if}
                                </Label>
                                <Input 
                                    bind:value={formData.suratTugasNumber} 
                                    disabled={!isEditing && !isEditable} 
                                    class={(!isEditing && !isEditable) ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : (isEditing ? 'bg-white border-blue-200 focus:border-blue-500' : 'bg-emerald-50/30 border-emerald-200 focus:border-emerald-500')} 
                                    placeholder="Cth: 1/B/2026/01" 
                                />
                                {#if !isEditing && isEditable}
                                    <p class="text-[10px] text-slate-400 italic">Nomor ini dapat diubah langsung tanpa menekan tombol "Edit Pengajuan".</p>
                                {/if}
                            </div>
                        </div>
                    </div>
                    
                    <div class="bg-white rounded-xl border border-slate-200 p-5 shadow-sm">
                        <h3 class="text-sm font-bold text-slate-800 flex items-center gap-2 mb-5">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
                            </svg>
                            Tujuan Kegiatan
                        </h3>
                        
                        <div class="space-y-4">
                            <div class="space-y-1.5">
                                <Label class="text-slate-600 text-xs">Tujuan Perjalanan</Label>
                                <Input bind:value={formData.purpose} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-amber-200 focus:border-amber-500'} />
                            </div>
                            
                            <div class="space-y-1.5">
                                <Label class="text-slate-600 text-xs">Stakeholder / Pejabat (Pendampingan)</Label>
                                <Select bind:value={formData.stakeholder} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-amber-200 focus:border-amber-500'}>
                                    <option value="">Tidak ada/Lainnya</option>
                                    {#each stakeholders as st}
                                        <option value={st}>{st}</option>
                                    {/each}
                                </Select>
                            </div>
                            
                            <div class="space-y-1.5">
                                <Label class="text-slate-600 text-xs">Agenda (Opsional)</Label>
                                <Textarea bind:value={formData.agenda} disabled={!isEditing} class="resize-y min-h-[80px] {!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-amber-200 focus:border-amber-500'}" />
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- Right Column: Employees & Cost -->
                <div class="lg:col-span-2 space-y-6">
                    <div class="bg-white rounded-xl border border-slate-200 p-5 shadow-sm flex flex-col h-full max-h-[350px]">
                        <h3 class="text-sm font-bold text-slate-800 flex items-center justify-between mb-4 pb-3 border-b border-slate-100">
                            <span class="flex items-center gap-2">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                                </svg>
                                Daftar Petugas
                            </span>
                            <span class="bg-indigo-100 text-indigo-700 text-xs font-bold px-2 py-0.5 rounded-full">{currentEmployeeCount}</span>
                        </h3>
                        <div class="overflow-y-auto pr-2 space-y-2 flex-1">
                            {#if !isEditing}
                                {#each records as record}
                                    <div class="p-3 bg-slate-50 rounded-lg border border-slate-100 flex items-center gap-3">
                                        <div class="h-9 w-9 rounded-full bg-indigo-100 border border-indigo-200 flex items-center justify-center text-indigo-700 font-bold text-xs shrink-0">
                                            {record.employee?.name ? (record.employee.name.split(' ').filter(Boolean).map(n=>n[0]).join('').substring(0,2).toUpperCase()) : '?'}
                                        </div>
                                        <div class="min-w-0">
                                            <p class="text-sm font-semibold text-slate-800 truncate">{record.employee?.name || '-'}</p>
                                            {#if (record.employee?.jabatan && record.employee.jabatan !== '-') || (record.employee?.golongan && record.employee.golongan !== '-') || (record.employee?.pangkat && record.employee.pangkat !== '-')}
                                                <p class="text-[11px] text-slate-500 truncate">
                                                    {(record.employee?.jabatan && record.employee.jabatan !== '-') ? record.employee.jabatan : 
                                                     ((record.employee?.pangkat && record.employee.pangkat !== '-') && (record.employee?.golongan && record.employee.golongan !== '-') ? `${record.employee.pangkat} (${record.employee.golongan})` : 
                                                     ((record.employee?.pangkat && record.employee.pangkat !== '-') ? record.employee.pangkat : 
                                                     ((record.employee?.golongan && record.employee.golongan !== '-') ? record.employee.golongan : '')))}
                                                </p>
                                            {:else}
                                                <p class="text-[11px] text-slate-500 truncate">Tidak ada jabatan</p>
                                            {/if}
                                        </div>
                                    </div>
                                {/each}
                            {:else}
                                <div class="mb-3 relative">
                                    <div class="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none text-slate-400">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                                        </svg>
                                    </div>
                                    <Input type="text" placeholder="Cari nama petugas..." bind:value={searchQuery} class="pl-8 bg-slate-50 border-slate-200 text-xs h-8" />
                                </div>
                                {#if filteredOfficers.length === 0}
                                    <div class="text-center py-4 text-xs text-slate-500 italic">Tidak ada petugas yang cocok.</div>
                                {:else}
                                    {#each filteredOfficers as officer}
                                        <label class="flex items-center space-x-3 p-2.5 rounded-lg border {selectedEmployeeIds.includes(officer.id) ? 'bg-indigo-50/50 border-indigo-200' : 'border-transparent hover:bg-slate-50 hover:border-slate-200'} cursor-pointer transition-all">
                                            <input type="checkbox"
                                                checked={selectedEmployeeIds.includes(officer.id)}
                                                on:change={(e) => {
                                                    if (e.target.checked) selectedEmployeeIds = [...selectedEmployeeIds, officer.id];
                                                    else selectedEmployeeIds = selectedEmployeeIds.filter(id => id !== officer.id);
                                                }}
                                                class="accent-indigo-600 h-4 w-4 rounded border-slate-300 shrink-0"
                                            />
                                            <div class="min-w-0">
                                                <p class="text-sm font-semibold text-slate-800 truncate">{officer.name}</p>
                                                {#if (officer.jabatan && officer.jabatan !== '-') || (officer.golongan && officer.golongan !== '-') || (officer.pangkat && officer.pangkat !== '-')}
                                                    <p class="text-[11px] text-slate-500 truncate">
                                                        {(officer.jabatan && officer.jabatan !== '-') ? officer.jabatan : 
                                                         ((officer.pangkat && officer.pangkat !== '-') && (officer.golongan && officer.golongan !== '-') ? `${officer.pangkat} (${officer.golongan})` : 
                                                         ((officer.pangkat && officer.pangkat !== '-') ? officer.pangkat : 
                                                         ((officer.golongan && officer.golongan !== '-') ? officer.golongan : '')))}
                                                    </p>
                                                {:else}
                                                    <p class="text-[11px] text-slate-500 truncate">Tidak ada jabatan</p>
                                                {/if}
                                            </div>
                                        </label>
                                    {/each}
                                {/if}
                            {/if}
                        </div>
                    </div>
                    
                    <div class="mt-4">
                        <CostEstimateCard 
                            breakdown={costBreakdown}
                            employeeCount={currentEmployeeCount}
                            totalCost={totalSBMCost}
                            readonly={true}
                        />
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Footer -->
        <div class="px-6 py-4 border-t border-slate-200 bg-white flex items-center justify-end gap-3 shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)] z-10">
            {#if !isEditing}
                <Button variant="outline" on:click={close} class="min-w-[100px]">Tutup</Button>
                {#if isEditable}
                    <Button variant="default" on:click={() => isEditing = true} class="bg-indigo-600 hover:bg-indigo-700 text-white min-w-[120px]">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                        Edit Pengajuan
                    </Button>
                {/if}
            {:else}
                <Button variant="ghost" on:click={() => isEditing = false} disabled={isLoading} class="min-w-[100px] hover:bg-slate-100">Batal</Button>
                <Button variant="default" class="bg-emerald-600 hover:bg-emerald-700 text-white min-w-[140px]" on:click={handleSave} disabled={isLoading}>
                    {#if isLoading}
                        <LottieLoader size="36px" className="brightness-0 invert -ml-1" />
                        Menyimpan...
                    {:else}
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                        </svg>
                        Simpan Perubahan
                    {/if}
                </Button>
            {/if}
        </div>
    </div>
  </div>

  <ConfirmationModal 
      bind:open={showSpdConfirmModal}
      title="Konfirmasi Perubahan Nomor SPJ Manual"
      description="PERHATIAN: Sistem penomoran SPJ bersifat dinamis. Mengubah nomor SPJ secara manual akan mengubah urutan pembuatan dan menyebabkan nomor urut petugas pada SPJ di bawahnya ikut bergeser secara otomatis (efek domino) untuk menyesuaikan. Apakah Anda yakin ingin melanjutkan?"
      confirmText="Ya, Ubah Nomor SPJ"
      cancelText="Batal"
      onConfirm={saveSpdNumberOnly}
  />
{/if}

