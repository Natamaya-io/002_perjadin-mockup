<script>
    import { onMount, onDestroy, untrack } from 'svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import LottieLoader from '../loader/LottieLoader.svelte';
    import pdfWorkerUrl from 'pdfjs-dist/build/pdf.worker.min.mjs?url';
    import { getBlobUrl } from '$lib/shared/utils/utils';

    // Manual polyfill for PDF.js v5 / Svelte 5 conflict
    // PDF.js v5 uses getOrInsertComputed (Stage 3 proposal) which Svelte 5's Proxy might not have
    if (typeof Map !== 'undefined' && !Map.prototype.getOrInsertComputed) {
        try {
            Object.defineProperty(Map.prototype, 'getOrInsertComputed', {
                value: function(key, callbackfn) {
                    if (this.has(key)) return this.get(key);
                    const value = callbackfn(key);
                    this.set(key, value);
                    return value;
                },
                configurable: true,
                writable: true
            });
        } catch (e) {
            console.error("Failed to polyfill Map.prototype.getOrInsertComputed", e);
        }
    }
    if (typeof WeakMap !== 'undefined' && !WeakMap.prototype.getOrInsertComputed) {
        try {
            Object.defineProperty(WeakMap.prototype, 'getOrInsertComputed', {
                value: function(key, callbackfn) {
                    if (this.has(key)) return this.get(key);
                    const value = callbackfn(key);
                    this.set(key, value);
                    return value;
                },
                configurable: true,
                writable: true
            });
        } catch (e) {
            console.error("Failed to polyfill WeakMap.prototype.getOrInsertComputed", e);
        }
    }
    
    let { url, type, filename = 'Dokumen', title = '' } = $props();
    let displayTitle = $derived(title || filename);

    // Non-reactive variables for DOM elements and PDF objects
    let container = null;
    let canvas = null;
    
    // Non-reactive references object
    const refs = {
        pdfDoc: null,
        pdfjsLib: null,
        renderAsync: null,
        currentRenderTask: null,
        pageNumPending: null
    };

    // Reactive state for UI only
    let loading = $state(true);
    let error = $state(null);
    let pageNum = $state(1);
    let pageRendering = $state(false);
    let scale = $state(1.5); // Default scale for better readability
    let totalPages = $state(0);
    
    let _rawUrl = $derived((typeof url === 'string' && url.startsWith('/uploads')) ? window.location.origin + url + '?t=' + new Date().getTime() : url);
    let finalUrl = $derived(getBlobUrl(_rawUrl));

    // Pan & Zoom CSS state
    let cssScale = $state(1.0);
    let panX = $state(0);
    let panY = $state(0);

    async function loadDocument() {
        if (!finalUrl) return;
        
        loading = true;
        error = null;
        refs.pdfDoc = null;
        totalPages = 0;
        resetCssZoom();

        try {
            if (type === 'pdf') {
                if (!refs.pdfjsLib) {
                    const mod = await import('pdfjs-dist');
                    refs.pdfjsLib = mod;
                    refs.pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorkerUrl;
                }
                
                let docParams;
                
                // Fetch the PDF robustly
                const res = await fetch(finalUrl);
                if (!res.ok) throw new Error(`Gagal mengambil file: ${res.statusText}`);
                
                const buf = await res.arrayBuffer();
                if (buf.byteLength === 0) {
                    throw new Error("Data dokumen PDF kosong (0 byte) diterima dari server.");
                }
                
                docParams = { data: new Uint8Array(buf) };

                // Use untrack to ensure getDocument and subsequent operations 
                // don't leak into Svelte's reactive context
                await untrack(async () => {
                    const loadingTask = refs.pdfjsLib.getDocument(docParams);
                    refs.pdfDoc = await loadingTask.promise;
                    totalPages = refs.pdfDoc.numPages;
                    await renderPage(pageNum);
                });
            } else if (type === 'docx') {
                if (!refs.renderAsync) {
                    const mod = await import('docx-preview');
                    refs.renderAsync = mod.renderAsync;
                }

                const response = await fetch(finalUrl);
                const blob = await response.blob();
                if (container) {
                    container.innerHTML = '';
                    await refs.renderAsync(blob, container, container, {
                        className: 'docx-viewer',
                        inWrapper: true
                    });
                }
                loading = false;
            } else if (type === 'image') {
                loading = false;
            }
        } catch (err) {
            console.error("Error loading document:", err);
            error = "Gagal memuat dokumen. " + (err.message || err.toString());
            loading = false;
        }
    }

    async function renderPage(num) {
        if (!refs.pdfDoc || !canvas) return;
        
        // Jika sedang merender, antrikan halaman ini dan keluar
        if (pageRendering) {
            refs.pageNumPending = num;
            return;
        }
        
        pageRendering = true;
        
        try {
            // Ensure rendering happens outside of any tracking context
            await untrack(async () => {
                const page = await refs.pdfDoc.getPage(num);
                const viewport = page.getViewport({ scale });
                
                const ctx = canvas.getContext('2d', { willReadFrequently: true });
                canvas.height = viewport.height;
                canvas.width = viewport.width;

                const renderContext = {
                    canvasContext: ctx,
                    viewport: viewport
                };
                
                // Batalkan task sebelumnya jika ada
                if (refs.currentRenderTask) {
                    await refs.currentRenderTask.cancel();
                }

                refs.currentRenderTask = page.render(renderContext);
                await refs.currentRenderTask.promise;
                
                refs.currentRenderTask = null;
                pageRendering = false;
                loading = false;

                // Jika ada antrian halaman baru saat kita sedang merender tadi, jalankan sekarang
                if (refs.pageNumPending !== null) {
                    const nextNum = refs.pageNumPending;
                    refs.pageNumPending = null;
                    renderPage(nextNum);
                }
            });
        } catch (err) {
            if (err.name !== 'RenderingCancelledException') {
                console.error("Page render error:", err);
            }
            pageRendering = false;
            refs.currentRenderTask = null;
        }
    }

    function queueRenderPage(num) {
        renderPage(num);
    }

    function onPrevPage() {
        if (pageNum <= 1) return;
        pageNum--;
        resetCssZoom();
        renderPage(pageNum);
    }

    function onNextPage() {
        if (pageNum >= totalPages) return;
        pageNum++;
        resetCssZoom();
        renderPage(pageNum);
    }

    function onZoomIn() {
        scale += 0.2;
        renderPage(pageNum);
    }

    function onZoomOut() {
        if (scale <= 0.4) return;
        scale -= 0.2;
        renderPage(pageNum);
    }

    function resetCssZoom() {
        cssScale = 1.0;
        panX = 0;
        panY = 0;
        if (scale !== 1.5) {
            scale = 1.5;
            if (type === 'pdf') {
                renderPage(pageNum);
            }
        }
    }

    function panzoom(node) {
        let isDraggingContent = false;
        let dragStartX = 0, dragStartY = 0;
        let initialDistance = null;
        let initialScale = 1;

        function updateTransform() {
            node.style.transform = `translate(${panX}px, ${panY}px) scale(${cssScale})`;
            node.style.transition = isDraggingContent ? 'none' : 'transform 0.1s ease-out';
        }

        function handleTouchStart(e) {
            if (e.touches.length === 2) {
                initialDistance = Math.hypot(
                    e.touches[0].clientX - e.touches[1].clientX,
                    e.touches[0].clientY - e.touches[1].clientY
                );
                initialScale = cssScale;
                isDraggingContent = false;
            } else if (e.touches.length === 1) {
                isDraggingContent = true;
                dragStartX = e.touches[0].clientX - panX;
                dragStartY = e.touches[0].clientY - panY;
                node.style.transition = 'none';
            }
        }

        function handleTouchMove(e) {
            if (e.touches.length === 2) {
                e.preventDefault();
                const currentDistance = Math.hypot(
                    e.touches[0].clientX - e.touches[1].clientX,
                    e.touches[0].clientY - e.touches[1].clientY
                );
                cssScale = Math.min(Math.max(0.5, initialScale * (currentDistance / initialDistance)), 6);
                updateTransform();
            } else if (e.touches.length === 1 && isDraggingContent) {
                if (cssScale > 1) {
                    e.preventDefault();
                    panX = e.touches[0].clientX - dragStartX;
                    panY = e.touches[0].clientY - dragStartY;
                    updateTransform();
                }
            }
        }

        function handleTouchEnd() {
            isDraggingContent = false;
            initialDistance = null;
            node.style.transition = 'transform 0.1s ease-out';
        }

        function handleMouseDown(e) {
            isDraggingContent = true;
            dragStartX = e.clientX - panX;
            dragStartY = e.clientY - panY;
            node.style.cursor = 'grabbing';
            node.style.transition = 'none';
        }

        function handleMouseMove(e) {
            if (isDraggingContent) {
                if (cssScale > 1 || type === 'image') {
                    e.preventDefault();
                }
                panX = e.clientX - dragStartX;
                panY = e.clientY - dragStartY;
                updateTransform();
            }
        }

        function handleMouseUp() {
            isDraggingContent = false;
            node.style.cursor = 'grab';
            node.style.transition = 'transform 0.1s ease-out';
        }

        function handleWheel(e) {
            if (e.ctrlKey || e.metaKey) {
                e.preventDefault();
                const delta = e.deltaY * -0.01;
                cssScale = Math.min(Math.max(0.5, cssScale + delta), 6);
                updateTransform();
            }
        }

        const containerElem = node.parentElement;
        containerElem.addEventListener('touchstart', handleTouchStart, { passive: false });
        containerElem.addEventListener('touchmove', handleTouchMove, { passive: false });
        containerElem.addEventListener('touchend', handleTouchEnd);
        containerElem.addEventListener('mousedown', handleMouseDown);
        window.addEventListener('mousemove', handleMouseMove, { passive: false });
        window.addEventListener('mouseup', handleMouseUp);
        containerElem.addEventListener('wheel', handleWheel, { passive: false });
        node.style.cursor = 'grab';
        updateTransform();

        return {
            destroy() {
                containerElem.removeEventListener('touchstart', handleTouchStart);
                containerElem.removeEventListener('touchmove', handleTouchMove);
                containerElem.removeEventListener('touchend', handleTouchEnd);
                containerElem.removeEventListener('mousedown', handleMouseDown);
                window.removeEventListener('mousemove', handleMouseMove);
                window.removeEventListener('mouseup', handleMouseUp);
                containerElem.removeEventListener('wheel', handleWheel);
            }
        };
    }

    onMount(() => {
        // Redundant call removed to prevent race conditions with the reactive block below
    });

    onDestroy(() => {
        // Prevent massive memory leaks when navigating away while PDF is open/rendering
        if (refs.currentRenderTask) {
            refs.currentRenderTask.cancel().catch(() => {});
        }
        if (refs.pdfDoc) {
            refs.pdfDoc.destroy().catch(() => {});
        }
    });

    $effect(() => {
        if (url) {
            pageNum = 1;
            loadDocument();
        }
    });

    async function downloadFile() {
        try {
            let href;
            let dlName = filename || 'dokumen';

            if (typeof finalUrl === 'string' && finalUrl.startsWith('data:')) {
                href = finalUrl;
            } else {
                const res = await fetch(finalUrl);
                const blob = await res.blob();
                href = URL.createObjectURL(blob);
            }

            const a = document.createElement('a');
            a.href = href;
            a.download = dlName;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);

            if (!finalUrl.startsWith('data:')) {
                setTimeout(() => URL.revokeObjectURL(href), 10000);
            }
        } catch (err) {
            console.error('Download failed:', err);
        }
    }
