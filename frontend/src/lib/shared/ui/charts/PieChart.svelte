<script>
    export let data = []; // { label: string, value: number, color: string }
    export let title = '';
    
    let total = 0;
    $: total = data.reduce((acc, d) => acc + d.value, 0);

    // Calculate angles
    let slices = [];
    $: {
        let currentAngle = 0;
        slices = data.map(d => {
            const angle = (d.value / total) * 360;
            const startAngle = currentAngle;
            currentAngle += angle;
            return { ...d, startAngle, endAngle: currentAngle };
        });
    }

    // Helper to calculate coordinates
    function getCoordinatesForPercent(percent) {
        const x = Math.cos(2 * Math.PI * percent);
        const y = Math.sin(2 * Math.PI * percent);
        return [x, y];
    }

    function getPath(slice) {
        // Start and end coordinates
        const start = getCoordinatesForPercent(slice.startAngle / 360);
        const end = getCoordinatesForPercent(slice.endAngle / 360);
        
        // Large arc flag: 1 if angle > 180 degrees
        const largeArcFlag = slice.endAngle - slice.startAngle > 180 ? 1 : 0;
        
        // Path command: Move to center, Line to start, Arc to end, Close path
        return `M 0 0 L ${start[0]} ${start[1]} A 1 1 0 ${largeArcFlag} 1 ${end[0]} ${end[1]} Z`;
    }
</script>

<div class="bg-white p-6 rounded-xl shadow-sm border border-slate-100 flex flex-col h-full">
    {#if title}
        <h3 class="text-lg font-semibold text-slate-800 mb-6">{title}</h3>
    {/if}

    <div class="flex-1 flex flex-col items-center justify-center">
        {#if total > 0}
            <div class="relative w-48 h-48 md:w-56 md:h-56">
                <!-- SVG rotated -90deg so 0 is at top -->
                <svg viewBox="-1.05 -1.05 2.1 2.1" class="w-full h-full transform -rotate-90 overflow-visible">
                    {#each slices as slice}
                        <path 
                            d={getPath(slice)} 
                            fill={slice.color} 
                            stroke="white" 
                            stroke-width="0.02" 
                            class="hover:opacity-80 transition-opacity cursor-pointer"
                        >
                            <title>{slice.label}: {slice.value} ({Math.round((slice.value / total) * 100)}%)</title>
                        </path>
                    {/each}
                    <!-- Optional: White circle in middle for Donut Chart effect -->
                    <!-- <circle cx="0" cy="0" r="0.6" fill="white" /> -->
                </svg>
            </div>

            <!-- Legend -->
            <div class="mt-8 flex flex-wrap justify-center gap-x-6 gap-y-2">
                {#each data as item}
                    <div class="flex items-center text-sm">
                        <span class="w-3 h-3 rounded-full mr-2 shadow-sm" style="background-color: {item.color}"></span>
                        <span class="text-slate-600 font-medium">{item.label}</span>
                        <span class="ml-1.5 text-slate-400 text-xs">({Math.round((item.value / total) * 100)}%)</span>
                    </div>
                {/each}
            </div>
        {:else}
            <div class="text-slate-400 text-sm italic py-10">Belum ada data untuk ditampilkan</div>
        {/if}
    </div>
</div>
