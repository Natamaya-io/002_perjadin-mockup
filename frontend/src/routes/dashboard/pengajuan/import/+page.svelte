<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { userStore } from '$lib/features/auth/store';
    import { api } from '$lib/shared/api';
    import { toast } from '$lib/shared/stores/toast';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';
    import type { ImportResult } from '$lib/shared/api/types';

    let fileInput: HTMLInputElement;
    let selectedFile: File | null = null;
    let isDragging = false;
    let isUploading = false;
    let importResult: ImportResult | null = null;

    onMount(() => {
        if ($userStore.role !== 'super_admin' && $userStore.role !== 'kasubag') {
            goto('/dashboard/pengajuan');
        }
    });

    function handleDragEnter(e: DragEvent) {
        e.preventDefault();
        isDragging = true;
    }

    function handleDragLeave(e: DragEvent) {
        e.preventDefault();
        isDragging = false;
    }

    function handleDrop(e: DragEvent) {
        e.preventDefault();
        isDragging = false;
        
        if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
            const file = e.dataTransfer.files[0];
            validateAndSetFile(file);
        }
    }

    function handleFileSelect(e: Event) {
        const target = e.target as HTMLInputElement;
        if (target.files && target.files.length > 0) {
            validateAndSetFile(target.files[0]);
        }
    }

    function validateAndSetFile(file: File) {
        if (!file.name.toLowerCase().endsWith('.xlsx')) {
            toast.error('Format file salah. Silakan upload file .xlsx');
            selectedFile = null;
            if (fileInput) fileInput.value = '';
            return;
        }
        selectedFile = file;
        importResult = null; // reset result if selecting new file
    }

    async function handleUpload() {
        if (!selectedFile) return;

        isUploading = true;
        try {
            importResult = await api.importExcel(selectedFile);
            toast.success(`Berhasil mengimport ${importResult.imported} data.`);
        } catch (error: any) {
            toast.error(error.message || 'Gagal melakukan import');
            console.error('Import error:', error);
        } finally {
            isUploading = false;
        }
    }

    function resetForm() {
        selectedFile = null;
        importResult = null;
        if (fileInput) fileInput.value = '';
    }
</script>

