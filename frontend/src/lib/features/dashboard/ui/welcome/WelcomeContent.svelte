<script>
    import { userStore } from '$lib/features/auth/store';
    export let role = '';
    
    $: greetingName = $userStore.name || (
        role === 'super_admin' ? 'Super Admin' : 
        role === 'kasubag' ? 'Kasubag' :
        role === 'protokol' ? 'Protokol' : 'Staf'
    );
</script>

<div class="flex flex-col space-y-2 max-w-2xl">
    <!-- Greeting Section -->
    <div class="flex flex-col space-y-1">
        <div class="flex items-center space-x-3">
            <span class="text-3xl md:text-4xl animate-wave origin-bottom-right inline-block">👋</span>
            <h1 class="text-3xl md:text-4xl font-bold tracking-tight text-white drop-shadow-md">
                Halo, <span class="text-transparent bg-clip-text bg-gradient-to-r from-blue-100 to-white">{greetingName}</span>
            </h1>
        </div>
        <p class="text-blue-100/80 text-sm md:text-base font-medium tracking-wide">
            Selamat datang di Dashboard Perjalanan Dinas
        </p>
    </div>
</div>

<style>
    @keyframes wave {
        0% { transform: rotate(0.0deg) }
        10% { transform: rotate(14.0deg) }
        20% { transform: rotate(-8.0deg) }
        30% { transform: rotate(14.0deg) }
        40% { transform: rotate(-4.0deg) }
        50% { transform: rotate(10.0deg) }
        60% { transform: rotate(0.0deg) }
        100% { transform: rotate(0.0deg) }
    }
    .animate-wave {
        animation-name: wave;
        animation-duration: 2.5s;
        animation-iteration-count: infinite;
        transform-origin: 70% 70%;
    }
</style>
