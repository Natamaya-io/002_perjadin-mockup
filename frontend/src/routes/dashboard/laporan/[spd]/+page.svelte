<script>
    function getDisplayName(name) {
        if (!name) return '';
        const parts = name.split('/');
        const cleanName = parts[parts.length - 1];
        if (cleanName.length > 36 && cleanName.includes('_')) {
            const splitName = cleanName.split('_');
            if (splitName[0].length === 36 || splitName[0].length === 32) {
                return 'Surat_Tugas_' + splitName.slice(1).join('_');
            }
        }
        return cleanName;
    }
    import { page } from '$app/stores';
    import { recordsStore, updateRecord, loadRecords, updateMultipleRecords, isFetchingRecords, clearStores } from '$lib/features/pengajuan/store';
    import { userStore } from '$lib/features/auth/store';
    import { loadingStore, startLoading, stopLoading } from '$lib/shared/stores/loading';
    import { provincesStore } from '$lib/shared/stores/master-data';
    import { toast } from '$lib/shared/stores/toast';
    import { api } from '$lib/shared/api';
    import { onMount } from 'svelte';
    import { goto, invalidateAll } from '$app/navigation';
    import { getInitials, toTitleCase, formatCurrency, compressImage, getBlobUrl, generateThumbnailUrl } from '$lib/shared/utils/utils';
    import { fade } from 'svelte/transition';
    import { cn } from '$lib/shared/utils/utils';
    
    // UI
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Textarea from '$lib/shared/ui/textarea/Textarea.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import DocumentViewer from '$lib/shared/ui/document-viewer/DocumentViewer.svelte';
    import Dialog from '$lib/shared/ui/dialog/Dialog.svelte';

    async function uploadAndGetPath(file) {
        if (!file) return null;
        startLoading('Mengupload file...');
        try {
            const res = await api.uploadFile(file);
            return res.path;
        } catch(err) {
            toast.error("Gagal mengupload file");
            return null;
        } finally {
            stopLoading();
        }
    }

    let spd = $page.params.spd || ''; 
    $: currentTab = $page.url.searchParams.get('tab') || 'laporan';
    
    $: allRecordsForSpd = spd ? $recordsStore.filter(r => r.spd === decodeURIComponent(spd)).sort((a,b) => (b.sequenceNumber || 0) - (a.sequenceNumber || 0)) : [];
    $: record = allRecordsForSpd.find(r => r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email)) || allRecordsForSpd[0];
    
    // Check if the current user is part of this SPD group
    $: isUserInSpdGroup = allRecordsForSpd.some(r => r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email));
    
    // Allow access to all records if user is admin/kasubag. 
    // If user is protokol, they can see and edit both 'laporan' and 'rincian' 
    // for all members in the same SPD group.
    $: recordsList = allRecordsForSpd;
    
    // Use the same deterministic sort as admin/perdin page for consistent SPD sub-numbers
    $: nomorSpdPetugas = record ? String(record.sequenceNumber || 0).padStart(3, '0') : '000';

    let reportText = '';
    let manualSuratTugasNumber = '';
    let manualSuratTugasDate = '';
    let tanggalMerahList = [];
    let newTanggalMerah = '';
    // Track record ID untuk deteksi navigasi antar SPD
    let loadedForRecordId = '';
    
    // Prevent false positive "Not Found" error during initial SWR fetch
    let initialLoad = true;
    
    // Signatories
    let ppkName = '';
    let ppkNip = '';
    let bendaharaName = '';
    let bendaharaNip = '';

    /** @type {Array<{name: string, type: string, data: string, timestamp: string}>} */
    let uploadedFiles = []; 
    let isDragging = false;
    
    /** @type {{name: string, type: string, data: string, timestamp: string} | null} */
    let sppdFile = null;
    /** @type {{name: string, type: string, data: string, timestamp: string} | null} */
    let suratTugasFile = null;

    // Preview Modal State
    let isPreviewOpen = false;
    /** @type {{name: string, type: string, data: string} | null} */
    let previewFile = null;

    // Split Hotel Modal State
    let showSplitHotelModal = false;
    /** @type {{ sourceEmpId: string | null, locationIndex: number, totalBill: number, days: any, selectedEmpIds: string[], isExtend: boolean, extendIdx: number }} */
    let splitHotelData = {
        sourceEmpId: null,
        locationIndex: 0,
        totalBill: 0,
        days: 0,
        selectedEmpIds: [],
        isExtend: false,
        extendIdx: -1
    };

    function openSplitHotelModal(empId, locationIndex, isExtend = false, extendIdx = -1) {
        try {
            const detail = localCosts[empId].costs.details[locationIndex];
            let totalBill = 0;
            let days = 0;
            
            if (isExtend && extendIdx > -1) {
                const extendCost = detail.additionalCosts[extendIdx];
                totalBill = (extendCost.hotelRate || 0) * (extendCost.hotelDays || 0);
                days = extendCost.hotelDays || 0;
            } else {
                totalBill = (detail.hotelRate || 0) * (detail.hotelDays || 0);
                days = detail.hotelDays || 0;
            }

            splitHotelData = {
                sourceEmpId: empId,
                locationIndex: locationIndex,
                totalBill: totalBill,
                days: days,
                selectedEmpIds: [empId], // Automatically select the initiator
                isExtend: isExtend,
                extendIdx: extendIdx
            };
            showSplitHotelModal = true;
        } catch (e) {
            console.error("Error opening split modal:", e);
            toast.error("Terjadi kesalahan sistem: " + e.message);
        }
    }

    function toggleSplitEmp(empId) {
        if (splitHotelData.selectedEmpIds.includes(empId)) {
            splitHotelData.selectedEmpIds = splitHotelData.selectedEmpIds.filter(id => id !== empId);
        } else {
            splitHotelData.selectedEmpIds = [...splitHotelData.selectedEmpIds, empId];
        }
    }

    function applySplitHotel() {
        const totalPeople = splitHotelData.selectedEmpIds.length;
        if (totalPeople === 0) {
            toast.error("Minimal 1 orang dipilih.");
            return;
        }
        if (splitHotelData.days <= 0) {
            toast.error("Durasi malam tidak boleh 0.");
            return;
        }

        const ratePerNightPerPerson = Math.round(splitHotelData.totalBill / totalPeople / splitHotelData.days);
        
        let sourceFile = null;
        if (splitHotelData.isExtend && splitHotelData.extendIdx > -1) {
            sourceFile = localCosts[splitHotelData.sourceEmpId].costs.details[splitHotelData.locationIndex].additionalCosts[splitHotelData.extendIdx].file;
            
            // Apply to source person
            const sourceCost = localCosts[splitHotelData.sourceEmpId].costs.details[splitHotelData.locationIndex].additionalCosts[splitHotelData.extendIdx];
            if (splitHotelData.selectedEmpIds.includes(splitHotelData.sourceEmpId)) {
                sourceCost.hotelRate = ratePerNightPerPerson;
                sourceCost.hotelDays = splitHotelData.days;
                sourceCost.amount = ratePerNightPerPerson * splitHotelData.days;
            } else {
                sourceCost.hotelRate = 0;
                sourceCost.hotelDays = 0;
                sourceCost.amount = 0;
            }
            recalculateTotal(splitHotelData.sourceEmpId);

            // Apply to selected others
            for (const empId of splitHotelData.selectedEmpIds.filter(id => id !== splitHotelData.sourceEmpId)) {
                if (localCosts[empId]) {
                    // Ensure additionalCosts array exists
                    if (!localCosts[empId].costs.details[splitHotelData.locationIndex].additionalCosts) {
                        localCosts[empId].costs.details[splitHotelData.locationIndex].additionalCosts = [];
                    }
                    
                    // We need to either update existing Extend Penginapan or add a new one
                    let targetExtendIdx = localCosts[empId].costs.details[splitHotelData.locationIndex].additionalCosts.findIndex(c => c.name === 'Extend Penginapan');
                    
                    if (targetExtendIdx === -1) {
                        localCosts[empId].costs.details[splitHotelData.locationIndex].additionalCosts.push({
                            name: 'Extend Penginapan',
                            hotelRate: ratePerNightPerPerson,
                            hotelDays: splitHotelData.days,
                            amount: ratePerNightPerPerson * splitHotelData.days,
                            file: sourceFile ? { ...sourceFile } : null
                        });
                    } else {
                        const targetCost = localCosts[empId].costs.details[splitHotelData.locationIndex].additionalCosts[targetExtendIdx];
                        targetCost.hotelRate = ratePerNightPerPerson;
                        targetCost.hotelDays = splitHotelData.days;
                        targetCost.amount = ratePerNightPerPerson * splitHotelData.days;
                        if (sourceFile) {
                            targetCost.file = { ...sourceFile };
                        }
                    }
                    recalculateTotal(empId);
                }
            }
        } else {
            sourceFile = localCosts[splitHotelData.sourceEmpId].costs.details[splitHotelData.locationIndex].hotelFile;

            // Apply to source person
            if (splitHotelData.selectedEmpIds.includes(splitHotelData.sourceEmpId)) {
                localCosts[splitHotelData.sourceEmpId].costs.details[splitHotelData.locationIndex].hotelRate = ratePerNightPerPerson;
                localCosts[splitHotelData.sourceEmpId].costs.details[splitHotelData.locationIndex].hotelDays = splitHotelData.days;
            } else {
                localCosts[splitHotelData.sourceEmpId].costs.details[splitHotelData.locationIndex].hotelRate = 0;
                localCosts[splitHotelData.sourceEmpId].costs.details[splitHotelData.locationIndex].hotelDays = 0;
            }
            recalculateTotal(splitHotelData.sourceEmpId);

            // Apply to selected others
            for (const empId of splitHotelData.selectedEmpIds.filter(id => id !== splitHotelData.sourceEmpId)) {
                if (localCosts[empId]) {
                    localCosts[empId].costs.details[splitHotelData.locationIndex].hotelRate = ratePerNightPerPerson;
                    localCosts[empId].costs.details[splitHotelData.locationIndex].hotelDays = splitHotelData.days;
                    if (sourceFile) {
                        localCosts[empId].costs.details[splitHotelData.locationIndex].hotelFile = { ...sourceFile };
                    }
                    recalculateTotal(empId);
                }
            }
        }

        localCosts = { ...localCosts };
        showSplitHotelModal = false;
        toast.success(`Biaya penginapan dibagi ke ${totalPeople} orang.`);
    }

    let localCosts = {};

    $: if (recordsList.length > 0 && Object.keys(localCosts).length === 0) {
        recordsList.forEach(r => {
            const locations = getLocations(r);
            let details = r.costs?.details || [];
            
            // Migration for legacy data if details is empty but costs exist
            if (details.length === 0 && r.costs && (r.costs.hotelDays || r.costs.transportAmount || r.costs.ticketGo)) {
                 details = [{ ...r.costs, details: undefined }];
            }
            
            // Ensure details array matches locations length
            if (details.length < locations.length) {
                for (let i = details.length; i < locations.length; i++) {
                    details.push({
                        transportMode: 'Pesawat/Kendaraan Umum',
                        ticketGo: 0,
                        ticketBack: 0,
                        hotelDays: 0,
                        hotelRate: 0,
                        transportAmount: 0,
                        additionalCosts: [],
                        boardingPassFiles: []
                    });
                }
            }

            localCosts[r.id] = { 
                costs: { 
                    ...r.costs, 
                    details: details 
                }, 
                totalCost: r.totalCost || 0,
                selectedLocationIndex: 0
            };
        });
        localCosts = { ...localCosts }; // Trigger reactivity
        updateTotals(recordsList, $provincesStore);
    }

    // Use a reactive statement that only depends on recordsList and provincesStore to prevent Svelte 5 infinite loops
    $: updateTotals(recordsList, $provincesStore);

    function updateTotals(records, provinces) {
        if (!records || Object.keys(localCosts).length === 0 || !provinces) return;
        let updated = false;
        for (let r of records) {
            const empId = r.id;
            if (localCosts[empId] && localCosts[empId].costs.details) {
                const locations = getLocations(r);
                let grandTotal = 0;

                // Calculate total across all locations
                localCosts[empId].costs.details.forEach((detail, idx) => {
                    const loc = locations[idx];
                    if (!loc) return;
                    
                    const days = getDaysForLocation(loc);
                    const rate = provinces.find(p => p.name === loc.province)?.luarKota || 0; // SBM Rate
                    const sbmTotal = days * rate;

                    const totalHotel = ((detail.hotelDays || 0) * (detail.hotelRate || 0)) + (detail.additionalCosts || []).reduce((sum, c) => c.name === 'Extend Penginapan' ? sum + (Number(c.amount) || 0) : sum, 0);
                    const totalTicket = Number(detail.ticketGo || 0) + Number(detail.ticketBack || 0);
                    const totalTransportAmount = Number(detail.transportAmount || 0);
                    const totalAdditional = (detail.additionalCosts || []).reduce((sum, cost) => {
                        if (cost.name === 'Extend Tiket') {
                            return sum + (Number(cost.ticketGo) || 0) + (Number(cost.ticketBack) || 0) + (Number(cost.amount) || 0);
                        } else if (cost.name === 'Extend Penginapan') {
                            return sum;
                        }
                        return sum + (Number(cost.amount) || 0);
                    }, 0);
                    
                    grandTotal += sbmTotal + totalHotel + totalTicket + totalTransportAmount + totalAdditional;
                });
                
                if (localCosts[empId].totalCost !== grandTotal) {
                    localCosts[empId].totalCost = grandTotal;
                    updated = true;
                }
            }
        }
        if (updated) {
            localCosts = { ...localCosts };
        }
    }

	function getTotalDays(empRecord) {
    	const locs = empRecord?.locations && empRecord.locations.length > 0 
       	 ? empRecord.locations 
        	: [{ startDate: empRecord?.startDate, endDate: empRecord?.endDate, province: empRecord?.province }];
    
    	return locs.reduce((total, loc) => {
        	const start = new Date(loc.startDate);
        	const end = new Date(loc.endDate);
        	let locDays = 0;
        	if (!isNaN(start.getTime()) && !isNaN(end.getTime())) {
            	const diffTime = end.getTime() - start.getTime();
            	const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
            	locDays = diffDays > 0 ? diffDays : 0;
        	}
        	return total + locDays;
    	}, 0);
	}

    function getSbmRate(empRecord) {
        const locs = empRecord?.locations && empRecord.locations.length > 0
            ? empRecord.locations
            : [{ startDate: empRecord?.startDate, endDate: empRecord?.endDate, province: empRecord?.province }];

        let totalSbm = 0;
        let totalDays = 0;

        locs.forEach(loc => {
            const provData = $provincesStore.find(p => p.name === loc.province);
            const rate = provData ? provData.luarKota : 0;
            const start = new Date(loc.startDate);
            const end = new Date(loc.endDate);
            let locDays = 0;

            if (!isNaN(start.getTime()) && !isNaN(end.getTime())) {
                locDays = Math.ceil((end.getTime() - start.getTime()) / (1000 * 3600 * 24)) + 1;
            } else {
                locDays = 1; // Fallback
            }

            totalSbm += rate * locDays;
            totalDays += locDays;
        });

        return totalDays > 0 ? totalSbm / totalDays : 0;
    }

	function getTotalDailyAllowance(empRecord) {
		const locs = empRecord?.locations && empRecord.locations.length > 0 
	   	 ? empRecord.locations 
	    	: [{ startDate: empRecord?.startDate, endDate: empRecord?.endDate, province: empRecord?.province }];

		return locs.reduce((total, loc) => {
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
	    	return total + (rate * locDays);
		}, 0);
	}

	function getLocations(empRecord) {
	     if (empRecord?.locations && empRecord.locations.length > 0) {
             return empRecord.locations;
         }
         // SOTA: Cache fallback array onto the record object to prevent breaking Svelte {#each} reference equality
         if (!empRecord._fallbackLocations) {
             empRecord._fallbackLocations = [{ startDate: empRecord?.startDate, endDate: empRecord?.endDate, province: empRecord?.province, location: empRecord?.location }];
         }
         return empRecord._fallbackLocations;
	}
	function getDaysForLocation(loc) {
	    if (!loc?.startDate || !loc?.endDate) return 0;
	    const start = new Date(loc.startDate);
	    const end = new Date(loc.endDate);
	    let locDays = 0;
	    if (!isNaN(start.getTime()) && !isNaN(end.getTime())) {
	        const diffTime = end.getTime() - start.getTime();
	        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
	        locDays = diffDays > 0 ? diffDays : 0;
	    }
	    return locDays;
	}

	function getRateForLocation(loc) {
	     const provData = $provincesStore.find(p => p.name === loc.province);
	     return provData ? provData.luarKota : 0;
	}

    function formatInputNumber(value) {
        if (value === undefined || value === null || value === '') return '';
        const numStr = value.toString().replace(/[^0-9]/g, '');
        if (!numStr) return '';
        return parseInt(numStr, 10).toLocaleString('id-ID');
    }

    function recalculateTotal(empId) {
        if (!localCosts[empId]) return;
        const record = recordsList.find(r => r.id === empId);
        if (!record) return;

        localCosts[empId].totalCost = (localCosts[empId].costs.details || []).reduce((acc, detail, idx) => {
            const loc = record.locations && record.locations[idx] ? record.locations[idx] : null;
            let sbmTotal = 0;
            if (loc) {
                const provData = $provincesStore.find(p => p.name === loc.province);
                const rate = provData ? provData.luarKota : (localCosts[empId].costs.dailyAllowanceRate || 0);
                const start = new Date(loc.startDate);
                const end = new Date(loc.endDate);
                if (!isNaN(start.getTime()) && !isNaN(end.getTime())) {
                    const diffDays = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)) + 1;
                    sbmTotal = rate * (diffDays > 0 ? diffDays : 0);
                }
            }
            const hotel = ((detail.hotelDays || 0) * (detail.hotelRate || 0)) + (detail.additionalCosts || []).reduce((sum, c) => c.name === 'Extend Penginapan' ? sum + (Number(c.amount) || 0) : sum, 0);
            const ticket = Number(detail.ticketGo || 0) + Number(detail.ticketBack || 0);
            const transport = Number(detail.transportAmount || 0);
            const addCosts = (detail.additionalCosts || []).reduce((sum, c) => {
                if (c.name === 'Extend Tiket') {
                    return sum + (Number(c.ticketGo) || 0) + (Number(c.ticketBack) || 0) + (Number(c.amount) || 0);
                } else if (c.name === 'Extend Penginapan') {
                    return sum;
                }
                return sum + (c.amount || 0);
            }, 0);
            return acc + sbmTotal + hotel + ticket + transport + addCosts;
        }, 0);
    }

    let typingTimer;

    function updateCost(empId, field, event, locationIndex) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        localCosts[empId].costs.details[locationIndex][field] = isNaN(num) ? undefined : num;
        
        // SOTA: Update input visually instantly via DOM, bypassing Svelte's heavy component diffing
        event.target.value = formatInputNumber(num);
        
        clearTimeout(typingTimer);
        typingTimer = setTimeout(() => {
            recalculateTotal(empId);
            localCosts = { ...localCosts };
        }, 300);
    }

    function updateAdditionalCostAmount(empId, index, event, locationIndex) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        localCosts[empId].costs.details[locationIndex].additionalCosts[index].amount = isNaN(num) ? undefined : num;
        
        event.target.value = formatInputNumber(num);
        
        clearTimeout(typingTimer);
        typingTimer = setTimeout(() => {
            recalculateTotal(empId);
            localCosts = { ...localCosts };
        }, 300);
    }

    function addAdditionalCost(empId, locationIndex) {
        if (!localCosts[empId].costs.details[locationIndex].additionalCosts) localCosts[empId].costs.details[locationIndex].additionalCosts = [];
        localCosts[empId].costs.details[locationIndex].additionalCosts = [...localCosts[empId].costs.details[locationIndex].additionalCosts, { name: '', amount: undefined, file: null }];
        localCosts = { ...localCosts };
    }

    function addExtendTicket(empId, locationIndex) {
        if (!localCosts[empId].costs.details[locationIndex].additionalCosts) localCosts[empId].costs.details[locationIndex].additionalCosts = [];
        localCosts[empId].costs.details[locationIndex].additionalCosts = [...localCosts[empId].costs.details[locationIndex].additionalCosts, { name: 'Extend Tiket', amount: undefined, file: null }];
        localCosts = { ...localCosts };
    }

    function addExtendPenginapan(empId, locationIndex) {
        if (!localCosts[empId].costs.details[locationIndex].additionalCosts) localCosts[empId].costs.details[locationIndex].additionalCosts = [];
        localCosts[empId].costs.details[locationIndex].additionalCosts = [...localCosts[empId].costs.details[locationIndex].additionalCosts, { name: 'Extend Penginapan', amount: undefined, file: null }];
        localCosts = { ...localCosts };
    }

    function removeAdditionalCost(empId, index, locationIndex) {
        localCosts[empId].costs.details[locationIndex].additionalCosts = localCosts[empId].costs.details[locationIndex].additionalCosts.filter((_, i) => i !== index);
        localCosts = { ...localCosts };
    }

    async function handleSpecificFileSelect(empId, e, field, locationIndex) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }
        file = await compressImage(file);
        const path = await uploadAndGetPath(file);
        if (path) {
            localCosts[empId].costs.details[locationIndex][field] = {
                name: file.name,
                size: file.size,
                type: file.type,
                path: path
            };
            localCosts[empId].costsChanged = true;
            localCosts = { ...localCosts };
        }
        e.target.value = '';
    }

    async function handleBoardingPassFileSelect(empId, e, locationIndex) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }
        
        const detail = localCosts[empId].costs.details[locationIndex];
        if (!detail.boardingPassFiles) {
            detail.boardingPassFiles = [];
            if (detail.boardingPassFile) {
                detail.boardingPassFiles.push(detail.boardingPassFile);
                delete detail.boardingPassFile;
            }
        }

        file = await compressImage(file);
        const path = await uploadAndGetPath(file);
        if (path) {
            localCosts[empId].costs.details[locationIndex].boardingPassFiles.push({
                name: file.name,
                size: file.size,
                type: file.type,
                path: path
            });
            localCosts[empId].costsChanged = true;
            localCosts = { ...localCosts };
        }
        e.target.value = '';
    }

    function removeBoardingPassFile(empId, index, locationIndex) {
        if (localCosts[empId].costs.details[locationIndex].boardingPassFiles) {
            localCosts[empId].costs.details[locationIndex].boardingPassFiles = localCosts[empId].costs.details[locationIndex].boardingPassFiles.filter((_, i) => i !== index);
        }
        localCosts[empId].costsChanged = true;
        localCosts = { ...localCosts };
    }

    function removeSpecificFile(empId, field, locationIndex) {
        localCosts[empId].costs.details[locationIndex][field] = null;
        localCosts[empId].costsChanged = true;
        localCosts = { ...localCosts };
    }

    function updateAdditionalExtendCost(empId, index, field, event, locationIndex) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        localCosts[empId].costs.details[locationIndex].additionalCosts[index][field] = isNaN(num) ? undefined : num;
        
        event.target.value = formatInputNumber(num);
        
        clearTimeout(typingTimer);
        typingTimer = setTimeout(() => {
            recalculateTotal(empId);
            localCosts[empId].costsChanged = true;
            localCosts = { ...localCosts };
        }, 300);
    }

    function updateExtendPenginapanCost(empId, index, field, event, locationIndex) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        const cost = localCosts[empId].costs.details[locationIndex].additionalCosts[index];
        cost[field] = isNaN(num) ? undefined : num;
        cost.amount = (cost.hotelDays || 0) * (cost.hotelRate || 0);
        
        event.target.value = formatInputNumber(num);
        
        clearTimeout(typingTimer);
        typingTimer = setTimeout(() => {
            recalculateTotal(empId);
            localCosts[empId].costsChanged = true;
            localCosts = { ...localCosts };
        }, 300);
    }


    async function handleAdditionalExtendSpecificFileSelect(empId, e, field, index, locationIndex) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }
        file = await compressImage(file);
        const path = await uploadAndGetPath(file);
        if (path) {
            localCosts[empId].costs.details[locationIndex].additionalCosts[index][field] = {
                name: file.name,
                size: file.size,
                type: file.type,
                path: path
            };
            localCosts[empId].costsChanged = true;
            localCosts = { ...localCosts };
        }
        e.target.value = '';
    }

    function removeAdditionalExtendSpecificFile(empId, field, index, locationIndex) {
        localCosts[empId].costs.details[locationIndex].additionalCosts[index][field] = null;
        localCosts[empId].costsChanged = true;
        localCosts = { ...localCosts };
    }

    async function handleAdditionalExtendBoardingPassSelect(empId, e, index, locationIndex) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }

        const detail = localCosts[empId].costs.details[locationIndex].additionalCosts[index];
        if (!detail.boardingPassFiles) {
            detail.boardingPassFiles = [];
        }

        file = await compressImage(file);
        const path = await uploadAndGetPath(file);
        if (path) {
            localCosts[empId].costs.details[locationIndex].additionalCosts[index].boardingPassFiles.push({
                name: file.name,
                size: file.size,
                type: file.type,
                path: path
            });
            localCosts[empId].costsChanged = true;
            localCosts = { ...localCosts };
        }
        e.target.value = '';
    }

    function removeAdditionalExtendBoardingPass(empId, bpIndex, index, locationIndex) {
        if (localCosts[empId].costs.details[locationIndex].additionalCosts[index].boardingPassFiles) {
            localCosts[empId].costs.details[locationIndex].additionalCosts[index].boardingPassFiles = localCosts[empId].costs.details[locationIndex].additionalCosts[index].boardingPassFiles.filter((_, i) => i !== bpIndex);
        }
        localCosts[empId].costsChanged = true;
        localCosts = { ...localCosts };
    }

    async function handleAdditionalFileSelect(empId, e, index, locationIndex) {
        let file = e.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 10MB.`);
            e.target.value = '';
            return;
        }
        file = await compressImage(file);
        const path = await uploadAndGetPath(file);
        if (path) {
            localCosts[empId].costs.details[locationIndex].additionalCosts[index].file = {
                name: file.name,
                size: file.size,
                type: file.type,
                path: path
            };
            localCosts[empId].costsChanged = true;
            localCosts = { ...localCosts };
        }
        e.target.value = '';
    }

    function removeAdditionalFile(empId, index, locationIndex) {
        localCosts[empId].costs.details[locationIndex].additionalCosts[index].file = null;
        localCosts[empId].costsChanged = true;
        localCosts = { ...localCosts };
    }

    function viewFilePreview(file) {
        if (file && file.data) {
            const win = window.open();
            if (win) {
                win.document.write(`<iframe src="${file.data}" frameborder="0" style="border:0; top:0px; left:0px; bottom:0px; right:0px; width:100%; height:100%;" allowfullscreen></iframe>`);
            }
        }
    }

    /** @param {{name: string, type: string, data: string}} file */
    function openPreview(file) {
        previewFile = file;
        isPreviewOpen = true;
    }

    let isDataLoaded = false;
    function addTanggalMerah() {
        if (newTanggalMerah && !tanggalMerahList.includes(newTanggalMerah)) {
            tanggalMerahList = [...tanggalMerahList, newTanggalMerah].sort();
            newTanggalMerah = '';
        }
    }
    
    function removeTanggalMerah(date) {
        tanggalMerahList = tanggalMerahList.filter(d => d !== date);
    }

    $: if (record && !manualSuratTugasNumber && record.suratTugasNumber) { manualSuratTugasNumber = record.suratTugasNumber; }
    $: if (record && !manualSuratTugasDate && record.suratTugasDate && record.suratTugasDate !== '0001-01-01T00:00:00Z') { 
        manualSuratTugasDate = new Date(record.suratTugasDate).toISOString().split('T')[0]; 
    }
    // Guard: load semua field form saat pertama kali atau saat navigasi ke record baru
    $: if (record && record.reportData && (!isDataLoaded || record.id !== loadedForRecordId)) {
        // Only overwrite if switching record OR if local state is empty
        if (record.id !== loadedForRecordId || !reportText) {
            reportText = record.reportData.text || '';
        }
        
        if (record.id !== loadedForRecordId || uploadedFiles.length === 0) {
            const rawFiles = record.reportData.files;
            uploadedFiles = Array.isArray(rawFiles) ? rawFiles : [];
        }

        // Normalisasi: {} dari toJsonb(nil) dianggap null agar tidak render 'file kosong'
        const _rawSppd = record.reportData.sppdFile;
        if (record.id !== loadedForRecordId || !sppdFile) {
            sppdFile = (_rawSppd && typeof _rawSppd === 'object' && !Array.isArray(_rawSppd) && Object.keys(_rawSppd).length > 0) ? _rawSppd : null;
        }

        const _rawST = record.reportData.suratTugasFile;
        if (record.id !== loadedForRecordId || !suratTugasFile) {
            suratTugasFile = (_rawST && typeof _rawST === 'object' && !Array.isArray(_rawST) && Object.keys(_rawST).length > 0) ? _rawST : null;
        }

        if (record.id !== loadedForRecordId || !manualSuratTugasNumber) {
            manualSuratTugasNumber = record.suratTugasNumber || '';
        }

        // Sync tanggalMerahList dari store (hanya saat load awal atau ganti record)
        if (record.id !== loadedForRecordId || tanggalMerahList.length === 0) {
            const _raw = record.reportData.tanggalMerah;
            try {
                if (Array.isArray(_raw)) {
                    tanggalMerahList = [..._raw];
                } else if (typeof _raw === 'string' && _raw.length > 0) {
                    tanggalMerahList = JSON.parse(_raw);
                } else {
                    tanggalMerahList = [];
                }
            } catch {
                tanggalMerahList = [];
            }
        }
        loadedForRecordId = record.id;
        isDataLoaded = true;
    }

    // Reset guard saat navigasi ke SPD berbeda
    $: if (record && record.id && record.id !== loadedForRecordId && isDataLoaded) {
        isDataLoaded = false;
    }

    // [ANTI-OOM FIX]: Single Fetcher reaktif untuk mengambil utuh Base64 yang dipotong oleh Bulk API
    let fullRecordFetchedFor = new Set();
    $: if (record && record.id && !fullRecordFetchedFor.has(record.id)) {
        fullRecordFetchedFor.add(record.id);
        fullRecordFetchedFor = fullRecordFetchedFor; // Svelte reactivity trigger
        fetchFullRecordData(record.id);
    }

    async function fetchFullRecordData(id) {
        try {
            const res = await api.getRecordById(id);
            if (res) {
                // 1. Pulihkan Dokumentasi Kegiatan (Files Base64/Uploads)
                if (record && record.id === id) {
                    const rawFiles = res.reportData?.files;
                    if (rawFiles && Array.isArray(rawFiles) && rawFiles.length > 0) {
                        const processedFiles = [];
                        for (let f of rawFiles) {
                            processedFiles.push({
                                ...f,
                                blobUrl: getBlobUrl(f),
                                thumbnailUrl: await generateThumbnailUrl(f)
                            });
                        }
                        uploadedFiles = processedFiles;
                    }
                    const _rawSppd = res.reportData?.sppdFile;
                    if (_rawSppd && typeof _rawSppd === 'object' && !Array.isArray(_rawSppd) && Object.keys(_rawSppd).length > 0) {
                        sppdFile = { ..._rawSppd, blobUrl: getBlobUrl(_rawSppd) };
                    }
                    const _rawST = res.reportData?.suratTugasFile;
                    if (_rawST && typeof _rawST === 'object' && !Array.isArray(_rawST) && Object.keys(_rawST).length > 0) {
                        suratTugasFile = { ..._rawST, blobUrl: getBlobUrl(_rawST) };
                    }
                }

                // 2. Pulihkan Kwitansi Rincian Biaya (Files Base64/Uploads)
                if (localCosts[id] && res.costs) {
                    const lCost = localCosts[id].costs;
                    lCost.receiptFiles = res.costs.receiptFiles;
                    lCost.ticketGoFile = res.costs.ticketGoFile;
                    lCost.ticketBackFile = res.costs.ticketBackFile;
                    lCost.boardingPassFile = res.costs.boardingPassFile;
                    lCost.hotelFile = res.costs.hotelFile;
                    lCost.transportFile = res.costs.transportFile;
                    lCost.additionalCosts = res.costs.additionalCosts;
                    
                    if (res.costs.details && lCost.details) {
                        res.costs.details.forEach((d, idx) => {
                            if (lCost.details[idx]) {
                                lCost.details[idx].boardingPassFiles = d.boardingPassFiles;
                                lCost.details[idx].ticketGoFile = d.ticketGoFile;
                                lCost.details[idx].ticketBackFile = d.ticketBackFile;
                                lCost.details[idx].hotelFile = d.hotelFile;
                                lCost.details[idx].transportFile = d.transportFile;
                                lCost.details[idx].additionalCosts = d.additionalCosts;
                            }
                        });
                    }
                    localCosts = { ...localCosts }; // Picu reaktivitas Svelte
                }
            }
        } catch(e) {
            console.warn("[AntiGravity] Gagal mengambil data tunggal Base64", e);
        }
    }

    onMount(async () => {
        try {
            const hasSpd = $recordsStore.some(r => r.spd === decodeURIComponent(spd));
            if (!hasSpd) {
                await loadRecords(spd);
            }
        } finally {
            initialLoad = false;
        }
    });

    /** @param {File[]} files */
    async function processFiles(files) {
        if (uploadedFiles.length + files.length > 6) {
            toast.error('Maksimal 6 file yang diperbolehkan.');
            return;
        }
        const allowedTypes = ['image/jpeg', 'image/png', 'image/jpg'];
        for (let file of files) {
            if (file.size > 10 * 1024 * 1024) {
                toast.error(`File ${file.name} terlalu besar. Maksimal 10MB.`);
                continue;
            }

            if (!file.type.startsWith('image/')) {
                toast.error(`File ${file.name} tidak didukung. Hanya file gambar yang diperbolehkan untuk dokumentasi.`);
                continue;
            }

            file = await compressImage(file);
            const path = await uploadAndGetPath(file);
            if (path) {
                const bUrl = getBlobUrl({ path: path });
                const thumbUrl = await generateThumbnailUrl({ path: path, type: file.type });
                uploadedFiles = [...uploadedFiles, {
                    name: file.name,
                    type: file.type,
                    path: path,
                    blobUrl: bUrl,
                    thumbnailUrl: thumbUrl,
                    timestamp: new Date(file.lastModified).toISOString()
                }];
            }
        }
    }
    /** @param {Event} event */
    function handleFileChange(event) {
        // @ts-ignore
        const selectedFiles = Array.from(event.target.files);
        processFiles(selectedFiles);
        // @ts-ignore
        event.target.value = ''; 
    }

    /** @param {Event} event */
    async function handleSppdFileChange(event) {
        // @ts-ignore
        let file = event.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error('File SPPD terlalu besar. Maksimal 10MB.');
            // @ts-ignore
            event.target.value = '';
            return;
        }
        file = await compressImage(file);
        const path = await uploadAndGetPath(file);
        if (path) {
            sppdFile = { 
                name: file.name, 
                type: file.type, 
                path: path, 
                timestamp: new Date(file.lastModified).toISOString()
            };
        }
        // @ts-ignore
        event.target.value = '';
    }

    /** @param {Event} event */
    async function handleSuratTugasFileChange(event) {
        // @ts-ignore
        let file = event.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error('File Surat Tugas terlalu besar. Maksimal 10MB.');
            // @ts-ignore
            event.target.value = '';
            return;
        }
        file = await compressImage(file);
        const path = await uploadAndGetPath(file);
        if (path) {
            suratTugasFile = { 
                name: file.name, 
                type: file.type, 
                path: path, 
                timestamp: new Date(file.lastModified).toISOString()
            };
        }
        // @ts-ignore
        event.target.value = '';
    }

    function handleDragOver(e) {
        e.preventDefault();
        isDragging = true;
    }

    /** @param {DragEvent} e */
    function handleDragLeave(e) {
        e.preventDefault();
        isDragging = false;
    }

    /** @param {DragEvent} e */
    function handleDrop(e) {
        e.preventDefault();
        isDragging = false;
        if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
            const droppedFiles = Array.from(e.dataTransfer.files).filter(f => 
                f.type.startsWith('image/') || f.type === 'application/pdf' || f.type === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
            );
            if (droppedFiles.length !== e.dataTransfer.files.length) {
                toast.warning('Beberapa file diabaikan. Hanya gambar, PDF, dan Word yang diperbolehkan.');
            }
            processFiles(droppedFiles);
        }
    }

    /** @param {number} index */
    function removeFile(index) {
        uploadedFiles = uploadedFiles.filter((_, i) => i !== index);
    }

    async function saveDraftLaporan() {
        if (!recordsList.length) return;
        
        const words = reportText.trim().split(/\s+/).length;
        if (reportText && words > 200) {
            toast.error(`Laporan terlalu panjang (${words} kata). Maksimal 200 kata.`);
            return;
        }

        // Pastikan snapshot selalu memuat input yang mungkin belum diklik "Tambahkan"
        let finalTanggalMerahList = [...tanggalMerahList];
        if (newTanggalMerah && !finalTanggalMerahList.includes(newTanggalMerah)) {
            finalTanggalMerahList = [...finalTanggalMerahList, newTanggalMerah].sort();
            // Update state UI juga agar sinkron
            tanggalMerahList = finalTanggalMerahList;
            newTanggalMerah = '';
        }
        
        // Snapshot sebelum await — guard bisa reset list saat loadRecords()
        const tmSnapshot = [...finalTanggalMerahList];
        startLoading();
        try {
            await updateRecord(record.id, {
                reportStatus: 'Draft',
                suratTugasNumber: manualSuratTugasNumber,
                suratTugasDate: manualSuratTugasDate ? new Date(manualSuratTugasDate).toISOString() : undefined,
                reportData: {
                    ...(record.reportData || {}),
                    text: reportText,
                    files: uploadedFiles,
                    sppdFile: sppdFile,
                    suratTugasFile: suratTugasFile,
                    tanggalMerah: tmSnapshot,
                    lastDraftSavedAt: new Date().toISOString()
                }
            });
            await loadRecords(spd);
            tanggalMerahList = tmSnapshot;
            toast.success('Draf Laporan Kegiatan berhasil disimpan!');
        } catch (error) {
            toast.error('Gagal menyimpan draf laporan. Silakan coba lagi.');
            console.error('Save draft report error:', error);
        } finally {
            stopLoading();
        }
    }

    async function submitLaporan() {
        if (!reportText) {
             toast.error('Isi laporan tidak boleh kosong.');
             return;
        }

        const words = reportText.trim().split(/\s+/).length;
        if (words > 200) {
            toast.error(`Laporan terlalu panjang (${words} kata). Maksimal 200 kata.`);
            return;
        }

        if (uploadedFiles.length === 0) {
            toast.error('Harap unggah minimal 1 dokumentasi.');
            return;
        }

        if (!recordsList.length) return;

        // Pastikan snapshot selalu memuat input yang mungkin belum diklik "Tambahkan"
        let finalTanggalMerahList = [...tanggalMerahList];
        if (newTanggalMerah && !finalTanggalMerahList.includes(newTanggalMerah)) {
            finalTanggalMerahList = [...finalTanggalMerahList, newTanggalMerah].sort();
            // Update state UI juga agar sinkron
            tanggalMerahList = finalTanggalMerahList;
            newTanggalMerah = '';
        }

        const tmSnapshot = [...finalTanggalMerahList];
        startLoading();
        try {
            const updates = recordsList.map(r => ({
                id: r.id,
                data: {
                    reportStatus: 'Completed',
                    suratTugasNumber: manualSuratTugasNumber,
                    suratTugasDate: manualSuratTugasDate ? new Date(manualSuratTugasDate).toISOString() : undefined,
                    reportData: {
                        ...(r.reportData || {}),
                        text: reportText,
                        files: uploadedFiles,
                        sppdFile: sppdFile,
                        suratTugasFile: suratTugasFile,
                        tanggalMerah: tmSnapshot,
                        submittedAt: new Date().toISOString()
                    }
                }
            }));
            await updateMultipleRecords(updates);
            await loadRecords(spd);
            tanggalMerahList = tmSnapshot;
            clearStores();
            toast.success('Laporan Kegiatan berhasil disubmit!');
            await invalidateAll();
            goto('/dashboard/laporan');
        } catch (error) {
            toast.error('Gagal mensubmit laporan. Silakan coba lagi.');
            console.error('Submit report error:', error);
        } finally {
            stopLoading();
        }
    }

    async function saveDraftRincian() {
        if (!recordsList.length) return;

        startLoading();
        try {
            const updates = recordsList.map(r => {
                const local = localCosts[r.id];
                const costsToSave = { ...local.costs };
                costsToSave.lastDraftSavedAt = new Date().toISOString();

                return {
                    id: r.id,
                    data: {
                        costs: costsToSave,
                        totalCost: local.totalCost
                    }
                };
            });
            await updateMultipleRecords(updates);

            toast.success('Draf Rincian Biaya berhasil disimpan!');
        } catch (error) {
            toast.error('Gagal menyimpan draf rincian biaya. Silakan coba lagi.');
            console.error('Save draft costs error:', error);
        } finally {
            stopLoading();
        }
    }

    async function submitRincian() {
        if (!recordsList.length) return;

        startLoading();
        try {
            const updatePromises = recordsList.map(r => {
                const local = localCosts[r.id];
                const costsToSave = { ...local.costs };
                
                costsToSave.dailyAllowanceDays = getTotalDays(r);
                costsToSave.dailyAllowanceRate = getSbmRate(r);

                return updateRecord(r.id, {
                    costs: costsToSave,
                    totalCost: local.totalCost,
                    status: 'Submitted'
                });
            });
            await Promise.all(updatePromises);
            await loadRecords(spd);
            clearStores();
            toast.success('Rincian Biaya berhasil disubmit!');
            goto('/dashboard/laporan');
        } catch (error) {
            toast.error('Gagal menyimpan rincian biaya. Silakan coba lagi.');
            console.error('Submit costs error:', error);
        } finally {
            stopLoading();
        }
    }
</script>

<div class="max-w-5xl mx-auto pb-20 space-y-8">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
            <h2 class="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">Form {currentTab === 'laporan' ? 'Laporan Kegiatan' : 'Rincian Biaya'}</h2>
            <p class="text-sm sm:text-base text-slate-500">
                {currentTab === 'laporan' ? 'Isi detail laporan dan unggah dokumentasi kegiatan perjalanan dinas.' : 'Lengkapi rincian komponen biaya untuk petugas terkait secara komprehensif.'}
            </p>
        </div>
        <Button variant="outline" class="w-full sm:w-auto border-slate-200 text-slate-600 hover:text-slate-900" on:click={() => goto('/dashboard/laporan')}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Kembali
        </Button>
    </div>

    <!-- Tab Navigation & Actions -->
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div class="flex items-center gap-2 p-1.5 bg-slate-100/80 rounded-xl border border-slate-200 w-full sm:w-max">
            <a href="?tab=laporan" class="flex-1 sm:flex-none">
                <button class={cn("w-full px-6 py-2.5 rounded-lg text-sm font-semibold transition-all flex items-center justify-center gap-2", currentTab === 'laporan' ? "bg-white text-blue-600 shadow-sm border border-slate-200" : "text-slate-500 hover:text-slate-700 hover:bg-slate-200/50")}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                    Laporan Kegiatan
                </button>
            </a>
            <a href="?tab=rincian" class="flex-1 sm:flex-none">
                <button class={cn("w-full px-6 py-2.5 rounded-lg text-sm font-semibold transition-all flex items-center justify-center gap-2", currentTab === 'rincian' ? "bg-white text-indigo-600 shadow-sm border border-slate-200" : "text-slate-500 hover:text-slate-700 hover:bg-slate-200/50")}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    Rincian Biaya
                </button>
            </a>
        </div>

        <!-- Print Button -->
        {#if record}
            <a href={`/print?type=laporan&id=${record.id}&spd=${encodeURIComponent(record.spd)}`} target="_blank" class="w-full sm:w-auto">
                <button class="w-full sm:w-auto px-4 py-2.5 bg-white border border-slate-200 text-slate-700 font-semibold rounded-xl hover:bg-slate-50 hover:text-blue-600 transition-colors shadow-sm flex items-center justify-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                    </svg>
                    Cetak Dokumen
                </button>
            </a>
        {/if}
    </div>

    {#if !record}
        <!-- Prevent error flash on first load by showing skeleton if records are empty and likely fetching -->
        {#if $isFetchingRecords || initialLoad}
            <div class="bg-white rounded-xl shadow-sm border border-slate-200 p-6 md:p-8 space-y-8 animate-pulse">
                <div class="flex flex-col gap-4">
                    <div class="h-8 bg-slate-200 rounded w-1/3"></div>
                    <div class="h-4 bg-slate-100 rounded w-1/4"></div>
                </div>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
                    <div class="space-y-4">
                        <div class="h-10 bg-slate-100 rounded"></div>
                        <div class="h-32 bg-slate-50 rounded border border-slate-100"></div>
                    </div>
                    <div class="space-y-4">
                        <div class="h-10 bg-slate-100 rounded"></div>
                        <div class="h-32 bg-slate-50 rounded border border-slate-100"></div>
                    </div>
                </div>
            </div>
        {:else}
            <div class="flex flex-col items-center justify-center p-12 text-center border-2 border-dashed border-red-200 rounded-xl bg-red-50">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-red-400 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <h3 class="text-lg font-medium text-red-800">Data Tidak Ditemukan</h3>
                <p class="text-red-600/80 max-w-sm mt-1">Surat Perjalanan Dinas dengan nomor <b>{decodeURIComponent(spd)}</b> tidak ditemukan atau Anda tidak memiliki akses.</p>
            </div>
        {/if}
    {:else}
        <!-- Header Info -->
        {#if currentTab === 'laporan'}
            <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden mb-8">
                <!-- ... existing header code ... -->
                <div class="p-4 sm:p-6 border-b border-slate-100">
                    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
                         <!-- Left: SPD Details -->
                         <div class="space-y-6">
                             <div class="flex items-start gap-4">
                                 <div class="p-3 bg-blue-50 text-blue-600 rounded-xl shrink-0 border border-blue-100">
                                     <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                         <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                     </svg>
                                 </div>
                                 <div>
                                     <h3 class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-1">ID SPJ</h3>
                                     <div class="text-xl font-mono font-bold tracking-widest text-slate-800">{record.spd}</div>
                                 </div>
                             </div>

                             <!-- Grid for Surat Tugas & Date -->
                             <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2">
                                 <div>
                                     <div class="flex justify-between items-center mb-1.5">
                                         <span class="block text-xs font-bold text-slate-400 uppercase tracking-widest">Nomor Surat Tugas</span>
                                     </div>
                                     {#if $userStore.role !== 'kasubag'}
                                         <input 
                                             type="text" 
                                             bind:value={manualSuratTugasNumber}
                                             placeholder="Input No. Surat Tugas..."
                                             class="font-semibold text-slate-800 bg-white px-3 py-1.5 rounded-lg border border-slate-200 focus:ring-2 focus:ring-blue-100 focus:border-blue-400 outline-none w-full text-sm transition-all shadow-sm"
                                         />
                                     {:else}
                                         <div class="font-medium text-slate-800 bg-slate-50 px-3 py-1.5 rounded-lg border border-slate-100 inline-block text-sm">
                                             {manualSuratTugasNumber || record.suratTugasNumber || '-'}
                                         </div>
                                     {/if}
                                 </div>
                                 <div>
                                     <div class="flex justify-between items-center mb-1.5">
                                         <span class="block text-xs font-bold text-slate-400 uppercase tracking-widest">Tanggal Surat Tugas</span>
                                     </div>
                                     {#if $userStore.role !== 'kasubag'}
                                         <input 
                                             type="date" 
                                             bind:value={manualSuratTugasDate}
                                             class="font-semibold text-slate-800 bg-white px-3 py-1.5 rounded-lg border border-slate-200 focus:ring-2 focus:ring-blue-100 focus:border-blue-400 outline-none w-full text-sm transition-all shadow-sm"
                                         />
                                     {:else}
                                         <div class="font-medium text-slate-800 bg-slate-50 px-3 py-1.5 rounded-lg border border-slate-100 inline-block text-sm">
                                             {manualSuratTugasDate ? new Date(manualSuratTugasDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' }) : '-'}
                                         </div>
                                     {/if}
                                 </div>
                             </div>
                         </div>

                         <!-- Right: Location & Date -->
                         <div class="bg-slate-50/50 rounded-2xl p-6 border border-slate-200/60 flex flex-col justify-center space-y-6 relative overflow-hidden">
                             <div class="absolute top-0 right-0 w-32 h-32 bg-blue-50 rounded-full blur-3xl -mr-16 -mt-16 opacity-50"></div>
                             
                             <div class="flex items-start gap-4 relative z-10">
                                 <div class="p-2 bg-white rounded-lg shadow-sm text-red-500 border border-slate-100 shrink-0">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                    </svg>
                                 </div>
                                 <div>
                                     <span class="block text-xs font-bold text-slate-400 uppercase tracking-widest mb-1">Lokasi Tujuan</span>
                                     {#each getLocations(record) as loc}
                                         <div class="font-bold text-lg text-slate-800 leading-snug mb-1 last:mb-0">
                                             {loc.location ? toTitleCase(loc.location) + ', ' : ''}{toTitleCase(loc.province)}
                                         </div>
                                     {/each}
                                 </div>
                             </div>
                             
                             <div class="flex items-start gap-4 pt-4 border-t border-slate-200 relative z-10">
                                 <div class="p-2 bg-white rounded-lg shadow-sm text-indigo-500 border border-slate-100 shrink-0">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                    </svg>
                                 </div>
                                 <div>
                                     <span class="block text-xs font-bold text-slate-400 uppercase tracking-widest mb-1">Periode Perjalanan</span>
                                     <div class="font-medium text-slate-700 text-sm">
                                         <span class="font-semibold text-slate-900">{new Date(record.startDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })}</span> 
                                         <span class="text-slate-400 mx-1.5">s/d</span> 
                                         <span class="font-semibold text-slate-900">{new Date(record.endDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })}</span>
                                     </div>
                                 </div>
                             </div>
                         </div>
                    </div>

                    <!-- Divider -->
                    <div class="border-t border-slate-100 mb-8"></div>

                    <!-- Bottom Section: Employee List -->
                    <div class="mb-8">
                        <h4 class="text-sm font-bold text-slate-700 uppercase tracking-wide mb-5 flex items-center gap-2">
                            <div class="w-1 h-5 bg-blue-500 rounded-full"></div>
                            Daftar Pegawai ({allRecordsForSpd.length})
                        </h4>
                        
                        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                            {#each allRecordsForSpd as emp, i}
                                <div class="flex items-center gap-3 p-3.5 rounded-xl border border-slate-200 hover:border-blue-300 hover:bg-blue-50/30 transition-all group bg-white shadow-sm hover:shadow-md">
                                     <!-- Initial Avatar -->
                                     <div class="w-10 h-10 rounded-full bg-slate-100 text-slate-600 font-bold flex items-center justify-center border border-slate-200 group-hover:bg-white group-hover:text-blue-600 group-hover:border-blue-200 transition-colors shadow-sm shrink-0">
                                         {getInitials(emp.employee?.name || '-')}
                                     </div>
                                     <div class="flex-1 min-w-0">
                                         <div class="text-sm font-bold text-slate-800 truncate leading-tight mb-1">{emp.employee?.name || '-'}</div>
                                         <div class="text-xs text-slate-500 flex items-center gap-1.5">
                                             <span class="bg-slate-50 px-2 py-0.5 rounded text-[10px] font-mono font-medium text-slate-500 border border-slate-200 group-hover:border-blue-200 group-hover:text-blue-600 transition-colors">SPD {String(emp.sequenceNumber || 0).padStart(3, '0')}</span>
                                         </div>
                                     </div>
                                </div>
                            {/each}
                        </div>
                    </div>

                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mt-8">
                        <!-- Upload SPPD & Surat Tugas -->
                        <!-- ... (same as before) ... -->
                        <div class="relative">
                        {#if sppdFile}
                            <div class="flex items-center justify-between p-3 border border-slate-200 rounded-xl bg-slate-50">
                                <div class="flex items-center gap-3 w-full">
                                    <div class="w-8 h-8 rounded bg-blue-100 flex items-center justify-center shrink-0">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <p class="text-sm font-medium text-slate-700 truncate">{sppdFile.name}</p>
                                    </div>
                                    <div class="flex gap-2">
                                        <button type="button" class="p-1.5 text-blue-600 hover:bg-blue-100 rounded-md transition-colors" on:click={() => openPreview(sppdFile)}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg></button>
                                        <button type="button" class="p-1.5 text-red-500 hover:bg-red-100 rounded-md transition-colors" on:click={() => sppdFile = null}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>
                                    </div>
                                </div>
                            </div>
                        {:else}
                            <label class="flex items-center justify-center p-3.5 border border-slate-200 rounded-xl cursor-pointer hover:bg-slate-50 transition-colors w-full group">
                                <span class="inline-flex items-center gap-2 text-slate-600 group-hover:text-blue-600 transition-colors text-sm font-semibold">
                                    <span class="flex items-center justify-center w-6 h-6 bg-blue-50 text-blue-600 rounded-md group-hover:bg-blue-600 group-hover:text-white transition-colors shadow-sm">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                    </span>
                                    Klik untuk unggah SPPD (PDF/Gambar)
                                </span>
                                <input type="file" accept="image/*,application/pdf" class="hidden" on:change={handleSppdFileChange} />
                            </label>
                        {/if}
                    </div>

                    <!-- Upload Surat Tugas -->
                    <div class="relative">
                        {#if suratTugasFile}
                            <!-- ... -->
                             <div class="flex items-center justify-between p-3 border border-slate-200 rounded-xl bg-slate-50">
                                <div class="flex items-center gap-3 w-full">
                                    <div class="w-8 h-8 rounded bg-emerald-100 flex items-center justify-center shrink-0">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <p class="text-sm font-medium text-slate-700 truncate" title={getDisplayName(suratTugasFile.name)}>{getDisplayName(suratTugasFile.name)}</p>
                                    </div>
                                    <div class="flex gap-2">
                                        <button type="button" class="p-1.5 text-blue-600 hover:bg-blue-100 rounded-md transition-colors" on:click={() => openPreview(suratTugasFile)}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg></button>
                                        <button type="button" class="p-1.5 text-red-500 hover:bg-red-100 rounded-md transition-colors" on:click={() => suratTugasFile = null}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>
                                    </div>
                                </div>
                            </div>
                        {:else}
                            <label class="flex items-center justify-center p-3.5 border border-slate-200 rounded-xl cursor-pointer hover:bg-slate-50 transition-colors w-full group">
                                <span class="inline-flex items-center gap-2 text-slate-600 group-hover:text-blue-600 transition-colors text-sm font-semibold">
                                    <span class="flex items-center justify-center w-6 h-6 bg-blue-50 text-blue-600 rounded-md group-hover:bg-blue-600 group-hover:text-white transition-colors shadow-sm">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                    </span>
                                    Klik untuk unggah Surat Tugas
                                </span>
                                <input type="file" accept="image/*,application/pdf" class="hidden" on:change={handleSuratTugasFileChange} />
                            </label>
                        {/if}
                    </div>
                </div>
                        <div class="p-4 sm:p-6 md:p-8 space-y-8">
                            <!-- Tanggal Merah Section -->
                            <div class="space-y-4">
                                <div class="flex justify-between items-center border-b border-slate-100 pb-3">
                                    <Label class="text-lg font-bold text-slate-800">Hari Libur / Tanggal Merah</Label>
                                    <span class="text-xs text-slate-400 font-medium px-2 py-1 bg-slate-50 rounded-full border border-slate-200">
                                        {tanggalMerahList.length} Tanggal
                                    </span>
                                </div>
                                <div class="flex items-center gap-3">
                                    <input 
                                        type="date" 
                                        bind:value={newTanggalMerah}
                                        disabled={$userStore.role === 'kasubag'}
                                        class="font-semibold text-slate-800 bg-white px-3 py-2 rounded-lg border border-slate-200 focus:ring-2 focus:ring-blue-100 focus:border-blue-400 outline-none flex-1 text-sm shadow-sm"
                                    />
                                    <Button type="button" disabled={$userStore.role === 'kasubag'} on:click={addTanggalMerah} class="bg-blue-600 hover:bg-blue-700 text-white shadow-sm px-6 h-[38px]">
                                        Tambahkan
                                    </Button>
                                </div>
                                {#if tanggalMerahList.length > 0}
                                    <div class="flex flex-wrap gap-2 mt-3">
                                        {#each tanggalMerahList as tm}
                                            <div class="flex items-center gap-1.5 bg-rose-50 text-rose-700 border border-rose-200 px-3 py-1.5 rounded-lg text-sm font-medium shadow-sm transition-all hover:bg-rose-100">
                                                <span>{new Date(tm).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })}</span>
                                                {#if $userStore.role !== 'kasubag'}
                                                    <button type="button" on:click={() => removeTanggalMerah(tm)} class="text-rose-400 hover:text-rose-700 transition-colors ml-1 p-0.5 rounded-full hover:bg-rose-200" title="Hapus">
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
                                                            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                                        </svg>
                                                    </button>
                                                {/if}
                                            </div>
                                        {/each}
                                    </div>
                                {/if}
                                <div class="mt-3">
                                    <p class="text-xs text-slate-400 italic">Tanggal merah akan dikecualikan dalam perhitungan tanggal laporan dan rincian pada Cetak Dokumen.</p>
                                </div>
                            </div>

                            <!-- Report Text Section -->
                            <div class="space-y-4">
                                <div class="flex justify-between items-center border-b border-slate-100 pb-3">
                                    <Label class="text-lg font-bold text-slate-800">Isi Laporan Kegiatan</Label>
                                    <span class="text-xs text-slate-400 font-medium px-2 py-1 bg-slate-50 rounded-full border border-slate-200">
                                        {reportText.trim().split(/\s+/).filter(w => w.length > 0).length} / 200 Kata
                                    </span>
                                </div>
                                <Textarea 
                                    rows="10" 
                                    class="resize-y min-h-[150px] text-base leading-relaxed p-4 border-slate-200 focus:border-blue-300 focus:ring-blue-100 placeholder:text-slate-300 shadow-sm"
                                    placeholder="Deskripsikan hasil kegiatan, kendala yang dihadapi, dan tindak lanjut yang diperlukan..." 
                                     bind:value={reportText} 
                                    disabled={$userStore.role === 'kasubag'}
                                />
                                <p class="text-xs text-slate-400 italic">Maksimal 200 kata. Gunakan bahasa yang baku dan jelas.</p>
                            </div>

                            <!-- File Upload Section -->
                            <div class="space-y-4 pt-4 border-t border-slate-100">
                                <div class="flex justify-between items-center border-b border-slate-100 pb-3">
                                    <Label class="text-lg font-bold text-slate-800">Dokumentasi Kegiatan</Label>
                                    
                                        <span class="text-xs text-slate-400">Max 6 File (JPG/PDF), Max 10MB</span>
                                </div>
                                
                                {#if $userStore.role !== 'kasubag'}
                                <label 
                                    class="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed rounded-xl cursor-pointer transition-all group relative
                                    {isDragging ? 'border-blue-500 bg-blue-50' : 'border-slate-300 bg-slate-50 hover:bg-slate-100 hover:border-blue-400'}"
                                    on:dragover={handleDragOver}
                                    on:dragleave={handleDragLeave}
                                    on:drop={handleDrop}
                                >
                                    <div class="flex flex-col items-center justify-center pt-5 pb-6 pointer-events-none">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 mb-3 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                        </svg>
                                        <p class="mb-1 text-sm text-slate-500"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file ke sini</p>
                                        <p class="text-xs text-slate-400">PNG, JPG (Maks. 10MB)</p>
                                    </div>
                                    {#if isDragging}
                                        <div class="absolute inset-0 flex items-center justify-center bg-blue-50/90 rounded-xl pointer-events-none">
                                            <p class="text-blue-600 font-bold text-lg animate-pulse">Lepaskan file di sini</p>
                                        </div>
                                    {/if}
                                    <input type="file" multiple class="hidden" on:change={handleFileChange} />
                                </label>
                                {:else}
                                <div class="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed rounded-xl border-slate-200 bg-slate-50/50">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-7 h-7 mb-2 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                                    </svg>
                                    <p class="text-xs text-slate-400 font-medium">Unggah dokumentasi hanya dapat dilakukan oleh Petugas Protokol</p>
                                </div>
                                {/if}
                                
                                {#if uploadedFiles.length > 0}
                                    <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4 mt-6">
                                        {#each uploadedFiles as file, i}
                                            <div class="group relative aspect-square bg-slate-100 rounded-lg overflow-hidden border border-slate-200 shadow-sm cursor-pointer" on:click={() => openPreview(file)}>
                                                {#if file.type.startsWith('image/')}
                                                    <img decoding="async" src={file.thumbnailUrl || file.blobUrl} alt="Preview" class="object-cover w-full h-full" />
                                                {:else if file.type === 'application/pdf'}
                                                    <div class="flex flex-col items-center justify-center h-full text-red-500 bg-red-50 p-4 text-center group-hover:bg-red-100 transition-colors">
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                                        </svg>
                                                        <span class="text-[10px] font-bold uppercase tracking-wider text-red-700">PDF Document</span>
                                                        <span class="text-[9px] truncate w-full mt-1 text-red-600/70">{file.name}</span>
                                                    </div>
                                                {:else if file.type === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'}
                                                    <div class="flex flex-col items-center justify-center h-full text-blue-600 bg-blue-50 p-4 text-center group-hover:bg-blue-100 transition-colors">
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                                        </svg>
                                                        <span class="text-[10px] font-bold uppercase tracking-wider text-blue-700">Word Document</span>
                                                        <span class="text-[9px] truncate w-full mt-1 text-blue-600/70">{file.name}</span>
                                                    </div>
                                                {:else}
                                                    <div class="flex flex-col items-center justify-center h-full text-slate-400 p-4 text-center">
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                                        </svg>
                                                        <span class="text-[10px] truncate w-full">{file.name}</span>
                                                    </div>
                                                {/if}
                                                
                                                <div class="absolute top-2 right-2 flex flex-col gap-2 z-10">
                                                    <button 
                                                        class="bg-white/95 text-slate-700 p-2 rounded-full hover:bg-white hover:text-blue-600 shadow-sm transition-all"
                                                        on:click|stopPropagation={() => openPreview(file)}
                                                        title="Lihat"
                                                    >
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                                        </svg>
                                                    </button>
                                                    
                                                    {#if $userStore.role !== 'kasubag'}
                                                    <button 
                                                        class="bg-white/95 text-slate-700 p-2 rounded-full hover:bg-red-500 hover:text-white shadow-sm transition-all"
                                                        on:click|stopPropagation={() => removeFile(i)}
                                                        title="Hapus file"
                                                    >
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                                                    </button>
                                                    {/if}
                                                </div>

                                                <div class="absolute bottom-2 right-2 bg-black/75 text-white text-[9px] px-1.5 py-0.5 rounded pointer-events-none z-0">
                                                    {new Date(file.timestamp).toLocaleDateString()}
                                                </div>
                                            </div>
                                        {/each}
                                    </div>
                                {/if}
                            </div>
                        </div>

                        {#if $userStore.role !== 'kasubag'}
                            <div class="bg-slate-50 border-t border-slate-100 p-4 sm:p-6 -mx-4 sm:-mx-6 md:-mx-8 -mb-4 sm:-mb-6 md:-mb-8 flex flex-col sm:flex-row justify-end gap-3 rounded-b-xl mt-8">
                                <Button variant="outline" size="lg" disabled={$loadingStore} class="border-blue-200 text-blue-700 hover:bg-blue-50 px-6 w-full sm:w-auto" on:click={saveDraftLaporan}>Simpan Draft</Button>
                                <Button size="lg" disabled={$loadingStore} class="bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20 px-8 w-full sm:w-auto" on:click={submitLaporan}>Submit Laporan Kegiatan</Button>
                            </div>
                        {/if}
                    </div>
                </div>
                {/if}

                {#if currentTab === 'rincian'}
                    <!-- Form Input Rincian Biaya (Integrated from CostModal) -->
                    <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden mt-8">
                        <div class="p-4 sm:p-6 md:p-8 space-y-4">
                            {#each recordsList as empRecord, index}
                                {@const empId = empRecord.id}
                                {#if localCosts[empId]}
                                    <div id="form-rincian-{empId}" class="bg-white rounded-xl shadow-sm overflow-hidden scroll-mt-24 border border-slate-200">
                                        <button 
                                            class="w-full flex items-center justify-between p-4 bg-indigo-600 hover:bg-indigo-700 transition-colors text-white font-bold text-left rounded-t-xl {localCosts[empId]._expanded ? 'rounded-b-none' : 'rounded-b-xl'}"
                                            on:click={() => {
                                                localCosts[empId]._expanded = !localCosts[empId]._expanded;
                                                localCosts = { ...localCosts };
                                            }}
                                        >
                                            <span class="text-sm">{empRecord.employee?.name || 'Petugas'}</span>
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 transform transition-transform duration-200 {localCosts[empId]._expanded ? 'rotate-180' : ''}" viewBox="0 0 20 20" fill="currentColor">
                                                <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
                                            </svg>
                                        </button>

                                        {#if localCosts[empId]._expanded}
                                            <div class="p-4 md:p-6 bg-slate-50/30 grid gap-4 md:gap-6 border-t border-slate-200">
                                            
                                            <!-- Location Tabs -->
                                            <div class="col-span-1 md:col-span-2 -mb-2">
                                                <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider mb-2 block">Pilih Lokasi</Label>
                                                <div class="flex gap-2 overflow-x-auto pb-2">
                                                    {#each getLocations(empRecord) as loc, idx}
                                                        <button 
                                                            class={cn(
                                                                "px-4 py-2 rounded-lg text-sm font-semibold whitespace-nowrap transition-colors border",
                                                                localCosts[empId].selectedLocationIndex === idx 
                                                                    ? "bg-indigo-600 text-white border-indigo-600 shadow-sm" 
                                                                    : "bg-white text-slate-600 border-slate-200 hover:bg-slate-50 hover:border-slate-300"
                                                            )}
                                                            on:click={() => { localCosts[empId].selectedLocationIndex = idx; localCosts = {...localCosts}; }}
                                                        >
                                                            {loc.province ? toTitleCase(loc.province) : toTitleCase(empRecord.province)}
                                                        </button>
                                                    {/each}
                                                </div>
                                            </div>

                                            {#each getLocations(empRecord) as loc, idx}
                                                {#if localCosts[empId].selectedLocationIndex === idx && localCosts[empId].costs.details && localCosts[empId].costs.details[idx]}
                                                {@const detail = localCosts[empId].costs.details[idx]}
                                                {@const sbmDays = getDaysForLocation(loc)}
                                                {@const sbmRate = getRateForLocation(loc)}
                                                {@const sbmTotal = sbmDays * sbmRate}

                                                <!-- Mode Transportasi -->
                                                <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm w-full space-y-1.5">
                                                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Mode Transportasi</Label>
                                                    <Select disabled={$userStore.role === 'kasubag'} bind:value={detail.transportMode} class="bg-slate-50 border-slate-200 h-9 md:h-10 text-sm">
                                                        <option value="Pesawat/Kendaraan Umum">Pesawat/Kendaraan Umum</option>
                                                        <option value="Mobil">Mobil</option>
                                                    </Select>
                                                </div>

                                                <!-- Uang Harian SBM (Specific to Location) -->
                                                <div class="p-3 md:p-4 bg-blue-50/50 rounded-xl border border-blue-100 space-y-2.5 md:space-y-3 w-full">
                                                    <div class="flex flex-wrap justify-between items-center gap-2 border-b border-blue-200 pb-2 mb-2">
                                                        <h4 class="text-xs md:text-sm font-semibold text-blue-800 flex items-center gap-1.5">
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                                                            </svg>
                                                            Uang Harian (SBM)
                                                        </h4>
                                                        <span class="text-base md:text-lg font-bold text-blue-700">{formatCurrency(sbmTotal)}</span>
                                                    </div>
                                                    
                                                    <div class="bg-white p-3 rounded-lg border border-blue-100 shadow-sm space-y-2">
                                                        <div class="flex flex-wrap md:flex-nowrap items-center gap-2 md:gap-3 text-xs text-slate-600">
                                                            <div class="flex-none pr-2 border-r border-slate-100">
                                                                <span class="block text-[9px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Durasi</span>
                                                                <span class="font-medium whitespace-nowrap">{sbmDays} Hari</span>
                                                            </div>
                                                            <div class="flex-1 min-w-0 overflow-hidden">
                                                                <span class="block text-[9px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Rate Estimasi</span>
                                                                <span class="font-medium truncate block w-full">{formatCurrency(sbmRate)} <span class="text-[9px] text-slate-400 font-normal">/ hari</span></span>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>

                                                <!-- Tiket & Boarding Pass -->
                                                <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-4 w-full">
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
                                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Tiket Berangkat</Label>
                                                            <div class="relative">
                                                                <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                                <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(detail.ticketGo)} on:input={(e) => updateCost(empId, 'ticketGo', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white" />
                                                            </div>
                                                            {#if !detail.ticketGoFile && $userStore.role !== 'kasubag'}
                                                                <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-1">
                                                                    <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                        </svg>
                                                                        <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file</p>
                                                                    </div>
                                                                    <input type="file" class="hidden" on:change={(e) => handleSpecificFileSelect(empId, e, 'ticketGoFile', idx)} />
                                                                </label>
                                                            {/if}
                                                            {#if detail.ticketGoFile}
                                                                <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md">
                                                                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{detail.ticketGoFile.name}</span>
                                                                    <div class="flex gap-2 shrink-0 text-[10px]">
                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(detail.ticketGoFile)}>Lihat</button>
                                                                        {#if $userStore.role !== 'kasubag'}
                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'ticketGoFile', idx)}>Hapus</button>
                                                                        {/if}
                                                                    </div>
                                                                </div>
                                                            {/if}
                                                        </div>

                                                        <!-- Tiket Pulang -->
                                                        <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Tiket Pulang</Label>
                                                            <div class="relative">
                                                                <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                                <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(detail.ticketBack)} on:input={(e) => updateCost(empId, 'ticketBack', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white" />
                                                            </div>
                                                            {#if !detail.ticketBackFile && $userStore.role !== 'kasubag'}
                                                                <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-1">
                                                                    <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                        </svg>
                                                                        <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file</p>
                                                                    </div>
                                                                    <input type="file" class="hidden" on:change={(e) => handleSpecificFileSelect(empId, e, 'ticketBackFile', idx)} />
                                                                </label>
                                                            {/if}
                                                            {#if detail.ticketBackFile}
                                                                <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md">
                                                                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{detail.ticketBackFile.name}</span>
                                                                    <div class="flex gap-2 shrink-0 text-[10px]">
                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(detail.ticketBackFile)}>Lihat</button>
                                                                        {#if $userStore.role !== 'kasubag'}
                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'ticketBackFile', idx)}>Hapus</button>
                                                                        {/if}
                                                                    </div>
                                                                </div>
                                                            {/if}
                                                        </div>

                                                        <!-- Boarding Pass -->
                                                        <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg md:col-span-2 flex flex-col justify-center">
                                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Boarding Pass</Label>
                                                            
                                                            {#if detail.boardingPassFiles && detail.boardingPassFiles.length > 0}
                                                                <div class="grid grid-cols-1 md:grid-cols-2 gap-2 mt-2">
                                                                    {#each detail.boardingPassFiles as bpFile, bpIdx}
                                                                        <div class="flex items-center justify-between p-2 bg-white border border-slate-200 rounded-md w-full">
                                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{bpFile.name}</span>
                                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(bpFile)}>Lihat</button>
                                                                                {#if $userStore.role !== 'kasubag'}
                                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeBoardingPassFile(empId, bpIdx, idx)}>Hapus</button>
                                                                                {/if}
                                                                            </div>
                                                                        </div>
                                                                    {/each}
                                                                </div>
                                                            {/if}
            

                                                            {#if $userStore.role !== 'kasubag'}
                                                            <label class="flex flex-col items-center justify-center w-full h-12 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-2">
                                                                <div class="flex items-center justify-center pointer-events-none gap-2">
                                                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                                                                    </svg>
                                                                    <p class="text-[10px] md:text-xs font-semibold text-blue-600">Tambah Boarding Pass</p>
                                                                </div>
                                                                <input type="file" class="hidden" on:change={(e) => handleBoardingPassFileSelect(empId, e, idx)} />
                                                            </label>
                                                            {/if}

                                                            <!-- Extend Tiket Items -->
                                                            {#if detail.additionalCosts && detail.additionalCosts.length > 0}
                                                                {#each detail.additionalCosts as cost, costIdx}
                                                                    {#if cost.name === 'Extend Tiket'}
                                                                        <div class="col-span-1 md:col-span-2 mt-4 relative pt-4 border-t border-slate-200">
                                                                            {#if $userStore.role !== 'kasubag'}
                                                                                <button type="button" class="absolute top-2 right-0 bg-red-100 text-red-600 rounded-full p-1 border border-red-200 hover:bg-red-200 transition-colors shadow-sm z-10" on:click={() => removeAdditionalCost(empId, costIdx, idx)}>
                                                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
                                                                                        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                                                                    </svg>
                                                                                </button>
                                                                            {/if}
                                                                            <Label class="text-[10px] md:text-xs font-bold uppercase text-blue-600 tracking-wider mb-3 block">Extend Tiket {costIdx + 1}</Label>
                                                                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                                                <!-- Extend Tiket Berangkat -->
                                                                                <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                                                                                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Tiket Berangkat</Label>
                                                                                    <div class="relative">
                                                                                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                                                        <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(cost.ticketGo)} on:input={(e) => updateAdditionalExtendCost(empId, costIdx, 'ticketGo', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white" />
                                                                                    </div>
                                                                                    {#if !cost.ticketGoFile && $userStore.role !== 'kasubag'}
                                                                                        <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-1">
                                                                                            <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                                                </svg>
                                                                                                <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file</p>
                                                                                            </div>
                                                                                            <input type="file" class="hidden" on:change={(e) => handleAdditionalExtendSpecificFileSelect(empId, e, 'ticketGoFile', costIdx, idx)} />
                                                                                        </label>
                                                                                    {/if}
                                                                                    {#if cost.ticketGoFile}
                                                                                        <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md">
                                                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{cost.ticketGoFile.name}</span>
                                                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(cost.ticketGoFile)}>Lihat</button>
                                                                                                {#if $userStore.role !== 'kasubag'}
                                                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeAdditionalExtendSpecificFile(empId, 'ticketGoFile', costIdx, idx)}>Hapus</button>
                                                                                                {/if}
                                                                                            </div>
                                                                                        </div>
                                                                                    {/if}
                                                                                </div>

                                                                                <!-- Extend Tiket Pulang -->
                                                                                <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                                                                                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Tiket Pulang</Label>
                                                                                    <div class="relative">
                                                                                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                                                        <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(cost.ticketBack)} on:input={(e) => updateAdditionalExtendCost(empId, costIdx, 'ticketBack', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white" />
                                                                                    </div>
                                                                                    {#if !cost.ticketBackFile && $userStore.role !== 'kasubag'}
                                                                                        <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-1">
                                                                                            <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                                                </svg>
                                                                                                <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file</p>
                                                                                            </div>
                                                                                            <input type="file" class="hidden" on:change={(e) => handleAdditionalExtendSpecificFileSelect(empId, e, 'ticketBackFile', costIdx, idx)} />
                                                                                        </label>
                                                                                    {/if}
                                                                                    {#if cost.ticketBackFile}
                                                                                        <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md">
                                                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{cost.ticketBackFile.name}</span>
                                                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(cost.ticketBackFile)}>Lihat</button>
                                                                                                {#if $userStore.role !== 'kasubag'}
                                                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeAdditionalExtendSpecificFile(empId, 'ticketBackFile', costIdx, idx)}>Hapus</button>
                                                                                                {/if}
                                                                                            </div>
                                                                                        </div>
                                                                                    {/if}
                                                                                </div>

                                                                                <!-- Extend Boarding Pass -->
                                                                                <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg md:col-span-2 flex flex-col justify-center">
                                                                                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Boarding Pass</Label>
                                                                                    
                                                                                    {#if cost.boardingPassFiles && cost.boardingPassFiles.length > 0}
                                                                                        <div class="grid grid-cols-1 md:grid-cols-2 gap-2 mt-2">
                                                                                            {#each cost.boardingPassFiles as bpFile, bpIdx}
                                                                                                <div class="flex items-center justify-between p-2 bg-white border border-slate-200 rounded-md w-full">
                                                                                                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{bpFile.name}</span>
                                                                                                    <div class="flex gap-2 shrink-0 text-[10px]">
                                                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(bpFile)}>Lihat</button>
                                                                                                        {#if $userStore.role !== 'kasubag'}
                                                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeAdditionalExtendBoardingPass(empId, bpIdx, costIdx, idx)}>Hapus</button>
                                                                                                        {/if}
                                                                                                    </div>
                                                                                                </div>
                                                                                            {/each}
                                                                                        </div>
                                                                                    {/if}

                                                                                    {#if $userStore.role !== 'kasubag'}
                                                                                    <label class="flex flex-col items-center justify-center w-full h-12 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-2">
                                                                                        <div class="flex items-center justify-center pointer-events-none gap-2">
                                                                                            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                                                                                            </svg>
                                                                                            <p class="text-[10px] md:text-xs font-semibold text-blue-600">Tambah Boarding Pass</p>
                                                                                        </div>
                                                                                        <input type="file" class="hidden" on:change={(e) => handleAdditionalExtendBoardingPassSelect(empId, e, costIdx, idx)} />
                                                                                    </label>
                                                                                    {/if}
                                                                                </div>
                                                                            </div>
                                                                        </div>
                                                                    {/if}
                                                                {/each}
                                                            {/if}

                                                            {#if $userStore.role !== 'kasubag'}
                                                                <button type="button" class="mt-3 w-full flex items-center justify-center gap-1.5 px-3 py-2 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-xs font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm" on:click={() => addExtendTicket(empId, idx)}>
                                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                                                                    Extend Tiket
                                                                </button>
                                                            {/if}
                                                        </div>
                                                    </div>
                                                </div>

                                                <!-- Hotel -->
                                                <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-3 w-full">
                                                    <div class="flex justify-between items-center mb-1">
                                                        <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5 md:gap-2">
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                                                            </svg>
                                                            Penginapan (Hotel)
                                                        </h4>
                                                        {#if $userStore.role !== 'kasubag'}
                                                            <button 
                                                                type="button" 
                                                                class="text-[10px] md:text-xs font-bold uppercase tracking-wider text-indigo-600 bg-indigo-50 hover:bg-indigo-100 hover:text-indigo-700 px-2 py-1 rounded-md border border-indigo-200 transition-colors flex items-center gap-1"
                                                                on:click={() => openSplitHotelModal(empId, idx)}
                                                            >
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" /></svg>
                                                                Bagi Biaya
                                                            </button>
                                                        {/if}
                                                    </div>

                                                    <div class="grid grid-cols-3 gap-3 md:gap-4 w-full">
                                                        <div class="space-y-1.5 col-span-1">
                                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Malam</Label>
                                                            <Input type="number" disabled={$userStore.role === 'kasubag'} bind:value={detail.hotelDays} class="h-9 md:h-10 text-sm bg-slate-50 border-slate-200 px-2" />
                                                        </div>
                                                        <div class="space-y-1.5 col-span-2">
                                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Rate per Malam</Label>
                                                            <div class="relative">
                                                                <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                                <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(detail.hotelRate)} on:input={(e) => updateCost(empId, 'hotelRate', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-slate-50 border-slate-200" />
                                                            </div>
                                                        </div>
                                                    </div>

                                                    {#if !detail.hotelFile && $userStore.role !== 'kasubag'}
                                                        <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-slate-50 hover:bg-slate-100 hover:border-blue-400 transition-all group">
                                                            <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                </svg>
                                                                <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah kwitansi hotel</span> atau seret file</p>
                                                            </div>
                                                            <input type="file" class="hidden" on:change={(e) => handleSpecificFileSelect(empId, e, 'hotelFile', idx)} />
                                                        </label>
                                                    {/if}
                                                    
                                                    {#if detail.hotelFile}
                                                        <div class="flex items-center justify-between p-2 mt-2 bg-slate-50 border border-slate-200 rounded-md w-full">
                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{detail.hotelFile.name}</span>
                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(detail.hotelFile)}>Lihat</button>
                                                                {#if $userStore.role !== 'kasubag'}
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'hotelFile', idx)}>Hapus</button>
                                                                {/if}
                                                            </div>
                                                        </div>
                                                    {/if}

                                                    <!-- Extend Penginapan Items -->
                                                    {#if detail.additionalCosts && detail.additionalCosts.length > 0}
                                                        {#each detail.additionalCosts as cost, costIdx}
                                                            {#if cost.name === 'Extend Penginapan'}
                                                                <div class="col-span-1 md:col-span-2 mt-4 relative pt-4 border-t border-slate-200 w-full">
                                                                    <div class="flex justify-between items-center mb-3 min-h-[32px] pr-8 relative">
                                                                        {#if $userStore.role !== 'kasubag'}
                                                                            <button type="button" class="absolute -top-1 -right-1 bg-red-100 text-red-600 rounded-full p-1 border border-red-200 hover:bg-red-200 transition-colors shadow-sm z-10" on:click={() => removeAdditionalCost(empId, costIdx, idx)}>
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
                                                                            {#if $userStore.role !== 'kasubag'}
                                                                                <button 
                                                                                    type="button" 
                                                                                    class="text-[10px] md:text-xs font-bold uppercase tracking-wider text-indigo-600 bg-indigo-50 hover:bg-indigo-100 hover:text-indigo-700 px-2 py-1 rounded-md border border-indigo-200 transition-colors flex items-center gap-1"
                                                                                    on:click={() => openSplitHotelModal(empId, idx, true, costIdx)}
                                                                                >
                                                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" /></svg>
                                                                                    Bagi Biaya
                                                                                </button>
                                                                            {/if}
                                                                            {#if !cost.file && $userStore.role !== 'kasubag'}
                                                                                <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm">
                                                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                                                                    Kwitansi
                                                                                    <input type="file" class="hidden" on:change={(e) => handleAdditionalFileSelect(empId, e, costIdx, idx)} />
                                                                                </label>
                                                                            {/if}
                                                                        </div>
                                                                    </div>
                                                                    
                                                                    <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                                                                        <div class="grid grid-cols-3 gap-3 md:gap-4 w-full">
                                                                            <div class="space-y-1.5 col-span-1">
                                                                                <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Malam</Label>
                                                                                <Input type="number" disabled={$userStore.role === 'kasubag'} bind:value={detail.additionalCosts[costIdx].hotelDays} on:input={(e) => updateExtendPenginapanCost(empId, costIdx, 'hotelDays', e, idx)} class="h-9 md:h-10 text-sm bg-white border-slate-200 px-2" />
                                                                            </div>
                                                                            <div class="space-y-1.5 col-span-2">
                                                                                <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Rate per Malam</Label>
                                                                                <div class="relative">
                                                                                    <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                                                    <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(cost.hotelRate)} on:input={(e) => updateExtendPenginapanCost(empId, costIdx, 'hotelRate', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white" />
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
                                                                                        {#if $userStore.role !== 'kasubag'}
                                                                                            <button type="button" class="inline-flex items-center justify-center px-2 py-1 text-[10px] font-bold uppercase tracking-wider bg-red-50 hover:bg-red-100 text-red-600 rounded border border-red-200 transition-colors" on:click={() => removeAdditionalFile(empId, costIdx, idx)}>Hapus</button>
                                                                                        {/if}
                                                                                    </div>
                                                                                </div>
                                                                            </div>
                                                                        {/if}
                                                                    </div>
                                                                </div>
                                                            {/if}
                                                        {/each}
                                                    {/if}

                                                    {#if $userStore.role !== 'kasubag'}
                                                        <button type="button" class="mt-3 w-full flex items-center justify-center gap-1.5 px-3 py-2 rounded-md bg-blue-50 text-blue-700 border border-blue-200 text-xs font-bold uppercase tracking-wider hover:bg-blue-100 transition-all shadow-sm" on:click={() => addExtendPenginapan(empId, idx)}>
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                                                            Extend Penginapan
                                                        </button>
                                                    {/if}
                                                </div>

                                                <!-- Bukti Transportasi atau Rental -->
                                                <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-3 w-full">
                                                    <div class="flex justify-between items-center mb-1">
                                                        <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                                                            </svg>
                                                            Transport Daerah
                                                        </h4>
                                                    </div>

                                                    <div class="space-y-1.5 w-full">
                                                        <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Total Biaya (Opsional)</Label>
                                                        <div class="relative">
                                                            <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                            <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(detail.transportAmount)} on:input={(e) => updateCost(empId, 'transportAmount', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-slate-50 border-slate-200" />
                                                        </div>
                                                    </div>

                                                    {#if !detail.transportFile && $userStore.role !== 'kasubag'}
                                                        <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-slate-50 hover:bg-slate-100 hover:border-blue-400 transition-all group">
                                                            <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                </svg>
                                                                <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah bukti transport</span> atau seret file</p>
                                                            </div>
                                                            <input type="file" class="hidden" on:change={(e) => handleSpecificFileSelect(empId, e, 'transportFile', idx)} />
                                                        </label>
                                                    {/if}

                                                    {#if detail.transportFile}
                                                        <div class="flex items-center justify-between p-2 mt-2 bg-slate-50 border border-slate-200 rounded-md w-full">
                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{detail.transportFile.name}</span>
                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(detail.transportFile)}>Lihat</button>
                                                                {#if $userStore.role !== 'kasubag'}
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'transportFile', idx)}>Hapus</button>
                                                                {/if}
                                                            </div>
                                                        </div>
                                                    {/if}
                                                </div>

                                                <!-- Transport Lokal -->
                                                <div class="p-3 md:p-4 bg-slate-50 rounded-xl border border-slate-200 shadow-sm space-y-4 w-full">
                                                    <div class="flex justify-between items-center border-b border-slate-200 pb-2">
                                                        <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                                                            </svg>
                                                            Transport Lokal
                                                        </h4>
                                                        {#if $userStore.role !== 'kasubag'}
                                                        <Button size="sm" class="h-8 px-4 text-xs font-bold tracking-wider bg-indigo-600 hover:bg-indigo-700 text-white shadow-md shadow-indigo-500/20 transition-all rounded-lg flex items-center gap-1.5 hover:scale-[1.02]" on:click={() => addAdditionalCost(empId, idx)}>
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                                                            ADD COST</Button>
                                                        {/if}
                                                    </div>

                                                    <div class="space-y-3">
                                                        {#if detail.additionalCosts && detail.additionalCosts.filter(c => c.name !== 'Extend Tiket' && c.name !== 'Extend Penginapan').length > 0}
                                                            {#each detail.additionalCosts as cost, costIdx}
                                                                {#if cost.name !== 'Extend Tiket' && cost.name !== 'Extend Penginapan'}
                                                                    <div class="bg-white p-3 border border-slate-200 rounded-lg relative group">
                                                                        {#if $userStore.role !== 'kasubag'}
                                                                        <button type="button" class="absolute -top-2 -right-2 bg-red-100 text-red-600 rounded-full p-1 border border-red-200 hover:bg-red-200 transition-colors shadow-sm" on:click={() => removeAdditionalCost(empId, costIdx, idx)}>
                                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
                                                                                <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                                                            </svg>
                                                                        </button>
                                                                        {/if}
                                                                        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                                                                            <div class="space-y-1.5">
                                                                                <Label class="text-[10px] font-semibold uppercase text-slate-500 tracking-wider">Nama Biaya</Label>
                                                                                <Input type="text" disabled={$userStore.role === 'kasubag'} placeholder="Cth: Taksi Bandara" bind:value={detail.additionalCosts[costIdx].name} class="h-8 text-xs bg-slate-50 border-slate-200" />
                                                                            </div>
                                                                            <div class="space-y-1.5">
                                                                                <Label class="text-[10px] font-semibold uppercase text-slate-500 tracking-wider">Nominal</Label>
                                                                                <div class="relative">
                                                                                    <span class="absolute left-2 top-1.5 text-slate-400 text-xs">Rp</span>
                                                                                    <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(cost.amount)} on:input={(e) => updateAdditionalCostAmount(empId, costIdx, e, idx)} class="pl-7 h-8 text-xs bg-slate-50 border-slate-200" />
                                                                                </div>
                                                                            </div>
                                                                            <div class="md:col-span-2 mt-1 pt-3 border-t border-slate-100">
                                                                                <span class="block text-[10px] font-semibold uppercase text-slate-500 tracking-wider mb-2">Kwitansi / Bukti (PDF/Gambar)</span>
                                                                                {#if !cost.file}
                                                                                    {#if $userStore.role !== 'kasubag'}
                                                                                    <label class="flex flex-col items-center justify-center w-full h-14 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-slate-50 hover:bg-slate-100 hover:border-blue-400 transition-all group">
                                                                                        <div class="flex flex-col items-center justify-center pt-1 pb-1 pointer-events-none">
                                                                                            <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                                            </svg>
                                                                                            <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik unggah</span> atau seret file</p>
                                                                                        </div>
                                                                                        <input type="file" class="hidden" on:change={(e) => handleAdditionalFileSelect(empId, e, costIdx, idx)} />
                                                                                    </label>
                                                                                    {/if}
                                                                                {:else}
                                                                                    <div class="flex items-center gap-2 bg-slate-50 p-2 rounded border border-slate-200 w-full">
                                                                                        <span class="text-[10px] text-slate-700 truncate max-w-[150px] flex-1">{cost.file.name}</span>
                                                                                        <div class="flex gap-2 shrink-0 text-[10px]">
                                                                                            <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(cost.file)}>Lihat</button>
                                                                                            {#if $userStore.role !== 'kasubag'}
                                                                                            <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeAdditionalFile(empId, costIdx, idx)}>Hapus</button>
                                                                                            {/if}
                                                                                        </div>
                                                                                    </div>
                                                                                {/if}
                                                                            </div>
                                                                        </div>
                                                                    </div>
                                                                {/if}
                                                            {/each}
                                                        {:else}
                                                            <div class="text-center p-4 border border-dashed border-slate-300 rounded-lg text-xs text-slate-500">
                                                                Belum ada transport lokal diinputkan untuk lokasi ini.
                                                            </div>
                                                        {/if}
                                                    </div>
                                                </div>

                                                <!-- Total Keseluruhan Biaya -->
                                                <div class="mt-6 pt-4 border-t border-slate-200">
                                                    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-2 bg-gradient-to-br from-indigo-50 to-blue-50 p-4 rounded-xl border border-indigo-100 shadow-sm">
                                                        <div class="flex items-center gap-3">
                                                            <div class="p-2.5 bg-indigo-100 text-indigo-600 rounded-lg">
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
                                                                </svg>
                                                            </div>
                                                            <div>
                                                                <h4 class="text-sm font-bold text-slate-800">Total Biaya Perjalanan Dinas</h4>
                                                                <p class="text-[10px] text-slate-500 font-medium">Akumulasi seluruh komponen biaya semua lokasi</p>
                                                            </div>
                                                        </div>
                                                        <div class="mt-2 sm:mt-0 text-left sm:text-right w-full sm:w-auto">
                                                            <span class="block text-2xl md:text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-blue-700 to-indigo-600 tracking-tight">
                                                                {formatCurrency(localCosts[empId].totalCost)}
                                                            </span>
                                                        </div>
                                                    </div>
                                                </div>
                                                
                                                {/if}
                                            {/each}
                                        </div>
                                        {/if}
                                    </div>
                                {/if}
                            {/each}
                            
                            {#if $userStore.role !== 'kasubag'}
                                <div class="bg-slate-50 border-t border-slate-100 p-4 sm:p-6 mt-8 -mx-4 sm:-mx-6 md:-mx-8 -mb-4 sm:-mb-6 md:-mb-8 flex flex-col sm:flex-row justify-end gap-3 rounded-b-xl">
                                    <Button variant="outline" size="lg" disabled={$loadingStore} class="border-indigo-200 text-indigo-700 hover:bg-indigo-50 px-6 w-full sm:w-auto" on:click={saveDraftRincian}>Simpan Draft</Button>
                                    <Button size="lg" disabled={$loadingStore} class="bg-indigo-600 hover:bg-indigo-700 text-white shadow-lg shadow-indigo-500/20 px-8 w-full sm:w-auto" on:click={submitRincian}>Simpan Rincian Biaya</Button>
                                </div>
                            {/if}
                        </div>
                    </div>
                {/if}
    {/if}
</div>

    <!-- Split Hotel Modal -->
    <Dialog open={showSplitHotelModal} on:close={() => showSplitHotelModal = false} class="w-[90vw] max-w-lg p-0 overflow-hidden rounded-2xl shadow-2xl border-0">
        <div class="flex flex-col bg-white">
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
                <!-- Inputs for total cost -->
                <div class="grid grid-cols-3 gap-4">
                    <div class="col-span-1 space-y-1.5">
                        <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Durasi (Malam)</Label>
                        <Input type="number" bind:value={splitHotelData.days} class="bg-slate-50 border-slate-200" min="1" />
                    </div>
                    <div class="col-span-2 space-y-1.5">
                        <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Total Tagihan (Rp)</Label>
                        <div class="relative">
                            <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                            <Input 
                                type="text" 
                                value={formatInputNumber(splitHotelData.totalBill)} 
                                on:input={(e) => {
                                    const target = Object(e.target);
                                    const raw = (target.value || '').replace(/[^0-9]/g, '');
                                    splitHotelData.totalBill = parseInt(raw, 10) || 0;
                                }}
                                class="pl-9 bg-white border-slate-200 font-semibold" 
                            />
                        </div>
                    </div>
                </div>

                <!-- Select employees to share with -->
                <div class="space-y-3">
                    <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Bagikan Dengan ({splitHotelData.selectedEmpIds.length} orang dipilih)</Label>
                    <div class="bg-slate-50 border border-slate-200 rounded-xl overflow-hidden max-h-48 overflow-y-auto custom-scrollbar">
                        {#each recordsList as emp (emp.id)}
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
                                    <div class="text-xs text-slate-500 truncate">{emp.employee?.nip || '-'}</div>
                                </div>
                            </label>
                        {/each}
                    </div>
                </div>

                <!-- Preview Calculation -->
                <div class="bg-indigo-50 border border-indigo-100 rounded-xl p-4 flex justify-between items-center">
                    <div>
                        <div class="text-xs font-bold uppercase text-indigo-500 tracking-wider mb-0.5">Total per Orang</div>
                        <div class="text-sm text-slate-600 font-medium">{splitHotelData.selectedEmpIds.length} Orang</div>
                    </div>
                    <div class="text-right">
                        <div class="text-xl font-bold text-indigo-700">
                            {formatCurrency(splitHotelData.selectedEmpIds.length > 0 ? (splitHotelData.totalBill / splitHotelData.selectedEmpIds.length) : 0)}
                        </div>
                    </div>
                </div>            </div>

            <div class="px-6 py-4 bg-slate-50 border-t border-slate-100 flex justify-end gap-3">
                <Button variant="outline" class="border-slate-200 text-slate-600 hover:bg-slate-100" on:click={() => showSplitHotelModal = false}>Batal</Button>
                <Button class="bg-indigo-600 hover:bg-indigo-700 text-white shadow-md shadow-indigo-500/20" on:click={applySplitHotel}>Terapkan Pembagian</Button>
            </div>
        </div>
    </Dialog>

    <!-- Document Preview Modal -->
    <Dialog open={isPreviewOpen} hideCloseButton={true} on:close={() => isPreviewOpen = false} class="!w-[95vw] md:!w-[90vw] !max-w-6xl !h-[90vh] md:!h-[85vh] !p-0 overflow-hidden rounded-xl shadow-2xl z-[60]">
        <div class="h-full flex flex-col">
            <div class="flex items-center justify-between px-4 md:px-6 py-3 md:py-4 border-b border-slate-100 bg-slate-50/50 flex-none">
                <div class="flex flex-col min-w-0 pr-4">
                    <h3 class="text-base md:text-lg font-bold text-slate-800 tracking-tight truncate">Pratinjau Dokumen</h3>
                    {#if previewFile}
                        <p class="text-[10px] md:text-xs text-slate-500 truncate max-w-[200px] sm:max-w-xs md:max-w-md" title={getDisplayName(previewFile.name)}>{getDisplayName(previewFile.name)}</p>
                    {/if}
                </div>
                <button type="button" class="p-2 -mr-2 text-slate-400 hover:text-red-500 hover:bg-slate-100 rounded-full transition-colors flex-shrink-0" on:click={() => isPreviewOpen = false}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>
            
            <div class="flex-1 overflow-hidden bg-slate-100 relative p-0">
                {#if previewFile}
                    <DocumentViewer 
                        url={previewFile.data || (previewFile.path ? '/uploads/' + previewFile.path.replace(/^\/?uploads\//, '') : '')} 
                        type={previewFile.type.startsWith('image/') ? 'image' : (previewFile.type === 'application/pdf' ? 'pdf' : 'docx')} 
                        filename={previewFile.name} 
                    />
                {/if}
            </div>
        </div>
    </Dialog>