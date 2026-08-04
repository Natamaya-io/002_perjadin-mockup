<script>
    import { createEventDispatcher } from 'svelte';
    import { toast } from '$lib/shared/stores/toast';
    import { api } from '$lib/shared/api';
    import { userStore } from '$lib/features/auth/store';
    import { provincesStore } from '$lib/shared/stores/master-data';
    import { formatCurrency, compressImage } from '$lib/shared/utils/utils';
    import Dialog from '$lib/shared/ui/dialog/Dialog.svelte';
    import DialogHeader from '$lib/shared/ui/dialog/DialogHeader.svelte';
    import DialogTitle from '$lib/shared/ui/dialog/DialogTitle.svelte';
    import DialogFooter from '$lib/shared/ui/dialog/DialogFooter.svelte';
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import DocumentViewer from '$lib/shared/ui/document-viewer/DocumentViewer.svelte';

    export let open = false;
    export let record = null;
    
    export let editingCosts = {};
    export let isReadOnlyOverride = false;

    const dispatch = createEventDispatcher();

    // The modal is read-only if the user is a kasubag or if the record is already Approved (meaning it's just for review/printing) or if overridden
    // Allow 'super_admin' to edit even if Approved
    $: isReadOnly = isReadOnlyOverride || $userStore.role === 'kasubag' || (record?.status === 'Approved' && $userStore.role !== 'super_admin');

    // Derived Calculations for SBM Uang Harian
    $: costBreakdown = (record?.locations && record.locations.length > 0 
        ? record.locations 
        : []
    ).map(loc => {
        const provData = $provincesStore.find(p => p.name === loc.province);
        const rate = provData ? provData.luarKota : 0;
        const start = new Date(loc.startDate);
        const end = new Date(loc.endDate);
        let locDays = 0;
        if (!isNaN(start.getTime()) && !isNaN(end.getTime())) {
            const diffTime = end.getTime() - start.getTime();
            const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
            locDays = diffDays > 0 ? diffDays : 0;
        }
        return { days: locDays, rate: rate, province: loc.province };
    });

    let selectedLocationIndex = 0;
    let editingSpd = ''; // Added here

    $: if (open && record) {
        editingSpd = record.spd || ''; // Set value here
        const locationsCount = costBreakdown.length;
        if (!editingCosts.details) {
            editingCosts.details = [];
        }
        
        // Migration from legacy flat structure
        if (editingCosts.details.length === 0 && (editingCosts.hotelDays || editingCosts.transportAmount || editingCosts.ticketGo || editingCosts.transportMode)) {
            editingCosts.details = [{
                ...editingCosts,
                details: undefined
            }];
        }

        // Fill missing locations
        while (editingCosts.details.length < locationsCount) {
            editingCosts.details.push({
                transportMode: 'Pesawat/Kendaraan Umum',
                ticketGo: 0, ticketBack: 0,
                hotelDays: 0, hotelRate: 0,
                transportAmount: 0,
                additionalCosts: [],
                boardingPassFiles: []
            });
        }
        
        // Reset selected index if out of bounds
        if (selectedLocationIndex >= locationsCount) {
            selectedLocationIndex = 0;
        }

        editingCosts = editingCosts;
    }

	$: days = costBreakdown.reduce((sum, item) => sum + item.days, 0);
	$: totalDailyAllowance = costBreakdown.reduce((sum, item) => sum + (item.rate * item.days), 0);
	$: sbmRateAvg = days > 0 ? totalDailyAllowance / days : 0;

    $: grandTotal = (editingCosts.details || []).reduce((acc, detail, idx) => {
        // Individual SBM for this person at this location
        const sbmTotal = costBreakdown[idx] ? (costBreakdown[idx].rate * costBreakdown[idx].days) : 0;
        const hotel = (detail.hotelDays || 0) * (detail.hotelRate || 0);
        const ticket = Number(detail.ticketGo || 0) + Number(detail.ticketBack || 0);
        const transport = Number(detail.transportAmount || 0);
        const addCosts = (detail.additionalCosts || []).reduce((sum, c) => {
            if (c.name === 'Extend Tiket') {
                return sum + (Number(c.ticketGo) || 0) + (Number(c.ticketBack) || 0) + (Number(c.amount) || 0);
            }
            return sum + (c.amount || 0);
        }, 0);
        return acc + sbmTotal + hotel + ticket + transport + addCosts;
    }, 0);

    // Current tab helpers
    $: detail = editingCosts.details && editingCosts.details[selectedLocationIndex] ? editingCosts.details[selectedLocationIndex] : {};
    $: currentLocSbm = costBreakdown[selectedLocationIndex] || { rate: 0, days: 0, province: '' };
    
    $: currentTotalHotel = (detail.hotelDays || 0) * (detail.hotelRate || 0) + (detail.additionalCosts || []).reduce((sum, cost) => cost.name === 'Extend Penginapan' ? sum + (Number(cost.amount) || 0) : sum, 0);
    $: currentTotalAdditional = (detail.additionalCosts || []).reduce((sum, cost) => {
        if (cost.name === 'Extend Tiket' || cost.name === 'Extend Penginapan') {
            return sum;
        }
        return sum + (cost.amount || 0);
    }, 0);

    export let allRecordsInSpd = [];

    // Preview State
    let previewFile = null;
    let showPreview = false;

    function openPreview(file) {
        previewFile = file;
        showPreview = true;
    }

    function closePreview() {
        showPreview = false;
        previewFile = null;
    }

    // Split Hotel State
    let showSplitHotelModal = false;
    let splitHotelData = {
        locationIndex: 0,
        totalBill: 0,
        days: 0,
        selectedEmpIds: [],
        isExtend: false,
        extendIdx: -1
    };
    let pendingOtherUpdatesToSave = [];

    function openSplitHotelModal(isExtend = false, extendIdx = -1) {
        const detail = editingCosts.details[selectedLocationIndex];
        let totalBill = 0;
        let days = 0;
        let previousSplitWith = [];
        
        if (isExtend && extendIdx > -1) {
            const extendCost = detail.additionalCosts[extendIdx];
            days = extendCost.hotelDays || 0;
            // Gunakan hotelOriginalRate jika sudah pernah di-split sebelumnya
            // agar tidak terjadi double-divide
            if (extendCost.hotelOriginalRate) {
                totalBill = extendCost.hotelOriginalRate * days;
                previousSplitWith = extendCost.hotelSplitWith || [];
            } else {
                totalBill = (extendCost.hotelRate || 0) * days;
            }
        } else {
            days = detail.hotelDays || 0;
            // Gunakan hotelOriginalRate jika sudah pernah di-split sebelumnya
            if (detail.hotelOriginalRate) {
                totalBill = detail.hotelOriginalRate * days;
                previousSplitWith = detail.hotelSplitWith || [];
            } else {
                totalBill = (detail.hotelRate || 0) * days;
            }
        }

        splitHotelData = {
            locationIndex: selectedLocationIndex,
            totalBill: totalBill,
            days: days,
            selectedEmpIds: previousSplitWith,  // restore siapa yang sudah di-split
            isExtend: isExtend,
            extendIdx: extendIdx
        };
        showSplitHotelModal = true;
    }

    function toggleSplitEmp(empId) {
        if (splitHotelData.selectedEmpIds.includes(empId)) {
            splitHotelData.selectedEmpIds = splitHotelData.selectedEmpIds.filter(id => id !== empId);
        } else {
            splitHotelData.selectedEmpIds = [...splitHotelData.selectedEmpIds, empId];
        }
    }

    function applySplitHotel() {
        const totalPeople = 1 + splitHotelData.selectedEmpIds.length;
        if (totalPeople === 0) {
            toast.error("Minimal 1 orang dipilih.");
            return;
        }
        if (splitHotelData.days <= 0) {
            toast.error("Durasi malam tidak boleh 0.");
            return;
        }

        // Harga per malam SEBELUM dibagi (total kamar penuh per malam)
        const originalRatePerNight = Math.round(splitHotelData.totalBill / splitHotelData.days);
        // Harga per malam PER ORANG setelah dibagi
        const ratePerNightPerPerson = Math.round(splitHotelData.totalBill / totalPeople / splitHotelData.days);
        
        let sourceFile = null;
        if (splitHotelData.isExtend && splitHotelData.extendIdx > -1) {
            sourceFile = editingCosts.details[splitHotelData.locationIndex].additionalCosts[splitHotelData.extendIdx].file;
            
            // Apply to current person + simpan state split
            const sourceCost = editingCosts.details[splitHotelData.locationIndex].additionalCosts[splitHotelData.extendIdx];
            sourceCost.hotelRate = ratePerNightPerPerson;
            sourceCost.hotelDays = splitHotelData.days;
            sourceCost.amount = ratePerNightPerPerson * splitHotelData.days;
            // Simpan original rate & siapa saja yang di-split
            sourceCost.hotelOriginalRate = originalRatePerNight;
            sourceCost.hotelSplitWith = [...splitHotelData.selectedEmpIds];
            editingCosts = editingCosts;

            // Prepare updates for others
            const newUpdates = splitHotelData.selectedEmpIds.map(empId => ({
                empId,
                locationIndex: splitHotelData.locationIndex,
                hotelRate: ratePerNightPerPerson,
                hotelDays: splitHotelData.days,
                hotelFile: sourceFile ? { ...sourceFile } : null,
                hotelOriginalRate: originalRatePerNight,
                hotelSplitWith: [...splitHotelData.selectedEmpIds, record.id],
                isExtend: true
            }));
            pendingOtherUpdatesToSave = [...pendingOtherUpdatesToSave, ...newUpdates];

        } else {
            sourceFile = editingCosts.details[splitHotelData.locationIndex].hotelFile;

            // Apply to current person + simpan state split
            editingCosts.details[splitHotelData.locationIndex].hotelRate = ratePerNightPerPerson;
            editingCosts.details[splitHotelData.locationIndex].hotelDays = splitHotelData.days;
            // Simpan original rate & siapa saja yang di-split
            editingCosts.details[splitHotelData.locationIndex].hotelOriginalRate = originalRatePerNight;
            editingCosts.details[splitHotelData.locationIndex].hotelSplitWith = [...splitHotelData.selectedEmpIds];
            editingCosts = editingCosts;

            // Prepare updates for others (termasuk originalRate agar modal mereka pun bisa restore)
            const newUpdates = splitHotelData.selectedEmpIds.map(empId => ({
                empId,
                locationIndex: splitHotelData.locationIndex,
                hotelRate: ratePerNightPerPerson,
                hotelDays: splitHotelData.days,
                hotelFile: sourceFile ? { ...sourceFile } : null,
                hotelOriginalRate: originalRatePerNight,
                hotelSplitWith: [...splitHotelData.selectedEmpIds, record.id],
                isExtend: false
            }));
            pendingOtherUpdatesToSave = [...pendingOtherUpdatesToSave, ...newUpdates];
        }

        showSplitHotelModal = false;
        toast.success(`Biaya dibagi ke ${totalPeople} orang (akan tersimpan saat klik Simpan).`);
    }

    function formatFileSize(bytes) {
        if (!bytes) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }


    function formatInputNumber(value) {
        if (value === undefined || value === null || value === '') return '';
        const numStr = value.toString().replace(/[^0-9]/g, '');
        if (!numStr) return '';
        return parseInt(numStr, 10).toLocaleString('id-ID');
    }

    function updateCost(field, event) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        editingCosts.details[selectedLocationIndex][field] = isNaN(num) ? undefined : num;
        editingCosts = editingCosts;
    }

    async function handleSpecificFileSelect(e, field) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (ev) => {
            editingCosts.details[selectedLocationIndex][field] = {
                name: file.name,
                size: file.size,
                type: file.type,
                data: ev.target.result
            };
            editingCosts = editingCosts;
        };
        file = await compressImage(file);
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    async function handleBoardingPassFileSelect(e) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }

        if (!editingCosts.details[selectedLocationIndex].boardingPassFiles) {
            editingCosts.details[selectedLocationIndex].boardingPassFiles = [];
        }

        const reader = new FileReader();
        reader.onload = (ev) => {
            editingCosts.details[selectedLocationIndex].boardingPassFiles = [
                ...editingCosts.details[selectedLocationIndex].boardingPassFiles, 
                {
                    name: file.name,
                    size: file.size,
                    type: file.type,
                    data: ev.target.result
                }
            ];
            editingCosts = editingCosts;
        };
        file = await compressImage(file);
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function removeBoardingPassFile(index) {
        if (editingCosts.details[selectedLocationIndex].boardingPassFiles) {
            editingCosts.details[selectedLocationIndex].boardingPassFiles = editingCosts.details[selectedLocationIndex].boardingPassFiles.filter((_, i) => i !== index);
            editingCosts = editingCosts;
        }
    }

    function removeSpecificFile(field) {
        editingCosts.details[selectedLocationIndex][field] = null;
        editingCosts = editingCosts;
    }

    function addAdditionalCost() {
        if (!editingCosts.details[selectedLocationIndex].additionalCosts) editingCosts.details[selectedLocationIndex].additionalCosts = [];
        editingCosts.details[selectedLocationIndex].additionalCosts = [...editingCosts.details[selectedLocationIndex].additionalCosts, { name: '', amount: undefined, file: null }];
        editingCosts = editingCosts;
    }

    function addExtendTicket() {
        if (!editingCosts.details[selectedLocationIndex].additionalCosts) editingCosts.details[selectedLocationIndex].additionalCosts = [];
        editingCosts.details[selectedLocationIndex].additionalCosts = [...editingCosts.details[selectedLocationIndex].additionalCosts, { name: 'Extend Tiket', amount: undefined, file: null }];
        editingCosts = editingCosts;
    }

    function addExtendPenginapan() {
        if (!editingCosts.details[selectedLocationIndex].additionalCosts) editingCosts.details[selectedLocationIndex].additionalCosts = [];
        editingCosts.details[selectedLocationIndex].additionalCosts = [...editingCosts.details[selectedLocationIndex].additionalCosts, { name: 'Extend Penginapan', amount: undefined, file: null }];
        editingCosts = editingCosts;
    }

    async function handleExtendTicketBackFileSelect(e, index) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }

        const reader = new FileReader();
        reader.onload = (ev) => {
            editingCosts.details[selectedLocationIndex].additionalCosts[index].ticketBackFile = {
                name: file.name,
                size: file.size,
                type: file.type,
                data: ev.target.result
            };
            editingCosts = editingCosts;
        };
        file = await compressImage(file);
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    async function handleExtendBoardingPassFileSelect(e, index) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }

        if (!editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles) {
            editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles = [];
        }

        const reader = new FileReader();
        reader.onload = (ev) => {
            editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles = [
                ...editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles, 
                {
                    name: file.name,
                    size: file.size,
                    type: file.type,
                    data: ev.target.result
                }
            ];
            editingCosts = editingCosts;
        };
        file = await compressImage(file);
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function removeExtendBoardingPassFile(costIdx, bpIdx) {
        if (editingCosts.details[selectedLocationIndex].additionalCosts[costIdx].boardingPassFiles) {
            editingCosts.details[selectedLocationIndex].additionalCosts[costIdx].boardingPassFiles = editingCosts.details[selectedLocationIndex].additionalCosts[costIdx].boardingPassFiles.filter((_, i) => i !== bpIdx);
            editingCosts = editingCosts;
        }
    }

    function removeAdditionalCost(index) {
        editingCosts.details[selectedLocationIndex].additionalCosts = editingCosts.details[selectedLocationIndex].additionalCosts.filter((_, i) => i !== index);
        editingCosts = editingCosts;
    }

    function updateAdditionalCostAmount(index, event) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        editingCosts.details[selectedLocationIndex].additionalCosts[index].amount = isNaN(num) ? undefined : num;
        editingCosts = editingCosts;
    }

    function updateAdditionalExtendCost(index, field, event) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        editingCosts.details[selectedLocationIndex].additionalCosts[index][field] = isNaN(num) ? undefined : num;
        editingCosts = editingCosts;
    }

    async function handleAdditionalExtendSpecificFileSelect(e, index, field) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (ev) => {
            editingCosts.details[selectedLocationIndex].additionalCosts[index][field] = {
                name: file.name,
                size: file.size,
                type: file.type,
                data: ev.target.result
            };
            editingCosts = editingCosts;
        };
        file = await compressImage(file);
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function removeAdditionalExtendSpecificFile(index, field) {
        editingCosts.details[selectedLocationIndex].additionalCosts[index][field] = null;
        editingCosts = editingCosts;
    }

    async function handleAdditionalExtendBoardingPassSelect(e, index) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }

        if (!editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles) {
            editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles = [];
        }

        const reader = new FileReader();
        reader.onload = (ev) => {
            editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles = [
                ...editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles, 
                {
                    name: file.name,
                    size: file.size,
                    type: file.type,
                    data: ev.target.result
                }
            ];
            editingCosts = editingCosts;
        };
        file = await compressImage(file);
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function removeAdditionalExtendBoardingPass(index, bpIndex) {
        if (editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles) {
            editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles = editingCosts.details[selectedLocationIndex].additionalCosts[index].boardingPassFiles.filter((_, i) => i !== bpIndex);
            editingCosts = editingCosts;
        }
    }

    async function handleAdditionalFileSelect(e, index) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (ev) => {
            editingCosts.details[selectedLocationIndex].additionalCosts[index].file = {
                name: file.name,
                size: file.size,
                type: file.type,
                data: ev.target.result
            };
            editingCosts = editingCosts;
        };
        file = await compressImage(file);
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function validateCosts() {
        for (let i = 0; i < (editingCosts.details || []).length; i++) {
            const d = editingCosts.details[i];
            const provName = costBreakdown[i] ? costBreakdown[i].province : 'Lokasi';
            if ((d.ticketGo || 0) < 0) return `Biaya Tiket Berangkat di ${provName} tidak boleh negatif`;
            if ((d.ticketBack || 0) < 0) return `Biaya Tiket Pulang di ${provName} tidak boleh negatif`;
            if ((d.hotelDays || 0) < 0) return `Durasi Penginapan di ${provName} tidak boleh negatif`;
            if ((d.hotelRate || 0) < 0) return `Rate Penginapan di ${provName} tidak boleh negatif`;
            if ((d.transportAmount || 0) < 0) return `Biaya Transportasi di ${provName} tidak boleh negatif`;
            for (const cost of (d.additionalCosts || [])) {
                if ((cost.amount || 0) < 0) return `Transport Lokal di ${provName} tidak boleh negatif`;
            }
        }
        return null;
    }

    function handleSave() {
        if (isReadOnly) return;
        const error = validateCosts();
        if (error) {
            toast.error(error);
            return;
        }

        // Map back primary location details to root for backend compatibility
        // because the backend only supports saving a flat cost structure per record.
        if (editingCosts.details && editingCosts.details.length > 0) {
            const primary = editingCosts.details[0];
            editingCosts = {
                ...editingCosts,
                transportMode: primary.transportMode,
                ticketGo: primary.ticketGo,
                ticketBack: primary.ticketBack,
                hotelDays: primary.hotelDays,
                hotelRate: primary.hotelRate,
                transportAmount: primary.transportAmount,
                additionalCosts: primary.additionalCosts,
                boardingPassFiles: primary.boardingPassFiles,
                ticketGoFile: primary.ticketGoFile,
                ticketBackFile: primary.ticketBackFile,
                hotelFile: primary.hotelFile,
                transportFile: primary.transportFile
            };
        }

        editingCosts.dailyAllowanceRate = sbmRateAvg;
        editingCosts.dailyAllowanceDays = days;

        dispatch('save', { editingCosts, grandTotal, pendingOtherUpdates: pendingOtherUpdatesToSave, spd: editingSpd });
    }
