<script>
    import { createEventDispatcher, onMount } from 'svelte';
    import { fade, scale } from 'svelte/transition';
    import { portal } from '$lib/shared/actions/portal';
    import { api } from '$lib/shared/api';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Label from '$lib/shared/ui/label/Label.svelte';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';

    export let isOpen = false;

    const dispatch = createEventDispatcher();
    let newPassword = '';
    let confirmPassword = '';
    let isLoading = false;
    let errorMessage = '';
    let showNewPassword = false;
    let showConfirmPassword = false;

    // Body scroll lock
    $: if (typeof document !== 'undefined') {
        if (isOpen) {
            document.body.style.overflow = 'hidden';
        } else {
            document.body.style.overflow = '';
        }
    }

    onMount(() => {
        return () => {
            if (typeof document !== 'undefined') {
                document.body.style.overflow = '';
            }
        };
    });

    $: passwordStrength = getPasswordStrength(newPassword);
    $: passwordsMatch = newPassword && confirmPassword && newPassword === confirmPassword;
    $: canSubmit = newPassword.length >= 8 && passwordsMatch && !isLoading;

    function getPasswordStrength(pw) {
        if (!pw) return { level: 0, label: '', color: '' };
        let score = 0;
        if (pw.length >= 8) score++;
        if (pw.length >= 12) score++;
        if (/[A-Z]/.test(pw)) score++;
        if (/[0-9]/.test(pw)) score++;
        if (/[^A-Za-z0-9]/.test(pw)) score++;

        if (score <= 1) return { level: 1, label: 'Lemah', color: 'bg-red-500' };
        if (score <= 2) return { level: 2, label: 'Cukup', color: 'bg-amber-500' };
        if (score <= 3) return { level: 3, label: 'Baik', color: 'bg-blue-500' };
        return { level: 4, label: 'Kuat', color: 'bg-emerald-500' };
    }

    async function handleSubmit() {
        errorMessage = '';

        if (!newPassword || !confirmPassword) {
            errorMessage = 'Semua field wajib diisi.';
            return;
        }

        if (newPassword.length < 8) {
            errorMessage = 'Password minimal 8 karakter.';
            return;
        }

        if (newPassword === '12345678') {
            errorMessage = 'Password baru tidak boleh sama dengan password default (12345678).';
            return;
        }

        if (newPassword !== confirmPassword) {
            errorMessage = 'Konfirmasi password tidak cocok.';
            return;
        }

        isLoading = true;
        try {
            await api.changePassword(newPassword);
            dispatch('success');
            isOpen = false;
        } catch (e) {
            errorMessage = e.message || 'Gagal mengubah password. Silakan coba lagi.';
        } finally {
            isLoading = false;
        }
    }
</script>

