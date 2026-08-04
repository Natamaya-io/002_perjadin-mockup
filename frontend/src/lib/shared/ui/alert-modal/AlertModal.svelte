<script>
    import Dialog from '$lib/shared/ui/dialog/Dialog.svelte';
    import DialogHeader from '$lib/shared/ui/dialog/DialogHeader.svelte';
    import DialogTitle from '$lib/shared/ui/dialog/DialogTitle.svelte';
    import DialogFooter from '$lib/shared/ui/dialog/DialogFooter.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';

    export let open = false;
    export let title = 'Pemberitahuan';
    export let description = '';
    export let buttonText = 'OK';
    export let onClose = () => {};

    function handleClose() {
        if (onClose) onClose();
        open = false;
    }
</script>

<Dialog bind:open on:close={handleClose}>
    <DialogHeader class="border-b border-slate-100 pb-4">
        <DialogTitle class="text-xl text-slate-900">{title}</DialogTitle>
        <p class="text-sm text-slate-500 mt-2">{description}</p>
    </DialogHeader>
    
    <DialogFooter class="border-t border-slate-100 pt-4 flex justify-end">
        <Button class="bg-blue-600 hover:bg-blue-700 text-white shadow-lg min-w-[80px]" on:click={handleClose}>{buttonText}</Button>
    </DialogFooter>
</Dialog>