</script>

<Dialog bind:open={open} class="w-[calc(100vw-2rem)] md:w-full max-w-[95vw] md:max-w-2xl lg:max-w-4xl overflow-hidden flex flex-col p-0 h-[calc(100vh-2rem)] max-h-[90vh] md:max-h-[85vh] mx-auto my-auto rounded-2xl shadow-2xl border-0" on:close={() => dispatch('close')}>
    <DialogHeader class="border-b border-slate-100 p-4 md:p-6 shrink-0 bg-white z-10">
        <DialogTitle class="text-lg md:text-xl font-bold text-slate-800">
            {$userStore.role === 'protokol' 
                ? (record?.status === 'Draft' ? 'Input Rincian Biaya' : 'Edit Rincian Biaya') 
                : 'Review Rincian Biaya'}
        </DialogTitle>
        <p class="text-xs md:text-sm text-slate-500 mt-1">Rincian komponen biaya untuk pegawai <span class="font-semibold text-blue-600 bg-blue-50 px-1.5 py-0.5 rounded">{record?.employee?.name}</span>.</p>
    </DialogHeader>
    
    <!-- Location Tabs -->
    {#if costBreakdown.length > 1}
    <div class="px-4 md:px-6 pt-4 pb-2 bg-slate-50/50 border-b border-slate-100 shrink-0 overflow-x-auto">
        <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider mb-2 block">Pilih Provinsi / Lokasi</Label>
        <div class="flex gap-2 w-max">
            {#each costBreakdown as loc, idx}
                <button
                    class="px-4 py-2 rounded-lg text-sm font-semibold whitespace-nowrap transition-colors border {selectedLocationIndex === idx ? 'bg-blue-600 text-white border-blue-600 shadow-sm' : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-50 hover:border-slate-300'}"
                    on:click={() => selectedLocationIndex = idx}
                >
                    {loc.province}
                </button>
            {/each}
        </div>
    </div>
    {/if}

    <div class="grid gap-4 md:gap-6 p-4 md:p-6 overflow-y-auto overflow-x-hidden bg-slate-50/50 custom-scrollbar flex-1 min-h-0 relative">

        {#if editingCosts.details && editingCosts.details[selectedLocationIndex]}

        <!-- Mode Transportasi -->
        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm w-full space-y-1.5 mb-2">
            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Mode Transportasi</Label>
            <Select bind:value={editingCosts.details[selectedLocationIndex].transportMode} disabled={isReadOnly} class="bg-slate-50 border-slate-200 h-9 md:h-10 text-sm {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}">
                <option value="Pesawat/Kendaraan Umum">Pesawat/Kendaraan Umum</option>
                <option value="Mobil">Mobil</option>
            </Select>
        </div>

        <!-- Uang Harian SBM -->
        <div class="p-3 md:p-4 bg-blue-50/50 rounded-xl border border-blue-100 space-y-2.5 md:space-y-3 w-full mb-2">
            <div class="flex flex-wrap justify-between items-center gap-2">
                <h4 class="text-xs md:text-sm font-semibold text-blue-800 flex items-center gap-1.5">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    Uang Harian (SBM)
                </h4>
                <span class="text-base md:text-lg font-bold text-blue-700">{formatCurrency(currentLocSbm.rate * currentLocSbm.days)}</span>
            </div>
            
            <div class="space-y-2">
                <div class="flex flex-wrap md:flex-nowrap items-center gap-2 md:gap-3 text-xs md:text-sm text-slate-600 bg-white p-2 md:p-3 rounded-lg border border-blue-50/50 shadow-sm w-full">
                    <div class="flex-none pr-2 border-r border-slate-100 min-w-[80px]">
                         <span class="block text-[9px] md:text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Provinsi</span>
                         <span class="font-medium whitespace-nowrap text-blue-600">{currentLocSbm.province}</span>
                    </div>
                    <div class="flex-none pr-2 border-r border-slate-100">
                        <span class="block text-[9px] md:text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Durasi</span>
                        <span class="font-medium whitespace-nowrap">{currentLocSbm.days} Hari</span>
                    </div>
                    <div class="flex-1 min-w-0 overflow-hidden">
                        <span class="block text-[9px] md:text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Rate SBM</span>
                        <span class="font-medium truncate block w-full">{formatCurrency(currentLocSbm.rate)} <span class="text-[10px] md:text-xs text-slate-400 font-normal">/ hari</span></span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Tiket & Boarding Pass -->
        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-4 w-full mb-2">
            <div class="flex justify-between items-center">
                <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    Tiket & Boarding Pass
                </h4>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Tiket Berangkat -->
                <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                    <div class="flex justify-between items-center min-h-[32px]">
                        <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Berangkat</Label>
                        {#if !detail.ticketGoFile && !isReadOnly}
                            <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                Kwitansi
                                <input type="file" class="hidden" on:change={(e) => handleSpecificFileSelect(e, 'ticketGoFile')} />
                            </label>
                        {/if}
                    </div>
                    <div class="relative">
                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                        <Input type="text" value={formatInputNumber(detail.ticketGo)} on:input={(e) => updateCost('ticketGo', e)} disabled={isReadOnly} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                    </div>
                    {#if detail.ticketGoFile}
                        <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md shadow-sm">
                            <div class="flex items-center gap-2 min-w-0 flex-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{detail.ticketGoFile.name}</span>
                            </div>
                            <div class="flex gap-1.5 shrink-0">
                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-blue-50 hover:bg-blue-100 text-blue-700 rounded border border-blue-200 transition-colors" on:click={() => openPreview(detail.ticketGoFile)}>Lihat</button>
                                {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => removeSpecificFile('ticketGoFile')}>Hapus</button>{/if}
                            </div>
                        </div>
                    {/if}
                </div>

                <!-- Tiket Pulang -->
                <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                    <div class="flex justify-between items-center min-h-[32px]">
                        <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Pulang</Label>
                        {#if !detail.ticketBackFile && !isReadOnly}
                            <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                Kwitansi
                                <input type="file" class="hidden" on:change={(e) => handleSpecificFileSelect(e, 'ticketBackFile')} />
                            </label>
                        {/if}
                    </div>
                    <div class="relative">
                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                        <Input type="text" value={formatInputNumber(detail.ticketBack)} on:input={(e) => updateCost('ticketBack', e)} disabled={isReadOnly} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                    </div>
                    {#if detail.ticketBackFile}
                        <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md shadow-sm w-full">
                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{detail.ticketBackFile.name}</span>
                            <div class="flex gap-1.5 shrink-0 ml-2">
                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-blue-50 hover:bg-blue-100 text-blue-700 rounded border border-blue-200 transition-colors" on:click={() => openPreview(detail.ticketBackFile)}>Lihat</button>
                                {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => removeSpecificFile('ticketBackFile')}>Hapus</button>{/if}
                            </div>
                        </div>
                    {/if}
                </div>
            </div>

            <!-- Boarding Pass Files -->
            <div class="pt-3 border-t border-slate-100 space-y-2 w-full">
                <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider flex items-center gap-1.5">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z" /></svg>
                    Boarding Pass ({detail.boardingPassFiles ? detail.boardingPassFiles.length : 0} file)
                </Label>
                
                {#if detail.boardingPassFiles && detail.boardingPassFiles.length > 0}
                    <div class="flex flex-col gap-1.5">
                        {#each detail.boardingPassFiles as bpFile, bpIdx}
                            <div class="flex items-center justify-between p-2 bg-blue-50 border border-blue-200 rounded-md shadow-sm w-full">
                                <div class="flex items-center gap-2 min-w-0 flex-1">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z" /></svg>
                                    <span class="text-[10px] md:text-xs text-blue-800 truncate">{bpFile.name || 'Boarding Pass ' + (bpIdx + 1)}</span>
                                </div>
                                <div class="flex gap-1.5 shrink-0 ml-2">
                                    <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-blue-100 hover:bg-blue-200 text-blue-700 rounded border border-blue-300 transition-colors" on:click={() => openPreview(bpFile)}>Lihat</button>
                                    {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => removeBoardingPassFile(bpIdx)}>Hapus</button>{/if}
                                </div>
                            </div>
                        {/each}
                    </div>
                {:else}
                    <div class="flex flex-col items-center justify-center p-4 bg-slate-50 border border-dashed border-slate-200 rounded-md w-full">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-slate-300 mb-1" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                        <p class="text-[10px] text-slate-400 font-medium text-center">Belum ada dokumen Boarding Pass</p>
                    </div>
                {/if}
            </div>

            <!-- Upload Boarding Pass (hidden from Kasubag) -->
            {#if !isReadOnly}
                <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm self-start">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                    Tambah Boarding Pass
                    <input type="file" multiple accept="image/*,application/pdf" class="hidden" on:change={handleBoardingPassFileSelect} />
                </label>
            {/if}

            <!-- Extend Tiket Items -->
            {#if detail.additionalCosts && detail.additionalCosts.length > 0}
                {#each detail.additionalCosts as cost, costIdx}
                    {#if cost.name === 'Extend Tiket'}
                        <div class="col-span-1 md:col-span-2 mt-4 relative pt-4 border-t border-slate-200 w-full">
                            <div class="flex justify-between items-center mb-3 min-h-[32px] pr-8 relative">
                                {#if !isReadOnly}
                                    <button type="button" class="absolute -top-1 -right-1 bg-red-100 text-red-600 rounded-full p-1 border border-red-200 hover:bg-red-200 transition-colors shadow-sm z-10" on:click={() => removeAdditionalCost(costIdx)}>
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" /></svg>
                                    </button>
                                {/if}
                                <h4 class="text-[10px] md:text-xs font-bold uppercase text-blue-600 tracking-wider flex items-center gap-1.5 md:gap-2">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z" /></svg>
                                    Extend Tiket
                                </h4>
                                <div class="flex items-center gap-2">
                                    {#if !isReadOnly}
                                        <button type="button" class="text-[10px] md:text-xs font-bold uppercase tracking-wider text-indigo-600 bg-indigo-50 hover:bg-indigo-100 hover:text-indigo-700 px-2 py-1 rounded-md border border-indigo-200 transition-colors flex items-center gap-1" on:click={() => openSplitHotelModal(true, costIdx)}>
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" /></svg>
                                            Bagi Biaya
                                        </button>
                                    {/if}
                                </div>
                            </div>
                            
                            <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <!-- Extend Tiket Berangkat -->
                                    <div class="space-y-2">
                                        <div class="flex justify-between items-center min-h-[32px]">
                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Berangkat</Label>
                                            {#if !cost.file && !isReadOnly}
                                                <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm">
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                                    Kwitansi
                                                    <input type="file" class="hidden" on:change={(e) => handleAdditionalFileSelect(e, costIdx)} />
                                                </label>
                                            {/if}
                                        </div>
                                        <div class="relative">
                                            <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                            <Input type="text" value={formatInputNumber(cost.amount)} on:input={(e) => {
                                                const val = e.detail ? e.detail.value : e.target.value;
                                                editingCosts.details[selectedLocationIndex].additionalCosts[costIdx].amount = val.replace(/[^0-9]/g, '');
                                                editingCosts = editingCosts;
                                            }} disabled={isReadOnly} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                                        </div>
                                        {#if cost.file}
                                            <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md shadow-sm">
                                                <div class="flex items-center gap-2 min-w-0 flex-1">
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{cost.file.name}</span>
                                                </div>
                                                <div class="flex gap-1.5 shrink-0">
                                                    <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-blue-50 hover:bg-blue-100 text-blue-700 rounded border border-blue-200 transition-colors" on:click={() => openPreview(cost.file)}>Lihat</button>
                                                    {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => { editingCosts.details[selectedLocationIndex].additionalCosts[costIdx].file = null; editingCosts = editingCosts; }}>Hapus</button>{/if}
                                                </div>
                                            </div>
                                        {/if}
                                    </div>
                                    
                                    <!-- Extend Tiket Kembali -->
                                    <div class="space-y-2">
                                        <div class="flex justify-between items-center min-h-[32px]">
                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Kembali</Label>
                                            {#if !cost.ticketBackFile && !isReadOnly}
                                                <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm">
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                                    Kwitansi
                                                    <input type="file" class="hidden" on:change={(e) => handleExtendTicketBackFileSelect(e, costIdx)} />
                                                </label>
                                            {/if}
                                        </div>
                                        <div class="relative">
                                            <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                            <Input type="text" value={formatInputNumber(cost.ticketBack)} on:input={(e) => {
                                                const val = e.detail ? e.detail.value : e.target.value;
                                                editingCosts.details[selectedLocationIndex].additionalCosts[costIdx].ticketBack = val.replace(/[^0-9]/g, '');
                                                editingCosts = editingCosts;
                                            }} disabled={isReadOnly} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                                        </div>
                                        {#if cost.ticketBackFile}
                                            <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md shadow-sm">
                                                <div class="flex items-center gap-2 min-w-0 flex-1">
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{cost.ticketBackFile.name}</span>
                                                </div>
                                                <div class="flex gap-1.5 shrink-0">
                                                    <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-blue-50 hover:bg-blue-100 text-blue-700 rounded border border-blue-200 transition-colors" on:click={() => openPreview(cost.ticketBackFile)}>Lihat</button>
                                                    {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => { editingCosts.details[selectedLocationIndex].additionalCosts[costIdx].ticketBackFile = null; editingCosts = editingCosts; }}>Hapus</button>{/if}
                                                </div>
                                            </div>
                                        {/if}
                                    </div>
                                    
                                    <!-- Extend Boarding Pass -->
                                    <div class="space-y-2 md:col-span-2 pt-2 border-t border-slate-100">
                                        <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider flex items-center gap-1.5">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z" /></svg>
                                            Extend Boarding Pass ({cost.boardingPassFiles ? cost.boardingPassFiles.length : 0} file)
                                        </Label>
                                        
                                        {#if cost.boardingPassFiles && cost.boardingPassFiles.length > 0}
                                            <div class="grid grid-cols-1 md:grid-cols-2 gap-1.5">
                                                {#each cost.boardingPassFiles as bpFile, bpIdx}
                                                    <div class="flex items-center justify-between p-2 bg-blue-50 border border-blue-200 rounded-md shadow-sm w-full">
                                                        <div class="flex items-center gap-2 min-w-0 flex-1">
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z" /></svg>
                                                            <span class="text-[10px] md:text-xs text-blue-800 truncate">{bpFile.name || 'Boarding Pass ' + (bpIdx + 1)}</span>
                                                        </div>
                                                        <div class="flex gap-1.5 shrink-0 ml-2">
                                                            <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-blue-100 hover:bg-blue-200 text-blue-700 rounded border border-blue-300 transition-colors" on:click={() => openPreview(bpFile)}>Lihat</button>
                                                            {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => removeExtendBoardingPassFile(costIdx, bpIdx)}>Hapus</button>{/if}
                                                        </div>
                                                    </div>
                                                {/each}
                                            </div>
                                        {/if}

                                        {#if !isReadOnly}
                                            <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                                Tambah Boarding Pass
                                                <input type="file" multiple accept="image/*,application/pdf" class="hidden" on:change={(e) => handleExtendBoardingPassFileSelect(e, costIdx)} />
                                            </label>
                                        {/if}
                                    </div>
                                </div>
                            </div>
                        </div>
                    {/if}
                {/each}
            {/if}

            {#if !isReadOnly}
                <div class="w-full pt-2">
                    <button type="button" class="w-full flex items-center justify-center gap-1.5 px-3 py-2 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-xs font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm" on:click={addExtendTicket}>
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                        Extend Tiket
                    </button>
                </div>
            {/if}
        </div>

        <!-- Hotel -->
        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-3 w-full mb-2">
            <div class="flex justify-between items-center mb-1 min-h-[32px]">
                <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5 md:gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                    </svg>
                    Penginapan (Hotel)
                </h4>
                <div class="flex items-center gap-2">
                    {#if !isReadOnly}
                        <button 
                            type="button" 
                            class="text-[10px] md:text-xs font-bold uppercase tracking-wider text-indigo-600 bg-indigo-50 hover:bg-indigo-100 hover:text-indigo-700 px-2 py-1 rounded-md border border-indigo-200 transition-colors flex items-center gap-1"
                            on:click={() => openSplitHotelModal()}
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" /></svg>
                            Bagi Biaya
                        </button>
                    {/if}
                    {#if !detail.hotelFile && !isReadOnly}
                        <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                            Kwitansi
                            <input type="file" class="hidden" on:change={(e) => handleSpecificFileSelect(e, 'hotelFile')} />
                        </label>
                    {/if}
                </div>
            </div>

            <div class="grid grid-cols-3 gap-3 md:gap-4 w-full">
                <div class="space-y-1.5 col-span-1">
                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Malam</Label>
                    <Input type="number" bind:value={detail.hotelDays} disabled={isReadOnly} class="h-9 md:h-10 text-sm bg-slate-50 border-slate-200 px-2 {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                </div>
                <div class="space-y-1.5 col-span-2">
                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Rate per Malam</Label>
                    <div class="relative">
                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                        <Input type="text" value={formatInputNumber(detail.hotelRate)} on:input={(e) => updateCost('hotelRate', e)} disabled={isReadOnly} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-slate-50 border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                    </div>
                </div>
            </div>
            
            {#if detail.hotelFile}
                <div class="flex items-center justify-between p-2 mt-2 bg-slate-50 border border-slate-200 rounded-md shadow-sm w-full">
                    <div class="flex items-center gap-2 min-w-0 flex-1">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                        <span class="text-[10px] md:text-xs text-slate-700 truncate">{detail.hotelFile.name}</span>
                    </div>
                    <div class="flex gap-1.5 shrink-0 ml-2">
                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-blue-50 hover:bg-blue-100 text-blue-700 rounded border border-blue-200 transition-colors" on:click={() => openPreview(detail.hotelFile)}>Lihat</button>
                        {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => removeSpecificFile('hotelFile')}>Hapus</button>{/if}
                    </div>
                </div>
            {/if}

            <!-- Extend Penginapan Items -->
            {#if detail.additionalCosts && detail.additionalCosts.length > 0}
                {#each detail.additionalCosts as cost, costIdx}
                    {#if cost.name === 'Extend Penginapan'}
                        <div class="col-span-1 md:col-span-3 mt-4 relative pt-4 border-t border-slate-200 w-full">
                            <div class="flex justify-between items-center mb-3 min-h-[32px] pr-8 relative">
                                {#if !isReadOnly}
                                    <button type="button" class="absolute -top-1 -right-1 bg-red-100 text-red-600 rounded-full p-1 border border-red-200 hover:bg-red-200 transition-colors shadow-sm z-10" on:click={() => removeAdditionalCost(costIdx)}>
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
                                            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                        </svg>
                                    </button>
                                {/if}
                                <h4 class="text-[10px] md:text-xs font-bold uppercase text-blue-600 tracking-wider flex items-center gap-1.5 md:gap-2">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                                    </svg>
                                    Extend Penginapan
                                </h4>
                                <div class="flex items-center gap-2">
                                    {#if !isReadOnly}
                                        <button
                                            type="button"
                                            class="text-[10px] md:text-xs font-bold uppercase tracking-wider text-indigo-600 bg-indigo-50 hover:bg-indigo-100 hover:text-indigo-700 px-2 py-1 rounded-md border border-indigo-200 transition-colors flex items-center gap-1"
                                            on:click={() => openSplitHotelModal(true, costIdx)}
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" /></svg>
                                            Bagi Biaya
                                        </button>
                                    {/if}
                                    {#if !cost.file && !isReadOnly}
                                        <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                            Kwitansi
                                            <input type="file" class="hidden" on:change={(e) => handleAdditionalFileSelect(e, costIdx)} />
                                        </label>
                                    {/if}
                                </div>
                            </div>
                            
                            <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                                <div class="grid grid-cols-3 gap-3 md:gap-4 w-full">
                                    <div class="space-y-1.5 col-span-1">
                                        <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Malam</Label>
                                        <Input type="number" disabled={isReadOnly} bind:value={editingCosts.details[selectedLocationIndex].additionalCosts[costIdx].hotelDays} on:input={(e) => updateExtendPenginapanCost(costIdx, 'hotelDays', e)} class="h-9 md:h-10 text-sm bg-white border-slate-200 px-2 {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                                    </div>
                                    <div class="space-y-1.5 col-span-2">
                                        <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Rate per Malam</Label>
                                        <div class="relative">
                                            <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                            <Input type="text" disabled={isReadOnly} value={formatInputNumber(cost.hotelRate)} on:input={(e) => updateExtendPenginapanCost(costIdx, 'hotelRate', e)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                                        </div>
                                    </div>
                                </div>
                                
                                {#if cost.file}
                                    <div class="mt-3">
                                        <div class="flex items-center justify-between p-2 bg-white border border-slate-200 rounded-md shadow-sm">
                                            <div class="flex items-center gap-2 min-w-0 flex-1">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                                <span class="text-[10px] md:text-xs text-slate-700 truncate">{cost.file.name}</span>
                                            </div>
                                            <div class="flex gap-1.5 shrink-0 ml-2">
                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-blue-50 hover:bg-blue-100 text-blue-700 rounded border border-blue-200 transition-colors" on:click={() => openPreview(cost.file)}>Lihat</button>
                                                {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => { editingCosts.details[selectedLocationIndex].additionalCosts[costIdx].file = null; editingCosts = editingCosts; }}>Hapus</button>{/if}
                                            </div>
                                        </div>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    {/if}
                {/each}
            {/if}

                {#if !isReadOnly}
                    <div class="w-full pt-2">
                        <button type="button" class="w-full flex items-center justify-center gap-1.5 px-3 py-2 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-xs font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm" on:click={addExtendPenginapan}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                            Extend Penginapan
                        </button>
                    </div>
                {/if}

                <div class="text-right text-xs md:text-sm font-mono font-medium text-slate-600 border-t border-slate-100 pt-2 mt-1">
                    Subtotal Hotel: <span class="text-slate-800">{formatCurrency(currentTotalHotel)}</span>
                </div>
            </div>

            <!-- Bukti Transportasi atau Rental -->
            <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-3 w-full mb-2">
                <div class="flex justify-between items-center mb-1 min-h-[32px]">
                    <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                        </svg>
                        Transport Daerah
                    </h4>
                    {#if !detail.transportFile && !isReadOnly}
                        <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                            Bukti
                            <input type="file" class="hidden" on:change={(e) => handleSpecificFileSelect(e, 'transportFile')} />
                        </label>
                    {/if}
                </div>

                <div class="space-y-1.5 w-full">
                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Total Biaya (Opsional)</Label>
                    <div class="relative">
                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                        <Input type="text" value={formatInputNumber(detail.transportAmount)} on:input={(e) => updateCost('transportAmount', e)} disabled={isReadOnly} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-slate-50 border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                    </div>
                </div>

                {#if detail.transportFile}
                    <div class="flex items-center justify-between p-2 mt-2 bg-slate-50 border border-slate-200 rounded-md shadow-sm w-full">
                        <div class="flex items-center gap-2 min-w-0 flex-1">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                            <span class="text-[10px] md:text-xs text-slate-700 truncate">{detail.transportFile.name}</span>
                        </div>
                        <div class="flex gap-1.5 shrink-0 ml-2">
                            <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-blue-50 hover:bg-blue-100 text-blue-700 rounded border border-blue-200 transition-colors" on:click={() => openPreview(detail.transportFile)}>Lihat</button>
                            {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => removeSpecificFile('transportFile')}>Hapus</button>{/if}
                        </div>
                    </div>
                {/if}
            </div>

            <!-- Add Cost / Additional Costs -->
            <div class="p-3 md:p-4 bg-slate-50 rounded-xl border border-slate-200 shadow-sm space-y-4 w-full mb-2">
                <div class="flex justify-between items-center border-b border-slate-200 pb-2 min-h-[32px]">
                    <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                        </svg>
                        Transport Lokal
                    </h4>
                    {#if !isReadOnly}
                        <button class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-indigo-600 text-white border border-indigo-700 text-[10px] font-bold uppercase tracking-wider hover:bg-indigo-700 transition-all shadow-sm" on:click={addAdditionalCost}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                            Tambah Biaya
                        </button>
                    {/if}
                </div>

                <div class="space-y-3">
                    {#if detail.additionalCosts && detail.additionalCosts.filter(c => c.name !== 'Extend Tiket' && c.name !== 'Extend Penginapan').length > 0}
                        {#each detail.additionalCosts as cost, index}
                            {#if cost.name !== 'Extend Tiket' && cost.name !== 'Extend Penginapan'}
                            <div class="bg-white p-3 border border-slate-200 rounded-lg relative group shadow-sm">
                                {#if !isReadOnly}
                                    <button type="button" class="absolute -top-2 -right-2 bg-red-100 text-red-600 rounded-full p-1 border border-red-200 hover:bg-red-200 transition-colors shadow-md z-10" on:click={() => removeAdditionalCost(index)}>
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
                                            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                        </svg>
                                    </button>
                                {/if}
                                <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                                    <div class="space-y-1.5">
                                        <Label class="text-[10px] font-semibold uppercase text-slate-500 tracking-wider">Nama Biaya</Label>
                                        <Input type="text" placeholder="Cth: Taksi Bandara" bind:value={cost.name} disabled={isReadOnly} class="h-8 text-xs bg-slate-50 border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                                    </div>
                                    <div class="space-y-1.5">
                                        <Label class="text-[10px] font-semibold uppercase text-slate-500 tracking-wider">Nominal</Label>
                                        <div class="relative">
                                            <span class="absolute left-2 top-1.5 text-slate-400 text-xs">Rp</span>
                                            <Input type="text" value={formatInputNumber(cost.amount)} on:input={(e) => updateAdditionalCostAmount(index, e)} disabled={isReadOnly} class="pl-7 h-8 text-xs bg-slate-50 border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                                        </div>
                                    </div>
                                    <div class="md:col-span-2 flex flex-col sm:flex-row sm:items-center justify-between mt-1 pt-2 border-t border-slate-100 gap-2">
                                        <span class="text-[10px] font-medium text-slate-500 uppercase tracking-tight">Kwitansi / Bukti (.pdf):</span>
                                        {#if !cost.file && !isReadOnly}
                                            <label class="cursor-pointer inline-flex items-center gap-1.5 px-2 py-1 rounded bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                                Upload Kwitansi
                                                <input type="file" class="hidden" on:change={(e) => handleAdditionalFileSelect(e, index)} />
                                            </label>
                                        {:else if cost.file}
                                            <div class="flex items-center gap-2 bg-slate-50 p-1.5 rounded border border-slate-200 w-full sm:w-auto shadow-sm">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                                <span class="text-[10px] text-slate-700 truncate max-w-[120px]">{cost.file.name}</span>
                                                <div class="flex gap-1.5 ml-auto sm:ml-2">
                                                    <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[9px] font-bold uppercase tracking-wider bg-blue-50 hover:bg-blue-100 text-blue-700 rounded border border-blue-200 transition-colors" on:click={() => openPreview(cost.file)}>Lihat</button>
                                                    {#if !isReadOnly}
                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[9px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => { cost.file = null; editingCosts = editingCosts; }}>Hapus</button>
                                                    {/if}
                                                </div>
                                            </div>
                                        {/if}
                                    </div>
                                </div>
                            </div>
                            {/if}
                        {/each}
                        <div class="text-right text-xs md:text-sm font-mono font-medium text-slate-600 pt-2">
                            Subtotal Tambahan: <span class="text-slate-800">{formatCurrency(currentTotalAdditional)}</span>
                        </div>
                    {:else}
                        <div class="text-center p-4 border border-dashed border-slate-300 rounded-lg text-xs text-slate-500">
                            Belum ada biaya tambahan diinputkan untuk provinsi ini.
                        </div>
                    {/if}
                </div>
            </div>

            {/if}
        </div>

        <DialogFooter class="p-4 md:p-6 bg-white shrink-0 z-10 border-t border-slate-100 shadow-[0_-10px_15px_-3px_rgba(0,0,0,0.05)]">
            <div class="w-full flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <div class="w-full lg:w-auto text-left flex-shrink-0 bg-slate-50 p-3 rounded-xl border border-slate-100">
                    <span class="block text-[10px] md:text-xs text-slate-500 font-bold uppercase tracking-widest mb-1">Grand Total Estimasi</span>
                    <span class="text-xl md:text-2xl font-black text-blue-700 font-mono tracking-tight">{formatCurrency(grandTotal)}</span>
                </div>
                <div class="flex flex-wrap items-center justify-end gap-2.5 w-full lg:w-auto">
                    <Button variant="outline" class="h-10 md:h-11 border-slate-200 text-slate-600 hover:text-slate-800 hover:bg-slate-50 bg-white px-4 flex-1 sm:flex-none justify-center items-center transition-all rounded-xl shadow-sm" on:click={() => {
                        window.open(`/print?type=spd&id=${record.id}&spd=${encodeURIComponent(record.spd)}`, '_blank');
                    }}>
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-2 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><line x1="16" x2="8" y1="13" y2="13"/><line x1="16" x2="8" y1="17" y2="17"/><line x1="10" x2="8" y1="9" y2="9"/></svg>
                        <span class="text-xs font-bold uppercase tracking-wider">Cetak SPD</span>
                    </Button>
                    <Button variant="outline" class="h-10 md:h-11 border-slate-200 text-slate-600 hover:text-slate-800 hover:bg-slate-50 bg-white px-4 flex-1 sm:flex-none justify-center items-center transition-all rounded-xl shadow-sm" on:click={() => {
                        window.open(`/print?type=rincian&id=${record.id}&spd=${encodeURIComponent(record.spd)}`, '_blank');
                    }}>
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-2 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="16" height="20" x="4" y="2" rx="2"/><line x1="8" x2="16" y1="6" y2="6"/><line x1="16" x2="16" y1="14" y2="18"/><path d="M16 10h.01"/><path d="M12 10h.01"/><path d="M8 10h.01"/><path d="M12 14h.01"/><path d="M8 14h.01"/><path d="M12 18h.01"/><path d="M8 18h.01"/></svg>
                        <span class="text-xs font-bold uppercase tracking-wider">Cetak Rincian</span>
                    </Button>
                    
                    <div class="w-px h-8 bg-slate-200 hidden sm:block mx-1"></div>

                    <Button variant="outline" class="h-10 md:h-11 border-slate-200 text-slate-500 hover:text-red-600 hover:bg-red-50 hover:border-red-200 bg-white px-4 flex-1 sm:flex-none justify-center items-center transition-all rounded-xl shadow-sm" on:click={() => dispatch('close')}>
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-2 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
                        <span class="text-xs font-bold uppercase tracking-wider">Tutup</span>
                    </Button>

                    {#if !isReadOnly}
                        <Button variant="outline" class="h-10 md:h-11 border-red-200 text-red-600 hover:text-white hover:bg-red-600 bg-red-50 px-4 flex-1 sm:flex-none justify-center items-center transition-all rounded-xl shadow-sm" on:click={() => dispatch('reject', { record })}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-2 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m18 6-12 12"/><path d="m6 6 12 12"/></svg>
                            <span class="text-xs font-bold uppercase tracking-wider">Tolak Pengajuan</span>
                        </Button>
                        <Button class="h-10 md:h-11 bg-blue-600 hover:bg-blue-700 active:bg-blue-800 text-white shadow-lg shadow-blue-500/25 px-6 flex-1 sm:flex-none justify-center items-center transition-all rounded-xl" on:click={handleSave}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-2 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
                            <span class="text-xs font-bold uppercase tracking-wider">Simpan Perubahan</span>
                        </Button>
                    {/if}
                </div>
            </div>
        </DialogFooter>

        {#if showPreview && previewFile}
            <div class="absolute inset-0 z-50 bg-white flex flex-col">
                <div class="flex items-center justify-between px-4 py-3 border-b border-slate-100 bg-slate-50/50 flex-none">
                    <div class="flex flex-col min-w-0 pr-4">
                        <h3 class="text-sm font-bold text-slate-800 tracking-tight truncate">Preview Dokumen</h3>
                        <p class="text-[10px] text-slate-500 truncate">{previewFile.name}</p>
                    </div>
                    <button type="button" class="p-2 -mr-2 text-slate-400 hover:text-red-500 hover:bg-slate-100 rounded-full transition-colors flex-shrink-0" on:click={closePreview}>
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                <div class="flex-1 overflow-hidden bg-slate-100 relative">
                    <DocumentViewer 
                        url={previewFile.data || (previewFile.path ? '/uploads/' + previewFile.path.replace(/^\/?uploads\//, '') : '')} 
                        type={previewFile.type && previewFile.type.startsWith('image/') ? 'image' : (previewFile.type === 'application/pdf' ? 'pdf' : 'docx')} 
                        filename={previewFile.name} 
                    />
                </div>
            </div>
        {/if}

        <!-- Split Hotel Modal -->
        {#if showSplitHotelModal}
            <div class="absolute inset-0 z-[60] bg-black/40 backdrop-blur-sm flex items-center justify-center p-4">
                <div class="w-full max-w-lg flex flex-col bg-white rounded-2xl shadow-2xl overflow-hidden border border-slate-100">
                    <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
                        <h3 class="text-lg font-bold text-slate-800 flex items-center gap-2">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                            </svg>
                            Bagi Biaya Kamar (Split Bill)
                        </h3>
                        <p class="text-xs text-slate-500 mt-1">Kalkulator otomatis untuk membagi biaya penginapan secara merata.</p>
                    </div>
                    
                    <div class="p-6 space-y-6">
                        <div class="grid grid-cols-3 gap-4">
                            <div class="col-span-1 space-y-1.5">
                                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Durasi (Malam)</Label>
                                <Input type="number" bind:value={splitHotelData.days} class="bg-slate-50 border-slate-200" min="1" />
                            </div>
                            <div class="col-span-2 space-y-1.5">
                                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Harga Kamar Per Malam (Rp)</Label>
                                <div class="relative">
                                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                                    <Input 
                                        type="text" 
                                        value={formatInputNumber(splitHotelData.totalBill)} 
                                        on:input={(e) => {
                                            const raw = e.target.value.replace(/[^0-9]/g, '');
                                            splitHotelData.totalBill = parseInt(raw, 10) || 0;
                                        }}
                                        class="pl-9 bg-white border-slate-200 font-semibold" 
                                    />
                                </div>
                            </div>
                        </div>

                        <div class="space-y-3">
                            <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Bagikan Dengan ({splitHotelData.selectedEmpIds.length} orang dipilih)</Label>
                            <div class="bg-slate-50 border border-slate-200 rounded-xl overflow-hidden max-h-48 overflow-y-auto custom-scrollbar">
                                {#each allRecordsInSpd.filter(r => r.id !== record.id) as emp (emp.id)}
                                    <label class="flex items-center gap-3 p-3 border-b border-slate-100 last:border-0 hover:bg-slate-100/50 cursor-pointer transition-colors">
                                        <div class="relative flex items-start">
                                            <div class="flex items-center h-5">
                                                <input 
                                                    type="checkbox" 
                                                    class="w-4 h-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer transition-all"
                                                    checked={splitHotelData.selectedEmpIds.includes(emp.id)}
                                                    on:change={() => toggleSplitEmp(emp.id)}
                                                />
                                            </div>
                                        </div>
                                        <div class="flex-1 min-w-0">
                                            <div class="text-sm font-semibold text-slate-800 truncate">{emp.employee?.name || 'Petugas'}</div>
                                        </div>
                                    </label>
                                {/each}
                                {#if allRecordsInSpd.length <= 1}
                                    <div class="p-4 text-center text-xs text-slate-500">
                                        Tidak ada petugas lain dalam perjalanan ini.
                                    </div>
                                {/if}
                            </div>
                        </div>

                        <div class="bg-indigo-50 border border-indigo-100 rounded-xl p-4 flex justify-between items-center">
                            <div>
                                <div class="text-xs font-bold uppercase text-indigo-500 tracking-wider mb-0.5">Total per Orang</div>
                                <div class="text-sm text-slate-600 font-medium">{1 + splitHotelData.selectedEmpIds.length} Orang</div>
                            </div>
                            <div class="text-right">
                                <div class="text-xl font-bold text-indigo-700">
                                    {formatCurrency(1 + splitHotelData.selectedEmpIds.length > 0 ? (splitHotelData.totalBill / (1 + splitHotelData.selectedEmpIds.length)) : 0)}
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="px-6 py-4 bg-slate-50 border-t border-slate-100 flex justify-end gap-3">
                        <Button variant="outline" class="border-slate-200 text-slate-600 hover:bg-slate-100" on:click={() => showSplitHotelModal = false}>Batal</Button>
                        <Button class="bg-indigo-600 hover:bg-indigo-700 text-white shadow-md shadow-indigo-500/20" on:click={applySplitHotel}>Terapkan Pembagian</Button>
                    </div>
                </div>
            </div>
        {/if}
    </Dialog>