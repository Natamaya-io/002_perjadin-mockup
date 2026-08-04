<script>
	import Input from '$lib/shared/ui/input/Input.svelte';
	import Select from '$lib/shared/ui/select/Select.svelte';

	export let searchQuery = '';
	export let statusFilter = 'all';
	export let sortOption = 'date-desc';
	export let startDate = '';
	export let endDate = '';
	export let statusOptions = [
		{ value: 'all', label: 'Semua Status' },
		{ value: 'In Progress', label: 'In Progress' },
		{ value: 'Completed', label: 'Completed' }
	];
</script>

<div class="flex flex-wrap items-center gap-3 w-full">
	<!-- Search -->
	<div class="relative w-full sm:w-64">
			<div
				class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-4 w-4"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
					/>
				</svg>
			</div>
			<Input
				type="text"
				placeholder="Cari No. SPJ atau Tujuan..."
				bind:value={searchQuery}
				class="pl-9 h-9 text-sm bg-white border-slate-200 w-full"
			/>
		</div>

		<!-- Status & Sort -->
		<div class="flex gap-2 w-full sm:w-auto">
			<div class="flex-1 sm:flex-none sm:w-32">
				<Select bind:value={statusFilter} class="h-9 text-xs w-full bg-white border-slate-200">
					{#each statusOptions as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</Select>
			</div>

			<div class="flex-1 sm:flex-none sm:w-40">
				<Select bind:value={sortOption} class="h-9 text-xs w-full bg-white border-slate-200">
					<option value="spj-desc">ID SPJ Terbaru</option>
					<option value="spj-asc">ID SPJ Terlama</option>
					<option value="date-desc">Tanggal Terbaru</option>
					<option value="date-asc">Tanggal Terlama</option>
					<option value="cost-desc">Biaya Tertinggi</option>
					<option value="cost-asc">Biaya Terendah</option>
				</Select>
			</div>
		</div>

	<!-- Date Filters -->
	<div class="flex items-center gap-2 w-full sm:w-auto overflow-x-auto pb-1">
		<div
			class="flex items-center gap-2 text-sm text-slate-600 whitespace-nowrap bg-white px-3 py-1.5 rounded-lg border border-slate-200 shadow-sm"
		>
			<span class="font-medium text-xs uppercase tracking-wider text-slate-500"
				>Filter Tanggal:</span
			>
			<div class="flex items-center gap-1">
				<input
					type="date"
					bind:value={startDate}
					class="h-7 text-xs border-slate-200 rounded px-2 focus:ring-blue-500 focus:border-blue-500 text-slate-600"
					placeholder="Dari"
				/>
				<span class="text-slate-400">-</span>
				<input
					type="date"
					bind:value={endDate}
					class="h-7 text-xs border-slate-200 rounded px-2 focus:ring-blue-500 focus:border-blue-500 text-slate-600"
					placeholder="Sampai"
				/>
			</div>
			{#if startDate || endDate}
				<button
					class="ml-2 text-xs text-red-500 hover:text-red-700 font-medium hover:underline transition-all"
					on:click={() => {
						startDate = '';
						endDate = '';
					}}
				>
					Reset
				</button>
			{/if}
		</div>
	</div>
</div>
