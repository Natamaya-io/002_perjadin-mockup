<script>
    export let stakeholders = [];
    export let selectedStakeholder = '';
    export let readonly = false;
</script>

<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
        <h3 class="font-semibold text-slate-800">Stakeholder Terkait</h3>
    </div>
    <div class="p-6 space-y-3">
        {#each stakeholders as sh}
            <label class="flex items-start space-x-3 p-3 rounded-lg {readonly ? 'cursor-default opacity-70' : 'hover:bg-slate-50 cursor-pointer'} transition-colors">
                <input type="radio" name="stakeholder" value={sh} bind:group={selectedStakeholder} disabled={readonly} class="mt-1 accent-blue-600 h-4 w-4 shrink-0" />
                <span class="text-sm text-slate-600 leading-snug">{sh}</span>
            </label>
        {/each}
    </div>
</div>
