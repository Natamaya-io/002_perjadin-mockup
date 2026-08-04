<script>
    import { onMount } from 'svelte';
    import { toast } from '$lib/shared/stores/toast';
    import { RealApiClient } from '$lib/shared/api/real';
    import { userStore } from '$lib/features/auth/store';
    import { loadingStore, startLoading, stopLoading } from '$lib/shared/stores/loading';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';

    import { settingsStore } from '$lib/shared/stores/settings';

    const api = new RealApiClient();
    
    // Form state seeded from SWR cache
    let settings = {
        ppk_name: $settingsStore.ppk_name || '',
        ppk_nip: $settingsStore.ppk_nip || '',
        bendahara_name: $settingsStore.bendahara_name || '',
        bendahara_nip: $settingsStore.bendahara_nip || ''
    };
    
    // Subscribe to store updates directly
    $: if ($settingsStore.isLoaded && !settings.ppk_name) {
        settings = { ...$settingsStore };
    }

    onMount(async () => {
        // Silently fetch data in the background (SWR pattern)
        try {
            const data = await api.getSettings();
            if (data) {
                const newSettings = {
                    ppk_name: data.ppk_name || '',
                    ppk_nip: data.ppk_nip || '',
                    bendahara_name: data.bendahara_name || '',
                    bendahara_nip: data.bendahara_nip || '',
                    isLoaded: true
                };
                settingsStore.set(newSettings);
                // Also update local state if it hasn't been edited
                settings = { ...newSettings };
            }
        } catch (error) {
            console.error("Failed to load settings in background:", error);
        }
    });

    async function handleSave() {
        startLoading();
        try {
            await api.updateSettings(settings);
            toast.success('Pengaturan global berhasil diperbarui!');
        } catch (error) {
            toast.error('Gagal memperbarui pengaturan.');
            console.error(error);
        } finally {
            stopLoading();
        }
    }
</script>

<div class="p-8 max-w-4xl mx-auto">
    <div class="mb-8">
        <h1 class="text-2xl font-bold text-slate-900">Pengaturan</h1>
        <p class="text-slate-500 mt-1">Kelola data pejabat penandatangan default (PPK & Bendahara) untuk seluruh sistem.</p>
    </div>

    <div class="space-y-6">
        <!-- Signatories Section -->
        <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
            <div class="p-6 border-b border-slate-100 bg-slate-50/50">
                <h3 class="font-bold text-slate-800 flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                    Pejabat Penandatangan Default
                </h3>
            </div>
            
            <div class="p-6">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
                    <!-- PPK -->
                    <div class="space-y-4">
                        <div class="pb-2 border-b border-slate-100">
                            <span class="text-xs font-bold text-indigo-600 uppercase tracking-wider">Pejabat Pembuat Komitmen (PPK)</span>
                        </div>
                        {#if !$settingsStore.isLoaded && !settings.ppk_name}
                            <!-- Skeleton loader for PPK -->
                            <div class="space-y-4 animate-pulse">
                                <div class="space-y-1.5"><div class="h-4 bg-slate-200 rounded w-24"></div><div class="h-10 bg-slate-100 rounded border border-slate-200"></div></div>
                                <div class="space-y-1.5"><div class="h-4 bg-slate-200 rounded w-16"></div><div class="h-10 bg-slate-100 rounded border border-slate-200"></div></div>
                            </div>
                        {:else}
                            <div class="space-y-4">
                                <div class="space-y-1.5">
                                    <Label for="ppk_name" class="text-xs font-semibold text-slate-600">Nama Lengkap</Label>
                                    <Input 
                                        id="ppk_name"
                                        bind:value={settings.ppk_name}
                                        placeholder="Contoh: Arief Hafidiyanto"
                                        class="focus:ring-indigo-500"
                                    />
                                </div>
                                <div class="space-y-1.5">
                                    <Label for="ppk_nip" class="text-xs font-semibold text-slate-600">NIP</Label>
                                    <Input 
                                        id="ppk_nip"
                                        bind:value={settings.ppk_nip}
                                        placeholder="Contoh: 19720827 200312 1 002"
                                        class="font-mono"
                                    />
                                </div>
                            </div>
                        {/if}
                    </div>

                    <!-- Bendahara -->
                    <div class="space-y-4">
                        <div class="pb-2 border-b border-slate-100">
                            <span class="text-xs font-bold text-indigo-600 uppercase tracking-wider">Bendahara Pengeluaran Pembantu</span>
                        </div>
                        {#if !$settingsStore.isLoaded && !settings.bendahara_name}
                            <!-- Skeleton loader for Bendahara -->
                            <div class="space-y-4 animate-pulse">
                                <div class="space-y-1.5"><div class="h-4 bg-slate-200 rounded w-24"></div><div class="h-10 bg-slate-100 rounded border border-slate-200"></div></div>
                                <div class="space-y-1.5"><div class="h-4 bg-slate-200 rounded w-16"></div><div class="h-10 bg-slate-100 rounded border border-slate-200"></div></div>
                            </div>
                        {:else}
                            <div class="space-y-4">
                                <div class="space-y-1.5">
                                    <Label for="bendahara_name" class="text-xs font-semibold text-slate-600">Nama Lengkap</Label>
                                    <Input 
                                        id="bendahara_name"
                                        bind:value={settings.bendahara_name}
                                        placeholder="Contoh: Liana Setyawati"
                                        class="focus:ring-indigo-500"
                                    />
                                </div>
                                <div class="space-y-1.5">
                                    <Label for="bendahara_nip" class="text-xs font-semibold text-slate-600">NIP</Label>
                                    <Input 
                                        id="bendahara_nip"
                                        bind:value={settings.bendahara_nip}
                                        placeholder="Contoh: 19800512 200901 2 001"
                                        class="font-mono"
                                    />
                                </div>
                            </div>
                        {/if}
                    </div>
                </div>
            </div>

            <div class="p-6 bg-slate-50 border-t border-slate-100 flex justify-end">
                <Button 
                    on:click={handleSave} 
                    class="bg-indigo-600 hover:bg-indigo-700 text-white px-8"
                >
                    Simpan Perubahan
                </Button>
            </div>
        </div>

        <!-- Info Box -->
        <div class="p-4 bg-blue-50 border border-blue-100 rounded-xl flex gap-3 text-blue-800">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div class="text-sm">
                <p class="font-bold">Informasi</p>
                <p class="mt-1 opacity-90">Data pejabat di atas akan digunakan sebagai default pada setiap dokumen (SPD, Rincian, Laporan). Pengguna masih dapat mengubah data pejabat secara manual pada form laporan jika diperlukan.</p>
            </div>
        </div>
    </div>
</div>
