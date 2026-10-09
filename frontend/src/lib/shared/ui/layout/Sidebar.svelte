<script>
    import { page } from '$app/stores';
    import { userStore, logout } from '$lib/features/auth/store';
    import { createEventDispatcher } from 'svelte';
    import { getInitials } from '$lib/shared/utils/utils';

    const dispatch = createEventDispatcher();

    export let mobileOpen = false;

    $: activeRoute = $page.url.pathname;

    let isSidebarOpen = false;
    let expandedMenus = {
        perdin: false,
        keuangan: false,
        dalkot: false
    };
    let lastActiveRoute = '';

    const icons = {
        dashboard: 'M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM14 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zM14 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z',
        chartPie: 'M10.5 6a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM10.5 6h9.75M10.5 6a1.5 1.5 0 11-3 0m3 0a1.5 1.5 0 10-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m-9.75 0h9.75',
        receipt: 'M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21l-2-2l-2 2l-2-2l-2 2l-2-2l-2 2l-2-2l-2 2l-2-2l-2 2l-2-2l-2 2L3 21V5a2 2 0 012-2h14a2 2 0 012 2v16z',
        map: 'M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7',
        cash: 'M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
        users: 'M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z',
        userCircle: 'M17.982 18.725A7.488 7.488 0 0012 15.75a7.488 7.488 0 00-5.982 2.975m11.963 0a9 9 0 10-11.963 0m11.963 0A8.966 8.966 0 0112 21a8.966 8.966 0 01-5.982-2.275M15 9.75a3 3 0 11-6 0 3 3 0 016 0z',
        plus: 'M12 4.5v15m7.5-7.5h-15',
        penSign: 'M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10',
        chatbot: 'M8.625 9.75a.375.375 0 11-.75 0 .375.375 0 01.75 0zm3.75 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm3.75 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM12 3c-4.97 0-9 3.806-9 8.5 0 2.438 1.087 4.636 2.828 6.187L4.5 21l3.69-1.64A9.42 9.42 0 0012 20c4.97 0 9-3.806 9-8.5S16.97 3 12 3z'
    };

    const dashboardItem = {
        label: 'Dashboard',
        href: '/dashboard',
        icon: icons.dashboard,
        exact: true,
        role: ['super_admin', 'kasubag', 'protokol']
    };

    const nestedMenus = [
        {
            id: 'perdin',
            label: 'Dinas Luar Kota',
            icon: icons.map,
            items: [
                {
                    label: 'Pengajuan Perjalanan',
                    href: '/dashboard/pengajuan',
                    icon: icons.plus,
                    role: ['super_admin', 'kasubag']
                },
                {
                    label: 'Rekap & Kalkulasi',
                    href: '/dashboard/admin/perdin',
                    icon: icons.cash,
                    role: ['super_admin', 'kasubag']
                },
                {
                    label: 'Laporan Perjadin',
                    href: '/dashboard/laporan',
                    icon: icons.map,
                    role: ['protokol', 'super_admin', 'kasubag']
                }
            ]
        },
        {
            id: 'keuangan',
            label: 'Keuangan',
            icon: icons.receipt,
            items: [
                {
                    label: 'Pengajuan GUP',
                    href: '/dashboard/gup',
                    icon: icons.receipt,
                    exact: true,
                    role: ['super_admin', 'kasubag']
                },
                {
                    label: 'Laporan dan Rekapitulasi',
                    href: '/dashboard/gup/laporan',
                    icon: icons.chartPie,
                    role: ['super_admin', 'kasubag']
                },
                {
                    label: 'LS',
                    href: '/dashboard/gup/ls',
                    icon: icons.cash,
                    role: ['super_admin', 'kasubag']
                },
                {
                    label: 'Sumber Dana',
                    href: '/dashboard/keuangan/sumber-dana',
                    icon: icons.cash,
                    role: ['super_admin', 'kasubag']
                }
            ]
        },
        {
            id: 'dalkot',
            label: 'Dalkot',
            icon: icons.map,
            items: [
                {
                    label: 'Daftar Dalkot',
                    href: '/dashboard/dalkot',
                    icon: icons.map,
                    exact: true,
                    role: ['super_admin', 'kasubag', 'protokol']
                },
                
                {
                    label: 'Pengajuan Dalkot',
                    href: '/dashboard/dalkot/pengajuan',
                    icon: icons.plus,
                    role: ['super_admin', 'kasubag', 'protokol']
                },
                {
                    label: 'Rekapitulasi Dalkot',
                    href: '/dashboard/dalkot/rekap',
                    icon: icons.receipt,
                    role: ['super_admin', 'kasubag', 'protokol']
                }
            ]
        }
    ];

    const chatbotItem = {
        label: 'Chatbot',
        href: '/dashboard/chatbot',
        icon: icons.chatbot,
        role: ['super_admin', 'kasubag', 'protokol']
    };

    const adminItems = [
        {
            label: 'User',
            href: '/dashboard/admin/users',
            icon: icons.userCircle,
            role: 'super_admin'
        },
        {
            label: 'Penandatanganan',
            href: '/dashboard/admin/settings',
            icon: icons.penSign,
            role: ['super_admin', 'kasubag']
        }
    ];

    function handleLinkClick() {
        dispatch('close');
    }

    function hasAccess(itemRole) {
        if (!itemRole) return true;
        if (Array.isArray(itemRole)) return itemRole.includes($userStore.role);
        return $userStore.role === itemRole;
    }

    function isItemActive(item, currentPath) {
        const normalizedCurrent = (currentPath || '').replace(/\/+$/, '') || '/';
        const targetPath = item.href.replace(/\/+$/, '') || '/';

        if (normalizedCurrent === targetPath) return true;
        if (item.exact || targetPath === '/dashboard') return false;

        return normalizedCurrent.startsWith(`${targetPath}/`);
    }

    function hasVisibleItems(menu) {
        return menu.items.some((item) => hasAccess(item.role));
    }

    function isMenuActive(menu) {
        return menu.items.some((item) => hasAccess(item.role) && isItemActive(item, activeRoute));
    }

    function toggleMenu(menuId) {
        const shouldOpen = !expandedMenus[menuId];
        expandedMenus = {
            perdin: false,
            keuangan: false,
            dalkot: false,
            [menuId]: shouldOpen
        };
    }

    $: if (activeRoute !== lastActiveRoute) {
        lastActiveRoute = activeRoute;
        const activeMenu = nestedMenus.find((menu) => hasVisibleItems(menu) && isMenuActive(menu));
        expandedMenus = {
            perdin: activeMenu?.id === 'perdin',
            keuangan: activeMenu?.id === 'keuangan',
            dalkot: activeMenu?.id === 'dalkot'
        };
    }
