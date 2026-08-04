<script>
    import { getInitials, getStatusBadge, toTitleCase, formatLocations } from '$lib/shared/utils/utils';
    import { userStore } from '$lib/features/auth/store';

    export let record;
    export let formatIDR;

    $: badge = getStatusBadge(record);
    $: isProtokol = $userStore.role !== 'super_admin' && $userStore.role !== 'kasubag';
</script>

<tr class="group hover:bg-slate-50/80 transition-all duration-200 cursor-default">
    {#if isProtokol}
        <!-- Nomor SPJ Column -->
        <td class="px-6 py-4 align-top">
            <span class="font-mono text-xs font-bold text-slate-700 bg-slate-100 px-2 py-1 rounded-md border border-slate-200">
                {record.nomorSpdPetugas || '001'}
            </span>
        </td>

        <!-- Destination Column -->
        <td class="px-6 py-4 align-top">
            <div class="flex flex-col space-y-1">
                <div class="flex items-start gap-1.5 font-medium text-slate-700">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400 mt-0.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd" />
                    </svg>
                    <span class="leading-tight">{formatLocations(record)}</span>
                </div>
            </div>
        </td>

        <!-- Date Column (Pergi dan Pulang) -->
        <td class="px-6 py-4 align-top">
            <div class="flex items-center gap-1.5 text-xs text-slate-700 font-medium bg-slate-50 w-fit px-2 py-1 rounded-md border border-slate-100">
                <span>{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })}</span>
                <span class="text-slate-400">&rarr;</span>
                <span>{new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}</span>
            </div>
        </td>

        <!-- Status Column -->
        <td class="px-6 py-4 align-top text-center">
            <span class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border ring-1 ring-inset {badge.class}">
                {badge.label}
            </span>
        </td>
    {:else}
        <!-- Employee Column -->
        <td class="px-6 py-4 align-top">
            <div class="flex items-center gap-3">
                <div class="relative flex-none">
                    <div class="h-10 w-10 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 text-white flex items-center justify-center text-xs font-bold shadow-md ring-2 ring-white">
                        {getInitials(record.employee?.name || '-')}
                    </div>
                    <div class="absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 bg-green-500 border-2 border-white rounded-full"></div>
                </div>
                <div class="flex flex-col min-w-0">
                    <span class="font-bold text-slate-800 truncate group-hover:text-blue-700 transition-colors">{record.employee?.name || '-'}</span>
                    <span class="text-[10px] text-slate-500 font-mono tracking-tight bg-slate-100 px-1.5 py-0.5 rounded-md w-fit mt-0.5 border border-slate-200/50">
                        {record.nomorSpdPetugas || record.spd}
                    </span>
                </div>
            </div>
        </td>

        <!-- Destination Column -->
        <td class="px-6 py-4 align-top">
            <div class="flex flex-col space-y-1">
                <div class="flex items-start gap-1.5 font-medium text-slate-700">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400 mt-0.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd" />
                    </svg>
                    <span class="leading-tight">{formatLocations(record)}</span>
                </div>
            </div>
        </td>

        <!-- Date Column -->
        <td class="px-6 py-4 align-top">
            <div class="flex flex-col space-y-1">
                <span class="text-sm text-slate-700 font-medium">
                    {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })}
                </span>
                <span class="text-xs text-slate-400">
                    {new Date(record.startDate).getFullYear()}
                </span>
            </div>
        </td>

        <!-- Status Column -->
        <td class="px-6 py-4 align-top text-center">
            <span class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border ring-1 ring-inset {badge.class}">
                {badge.label}
            </span>
        </td>

        <!-- Cost Column -->
        <td class="px-6 py-4 align-top text-right">
            <div class="flex flex-col items-end gap-0.5">
                <span class="font-bold text-slate-700 font-feature-settings-tnum">
                    {formatIDR(record.totalCost)}
                </span>
                {#if record.status === 'Approved' || record.paymentStatus === 'Paid'}
                     <span class="text-[10px] text-emerald-600 font-medium bg-emerald-50 px-1 rounded">Realisasi</span>
                {:else}
                     <span class="text-[10px] text-slate-400 font-medium">Estimasi</span>
                {/if}
            </div>
        </td>
    {/if}
</tr>

<style>
    .font-feature-settings-tnum {
        font-feature-settings: "tnum";
        font-variant-numeric: tabular-nums;
    }
</style>
