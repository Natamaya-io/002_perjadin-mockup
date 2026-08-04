<script>
    import { page } from '$app/stores';
    import { userStore, logout } from '$lib/features/auth/store';

    $: activeRoute = $page.url?.pathname || '';
</script>

<header class="hidden md:block sticky top-0 z-50 w-full border-b border-slate-200 bg-white shadow-sm transition-all flex-none print:hidden">
    <div class="container flex h-16 items-center max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="mr-8 flex items-center">
        <div class="flex items-center space-x-3 select-none">
          <img src="/kemnaker-ri.webp" alt="Logo" class="h-10 w-auto object-contain" />
          <div class="hidden sm:flex flex-col leading-none">
            <span class="font-serif font-bold text-lg text-slate-900 tracking-tight">Perjadin</span>
            <span class="font-sans text-xs italic text-slate-500 tracking-widest uppercase">Protokol</span>
          </div>
        </div>
      </div>
      
      <nav class="flex items-center space-x-1 text-sm font-medium flex-1">
        <a class="px-4 py-2 rounded-lg transition-all hover:bg-slate-100/80 {activeRoute === '/dashboard' ? 'text-primary bg-blue-50/50 font-semibold' : 'text-slate-600'}" href="/dashboard">Dashboard</a>
        
        <a class="px-4 py-2 rounded-lg transition-all hover:bg-slate-100/80 {activeRoute.includes('/dashboard/pengajuan') ? 'text-primary bg-blue-50/50 font-semibold' : 'text-slate-600'}" href="/dashboard/pengajuan/new">Pengajuan</a>
        
        {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}
            <a class="px-4 py-2 rounded-lg transition-all hover:bg-slate-100/80 {activeRoute.includes('/dashboard/admin/perdin') ? 'text-primary bg-blue-50/50 font-semibold' : 'text-slate-600'}" href="/dashboard/admin/perdin">Rekap & Kalkulasi</a>
        {/if}
        {#if $userStore.role === 'super_admin'}
            <a class="px-4 py-2 rounded-lg transition-all hover:bg-slate-100/80 {activeRoute.includes('/dashboard/admin/users') ? 'text-primary bg-blue-50/50 font-semibold' : 'text-slate-600'}" href="/dashboard/admin/users">User</a>
        {/if}
        <a class="px-4 py-2 rounded-lg transition-all hover:bg-slate-100/80 {activeRoute.includes('/dashboard/laporan') ? 'text-primary bg-blue-50/50 font-semibold' : 'text-slate-600'}" href="/dashboard/laporan">Laporan</a>
      </nav>

      <div class="flex items-center space-x-4">
        <div class="flex items-center space-x-3 text-sm bg-slate-100/50 px-3 py-1.5 rounded-full border border-slate-200/60">
            <span class="text-xs font-semibold text-slate-500 uppercase tracking-wider">Role</span>
            <span class="text-xs font-bold text-slate-800 uppercase">{$userStore.role}</span>
        </div>
        <button class="text-slate-400 hover:text-red-600 transition-colors p-2 rounded-full hover:bg-red-50" title="Logout" on:click={logout}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
        </button>
      </div>
    </div>
  </header>
