<script>
    import { formatCurrency, toTitleCase } from '$lib/shared/utils/utils';

    export let breakdown = [];
    export let employeeCount = 0;
    export let totalCost = 0;
    export let readonly = false;

    function formatToElegantStyle(amount) {
        return formatCurrency(amount);
    }

    function formatRawNumber(amount) {
        return formatCurrency(amount).replace('Rp', '').trim();
    }
</script>

<div class="bg-white rounded-2xl shadow-lg shadow-slate-200/40 border border-slate-100 overflow-hidden mt-6 transition-all duration-300 hover:shadow-xl group">
    <!-- Compact Header -->
    <div class="px-6 py-4 bg-slate-900 text-white flex items-center justify-between">
        <div class="flex items-center gap-3">
            <div class="p-2 bg-white/10 rounded-lg backdrop-blur-sm border border-white/10">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-indigo-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                </svg>
            </div>
            <h3 class="text-sm font-bold tracking-wide uppercase">Estimasi Biaya</h3>
        </div>
        <span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest bg-white/5 px-2 py-1 rounded-md">Luar Kota</span>
    </div>
    
    <div class="p-5 space-y-4 bg-white">
        <!-- Breakdown List - Compact version -->
        <div class="space-y-2">
            {#each breakdown as item (item.id)}
                <div class="flex justify-between items-center p-3 rounded-xl hover:bg-slate-50 transition-colors border border-transparent hover:border-slate-100">
                    <div class="flex items-center gap-3">
                        <div class="w-8 h-8 rounded-lg bg-indigo-50 flex items-center justify-center text-indigo-600 shrink-0">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                            </svg>
                        </div>
                        <div class="min-w-0">
                            <div class="text-sm font-bold text-slate-800 truncate">{toTitleCase(item.province)}</div>
                            <div class="text-[10px] text-slate-500 font-medium">{item.days} hari × {formatRawNumber(item.rate)}</div>
                        </div>
                    </div>
                    <div class="text-sm font-bold text-slate-900 bg-slate-50 px-3 py-1.5 rounded-lg border border-slate-100 shrink-0">
                        {formatToElegantStyle(item.rate * item.days)}
                    </div>
                </div>
            {/each}

            <!-- Officers Info - Compact -->
            <div class="flex justify-between items-center p-3 rounded-xl bg-slate-50/50 border border-slate-100">
                <div class="flex items-center gap-3">
                    <div class="w-8 h-8 rounded-lg bg-emerald-50 flex items-center justify-center text-emerald-600 shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                        </svg>
                    </div>
                    <div class="text-sm font-semibold text-slate-700">Jumlah Petugas</div>
                </div>
                <div class="text-sm font-bold text-slate-800">{employeeCount} Orang</div>
            </div>
        </div>

        <!-- Total Section - Streamlined -->
        <div class="pt-4 border-t border-slate-100">
            <div class="bg-indigo-600 p-4 rounded-xl flex items-center justify-between text-white shadow-lg shadow-indigo-200">
                <div class="flex flex-col">
                    <span class="text-[10px] font-bold text-indigo-200 uppercase tracking-widest">Total Estimasi</span>
                    <span class="text-xl font-black tracking-tight">{formatToElegantStyle(totalCost)}</span>
                </div>
                <div class="p-2 bg-white/10 rounded-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                    </svg>
                </div>
            </div>
            
            <div class="mt-4 flex gap-2 p-3 bg-amber-50 rounded-xl border border-amber-100/50">
                <div class="shrink-0 text-amber-500">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                    </svg>
                </div>
                <p class="text-[10px] text-slate-600 leading-normal">
                    Hanya mencakup <span class="font-bold">Uang Harian (SBM)</span>. Biaya transpor & hotel dihitung saat laporan kegiatan.
                </p>
            </div>
        </div>
    </div>
</div>