</script>

<aside
    class="sidebar-root fixed inset-y-0 left-0 z-[60] flex h-[100dvh] max-w-[85vw] transform flex-col border-r border-slate-200 bg-white shadow-xl transition-transform duration-200 ease-out md:shadow-[2px_0_8px_-3px_rgba(0,0,0,0.05)] print:hidden {mobileOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'} {isSidebarOpen || mobileOpen ? 'w-[280px]' : 'w-[80px]'}"
    style="will-change: transform;"
    on:mouseenter={() => {
        if (!mobileOpen) isSidebarOpen = true;
    }}
    on:mouseleave={() => {
        if (!mobileOpen) isSidebarOpen = false;
    }}
>
    <div class="sidebar-header-wrapper relative z-50 w-full flex-none">
        <div
            class="sidebar-header-container flex h-[88px] items-center border-b border-slate-100/80 {isSidebarOpen || mobileOpen
                ? 'justify-between pl-6 pr-5'
                : 'justify-center px-0'}"
        >
            <div
                class="logo-group flex items-center overflow-hidden {isSidebarOpen || mobileOpen ? 'gap-3' : 'gap-0'}"
            >
                <div class="logo-image-wrapper flex-shrink-0">
                    <img src="/kemnaker-ri.webp" alt="Logo" class="h-10 w-auto object-contain" />
                </div>
                <div
                    class="logo-text-wrapper flex flex-col leading-none {isSidebarOpen || mobileOpen
                        ? 'opacity-100'
                        : 'hidden opacity-0'}"
                >
                    <span class="logo-title whitespace-nowrap font-serif text-lg font-bold tracking-tight text-slate-900"
                        >Perjadin</span
                    >
                    <span
                        class="logo-subtitle whitespace-nowrap font-sans text-[10px] font-medium uppercase tracking-[0.2em] text-slate-500"
                        >Protokol</span
                    >
                </div>
            </div>
        </div>

        <button
            class="absolute right-4 top-[30px] flex h-8 w-8 items-center justify-center text-slate-400 hover:text-slate-600 focus:outline-none md:hidden"
            on:click={() => dispatch('close')}
            title="Tutup sidebar"
            aria-label="Tutup sidebar"
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
        </button>
    </div>

    <div class="custom-scrollbar flex flex-1 flex-col overflow-y-auto overflow-x-hidden">
        <nav class="sidebar-nav-root flex flex-1 flex-col space-y-1.5 px-3 py-6" aria-label="Navigasi utama">
            {#if hasAccess(dashboardItem.role)}
                <div class="nav-item-wrapper w-full">
                    <div class="nav-item-container" title={!isSidebarOpen && !mobileOpen ? dashboardItem.label : ''}>
                        <a
                            class="nav-link group flex items-center whitespace-nowrap rounded-xl px-3 py-2.5 {isSidebarOpen || mobileOpen
                                ? ''
                                : 'justify-center'} {isItemActive(dashboardItem, activeRoute)
                                ? 'bg-gradient-to-r from-blue-50 to-blue-50/50 text-blue-700 shadow-sm ring-1 ring-blue-100'
                                : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'}"
                            href={dashboardItem.href}
                            on:click={handleLinkClick}
                        >
                            <span
                                class="icon-wrapper flex-shrink-0 {isItemActive(dashboardItem, activeRoute)
                                    ? 'text-blue-600'
                                    : 'text-slate-400 group-hover:text-slate-600'}"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                                    <path stroke-linecap="round" stroke-linejoin="round" d={dashboardItem.icon} />
                                </svg>
                            </span>
                            <span class="text-wrapper {isSidebarOpen || mobileOpen ? 'ml-3 opacity-100' : 'hidden opacity-0'}">
                                <span class="text-inner text-sm font-medium">{dashboardItem.label}</span>
                            </span>
                        </a>
                    </div>
                </div>
            {/if}

            <div class="my-2 h-px bg-slate-100"></div>

            {#each nestedMenus as menu}
                {#if hasVisibleItems(menu)}
                    <div class="nav-group w-full">
                        <button
                            type="button"
                            class="group flex w-full items-center rounded-xl px-3 py-2.5 text-left transition {isSidebarOpen || mobileOpen
                                ? ''
                                : 'justify-center'} {isMenuActive(menu)
                                ? 'bg-blue-50/80 text-blue-700 ring-1 ring-blue-100'
                                : 'text-slate-700 hover:bg-slate-50 hover:text-slate-950'}"
                            title={!isSidebarOpen && !mobileOpen ? menu.label : ''}
                            aria-expanded={expandedMenus[menu.id]}
                            aria-controls={`sidebar-submenu-${menu.id}`}
                            on:click={() => toggleMenu(menu.id)}
                        >
                            <span
                                class="flex-shrink-0 {isMenuActive(menu)
                                    ? 'text-blue-600'
                                    : 'text-slate-400 group-hover:text-slate-600'}"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                                    <path stroke-linecap="round" stroke-linejoin="round" d={menu.icon} />
                                </svg>
                            </span>

                            {#if isSidebarOpen || mobileOpen}
                                <span class="ml-3 min-w-0 flex-1 truncate text-sm font-semibold">{menu.label}</span>
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    class="h-4 w-4 flex-shrink-0 text-slate-400 transition-transform duration-200 {expandedMenus[menu.id]
                                        ? 'rotate-180'
                                        : ''}"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                    stroke="currentColor"
                                    stroke-width="2"
                                >
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
                                </svg>
                            {/if}
                        </button>

                        {#if expandedMenus[menu.id] && (isSidebarOpen || mobileOpen)}
                            <div
                                id={`sidebar-submenu-${menu.id}`}
                                class="relative ml-5 mt-1 space-y-1 border-l border-slate-200 pl-3"
                            >
                                {#each menu.items as item}
                                    {#if hasAccess(item.role)}
                                        <a
                                            class="group flex items-center rounded-lg px-3 py-2 text-sm transition {isItemActive(item, activeRoute)
                                                ? 'bg-blue-600 font-semibold text-white shadow-sm shadow-blue-200'
                                                : 'text-slate-500 hover:bg-slate-50 hover:text-slate-900'}"
                                            href={item.href}
                                            on:click={handleLinkClick}
                                        >
                                            <span
                                                class="mr-2.5 flex-shrink-0 {isItemActive(item, activeRoute)
                                                    ? 'text-white'
                                                    : 'text-slate-400 group-hover:text-slate-600'}"
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-[18px] w-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                                                    <path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
                                                </svg>
                                            </span>
                                            <span class="min-w-0 truncate">{item.label}</span>
                                        </a>
                                    {/if}
                                {/each}
                            </div>
                        {/if}
                    </div>
                {/if}
            {/each}

            <div class="my-2 h-px bg-slate-100"></div>

            {#if hasAccess(chatbotItem.role)}
                <div class="nav-item-wrapper w-full">
                    <div class="nav-item-container" title={!isSidebarOpen && !mobileOpen ? chatbotItem.label : ''}>
                        <a
                            class="nav-link group flex items-center whitespace-nowrap rounded-xl px-3 py-2.5 {isSidebarOpen || mobileOpen
                                ? ''
                                : 'justify-center'} {isItemActive(chatbotItem, activeRoute)
                                ? 'bg-gradient-to-r from-violet-50 to-blue-50 text-violet-700 shadow-sm ring-1 ring-violet-100'
                                : 'text-slate-600 hover:bg-violet-50/70 hover:text-violet-700'}"
                            href={chatbotItem.href}
                            on:click={handleLinkClick}
                        >
                            <span
                                class="icon-wrapper flex-shrink-0 {isItemActive(chatbotItem, activeRoute)
                                    ? 'text-violet-600'
                                    : 'text-slate-400 group-hover:text-violet-600'}"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                                    <path stroke-linecap="round" stroke-linejoin="round" d={chatbotItem.icon} />
                                </svg>
                            </span>
                            <span class="text-wrapper {isSidebarOpen || mobileOpen ? 'ml-3 opacity-100' : 'hidden opacity-0'}">
                                <span class="text-inner text-sm font-semibold">{chatbotItem.label}</span>
                            </span>
                            {#if isSidebarOpen || mobileOpen}
                                <span class="ml-auto rounded-full bg-violet-100 px-2 py-0.5 text-[9px] font-bold uppercase tracking-wider text-violet-700">AI</span>
                            {/if}
                        </a>
                    </div>
                </div>
            {/if}

            {#if adminItems.some((item) => hasAccess(item.role))}
                <div class="nav-item-wrapper relative flex h-8 w-full items-center justify-center pb-2 pt-4 {isSidebarOpen || mobileOpen ? 'px-3' : 'px-0'}">
                    <span
                        class="absolute whitespace-nowrap text-[10px] font-bold uppercase tracking-widest text-slate-400 {isSidebarOpen || mobileOpen
                            ? 'left-3 opacity-100'
                            : 'hidden opacity-0'}"
                    >
                        Zona Admin
                    </span>
                    <div class="h-px bg-slate-200 {isSidebarOpen || mobileOpen ? 'hidden' : 'w-8 opacity-100'}"></div>
                </div>

                {#each adminItems as item}
                    {#if hasAccess(item.role)}
                        <div class="nav-item-wrapper w-full">
                            <div class="nav-item-container" title={!isSidebarOpen && !mobileOpen ? item.label : ''}>
                                <a
                                    class="nav-link group flex items-center whitespace-nowrap rounded-xl px-3 py-2.5 {isSidebarOpen || mobileOpen
                                        ? ''
                                        : 'justify-center'} {isItemActive(item, activeRoute)
                                        ? 'bg-gradient-to-r from-blue-50 to-blue-50/50 text-blue-700 shadow-sm ring-1 ring-blue-100'
                                        : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'}"
                                    href={item.href}
                                    on:click={handleLinkClick}
                                >
                                    <span
                                        class="icon-wrapper flex-shrink-0 {isItemActive(item, activeRoute)
                                            ? 'text-blue-600'
                                            : 'text-slate-400 group-hover:text-slate-600'}"
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                                            <path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
                                        </svg>
                                    </span>
                                    <span class="text-wrapper {isSidebarOpen || mobileOpen ? 'ml-3 opacity-100' : 'hidden opacity-0'}">
                                        <span class="text-inner text-sm font-medium">{item.label}</span>
                                    </span>
                                </a>
                            </div>
                        </div>
                    {/if}
                {/each}
            {/if}
        </nav>
    </div>

    <div class="sidebar-footer-wrapper z-50 flex-none border-t border-slate-100 bg-slate-50/50">
        <div class="sidebar-footer-content p-4">
            <div class="user-session-card flex items-center px-1 py-1 {isSidebarOpen || mobileOpen ? 'justify-between' : 'justify-center'}">
                <div
                    class="user-identity-section flex items-center overflow-hidden {isSidebarOpen || mobileOpen
                        ? 'gap-3 opacity-100'
                        : 'hidden opacity-0'}"
                >
                    <div class="avatar-wrapper flex-shrink-0">
                        <div class="avatar-container relative">
                            <div
                                class="avatar-circle flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 text-white shadow-md shadow-blue-500/20 ring-2 ring-white"
                            >
                                <span class="avatar-text select-none text-xs font-bold uppercase">
                                    {getInitials($userStore.name || $userStore.role || 'US')}
                                </span>
                            </div>
                            <div class="status-indicator-wrapper absolute bottom-0 right-0">
                                <div class="status-dot h-2.5 w-2.5 rounded-full border-2 border-white bg-green-500"></div>
                            </div>
                        </div>
                    </div>

                    <div class="user-details-wrapper flex min-w-0 flex-col justify-center">
                        <span class="whitespace-nowrap text-[10px] font-semibold uppercase leading-tight tracking-wider text-slate-400">
                            Logged in as
                        </span>
                        <span
                            class="mt-0.5 max-w-[120px] truncate text-sm font-bold capitalize leading-tight text-slate-800"
                            title={$userStore.email}
                        >
                            {$userStore.name || $userStore.role?.replace('_', ' ') || 'User'}
                        </span>
                    </div>
                </div>

                <div class="session-actions-section flex items-center {isSidebarOpen || mobileOpen ? 'ml-0' : 'mx-auto w-full justify-center'}">
                    <button
                        class="logout-btn group relative flex items-center justify-center rounded-lg p-2 text-slate-400 transition-all duration-75 hover:bg-red-50 hover:text-red-600 focus:outline-none focus:ring-2 focus:ring-red-100"
                        title="Logout"
                        aria-label="Logout"
                        on:click={logout}
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 transition-transform duration-75 group-hover:scale-110" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                        </svg>
                    </button>
                </div>
            </div>
        </div>
    </div>
</aside>
