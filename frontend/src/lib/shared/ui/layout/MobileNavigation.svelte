<script>
    import { page } from '$app/stores';
    import { userStore } from '$lib/features/auth/store';

    $: activeRoute = $page.url?.pathname || '';
</script>

<nav class="mobile-nav-root md:hidden sticky bottom-0 left-0 right-0 z-50 bg-white border-t border-slate-200 pb-[env(safe-area-inset-bottom)] shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)] flex-none print:hidden">
    <div class="mobile-nav-grid grid {$userStore.role === 'super_admin' ? 'grid-cols-5' : ($userStore.role === 'kasubag' ? 'grid-cols-4' : 'grid-cols-3')} h-16 w-full">
        
        <!-- Home Link -->
        <div class="nav-item-wrapper h-full w-full">
            <a href="/dashboard" class="nav-link flex flex-col items-center justify-center space-y-1 h-full w-full group {activeRoute === '/dashboard' ? 'text-blue-600' : 'text-slate-400 hover:text-slate-600'}">
                <div class="icon-container relative">
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon h-6 w-6 transition-transform group-active:scale-90" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                    </svg>
                </div>
                <div class="label-container">
                    <span class="label text-[10px] font-medium">Home</span>
                </div>
            </a>
        </div>

        <!-- Pengajuan Link -->
        <div class="nav-item-wrapper h-full w-full">
            <a href="/dashboard/pengajuan/new" class="nav-link flex flex-col items-center justify-center space-y-1 h-full w-full group {activeRoute.includes('/dashboard/pengajuan') ? 'text-blue-600' : 'text-slate-400 hover:text-slate-600'}">
                <div class="icon-container relative">
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon h-6 w-6 transition-transform group-active:scale-90" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                </div>
                <div class="label-container">
                    <span class="label text-[10px] font-medium">Pengajuan</span>
                </div>
            </a>
        </div>

        {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}
        <!-- Admin Link -->
        <div class="nav-item-wrapper h-full w-full">
            <a href="/dashboard/admin/perdin" class="nav-link flex flex-col items-center justify-center space-y-1 h-full w-full group {activeRoute.includes('/dashboard/admin/perdin') ? 'text-blue-600' : 'text-slate-400 hover:text-slate-600'}">
                <div class="icon-container relative">
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon h-6 w-6 transition-transform group-active:scale-90" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 3.666A5.002 5.002 0 0112 21a5.002 5.002 0 01-3-9.666V5.5A2.5 2.5 0 0012 3a2.5 2.5 0 002.5 2.5V7z" />
                    </svg>
                </div>
                <div class="label-container">
                    <span class="label text-[10px] font-medium text-center leading-tight">Rekap & Kalkulasi</span>
                </div>
            </a>
        </div>
        {/if}

        {#if $userStore.role === 'super_admin'}
          <!-- User Link -->
          <div class="nav-item-wrapper h-full w-full">
              <a href="/dashboard/admin/users" class="nav-link flex flex-col items-center justify-center space-y-1 h-full w-full group {activeRoute.includes('/dashboard/admin/users') ? 'text-blue-600' : 'text-slate-400 hover:text-slate-600'}">
                  <div class="icon-container relative">
                      <svg xmlns="http://www.w3.org/2000/svg" class="icon h-6 w-6 transition-transform group-active:scale-90" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                  </div>
                  <div class="label-container">
                      <span class="label text-[10px] font-medium">User</span>
                  </div>
              </a>
          </div>
        {/if}

        <!-- Laporan Perjadin Link -->
        <div class="nav-item-wrapper h-full w-full">
            <a href="/dashboard/laporan" class="nav-link flex flex-col items-center justify-center space-y-1 h-full w-full group {activeRoute.includes('/dashboard/laporan') ? 'text-blue-600' : 'text-slate-400 hover:text-slate-600'}">
                <div class="icon-container relative">
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon h-6 w-6 transition-transform group-active:scale-90" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                </div>
                <div class="label-container">
                    <span class="label text-[10px] font-medium text-center leading-tight">Laporan<br/>Perjadin</span>
                </div>
            </a>
        </div>
    </div>
</nav>
