<script>
    import { onMount } from 'svelte';
    export let records = [];

    // State for current viewing month
    let currentDate = new Date();
    $: currentMonth = currentDate.getMonth();
    $: currentYear = currentDate.getFullYear();
    
    // Calendar Generation
    $: daysInMonth = new Date(currentYear, currentMonth + 1, 0).getDate();
    $: firstDayOfMonth = new Date(currentYear, currentMonth, 1).getDay(); // 0 is Sunday
    
    $: blanks = Array(firstDayOfMonth).fill(null);
    $: days = Array.from({ length: daysInMonth }, (_, i) => i + 1);
    
    // Process records into a hashmap for quick lookup O(1)
    // Key: YYYY-MM-DD
    $: scheduledDates = records.reduce((acc, record) => {
        if (!record.startDate || !record.endDate) return acc;
        
        let start = new Date(record.startDate);
        let end = new Date(record.endDate);
        
        // Normalize time to 00:00:00 to avoid timezone shift issues
        start.setHours(0,0,0,0);
        end.setHours(0,0,0,0);
        
        let loop = new Date(start);
        while (loop <= end) {
            // Format to YYYY-M-D (months are 0-indexed in JS but 1-indexed in string format, here we just use what matches our grid)
            let dateKey = `${loop.getFullYear()}-${loop.getMonth()}-${loop.getDate()}`;
            
            // Priority to 'Completed' (green) over 'Draft/Approved' (red/amber) if there are overlaps
            if (!acc[dateKey]) {
                acc[dateKey] = record;
            } else if (acc[dateKey].reportStatus !== 'Completed' && record.reportStatus === 'Completed') {
                acc[dateKey] = record;
            }
            
            let newDate = loop.setDate(loop.getDate() + 1);
            loop = new Date(newDate);
        }
        return acc;
    }, {});

    const monthNames = [
        "Januari", "Februari", "Maret", "April", "Mei", "Juni", 
        "Juli", "Agustus", "September", "Oktober", "November", "Desember"
    ];
    const dayNames = ["Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"];

    function prevMonth() {
        if (currentMonth === 0) {
            currentDate = new Date(currentYear - 1, 11, 1);
        } else {
            currentDate = new Date(currentYear, currentMonth - 1, 1);
        }
    }

    function nextMonth() {
        if (currentMonth === 11) {
            currentDate = new Date(currentYear + 1, 0, 1);
        } else {
            currentDate = new Date(currentYear, currentMonth + 1, 1);
        }
    }
    
    function resetToToday() {
        currentDate = new Date();
    }

    function isToday(day, month, year) {
        const today = new Date();
        return day === today.getDate() && month === today.getMonth() && year === today.getFullYear();
    }
</script>

<div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between p-4 border-b border-slate-100 bg-slate-50 gap-4">
        <div class="flex items-center gap-3">
            <div class="p-2 bg-blue-100 text-blue-600 rounded-lg shrink-0">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
            </div>
            <div>
                <h3 class="text-base font-bold text-slate-800">Kalender Penugasan</h3>
                <p class="text-xs text-slate-500">Jadwal Perjalanan Dinas</p>
            </div>
        </div>

        <div class="flex items-center justify-between w-full sm:w-auto gap-2">
            <button class="p-1.5 rounded-md bg-white border border-slate-200 hover:bg-slate-100 text-slate-600 transition-colors shadow-sm" on:click={prevMonth}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
            </button>
            <span class="text-sm font-bold text-slate-800 flex-1 text-center sm:w-32 hover:text-blue-600 transition-colors" on:click={resetToToday} style="cursor: pointer;" title="Kembali ke hari ini">
                {monthNames[currentMonth]} {currentYear}
            </span>
            <button class="p-1.5 rounded-md bg-white border border-slate-200 hover:bg-slate-100 text-slate-600 transition-colors shadow-sm" on:click={nextMonth}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </button>
        </div>
    </div>

    <!-- Legends -->
    <div class="flex flex-wrap items-center justify-center gap-x-4 gap-y-2 py-3 px-2 bg-white border-b border-slate-100 text-[9px] sm:text-[10px] md:text-xs">
        <div class="flex items-center gap-1.5">
            <div class="w-2.5 h-2.5 sm:w-3 sm:h-3 rounded bg-red-100 border border-red-300"></div>
            <span class="text-slate-600 whitespace-nowrap">Terjadwal</span>
        </div>
        <div class="flex items-center gap-1.5">
            <div class="w-2.5 h-2.5 sm:w-3 sm:h-3 rounded bg-emerald-100 border border-emerald-300"></div>
            <span class="text-slate-600 whitespace-nowrap">Laporan Selesai</span>
        </div>
    </div>

    <!-- Calendar Grid -->
    <div class="p-1.5 sm:p-4">
        <!-- Days Header -->
        <div class="grid grid-cols-7 gap-0.5 sm:gap-1 mb-1 sm:mb-2">
            {#each dayNames as day, i}
                <div class="text-center text-[8px] sm:text-[10px] md:text-xs font-bold uppercase tracking-tight sm:tracking-wider {i === 0 ? 'text-red-500' : 'text-slate-500'}">
                    {day}
                </div>
            {/each}
        </div>

        <!-- Days Grid -->
        <div class="grid grid-cols-7 gap-0.5 sm:gap-1.5 md:gap-2">
            {#each blanks as _}
                <div class="aspect-square bg-slate-50/50 rounded-md sm:rounded-lg border border-slate-100 border-dashed"></div>
            {/each}

            {#each days as day}
                {@const dateKey = `${currentYear}-${currentMonth}-${day}`}
                {@const hasRecord = scheduledDates[dateKey]}
                {@const isCompleted = hasRecord && hasRecord.reportStatus === 'Completed'}
                
                <div class="relative aspect-square rounded-md sm:rounded-lg flex items-center justify-center text-[10px] sm:text-xs md:text-sm transition-all duration-200
                    {hasRecord ? (isCompleted ? 'bg-emerald-50 border border-emerald-200 text-emerald-800 shadow-sm font-bold' : 'bg-red-50 border border-red-200 text-red-800 shadow-sm font-bold') : 'bg-white border border-slate-100 text-slate-700 hover:bg-slate-50 hover:border-blue-200 font-medium'}
                    {isToday(day, currentMonth, currentYear) && !hasRecord ? 'ring-2 ring-blue-400 ring-offset-1 font-bold text-blue-700 bg-blue-50' : ''}
                " title={hasRecord ? `${hasRecord.purpose} (${hasRecord.location})` : ''}>
                    
                    <span>{day}</span>
                    
                    {#if hasRecord}
                        <div class="absolute bottom-1 right-1 sm:bottom-1.5 sm:right-1.5 w-1 h-1 sm:w-1.5 sm:h-1.5 rounded-full {isCompleted ? 'bg-emerald-500' : 'bg-red-500'}"></div>
                    {/if}
                    {#if isToday(day, currentMonth, currentYear)}
                        <div class="absolute top-1 right-1 sm:top-1.5 sm:right-1.5 w-1 h-1 sm:w-1.5 sm:h-1.5 rounded-full bg-blue-500" title="Hari ini"></div>
                    {/if}
                </div>
            {/each}
        </div>
    </div>
</div>