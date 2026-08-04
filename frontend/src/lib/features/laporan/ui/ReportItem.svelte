<script>
    import { cn, getInitials } from '$lib/shared/utils/utils';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import { userStore } from '$lib/features/auth/store';

    export let record;
</script>

<div class="group bg-white rounded-xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-200 overflow-hidden flex flex-col">
    <div class="p-6 flex-1 space-y-4">
        <div class="flex justify-between items-start">
            <div class="flex flex-wrap gap-1.5">
                <span class={cn("px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wide border", 
                    record.reportStatus === 'Completed' ? "bg-emerald-50 text-emerald-700 border-emerald-100" : "bg-amber-50 text-amber-700 border-amber-100")}>
                    {record.reportStatus === 'Completed' ? 'Selesai' : 'Pending'}
                </span>
                <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                    {record.type ? record.type.split('_').join(' ') : 'Dalam Kota'}
                </span>
            </div>
            <span class="inline-flex items-center font-mono text-[11px] font-bold tracking-widest text-slate-500">{record.spd}</span>
        </div>
        
        <div>
            <h3 class="font-bold text-lg text-slate-800 line-clamp-2 group-hover:text-blue-600 transition-colors">{record.purpose}</h3>
            <p class="text-sm text-slate-500 mt-1 flex items-center gap-1">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                {record.location}, {record.province}
            </p>
        </div>
        
        <div class="flex items-center gap-3 text-xs text-slate-500 pt-2 border-t border-slate-50">
            <div class="flex items-center gap-1">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
            </div>
            <span>&mdash;</span>
            <div>{new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</div>
        </div>
    </div>
    
    <div class="px-6 py-4 bg-slate-50 border-t border-slate-100 flex items-center justify-between">
         <div class="flex -space-x-2 overflow-hidden">
            <!-- Avatar placeholders -->
            <div class="h-6 w-6 rounded-full ring-2 ring-white bg-slate-200 flex items-center justify-center text-[10px] font-bold text-slate-500 leading-none">
                {getInitials(record.employee.name)}
            </div>
         </div>
         <div class="flex gap-2">
             {#if record.status === 'Approved' || record.status === 'Submitted' || record.status === 'Draft'}
                 <a href={`/print?type=spd&spd=${encodeURIComponent(record.spd)}`} target="_blank" class="contents">
                     <Button variant="outline" size="sm" class="h-8 p-2 rounded-md hover:bg-slate-100 text-slate-600 border-slate-200" title="Cetak SPD">
                         <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                     </Button>
                 </a>
                 <a href={`/print?type=rincian&spd=${encodeURIComponent(record.spd)}`} target="_blank" class="contents">
                     <Button variant="outline" size="sm" class="h-8 p-2 rounded-md hover:bg-slate-100 text-slate-600 border-slate-200" title="Cetak Rincian">
                         <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><rect width="16" height="20" x="4" y="2" rx="2" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"/><line x1="8" x2="16" y1="6" y2="6" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"/><line x1="16" x2="16" y1="14" y2="18" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"/><path d="M16 10h.01M12 10h.01M8 10h.01M12 14h.01M8 14h.01M12 18h.01M8 18h.01" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"/></svg>
                     </Button>
                 </a>
             {/if}
             {#if record.reportStatus === 'Completed'}
                <a href={`/print?type=laporan&spd=${encodeURIComponent(record.spd)}`} target="_blank" class="contents">
                    <Button variant="ghost" size="sm" class="h-8 w-8 p-0 rounded-full hover:bg-slate-200 text-slate-500" title="Cetak Laporan">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2-2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                        </svg>
                    </Button>
                </a>
             {/if}
             <a href={`/dashboard/laporan/${encodeURIComponent(record.spd)}`}>
                <Button size="sm" class={cn("h-8 text-xs font-medium shadow-sm transition-all", 
                    record.reportStatus === 'Completed' || $userStore.role === 'kasubag'
                    ? "bg-white text-slate-700 border border-slate-200 hover:bg-slate-50 hover:text-blue-600" 
                    : "bg-blue-600 hover:bg-blue-700 text-white shadow-blue-500/20")}>
                    {$userStore.role === 'kasubag' ? 'Lihat Laporan' : (record.reportStatus === 'Completed' ? 'Edit Laporan' : 'Input Laporan')}
                </Button>
             </a>
         </div>
    </div>
</div>
