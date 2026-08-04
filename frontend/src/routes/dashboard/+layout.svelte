<script>
    import { userStore } from '$lib/features/auth/store';
    import { goto } from '$app/navigation';
    import { onMount, onDestroy } from 'svelte';
    import { browser } from '$app/environment';
    import { page, navigating } from '$app/stores';
    import ChangePasswordModal from '$lib/features/auth/ui/ChangePasswordModal.svelte';

    let unsubscribeUser;
    let unsubscribePage;
    let isPasswordModalVisible = false;

    // Trigger modal visibility only once per session when requirement is detected
    $: if ($userStore.requirePasswordChange && !$userStore.passwordModalDismissed && !isPasswordModalVisible && browser) {
        isPasswordModalVisible = true;
    }

    // Persist the dismissed state to the store so it survives navigation
    $: if (!isPasswordModalVisible && $userStore.requirePasswordChange && browser) {
        userStore.update(u => ({ ...u, passwordModalDismissed: true }));
    }

    const accessRules = [
        { path: '/dashboard/admin/users', roles: ['super_admin', 'kasubag', 'protokol'] },
        { path: '/dashboard/pengajuan', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/pengajuan/new', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/admin/perdin', roles: ['super_admin', 'kasubag'] },
        { path: "/dashboard/gup", roles: ["super_admin", "kasubag"] },
        { path: "/dashboard/dalkot/rekap", roles: ["super_admin", "kasubag"] },
        { path: "/dashboard/dalkot", roles: ["super_admin", "kasubag", "protokol"] },
        { path: '/dashboard/billing-protokol', roles: ['protokol', 'super_admin', 'kasubag'] },
        { path: '/dashboard/laporan', roles: ['protokol', 'super_admin', 'kasubag'] },
        { path: '/dashboard/vip-bandara', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/total-penarikan', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/sewa-kendaraan', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/pemeliharaan', roles: ['super_admin', 'kasubag'] }
    ];

    function checkAccess(pathname, role) {
        // Find if any rule matches the current path
        for (const rule of accessRules) {
            if (pathname.startsWith(rule.path)) {
                if (!rule.roles.includes(role)) {
                    return false; // Access denied
                }
            }
        }
        return true; // No rule restricting this path, or role is allowed
    }

    onMount(() => {
        if (browser && !$userStore.loggedIn) {
            goto('/login');
            return;
        }

        if (browser && $userStore.loggedIn) {
             if (!checkAccess($page.url.pathname, $userStore.role)) {
                 goto('/dashboard');
             }
        }

        unsubscribeUser = userStore.subscribe(user => {
            if (browser) {
                if (!user.loggedIn) {
                    goto('/login');
                } else if (!checkAccess($page.url.pathname, user.role)) {
                    goto('/dashboard');
                }
            }
        });

        unsubscribePage = page.subscribe(p => {
             if (browser && $userStore.loggedIn) {
                  if (!checkAccess(p.url.pathname, $userStore.role)) {
                      goto('/dashboard');
                  }
             }
        });
    });

    onDestroy(() => {
        if (unsubscribeUser) unsubscribeUser();
        if (unsubscribePage) unsubscribePage();
    });
</script>

<slot />

<ChangePasswordModal 
    bind:isOpen={isPasswordModalVisible} 
    on:close={() => {
        userStore.update(u => ({ ...u, passwordModalDismissed: true }));
        isPasswordModalVisible = false;
    }}
    on:success={() => {
        userStore.update(u => ({ ...u, requirePasswordChange: false }));
        isPasswordModalVisible = false;
    }} 
/>

{#if $navigating}
    <div class="global-progress-bar"></div>
{/if}

<style>
    .global-progress-bar {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        height: 3px;
        background-color: #3b82f6; /* Tailwind blue-500 */
        z-index: 99999;
        animation: loader-progress 1.5s cubic-bezier(0.4, 0, 0.2, 1) infinite;
        transform-origin: left;
    }

    @keyframes loader-progress {
        0% { transform: scaleX(0); opacity: 1; }
        50% { transform: scaleX(0.7); opacity: 1; }
        100% { transform: scaleX(1); opacity: 0; }
    }
</style>