{#if isOpen}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div 
        use:portal
        class="fixed inset-0 z-[9999] flex items-center justify-center p-4 bg-slate-900/90"
    >
        <div 
            class="bg-white rounded-3xl shadow-2xl w-full max-w-md overflow-hidden ring-1 ring-slate-900/5 relative"
        >
            <!-- Header with gradient -->
            <div class="bg-gradient-to-br from-blue-600 to-indigo-700 px-6 py-5 text-white relative overflow-hidden">
                <button 
                    type="button"
                    aria-label="Tutup"
                    class="absolute right-4 top-4 z-20 text-white/70 hover:text-white transition-colors focus:outline-none focus:ring-2 focus:ring-white/20 rounded-lg p-1"
                    on:click={() => { isOpen = false; dispatch('close'); }}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
                <div class="absolute -right-6 -top-6 h-24 w-24 rounded-full bg-white/10 blur-xl pointer-events-none"></div>
                <div class="absolute -left-4 -bottom-4 h-20 w-20 rounded-full bg-white/10 blur-xl pointer-events-none"></div>
                <div class="relative z-10 flex items-start gap-4">
                    <div class="h-12 w-12 rounded-xl bg-white/20 flex items-center justify-center backdrop-blur-sm shrink-0 border border-white/20 shadow-inner">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                        </svg>
                    </div>
                    <div>
                        <h3 class="text-lg font-bold tracking-tight">Ganti Password Wajib</h3>
                        <p class="text-sm text-blue-100 mt-0.5">Demi keamanan akun, harap ganti password default Anda sebelum melanjutkan.</p>
                    </div>
                </div>
            </div>

            <!-- Body -->
            <div class="p-6">
                {#if errorMessage}
                    <div class="mb-4 text-xs font-medium text-red-600 bg-red-50 p-3 rounded-xl border border-red-100 flex items-start gap-2 animate-shake">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 shrink-0 mt-0.5 text-red-500" viewBox="0 0 20 20" fill="currentColor">
                            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                        </svg>
                        <span>{errorMessage}</span>
                    </div>
                {/if}

                <form on:submit|preventDefault={handleSubmit} class="space-y-4">
                    <!-- New Password -->
                    <div class="space-y-1.5">
                        <Label for="newPassword" class="text-slate-700 text-sm font-semibold">Password Baru</Label>
                        <div class="relative">
                            <Input 
                                id="newPassword" 
                                type={showNewPassword ? 'text' : 'password'} 
                                bind:value={newPassword}
                                placeholder="Minimal 8 karakter" 
                                required 
                                class="pr-10 h-11 border-slate-200 focus:border-blue-400 focus:ring-blue-100"
                            />
                            <button 
                                type="button" 
                                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 transition-colors"
                                on:click={() => showNewPassword = !showNewPassword}
                            >
                                {#if showNewPassword}
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>
                                {:else}
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                                {/if}
                            </button>
                        </div>
                        
                        <!-- Password Strength Indicator -->
                        {#if newPassword}
                            <div class="flex items-center gap-2 mt-2">
                                <div class="flex-1 flex gap-1">
                                    {#each [1, 2, 3, 4] as level}
                                        <div class="h-1.5 flex-1 rounded-full transition-all duration-300 {passwordStrength.level >= level ? passwordStrength.color : 'bg-slate-200'}"></div>
                                    {/each}
                                </div>
                                <span class="text-[10px] font-semibold uppercase tracking-wider {passwordStrength.level <= 1 ? 'text-red-500' : passwordStrength.level <= 2 ? 'text-amber-500' : passwordStrength.level <= 3 ? 'text-blue-500' : 'text-emerald-500'}">
                                    {passwordStrength.label}
                                </span>
                            </div>
                        {/if}
                    </div>
                    
                    <!-- Confirm Password -->
                    <div class="space-y-1.5">
                        <Label for="confirmPassword" class="text-slate-700 text-sm font-semibold">Konfirmasi Password Baru</Label>
                        <div class="relative">
                            <Input 
                                id="confirmPassword" 
                                type={showConfirmPassword ? 'text' : 'password'} 
                                bind:value={confirmPassword}
                                placeholder="Ketik ulang password baru" 
                                required 
                                class="pr-10 h-11 border-slate-200 focus:border-blue-400 focus:ring-blue-100 {confirmPassword && !passwordsMatch ? 'border-red-300 focus:border-red-400 focus:ring-red-100' : ''} {passwordsMatch ? 'border-emerald-300 focus:border-emerald-400 focus:ring-emerald-100' : ''}"
                            />
                            <button 
                                type="button" 
                                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 transition-colors"
                                on:click={() => showConfirmPassword = !showConfirmPassword}
                            >
                                {#if showConfirmPassword}
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>
                                {:else}
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                                {/if}
                            </button>
                        </div>
                        {#if confirmPassword && !passwordsMatch}
                            <p class="text-[11px] text-red-500 font-medium flex items-center gap-1 mt-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" /></svg>
                                Password tidak cocok
                            </p>
                        {/if}
                        {#if passwordsMatch}
                            <p class="text-[11px] text-emerald-500 font-medium flex items-center gap-1 mt-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                                Password cocok
                            </p>
                        {/if}
                    </div>

                    <!-- Tips -->
                    <div class="bg-slate-50 rounded-xl p-3 border border-slate-100">
                        <p class="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1.5">Tips Password Kuat</p>
                        <ul class="text-[11px] text-slate-500 space-y-0.5">
                            <li class="flex items-center gap-1.5">
                                <span class="w-1 h-1 rounded-full {newPassword.length >= 8 ? 'bg-emerald-500' : 'bg-slate-300'}"></span>
                                Minimal 8 karakter
                            </li>
                            <li class="flex items-center gap-1.5">
                                <span class="w-1 h-1 rounded-full {/[A-Z]/.test(newPassword) ? 'bg-emerald-500' : 'bg-slate-300'}"></span>
                                Mengandung huruf besar
                            </li>
                            <li class="flex items-center gap-1.5">
                                <span class="w-1 h-1 rounded-full {/[0-9]/.test(newPassword) ? 'bg-emerald-500' : 'bg-slate-300'}"></span>
                                Mengandung angka
                            </li>
                            <li class="flex items-center gap-1.5">
                                <span class="w-1 h-1 rounded-full {/[^A-Za-z0-9]/.test(newPassword) ? 'bg-emerald-500' : 'bg-slate-300'}"></span>
                                Mengandung karakter spesial
                            </li>
                        </ul>
                    </div>

                    <div class="pt-1">
                        <Button
                            type="submit"
                            disabled={!canSubmit}
                            class="w-full text-sm font-semibold h-12 rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white shadow-lg shadow-blue-500/25 transition-all active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:shadow-none"
                        >
                            {#if isLoading}
                                <LottieLoader size="36px" className="brightness-0 invert -ml-1" />
                                Menyimpan...
                            {:else}
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                                </svg>
                                Simpan Password Baru
                            {/if}
                        </Button>
                    </div>
                </form>
            </div>
        </div>
    </div>
{/if}

<style>
    @keyframes shake {
        0%, 100% { transform: translateX(0); }
        25% { transform: translateX(-4px); }
        75% { transform: translateX(4px); }
    }
    .animate-shake {
        animation: shake 0.3s ease-in-out;
    }
</style>
