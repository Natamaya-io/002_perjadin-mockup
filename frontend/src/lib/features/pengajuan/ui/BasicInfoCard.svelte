<script>
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';

    export let email = '';
    export let spdNumberInput = '';
    /** @type {File | null} */
    export let suratTugas = null;
    export let suratTugasNumber = '';
    export let readonly = false;

    // File handling logic
    let fileInput;
    let dragOver = false;

    function handleFileDrop(e) {
        e.preventDefault();
        dragOver = false;
        if (readonly) return;
        
        if (e.dataTransfer.items) {
            const item = e.dataTransfer.items[0];
            if (item.kind === 'file') {
                const file = item.getAsFile();
                if (validateFile(file)) {
                    suratTugas = file;
                }
            }
        }
    }

    function handleDragOver(e) {
        e.preventDefault();
        if (!readonly) dragOver = true;
    }

    function handleDragLeave() {
        dragOver = false;
    }

    function handleFileSelect(e) {
        if (readonly) return;
        const file = e.target.files[0];
        if (validateFile(file)) {
            suratTugas = file;
        }
    }

    function removeFile() {
        if (readonly) return;
        suratTugas = null;
        if (fileInput) fileInput.value = '';
    }

    function validateFile(file) {
        if (!file) return false;
        
        // Allowed types: PDF, JPG, PNG
        const allowedTypes = ['application/pdf', 'image/jpeg', 'image/png'];
        if (!allowedTypes.includes(file.type)) {
            alert('Format file tidak didukung. Harap unggah PDF, JPG, atau PNG.');
            return false;
        }

        // Max size: 5MB
        const maxSize = 5 * 1024 * 1024;
        if (file.size > maxSize) {
            alert('Ukuran file terlalu besar. Maksimal 5MB.');
            return false;
        }

        return true;
    }

    function formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
</script>

<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
        <h3 class="font-semibold text-slate-800 flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Informasi Dasar
        </h3>
    </div>
    <div class="p-6 space-y-6">
        <div class="space-y-2">
            <Label class="text-slate-600">Email Pengaju</Label>
            <Input value={email} readonly disabled class="bg-slate-50 text-slate-500 border-slate-200" />
        </div>

        <div class="space-y-2">
            <Label class="text-slate-600">Nomor SPJ <span class="text-xs text-slate-400 font-normal">(Opsional - Kosongkan untuk Otomatis)</span></Label>
            <div class="flex items-center">
                <span class="inline-flex items-center px-3 border border-r-0 border-slate-300 bg-slate-100 text-slate-500 text-sm rounded-l-md h-11">
                    ID-SPJ-
                </span>
                <Input type="number" min="1" bind:value={spdNumberInput} placeholder="Cth: 005" disabled={readonly} class="flex-1 rounded-none rounded-r-md h-11 border-slate-300 focus:ring-blue-500 focus:border-blue-500 {readonly ? 'bg-slate-50 opacity-70 cursor-not-allowed' : ''}" />
            </div>
            <p class="text-[10px] text-slate-400">Hanya angka. Isi jika ingin menentukan nomor urut secara manual.</p>
        </div>
        
        <div class="space-y-2 pt-2">
            <Label class="text-slate-600">Surat Tugas <span class="text-xs text-slate-400 font-normal">(Opsional)</span></Label>
            
            <div 
                class="mt-2 flex justify-center px-6 pt-5 pb-6 border-2 border-dashed rounded-xl transition-colors duration-200 {dragOver ? 'border-indigo-500 bg-indigo-50/50' : 'border-slate-300 hover:border-slate-400 bg-slate-50/30'} {readonly ? 'opacity-70 cursor-not-allowed' : 'cursor-pointer'}"
                on:drop={handleFileDrop}
                on:dragover={handleDragOver}
                on:dragleave={handleDragLeave}
                on:click={() => !readonly && fileInput.click()}
                on:keydown={(e) => !readonly && e.key === 'Enter' && fileInput.click()}
                role="button"
                tabindex="0"
                aria-label="Upload Surat Tugas"
            >
                <div class="space-y-2 text-center w-full">
                    {#if !suratTugas}
                        <svg class="mx-auto h-10 w-10 text-slate-400" stroke="currentColor" fill="none" viewBox="0 0 48 48" aria-hidden="true">
                            <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                        </svg>
                        <div class="flex text-sm text-slate-600 justify-center flex-col sm:flex-row items-center gap-1">
                            <span class="relative cursor-pointer rounded-md font-medium text-indigo-600 hover:text-indigo-500 focus-within:outline-none focus-within:ring-2 focus-within:ring-offset-2 focus-within:ring-indigo-500">
                                <span>Unggah file</span>
                                <input id="file-upload" name="file-upload" type="file" class="sr-only" bind:this={fileInput} on:change={handleFileSelect} disabled={readonly} accept=".pdf,.jpg,.jpeg,.png">
                            </span>
                            <p>atau drag and drop</p>
                        </div>
                        <p class="text-xs text-slate-500">PDF, PNG, JPG hingga 5MB</p>
                    {:else}
                        <div class="flex items-center justify-between p-3 bg-white border border-indigo-200 rounded-lg shadow-sm w-full" on:click|stopPropagation on:keydown|stopPropagation role="presentation">
                            <div class="flex items-center space-x-3 overflow-hidden">
                                <div class="flex-shrink-0">
                                    <svg class="h-8 w-8 text-indigo-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                    </svg>
                                </div>
                                <div class="flex-1 min-w-0 text-left">
                                    <p class="text-sm font-medium text-slate-900 truncate">
                                        {suratTugas.name}
                                    </p>
                                    <p class="text-xs text-slate-500">
                                        {formatFileSize(suratTugas.size)}
                                    </p>
                                </div>
                            </div>
                            {#if !readonly}
                                <button type="button" class="ml-4 flex-shrink-0 p-1 rounded-full text-slate-400 hover:text-red-500 hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-colors" on:click={removeFile} aria-label="Hapus file">
                                    <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                    </svg>
                                </button>
                            {/if}
                        </div>
                    {/if}
                </div>
            </div>
        </div>
        
        <div class="space-y-2">
            <Label class="text-slate-600">Nomor Surat Tugas <span class="text-xs text-slate-400 font-normal">(Opsional)</span></Label>
            <Input type="text" bind:value={suratTugasNumber} placeholder="Cth: 1/B/2026/01" class="h-11 {readonly ? 'bg-slate-50 opacity-70 cursor-not-allowed' : ''}" disabled={readonly} />
        </div>
    </div>
</div>
