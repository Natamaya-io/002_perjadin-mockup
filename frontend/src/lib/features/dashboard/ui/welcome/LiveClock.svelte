<script>
    import { onMount, onDestroy } from 'svelte';

    let time = new Date();
    let interval;

    const optionsTime = { 
        timeZone: 'Asia/Jakarta', 
        hour: '2-digit', 
        minute: '2-digit', 
        second: '2-digit',
        hour12: false 
    };

    const optionsDate = { 
        timeZone: 'Asia/Jakarta', 
        weekday: 'long', 
        year: 'numeric', 
        month: 'long', 
        day: 'numeric' 
    };

    onMount(() => {
        interval = setInterval(() => {
            time = new Date();
        }, 1000);
    });

    onDestroy(() => {
        clearInterval(interval);
    });

    $: timeString = time.toLocaleTimeString('id-ID', optionsTime).replace(/\./g, ':');
    $: dateString = time.toLocaleDateString('id-ID', optionsDate);
</script>

<div class="flex flex-col items-end text-right text-white/90">
    <div class="text-3xl md:text-4xl font-bold font-mono tracking-wider tabular-nums drop-shadow-md">
        {timeString} <span class="text-base md:text-lg font-sans font-medium text-blue-200">WIB</span>
    </div>
    <div class="text-sm md:text-base font-medium text-blue-100/80 mt-1 capitalize">
        {dateString}
    </div>
</div>
