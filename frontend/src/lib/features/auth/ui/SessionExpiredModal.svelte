<script lang="ts">
    import { fade, scale } from 'svelte/transition';
    import { sessionExpiredReason } from '../store';
    import { goto } from '$app/navigation';
    import { LogOut, AlertCircle } from 'lucide-svelte';

    export let reason: string = "Sesi Anda telah berakhir. Silakan login kembali.";

    // Translate and enhance the backend reason into informative Indonesian
    $: displayMessage = (() => {
        const r = reason.toLowerCase();
        if (r.includes('another device')) {
            return "Akun Anda baru saja digunakan untuk login di perangkat atau browser lain. Demi keamanan perlindungan data, akses Anda pada perangkat ini telah dihentikan secara otomatis.";
        }
        if (r.includes('expired') || r.includes('invalid token')) {
            return "Waktu sesi login Anda telah habis atau tidak lagi valid. Untuk alasan keamanan, silakan masuk kembali menggunakan kredensial Anda.";
        }
        return "Sesi Anda telah berakhir atau terjadi kendala autentikasi. Silakan login kembali untuk melanjutkan aktivitas.";
    })();

    function handleLogin() {
        sessionExpiredReason.set(null); // Clear reason
        goto('/login', { replaceState: true }); // Redirect to login
    }
</script>

<!-- Backdrop Blur Overlay -->
<div 
    class="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/90 p-4"
    in:fade={{ duration: 300 }}
    out:fade={{ duration: 200 }}
>
    <!-- Modal Container -->
    <div 
        class="bg-white rounded-3xl shadow-2xl overflow-hidden max-w-sm w-full border border-slate-100 flex flex-col relative"
        in:scale={{ duration: 400, start: 0.95, opacity: 0 }}
        out:scale={{ duration: 200, start: 0.95, opacity: 0 }}
    >
        <!-- Top Graphic / Accent Header -->
        <div class="h-32 bg-gradient-to-br from-rose-500 to-rose-600 relative overflow-hidden flex items-center justify-center">
            <!-- Decorative circles -->
            <div class="absolute -top-10 -right-10 w-32 h-32 rounded-full bg-white/10 blur-xl"></div>
            <div class="absolute -bottom-10 -left-10 w-24 h-24 rounded-full bg-white/10 blur-xl"></div>
            
            <!-- Icon -->
            <div class="bg-white/20 p-4 rounded-full border border-white/20 shadow-inner z-10">
                <LogOut class="w-10 h-10 text-white" strokeWidth={2.5} />
            </div>
        </div>

        <!-- Content -->
        <div class="p-8 text-center space-y-4">
            <h2 class="text-2xl font-bold text-slate-800 tracking-tight">Sesi Berakhir</h2>
            
            <div class="bg-rose-50 border border-rose-100 rounded-xl p-4 flex items-start text-left gap-3 shadow-inner">
                <AlertCircle class="w-6 h-6 text-rose-500 shrink-0 mt-0.5" />
                <p class="text-sm font-medium text-rose-800 leading-relaxed">
                    {displayMessage}
                </p>
            </div>
        </div>

        <!-- Actions -->
        <div class="p-6 pt-2 pb-8 flex flex-col gap-3">
            <button 
                on:click={handleLogin}
                class="w-full bg-rose-600 hover:bg-rose-700 active:scale-95 text-white font-semibold rounded-xl py-3.5 transition-all shadow-lg shadow-rose-600/30 flex items-center justify-center gap-2"
            >
                Login Kembali
            </button>
        </div>
    </div>
</div>
