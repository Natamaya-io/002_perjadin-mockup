<script>
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Label from '$lib/shared/ui/label/Label.svelte';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';
    import { createEventDispatcher } from 'svelte';

    export let email = '';
    export let password = '';
    export let isLoading = false;

    const dispatch = createEventDispatcher();
    let showPassword = false;

    function handleSubmit() {
        dispatch('submit');
    }
</script>

<form class="space-y-5" on:submit|preventDefault={handleSubmit}>
    <div class="space-y-1.5 group">
        <Label class="text-xs font-bold text-slate-500 uppercase tracking-wider ml-1 group-focus-within:text-blue-600 transition-colors">Username</Label>
        <div class="relative">
            <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.207" />
                </svg>
            </div>
            <Input 
                type="text" 
                bind:value={email} 
                placeholder="nama" 
                class="pl-11 h-12 text-sm bg-slate-50 border-slate-200 focus:bg-white focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 transition-all rounded-xl shadow-sm placeholder:text-slate-300" 
            />
        </div>
    </div>
    
    <div class="space-y-1.5 group">
        <Label class="text-xs font-bold text-slate-500 uppercase tracking-wider ml-1 group-focus-within:text-blue-600 transition-colors">Password</Label>
        <div class="relative">
            <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
            </div>
            <Input 
                type={showPassword ? 'text' : 'password'} 
                bind:value={password} 
                placeholder="••••••••" 
                class="pl-11 pr-11 h-12 text-sm bg-slate-50 border-slate-200 focus:bg-white focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 transition-all rounded-xl shadow-sm placeholder:text-slate-300" 
            />
            <button 
                type="button"
                class="absolute inset-y-0 right-0 pr-3.5 flex items-center text-slate-400 hover:text-slate-600 focus:outline-none transition-colors"
                on:click={() => showPassword = !showPassword}
                tabindex="-1"
            >
                {#if showPassword}
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                {:else}
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                {/if}
            </button>
        </div>
    </div>

    <Button type="submit" class="w-full h-12 text-sm font-bold text-white shadow-lg shadow-blue-500/30 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 transition-all transform active:scale-[0.98] rounded-xl" disabled={isLoading}>
        {#if isLoading}
            <div class="flex items-center justify-center gap-2">
                <LottieLoader size="40px" className="brightness-0 invert" />
                <span>Memproses...</span>
            </div>
        {:else}
            Masuk Dashboard
        {/if}
    </Button>
</form>