</script>

<div class="flex flex-col h-full bg-slate-200/50 rounded-lg overflow-hidden border border-slate-300 shadow-inner relative group select-none">
    
    <!-- Toolbar -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 sm:gap-4 px-3 sm:px-4 py-3 bg-slate-800 text-white shadow-md z-20">
        <div class="flex items-center justify-between w-full sm:w-auto gap-4">
            <div class="flex items-center gap-2 sm:gap-3 truncate">
                <div class="bg-white/10 p-1.5 rounded-md shrink-0">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 sm:h-4 sm:w-4 text-white/90" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clip-rule="evenodd" />
                    </svg>
                </div>
                <span class="text-xs sm:text-sm font-medium text-slate-100 truncate tracking-wide" title={displayTitle}>{displayTitle}</span>
            </div>
        </div>
        
        <div class="flex items-center justify-center gap-1.5 sm:gap-2 bg-slate-700/50 rounded-lg p-1 border border-slate-600/50 w-full sm:w-auto overflow-x-auto">
            {#if type === 'pdf'}
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors disabled:opacity-30 shrink-0" onclick={onPrevPage} disabled={pageNum <= 1}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                </button>
                <span class="text-[10px] sm:text-xs font-mono text-slate-200 min-w-[3.5rem] sm:min-w-[4rem] text-center font-semibold select-none shrink-0">{pageNum} / {totalPages}</span>
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors disabled:opacity-30 shrink-0" onclick={onNextPage} disabled={pageNum >= totalPages}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                </button>
                <div class="h-3 sm:h-4 w-px bg-slate-600 mx-1 shrink-0"></div>
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors shrink-0" onclick={onZoomOut}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5 10a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1z" clip-rule="evenodd" /></svg>
                </button>
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors shrink-0" onclick={onZoomIn}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                </button>
            {/if}
            <div class="h-3 sm:h-4 w-px bg-slate-600 mx-1 shrink-0"></div>
            <button class="flex items-center gap-1 p-1.5 sm:p-2 hover:bg-slate-600 rounded text-amber-400 hover:text-amber-300 transition-colors shrink-0" onclick={resetCssZoom}>
                <span class="text-[10px] font-bold">Reset</span>
            </button>
            <div class="h-3 sm:h-4 w-px bg-slate-600 mx-1 shrink-0"></div>
            <button
                class="flex items-center gap-1.5 px-2.5 py-1.5 hover:bg-emerald-600 bg-emerald-500/20 border border-emerald-500/40 rounded-md text-emerald-300 hover:text-white transition-all shrink-0"
                onclick={downloadFile}
                title="Unduh file ini"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                <span class="text-[10px] font-bold uppercase tracking-wider hidden sm:inline">Unduh</span>
            </button>
        </div>

        <div class="hidden sm:flex items-center gap-2">
            <!-- Tombol unduh di sini dihapus karena sudah ada di toolbar -->
        </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-hidden relative flex justify-center items-center bg-slate-200/50 touch-none">
        {#if loading}
            <div class="absolute inset-0 flex items-center justify-center bg-white/80 z-20 backdrop-blur-[2px]">
                <div class="flex flex-col items-center gap-1">
                    <LottieLoader size="60px" />
                    <span class="text-sm font-semibold text-slate-600 animate-pulse tracking-wide mt-4">Memproses Dokumen...</span>
                </div>
            </div>
        {/if}

        {#if error}
            <div class="flex flex-col items-center justify-center h-full text-slate-400 z-10">
                <p class="text-sm font-medium text-slate-600">{error}</p>
                <button class="mt-4 px-4 py-2 text-xs font-bold text-blue-600 hover:bg-blue-50 rounded-lg" onclick={loadDocument}>Coba Lagi</button>
            </div>
        {:else}
            <div use:panzoom class="relative flex justify-center items-center z-10 transform-origin-center will-change-transform w-full h-full p-4 md:p-8">
                {#if type === 'pdf'}
                    <div class="shadow-2xl shadow-slate-400/20 rounded-sm overflow-hidden bg-white flex max-h-full">
                        <canvas bind:this={canvas} class="block max-w-full max-h-[85vh] object-contain"></canvas>
                    </div>
                {:else if type === 'docx'}
                    <div bind:this={container} class="bg-white shadow-2xl p-8 min-h-[800px] w-full max-w-[800px] docx-wrapper"></div>
                {:else if type === 'image'}
                    <img decoding="async" loading="lazy" src={finalUrl} alt={filename} class="max-w-full h-auto shadow-xl rounded-lg" />
                {/if}
            </div>
        {/if}
    </div>
</div>

<style>
    :global(.docx-wrapper) { background: white !important; }
    :global(.docx-wrapper section) { box-shadow: none !important; margin-bottom: 0 !important; }
</style>
