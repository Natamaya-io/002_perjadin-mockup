<script>
    import LottieLoader from './LottieLoader.svelte';
</script>

<div class="fixed inset-0 z-[100] flex items-center justify-center bg-white/90">
    <div class="flex flex-col items-center gap-1">
        <!-- Lottie Animation -->
        <LottieLoader size="64px" />
        
        <!-- Text -->
        <div class="flex flex-col items-center animate-pulse mt-4">
            <span class="text-sm font-semibold text-slate-700 tracking-wide uppercase">Memuat</span>
            <span class="text-[10px] text-slate-500 font-medium">Mohon Tunggu...</span>
        </div>
    </div>
</div>
