<script>
    export let title = "Dalam Pengembangan";
    export let description = "Halaman ini sedang dalam tahap pengembangan dan akan segera tersedia.";
</script>

<div class="flex flex-col items-center justify-center min-h-[60vh] text-center p-8 bg-slate-50 rounded-3xl border border-slate-100">
    <div class="bg-blue-50 p-6 rounded-full mb-6 animate-pulse">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
        </svg>
    </div>
    
    <h1 class="text-2xl font-bold text-slate-800 mb-2">{title}</h1>
    <p class="text-slate-500 max-w-md mb-8 leading-relaxed">
        {description}
    </p>

    <div class="flex gap-3">
        <a href="/dashboard" class="px-6 py-2.5 bg-white border border-slate-200 text-slate-600 font-medium rounded-xl hover:bg-slate-50 hover:text-slate-900 transition-all shadow-sm">
            Kembali ke Dashboard
        </a>
    </div>
</div>
