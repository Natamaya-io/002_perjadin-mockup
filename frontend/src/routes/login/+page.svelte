<script>
    import { login } from '$lib/features/auth/store';
    import { goto } from '$app/navigation';
    import { fly } from 'svelte/transition';
    import { loadRecords } from '$lib/features/pengajuan/store';
    import { loadMasterData } from '$lib/shared/stores/master-data';
    
    // Login Components
    import LoginBackground from '$lib/features/auth/ui/LoginBackground.svelte';
    import LoginCard from '$lib/features/auth/ui/LoginCard.svelte';
    import LoginHeader from '$lib/features/auth/ui/LoginHeader.svelte';
    import LoginBody from '$lib/features/auth/ui/LoginBody.svelte';
    import DemoBanner from '$lib/features/auth/ui/DemoBanner.svelte';
    import LoginForm from '$lib/features/auth/ui/LoginForm.svelte';
    import LoginFooter from '$lib/features/auth/ui/LoginFooter.svelte';
    import ChangePasswordModal from '$lib/features/auth/ui/ChangePasswordModal.svelte';

    let email = '';
    let password = '';
    let isLoading = false;
    let errorMessage = '';
    
    let isChangePasswordMode = false;

    // Check toggle from environment variable
    const showDemoBanner = import.meta.env.VITE_SHOW_DEMO_BANNER === 'true';

    async function handleLogin() {
        errorMessage = '';
        if (!email || !password) {
            errorMessage = 'Username dan Password tidak boleh kosong';
            return;
        }

        isLoading = true;
        
        // Use email as username
        const result = await login(email, password);
        
        isLoading = false;
        
        if (result.success) {
            goto('/dashboard');
        } else {
            errorMessage = 'Username atau password salah.';
        }
    }

    function onPasswordChanged() {
        // Password changed successfully — now load deferred data and redirect
        loadMasterData();
        goto('/dashboard');
    }
</script>

<div class="relative min-h-[100dvh] w-full flex flex-col items-center justify-center p-4 sm:p-6 overflow-x-hidden">
    <LoginBackground />

    <div class="relative z-10 w-full max-w-[400px] flex flex-col my-auto" in:fly={{ y: 20, duration: 600, delay: 100 }}>
        <LoginCard>
            <LoginHeader />
            <LoginBody>
                {#if showDemoBanner}
                    <DemoBanner on:fill={(e) => { email = e.detail.email; password = e.detail.password; }} />
                {/if}
                {#if errorMessage}
                    <div class="text-xs font-medium text-red-600 bg-red-50 p-3 rounded-lg border border-red-100 text-center animate-shake">
                        {errorMessage}
                    </div>
                {/if}
                <LoginForm bind:email bind:password {isLoading} on:submit={handleLogin} />
            </LoginBody>
            <LoginFooter />
        </LoginCard>
    </div>

    <ChangePasswordModal bind:isOpen={isChangePasswordMode} on:success={onPasswordChanged} />
</div>

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