<div class="space-y-6 pb-20 max-w-5xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <div class="flex items-center gap-2 mb-2 text-sm text-slate-500">
                <a href="/dashboard/pengajuan" class="hover:text-blue-600 transition-colors">Pengajuan</a>
                <span>/</span>
                <span class="text-slate-800 font-medium">Import Excel</span>
            </div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Import Data SPJ Protokol</h1>
            <p class="text-sm text-slate-500 mt-1">Upload file Excel (.xlsx) untuk memasukkan data SPJ secara massal.</p>
        </div>
        <a href="/dashboard/pengajuan" class="inline-flex items-center text-slate-600 bg-slate-100 hover:bg-slate-200 font-medium px-4 py-2.5 rounded-xl transition-colors text-sm">
            Kembali
        </a>
    </div>

    <!-- Upload Zone -->
    {#if !importResult}
        <div class="bg-white p-8 rounded-2xl shadow-sm border border-slate-100">
            <!-- svelte-ignore a11y-no-static-element-interactions -->
            <div class="border-2 border-dashed rounded-xl p-10 text-center transition-colors {isDragging ? 'border-blue-500 bg-blue-50' : 'border-slate-300 bg-slate-50 hover:bg-slate-100'}"
                on:dragenter={handleDragEnter}
                on:dragleave={handleDragLeave}
                on:dragover={(e) => e.preventDefault()}
                on:drop={handleDrop}
            >
                <!-- Icon -->
                <div class="mx-auto w-16 h-16 mb-4 rounded-full bg-green-100 text-green-600 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                </div>
                
                <h3 class="text-lg font-semibold text-slate-700 mb-1">
                    {selectedFile ? 'File Terpilih' : 'Upload file Excel Anda'}
                </h3>
                <p class="text-sm text-slate-500 mb-6">
                    {selectedFile ? selectedFile.name : 'Tarik dan lepas file .xlsx di sini, atau klik untuk memilih.'}
                </p>

                <input type="file" accept=".xlsx" class="hidden" bind:this={fileInput} on:change={handleFileSelect} />
                
                <div class="flex items-center justify-center gap-3">
                    <button type="button" class="px-5 py-2.5 bg-white border border-slate-300 text-slate-700 rounded-lg hover:bg-slate-50 transition-colors font-medium cursor-pointer text-sm shadow-sm" on:click={() => fileInput.click()}>
                        Pilih File
                    </button>
                    
                    {#if selectedFile}
                        <button type="button" class="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium flex items-center shadow-sm shadow-blue-200 text-sm disabled:opacity-70 disabled:cursor-not-allowed" on:click={handleUpload} disabled={isUploading}>
                            {#if isUploading}
                                <LottieLoader size="32px" className="brightness-0 invert -ml-1" />
                                Mengupload...
                            {:else}
                                Import Sekarang
                            {/if}
                        </button>
                    {/if}
                </div>
            </div>
            
            <div class="mt-6 flex items-start gap-3 p-4 bg-amber-50 rounded-lg border border-amber-100">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-amber-500 flex-shrink-0 mt-0.5" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                </svg>
                <div class="text-sm text-amber-800">
                    <p class="font-medium mb-1">Catatan Penting:</p>
                    <ul class="list-disc pl-4 space-y-1">
                        <li>Sistem mendeteksi dan melewati data ganda (berdasarkan ID SPJ + Nama).</li>
                        <li>Data Pegawai baru dan THR akan otomatis disimpan ke sistem untuk konsistensi.</li>
                    </ul>
                </div>
            </div>
        </div>
    {/if}

    <!-- Result View -->
    {#if importResult}
        <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
            <!-- Result Summary Cards -->
            <div class="grid grid-cols-2 md:grid-cols-4 border-b border-slate-100">
                <div class="p-6 border-r border-slate-100">
                    <p class="text-sm font-medium text-slate-500 mb-1">Total Baris dibaca</p>
                    <p class="text-2xl font-bold text-slate-800">{importResult.totalRows}</p>
                </div>
                <div class="p-6 border-r border-slate-100">
                    <p class="text-sm font-medium text-slate-500 mb-1">Berhasil Diimport</p>
                    <p class="text-2xl font-bold text-green-600">{importResult.imported}</p>
                </div>
                <div class="p-6 border-r border-slate-100">
                    <p class="text-sm font-medium text-slate-500 mb-1">Dilewati (Skip)</p>
                    <p class="text-2xl font-bold text-amber-600">{importResult.skipped}</p>
                </div>
                <div class="p-6">
                    <p class="text-sm font-medium text-slate-500 mb-1">Gagal</p>
                    <p class="text-2xl font-bold text-red-600">{importResult.failed}</p>
                </div>
            </div>

            <!-- Header Actions -->
            <div class="p-4 border-b border-slate-100 bg-slate-50 flex justify-between items-center">
                <h3 class="font-semibold text-slate-700">Detail Status Import</h3>
                <div class="flex gap-2">
                    <a href="/dashboard/pengajuan" class="px-4 py-2 border border-slate-300 text-slate-700 hover:bg-slate-100 rounded-lg text-sm font-medium transition-colors">
                        Lihat Data Pengajuan
                    </a>
                    <button class="px-4 py-2 bg-blue-600 text-white hover:bg-blue-700 rounded-lg text-sm font-medium shadow-sm transition-colors" on:click={resetForm}>
                        Import File Lain
                    </button>
                </div>
            </div>

            <!-- Result Table -->
            <div class="overflow-x-auto max-h-[500px] table-scrollbar table-scroll-shadows">
                <table class="w-full text-left text-sm whitespace-nowrap">
                    <thead class="bg-white sticky top-0 z-10 shadow-sm">
                        <tr class="text-slate-600 text-xs uppercase tracking-wider border-b border-slate-200 block md:table-row">
                            <th class="px-4 py-3 font-semibold w-16">Baris</th>
                            <th class="px-4 py-3 font-semibold">ID SPJ</th>
                            <th class="px-4 py-3 font-semibold">Nama Pegawai</th>
                            <th class="px-4 py-3 font-semibold">Status</th>
                            <th class="px-4 py-3 font-semibold">Keterangan</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100 block md:table-row-group">
                        {#if importResult.details.length === 0}
                            <tr>
                                <td colspan="5" class="px-4 py-8 text-center text-slate-500 text-sm">Tidak ada detail tersedia</td>
                            </tr>
                        {:else}
                            {#each importResult.details as detail}
                                <tr class="hover:bg-slate-50/50 block md:table-row">
                                    <td class="px-4 py-2.5 font-mono text-xs text-slate-500">{detail.row}</td>
                                    <td class="px-4 py-2.5 font-medium text-slate-700">{detail.spjId}</td>
                                    <td class="px-4 py-2.5 text-slate-600">{detail.name}</td>
                                    <td class="px-4 py-2.5">
                                        {#if detail.status === 'imported'}
                                            <span class="inline-flex items-center rounded-full bg-green-50 px-2.5 py-0.5 text-xs font-medium text-green-700 border border-green-200">Berhasil</span>
                                        {:else if detail.status === 'skipped_duplicate'}
                                            <span class="inline-flex items-center rounded-full bg-slate-50 px-2.5 py-0.5 text-xs font-medium text-slate-600 border border-slate-200">Duplikat</span>
                                        {:else if detail.status === 'skipped_thr'}
                                            <span class="inline-flex items-center rounded-full bg-slate-50 px-2.5 py-0.5 text-xs font-medium text-slate-600 border border-slate-200">Bukan SPJ (THR)</span>
                                        {:else if detail.status === 'skipped_no_user'}
                                            <span class="inline-flex items-center rounded-full bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-700 border border-amber-200">User Tidak Ada</span>
                                        {:else if detail.status === 'failed'}
                                            <span class="inline-flex items-center rounded-full bg-red-50 px-2.5 py-0.5 text-xs font-medium text-red-700 border border-red-200">Gagal</span>
                                        {/if}
                                    </td>
                                    <td class="px-4 py-2.5 text-slate-500 text-xs max-w-xs truncate" title={detail.message}>{detail.message}</td>
                                </tr>
                            {/each}
                        {/if}
                    </tbody>
                </table>
            </div>
        </div>
    {/if}
</div>
