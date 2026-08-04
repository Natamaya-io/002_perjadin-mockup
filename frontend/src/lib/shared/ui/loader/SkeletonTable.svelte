<script>
    export let rows = 8;
</script>

<div class="w-full animate-pulse">
    {#each Array(rows) as _}
        <div class="flex items-center border-b border-slate-100 bg-white">
            <!-- ID SPJ -->
            <div class="min-w-[120px] pl-4 py-4 space-y-2">
                <div class="h-4 bg-slate-200 rounded w-20"></div>
                <div class="h-3 bg-slate-100 rounded-full w-14"></div>
                <div class="h-2 bg-slate-50 rounded w-10"></div>
            </div>

            <!-- Tujuan & Lokasi -->
            <div class="min-w-[250px] flex-1 py-4 space-y-2.5 px-2">
                <div class="h-4 bg-slate-100 rounded w-5/6"></div>
                <div class="flex items-center gap-1.5">
                    <div class="h-3 w-3 bg-slate-50 rounded-full"></div>
                    <div class="h-3 bg-slate-50 rounded w-1/2"></div>
                </div>
            </div>

            <!-- Tanggal -->
            <div class="min-w-[130px] py-4 space-y-1.5 px-2">
                <div class="h-3 bg-slate-100 rounded w-16"></div>
                <div class="h-2 bg-slate-50 rounded w-4 mx-auto"></div>
                <div class="h-3 bg-slate-100 rounded w-16"></div>
            </div>

            <!-- Status -->
            <div class="min-w-[160px] py-4 flex flex-col items-center px-2">
                <div class="h-5 bg-slate-100 rounded-full w-24 border border-slate-200/50"></div>
            </div>

            <!-- Total SBM / Column 5 -->
            <div class="min-w-[180px] py-4 flex justify-center px-2">
                <div class="h-4 bg-slate-100 rounded w-24"></div>
            </div>

            <!-- Aksi -->
            <div class="w-[140px] pr-4 py-4 flex justify-center gap-2">
                <div class="h-8 w-8 bg-slate-100 rounded"></div>
                <div class="h-8 w-8 bg-slate-100 rounded"></div>
            </div>
        </div>
    {/each}
</div>

