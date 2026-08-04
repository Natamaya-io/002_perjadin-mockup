<script>
    import Dialog from '$lib/shared/ui/dialog/Dialog.svelte';
    import DialogHeader from '$lib/shared/ui/dialog/DialogHeader.svelte';
    import DialogTitle from '$lib/shared/ui/dialog/DialogTitle.svelte';
    import DialogFooter from '$lib/shared/ui/dialog/DialogFooter.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';

    export let open = false;
    export let title = 'Konfirmasi';
    export let description = 'Apakah Anda yakin ingin melanjutkan?';
    export let confirmText = 'Ya, Lanjutkan';
    export let cancelText = 'Batal';
    export let onConfirm = () => {};
    export let onCancel = () => {};
    export let variant = 'default'; // 'default' | 'destructive'

    function handleConfirm() {
        onConfirm();
        open = false;
    }

    function handleCancel() {
        onCancel();
        open = false;
    }
</script>

<Dialog bind:open on:close={handleCancel}>
    <DialogHeader class="border-b border-slate-100 pb-4">
        <DialogTitle class="text-xl text-slate-900">{title}</DialogTitle>
        <p class="text-sm text-slate-500">{description}</p>
    </DialogHeader>
    
    <DialogFooter class="border-t border-slate-100 pt-4 flex justify-end gap-2">
        <Button variant="outline" on:click={handleCancel}>{cancelText}</Button>
        <Button class="{variant === 'destructive' ? 'bg-red-600 hover:bg-red-700' : 'bg-blue-600 hover:bg-blue-700'} text-white shadow-lg" on:click={handleConfirm}>{confirmText}</Button>
    </DialogFooter>
</Dialog>
