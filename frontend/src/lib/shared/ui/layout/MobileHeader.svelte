<script>
    import { logout } from '$lib/features/auth/store';
    import { createEventDispatcher } from 'svelte';

    const dispatch = createEventDispatcher();
</script>

<header class="md:hidden sticky top-0 z-50 w-full bg-white border-b border-slate-200 px-4 py-3 flex justify-between items-center shadow-sm flex-none print:hidden">
    <div class="flex items-center gap-3">
        <button 
            class="text-slate-500 hover:text-slate-700 focus:outline-none" 
            on:click={() => dispatch('toggleSidebar')}
            aria-label="Toggle Sidebar"
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
        </button>
    </div>

    <button class="text-slate-500 hover:text-red-600 transition-colors" on:click={logout} title="Logout">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
      </svg>
    </button>
</header>
