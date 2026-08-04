<script lang="ts">
    import { page } from '$app/stores';
    import { recordsStore, loadRecords } from '$lib/features/pengajuan/store';
    import { onMount, onDestroy } from 'svelte';
    import { toast } from '$lib/shared/stores/toast';
    import { api } from '$lib/shared/api';
    import DocumentViewer from '$lib/shared/ui/document-viewer/DocumentViewer.svelte';
    import type { TravelRecord } from '$lib/shared/api/types';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';

    let type: string = '';
    let spd: string = '';
    let recordId: string = '';
    let record: TravelRecord | null = null;

    let pdfUrl: string = '';
    let isGeneratingPdf: boolean = false;
    let isDataLoaded: boolean = false;
    let pdfError: string = '';
    let currentRecordIdForPdf: string | null = null;

    let unsubPage: () => void;
    let unsubRecords: () => void;

    onMount(async () => {
        unsubPage = page.subscribe(($page) => {
            type = $page.url.searchParams.get('type') || ''; 
            spd = $page.url.searchParams.get('spd') || '';
            recordId = $page.url.searchParams.get('id') || '';
            updateState();
        });

        unsubRecords = recordsStore.subscribe(() => {
            updateState();
        });

        try {
            const initialSpd = $page.url.searchParams.get('spd') || '';
            const initialId = $page.url.searchParams.get('id') || '';

            // Check if we need to load records
            let needsLoad = true;
            const unsubscribeCheck = recordsStore.subscribe(recs => {
                if (recs.length > 0) {
                    if (initialId) {
                        needsLoad = !recs.some(r => r.id === initialId);
                    } else if (initialSpd) {
                        needsLoad = !recs.some(r => r.spd === initialSpd);
                    } else {
                        needsLoad = false;
                    }
                }
            });
            unsubscribeCheck();

            if (needsLoad && (initialSpd || initialId)) {
                await loadRecords(initialSpd || initialId); // loadRecords takes spd, but id is better than nothing although loadRecords uses spd. We just don't want undefined.
            }
            isDataLoaded = true;
            updateState();
        } catch (e) {
            console.error(e);
            toast.error('Gagal memuat data perjalanan.');
        }
    });

    onDestroy(() => {
        if (unsubPage) unsubPage();
        if (unsubRecords) unsubRecords();
        if (pdfUrl && pdfUrl.startsWith('blob:')) window.URL.revokeObjectURL(pdfUrl);
    });

    function updateState() {
        if (!isDataLoaded) return;

        let recs: TravelRecord[] = [];
        const unsub = recordsStore.subscribe(val => recs = val);
        unsub();

        const allRecordsForSpd = spd ? recs.filter(r => r.spd === spd) : [];
        record = recordId 
            ? recs.find(r => r.id === recordId) || (allRecordsForSpd.length > 0 ? allRecordsForSpd[0] : null)
            : (allRecordsForSpd.length > 0 ? allRecordsForSpd[0] : null);

        if (record && record.id !== currentRecordIdForPdf) {
            currentRecordIdForPdf = record.id;
            loadPdfPreview();
        }
    }

    async function loadPdfPreview() {
        if (!record) return;

        isGeneratingPdf = true;
        pdfError = '';
        try {
            let pdfBlob: Blob | undefined;
            if (type === 'spd') {
                pdfBlob = await api.exportSpdPdf(record.id);
            } else if (type === 'laporan') {
                pdfBlob = await api.exportLaporanPdf(record.id);
            } else if (type === 'rincian') {
                pdfBlob = await api.exportRincianPdf(record.id);
            }

            if (pdfBlob) {
                if (pdfUrl && pdfUrl.startsWith('blob:')) window.URL.revokeObjectURL(pdfUrl);
                pdfUrl = window.URL.createObjectURL(pdfBlob);
            }
            isGeneratingPdf = false;
        } catch (e: any) {
            console.error('Error loading PDF preview:', e);
            pdfError = e.message || 'Terjadi kesalahan saat memuat dokumen.';
            toast.error('Gagal membuat pratinjau PDF.');
            isGeneratingPdf = false;
        }
    }

    async function downloadPdf() {
        if (!pdfUrl || !record || !record.employee) return;

        const filename = type === 'spd' 
            ? `SPD_${record.employee.name}_${spd.replace(/\//g, '_')}` 
            : type === 'laporan' 
                ? `Laporan_${record.employee.name}_${spd.replace(/\//g, '_')}`
                : `Rincian_Biaya_${record.employee.name}_${spd.replace(/\//g, '_')}`;

        const a = document.createElement('a');
        a.href = pdfUrl;
        a.download = `${filename}.pdf`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);

        toast.success('PDF berhasil diunduh!');
    }
</script>

<svelte:head>
    <title>Pratinjau {type ? type.toUpperCase() : 'Dokumen'} - {spd || ''}</title>
</svelte:head>

<div class="min-h-screen bg-slate-100 flex flex-col">
    <!-- Toolbar -->
    <div class="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-6 shadow-sm z-10">
        <div class="flex items-center gap-4">
            <div>
                <h2 class="text-sm font-bold text-slate-800">Pratinjau Dokumen Resmi</h2>
                <p class="text-xs text-slate-500">{spd || 'Memuat...'}</p>
            </div>
        </div>
        
        <div class="flex items-center gap-3">
        </div>
    </div>

    <!-- Document Viewer Area -->
    <div class="flex-1 overflow-hidden relative">
        {#if isGeneratingPdf && !pdfUrl}
            <div class="absolute inset-0 flex flex-col items-center justify-center bg-slate-50 z-20">
                <LottieLoader size="64px" />
                <p class="text-slate-600 font-medium mt-6">Sedang menyiapkan pratinjau PDF...</p>
                <p class="text-slate-400 text-sm mt-2">Ini mungkin memakan waktu beberapa detik karena merender via LibreOffice.</p>
            </div>
        {/if}

        {#if pdfUrl}
            <DocumentViewer url={pdfUrl} type="pdf" title="Pratinjau {type.toUpperCase()}" />
        {:else if !isGeneratingPdf && isDataLoaded}
            <div class="flex flex-col items-center justify-center h-full text-slate-400 bg-white">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mb-4 opacity-20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p class="text-lg font-medium">Pratinjau tidak tersedia.</p>
                {#if pdfError}
                    <p class="text-sm text-red-500 mt-2 max-w-md text-center">{pdfError}</p>
                {/if}
            </div>
        {/if}
    </div>
</div>

<style>
    /* Remove Paged.js global styles that might interfere */
    :global(.pagedjs_pages) {
        display: none !important;
    }
</style>
