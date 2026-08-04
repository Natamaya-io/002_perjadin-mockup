<script>
	import '../app.css';
	import { userStore } from '$lib/features/auth/store';
	import { loadingStore } from '$lib/shared/stores/loading';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import { fly, fade } from 'svelte/transition';
	import { onMount } from 'svelte';
	import { goto, beforeNavigate, afterNavigate } from '$app/navigation';

	import GlobalLoader from '$lib/shared/ui/loader/GlobalLoader.svelte';
	import Toaster from '$lib/shared/ui/toast/Toaster.svelte';

	// Layout Components
	import Sidebar from '$lib/shared/ui/layout/Sidebar.svelte';
	import MobileHeader from '$lib/shared/ui/layout/MobileHeader.svelte';
	import SessionExpiredModal from '$lib/features/auth/ui/SessionExpiredModal.svelte';
	import { sessionExpiredReason } from '$lib/features/auth/store';

	$: activeRoute = $page.url?.pathname || '';
	$: isBlankPage = activeRoute.startsWith('/login') || activeRoute.startsWith('/print');

	// Set ke true untuk mengaktifkan mode maintenance
	export let MAINTENANCE_MODE = false;

	// Auth Redirect
	$: if (!MAINTENANCE_MODE && !$userStore.loggedIn && !isBlankPage) {
		if (browser && !$sessionExpiredReason) {
			goto('/login');
		}
	}

	let navLoading = false;
	let navTimer;
	let storeLoading = false;
	let storeTimer;
	let mobileSidebarOpen = false;

	function toggleMobileSidebar() {
		mobileSidebarOpen = !mobileSidebarOpen;
	}

	function closeMobileSidebar() {
		mobileSidebarOpen = false;
	}

	// Navigation handling - We removed navLoading completely to prevent phantom loaders.
	// SvelteKit's client-side routing is fast enough that a full-screen blocker is jarring.
	// Instead, we just close the sidebar on navigation.
	beforeNavigate(({ to, from }) => {
		if (to?.url.pathname !== from?.url.pathname) {
			mobileSidebarOpen = false;
		}
	});

	// Handle loadingStore securely without reactivity loops
	function handleStoreLoading(isLoadingState) {
		if (isLoadingState) {
			clearTimeout(storeTimer);
			storeTimer = setTimeout(() => {
				storeLoading = true;
			}, 300);
		} else {
			clearTimeout(storeTimer);
			storeLoading = false;
		}
	}
	$: handleStoreLoading($loadingStore);

	// Handle initial hydration loading
	onMount(() => {
		// Initialization logic if any
	});

	$: showLoader = storeLoading;
</script>

{#if showLoader && !MAINTENANCE_MODE}
	<GlobalLoader />
{/if}

<Toaster />

{#if MAINTENANCE_MODE}
	<div
		class="h-[100dvh] w-screen bg-slate-900 flex flex-col items-center justify-center p-6 selection:bg-cyan-500/30"
	>
		<div class="text-center space-y-8 max-w-2xl mx-auto">
			<!-- Animated Gear Icon -->
			<div
				class="relative inline-flex items-center justify-center w-28 h-28 rounded-full bg-slate-800/50 shadow-2xl shadow-cyan-900/20 border border-slate-700/50 backdrop-blur-xl mb-2"
			>
				<div
					class="absolute inset-0 rounded-full border border-cyan-500/20 animate-ping"
					style="animation-duration: 3s;"
				></div>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-14 w-14 text-cyan-400 animate-[spin_4s_linear_infinite]"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
					/>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
					/>
				</svg>
			</div>

			<div class="space-y-4">
				<h1 class="text-4xl md:text-5xl font-extrabold tracking-tight text-white">
					Sistem Sedang <span
						class="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-blue-500"
						>Maintanance</span
					>
				</h1>
				<p class="text-lg md:text-xl text-slate-400 leading-relaxed max-w-xl mx-auto">
					Mohon maaf atas ketidaknyamanan ini. Kami sedang melakukan pembaruan dan peningkatan
					sistem untuk memberikan pengalaman yang lebih baik.
				</p>
			</div>

			<!-- Pulsing Status Pill -->
			<div class="pt-4">
				<div
					class="inline-flex items-center gap-2 px-5 py-2.5 rounded-full bg-slate-800/80 border border-slate-700/50 text-sm font-medium text-slate-300"
				>
					<span class="flex h-2.5 w-2.5 relative">
						<span
							class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"
						></span>
						<span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-cyan-500"></span>
					</span>
					Silakan kembali beberapa saat lagi !!!
				</div>
			</div>
		</div>
	</div>
{:else}
	<div
		class="min-h-screen w-full flex font-sans antialiased text-slate-900 {activeRoute.startsWith(
			'/login'
		)
			? 'bg-slate-900'
			: 'bg-slate-50'}"
	>
		{#if $userStore.loggedIn && !isBlankPage}
			{#if mobileSidebarOpen}
				<div
					class="fixed inset-0 z-40 bg-slate-900/80 md:hidden"
					on:click={closeMobileSidebar}
					role="button"
					tabindex="0"
					on:keydown={(e) => e.key === 'Escape' && closeMobileSidebar()}
				></div>
			{/if}
			<Sidebar mobileOpen={mobileSidebarOpen} on:close={closeMobileSidebar} />
		{/if}

		<div class="flex flex-col flex-1 relative w-full {($userStore.loggedIn && !isBlankPage) ? 'md:w-[calc(100vw-80px)] md:ml-[80px]' : ''}">
			{#if $userStore.loggedIn && !isBlankPage}
				<MobileHeader on:toggleSidebar={toggleMobileSidebar} />
			{/if}

			<!-- Main Content Area (Native Browser Hardware-Accelerated Scroll) -->
			<main class="flex-1 w-full relative">
				<div
					class="{$userStore.loggedIn && !isBlankPage
						? 'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'
						: ''} min-h-full"
				>
						<div class="min-h-full">
							<slot />
						</div>
				</div>
			</main>
		</div>
	</div>

	<!-- Render the Session Expired Modal if a reason exists -->
	{#if $sessionExpiredReason}
		<SessionExpiredModal reason={$sessionExpiredReason} />
	{/if}
{/if}
