<script>
    import { usersStore, addUser, updateUser, removeUser, loadUsers, isFetchingUsers } from '$lib/features/auth/store';
    import { userStore } from '$lib/features/auth/store';
    import { loadingStore, startLoading, stopLoading } from '$lib/shared/stores/loading';
    import { cn } from '$lib/shared/utils/utils';
    import { onMount } from 'svelte';
    
    onMount(() => {
        if ($usersStore.length === 0) {
            loadUsers();
        }
    });
    
    // UI Components
    import ChangePasswordModal from '$lib/features/auth/ui/ChangePasswordModal.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Dialog from '$lib/shared/ui/dialog/Dialog.svelte';
    import DialogHeader from '$lib/shared/ui/dialog/DialogHeader.svelte';
    import DialogTitle from '$lib/shared/ui/dialog/DialogTitle.svelte';
    import DialogFooter from '$lib/shared/ui/dialog/DialogFooter.svelte';
    import Table from '$lib/shared/ui/table/Table.svelte';
    import TableHeader from '$lib/shared/ui/table/TableHeader.svelte';
    import TableRow from '$lib/shared/ui/table/TableRow.svelte';
    import TableHead from '$lib/shared/ui/table/TableHead.svelte';
    import TableBody from '$lib/shared/ui/table/TableBody.svelte';
    import TableCell from '$lib/shared/ui/table/TableCell.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import { ConfirmationModal } from '$lib/shared/ui/confirmation-modal';
    import AlertModal from '$lib/shared/ui/alert-modal/AlertModal.svelte';

    let searchQuery = '';
    let isModalOpen = false;
    let isConfirmOpen = false;
    /** @type {number | string | null} */
    let editingId = null; // null for add mode, number for edit mode

    // Alert & Delete State
    let isAlertOpen = false;
    let alertTitle = '';
    let alertDescription = '';
    let isDeleteConfirmOpen = false;
    /** @type {number | string | null} */
    let deleteId = null;

    // Password Change UI state
    let isPasswordModalVisible = $userStore.requirePasswordChange;

    import { toast } from '$lib/shared/stores/toast';

    // Form State
    let formData = {
        name: '',
        email: '',
        password: '',
        role: 'protokol',
        nip: '',
        nomorHp: '',
        pangkat: '',
        golongan: '',
        jabatan: '',
        tingkatBiaya: ''
    };

    let currentPage = 1;
    let itemsPerPage = 12;

    $: filteredUsers = $usersStore.filter(u => 
        u.role !== 'keuangan' && (
            (u.name?.toLowerCase() || '').includes(searchQuery.toLowerCase()) || 
            (u.email?.toLowerCase() || '').includes(searchQuery.toLowerCase()) ||
            (u.nip || '').includes(searchQuery)
        )
    );

    $: totalPages = Math.ceil(filteredUsers.length / itemsPerPage) || 1;
    $: paginatedUsers = filteredUsers.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

    // Reset pagination to page 1 when search query changes
    $: if (searchQuery !== undefined) {
        currentPage = 1;
    }

    function openAddModal() {
        editingId = null;
        formData = { name: '', email: '', password: '', role: 'protokol', nip: '', nomorHp: '', pangkat: '', golongan: '', jabatan: '', tingkatBiaya: '' };
        isModalOpen = true;
    }

    /** @param {{ id: any; name?: string; email?: string; password?: string; role?: any; nip?: string; nomorHp?: string; }} user */
    function openEditModal(user) {
        editingId = user.id;
        formData = { ...user, password: '' }; // Clear password on edit init
        isModalOpen = true;
    }

    // ... (rest of functions) ...

    /* IN HTML Template: Update Table and Modal */
    /* I need to replace the Table and Modal blocks entirely to insert the new columns/fields cleanly */
    
    // ... inside <TableHead> ...
    // Add NIP and HP columns
    
    // ... inside <TableRow> ...
    // Add cells
    
    // ... inside Modal ...
    // Add Inputs


    function confirmSubmit() {
        if (!formData.name || !formData.email || !formData.nomorHp || (!editingId && !formData.password)) {
            toast.error('Nama, Email, Nomor WA, dan Password (untuk user baru) wajib diisi');
            return;
        }
        isConfirmOpen = true;
    }

    async function processSubmit() {
        startLoading();
        try {
            // @ts-ignore
            if (editingId) {
                // @ts-ignore
                await updateUser(editingId, formData);
                toast.success('Data user berhasil diperbarui');
            } else {
                // @ts-ignore
                await addUser(formData);
                toast.success('User baru berhasil ditambahkan');
            }
            isModalOpen = false;
            isConfirmOpen = false;
        } catch (e) {
            toast.error('Gagal menyimpan data user.');
        } finally {
            stopLoading();
        }
    }

    /** @param {number | string} id */
    function handleDelete(id) {
        deleteId = id;
        isDeleteConfirmOpen = true;
    }

    async function processDelete() {
        if (deleteId) {
            startLoading();
            try {
                await removeUser(deleteId);
                toast.success('User berhasil dihapus');
            } catch (e) {
                toast.error('Gagal menghapus user');
            } finally {
                stopLoading();
            }
            deleteId = null;
        }
        isDeleteConfirmOpen = false;
    }
</script>

<div class="space-y-6 pb-20">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
            <h2 class="text-2xl font-bold tracking-tight text-slate-900">Manajemen User</h2>
            <p class="text-slate-500">Kelola akun pengguna dan hak akses aplikasi.</p>
        </div>
        <div class="flex items-center space-x-2 w-full md:w-auto">
            <Button on:click={openAddModal} class="bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Tambah User
            </Button>
        </div>
    </div>

    {#if $userStore.role !== 'super_admin'}
        <div class="flex flex-col items-center justify-center p-12 text-center border-2 border-dashed border-slate-200 rounded-xl bg-slate-50">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            <h3 class="text-lg font-medium text-slate-900">Akses Dibatasi</h3>
            <p class="text-slate-500 max-w-sm mt-1">Halaman ini khusus untuk Super Admin. Silakan login dengan akun yang sesuai.</p>
        </div>
    {:else}
        <!-- Search Bar -->
        <div class="relative w-full md:w-96">
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
            </div>
            <Input type="text" placeholder="Cari Nama atau Email..." bind:value={searchQuery} class="pl-9 bg-white border-slate-200" />
        </div>

        <!-- Card Grid View (Responsive) -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {#if $isFetchingUsers && $usersStore.length === 0}
                {#each Array(8) as _}
                    <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm space-y-3 relative animate-pulse">
                        <div class="absolute top-4 right-4 flex gap-2">
                             <div class="h-4 w-4 bg-slate-200 rounded"></div>
                             <div class="h-4 w-4 bg-slate-200 rounded"></div>
                        </div>
                        <div>
                            <div class="h-5 bg-slate-200 rounded w-3/4 mb-2"></div>
                            <div class="h-3 bg-slate-100 rounded w-1/2 mt-1 mb-2"></div>
                            
                            <div class="grid grid-cols-2 gap-y-3 gap-x-4 mt-3">
                                <div>
                                    <div class="h-2 bg-slate-200 rounded w-8 mb-1"></div>
                                    <div class="h-3 bg-slate-200 rounded w-16"></div>
                                </div>
                                <div>
                                    <div class="h-2 bg-slate-200 rounded w-16 mb-1"></div>
                                    <div class="h-3 bg-slate-200 rounded w-12"></div>
                                </div>
                                <div class="col-span-2">
                                    <div class="h-2 bg-slate-200 rounded w-24 mb-1"></div>
                                    <div class="h-3 bg-slate-200 rounded w-full"></div>
                                    <div class="h-3 bg-slate-100 rounded w-3/4 mt-1"></div>
                                </div>
                            </div>
                        </div>
                        <div class="flex gap-4 items-center justify-between border-t border-slate-100 pt-3 mt-2">
                             <div class="h-5 bg-slate-200 rounded-full w-20"></div>
                             <div class="flex items-center gap-1.5">
                                 <div class="h-4 w-4 bg-slate-200 rounded"></div>
                                 <div class="h-3 bg-slate-200 rounded w-24"></div>
                             </div>
                        </div>
                    </div>
                {/each}
            {:else if paginatedUsers.length === 0}
                <div class="col-span-full text-center py-12 text-slate-500 italic bg-white rounded-xl border border-slate-200 shadow-sm">
                    <div class="flex flex-col items-center justify-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                        </svg>
                        <span>Tidak ada user ditemukan.</span>
                    </div>
                </div>
            {:else}
                {#each paginatedUsers as user (user.id)}
                    <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm space-y-3 relative hover:shadow-md hover:border-slate-300 transition-all">
                        <div class="absolute top-4 right-4 flex gap-2">
                             <button class="text-slate-400 hover:text-blue-600 transition-colors p-1" title="Edit" on:click={() => openEditModal(user)}>
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                </svg>
                             </button>
                             <button class="text-slate-400 hover:text-red-600 transition-colors p-1" title="Hapus" on:click={() => handleDelete(user.id)}>
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                             </button>
                        </div>
                        <div>
                            <h3 class="font-bold text-slate-800 pr-12 line-clamp-1" title={user.name}>{user.name}</h3>
                            <p class="text-[11px] font-mono text-slate-500 mt-1 mb-2 truncate" title={user.email}>{user.email}</p>
                            
                            <div class="grid grid-cols-2 gap-y-3 gap-x-4 mt-3">
                                <div>
                                    <p class="text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">NIP</p>
                                    <p class="text-xs font-mono text-slate-700 truncate" title={user.nip || '-'}>{user.nip || '-'}</p>
                                </div>
                                <div>
                                    <p class="text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Tingkat Biaya</p>
                                    <p class="text-xs font-bold text-slate-700">{user.tingkatBiaya || '-'}</p>
                                </div>
                                <div class="col-span-2">
                                    <p class="text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Jabatan & Pangkat</p>
                                    <p class="text-xs text-slate-700 line-clamp-1" title={user.jabatan || '-'}>{user.jabatan || '-'}</p>
                                    {#if (user.pangkat && user.pangkat !== '-') || (user.golongan && user.golongan !== '-')}
                                        <p class="text-[11px] text-slate-500 mt-0.5 line-clamp-1" title={(user.pangkat && user.pangkat !== '-') && (user.golongan && user.golongan !== '-') ? `${user.pangkat} (${user.golongan})` : ((user.pangkat !== '-' ? user.pangkat : '') || (user.golongan !== '-' ? user.golongan : ''))}>
                                            {(user.pangkat && user.pangkat !== '-') && (user.golongan && user.golongan !== '-') ? `${user.pangkat} (${user.golongan})` : ((user.pangkat !== '-' ? user.pangkat : '') || (user.golongan !== '-' ? user.golongan : ''))}
                                        </p>
                                    {/if}
                                </div>
                            </div>
                        </div>
                        <div class="flex gap-4 items-center justify-between border-t border-slate-100 pt-3 mt-2">
                             <span class={cn("px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider",
                                user.role === 'super_admin' ? "bg-purple-50 text-purple-700 border border-purple-100" :
                                user.role === 'kasubag' ? "bg-amber-50 text-amber-700 border border-amber-100" :
                                "bg-blue-50 text-blue-700 border border-blue-100")}>
                                {user.role}
                            </span>
                        </div>
                    </div>
                {/each}
            {/if}
        </div>

        <!-- Pagination Controls -->
        {#if totalPages > 1}
            <div class="flex items-center justify-between px-4 py-3 bg-white border border-slate-200 mt-6 rounded-xl shadow-sm">
                <div class="flex flex-1 justify-between sm:hidden">
                    <Button variant="outline" size="sm" disabled={currentPage === 1} on:click={() => currentPage--}>
                        Sebelumnya
                    </Button>
                    <Button variant="outline" size="sm" disabled={currentPage === totalPages} on:click={() => currentPage++}>
                        Selanjutnya
                    </Button>
                </div>
                <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
                    <div>
                        <p class="text-sm text-slate-700">
                            Menampilkan <span class="font-medium">{(currentPage - 1) * itemsPerPage + 1}</span> hingga <span class="font-medium">{Math.min(currentPage * itemsPerPage, filteredUsers.length)}</span> dari <span class="font-medium">{filteredUsers.length}</span> hasil
                        </p>
                    </div>
                    <div>
                        <nav class="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
                            <button on:click={() => currentPage--} disabled={currentPage === 1} class="relative inline-flex items-center rounded-l-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed">
                                <span class="sr-only">Previous</span>
                                <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                    <path fill-rule="evenodd" d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z" clip-rule="evenodd" />
                                </svg>
                            </button>
                            {#each Array(totalPages) as _, i}
                                {#if totalPages <= 7 || (i === 0 || i === totalPages - 1 || (i >= currentPage - 2 && i <= currentPage))}
                                    <button on:click={() => currentPage = i + 1} class="relative inline-flex items-center px-4 py-2 text-sm font-semibold {currentPage === i + 1 ? 'z-10 bg-blue-600 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600' : 'text-slate-900 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0'}">
                                        {i + 1}
                                    </button>
                                {:else if i === 1 || i === totalPages - 2}
                                    <span class="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-slate-700 ring-1 ring-inset ring-slate-300 focus:outline-offset-0">...</span>
                                {/if}
                            {/each}
                            <button on:click={() => currentPage++} disabled={currentPage === totalPages} class="relative inline-flex items-center rounded-r-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed">
                                <span class="sr-only">Next</span>
                                <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                    <path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
                                </svg>
                            </button>
                        </nav>
                    </div>
                </div>
            </div>
        {/if}
    {/if}

    <Dialog open={isModalOpen} on:close={() => isModalOpen = false}>
        <DialogHeader class="border-b border-slate-100 pb-4">
            <DialogTitle class="text-xl">{editingId ? 'Edit' : 'Tambah'} User</DialogTitle>
            <p class="text-sm text-slate-500">Lengkapi data akun pengguna.</p>
        </DialogHeader>

        <div class="space-y-4 py-4">
            <div class="space-y-2">
                <Label>Nama Lengkap <span class="text-red-500">*</span></Label>
                <Input type="text" placeholder="Contoh: Staf Pengaju" bind:value={formData.name} />
            </div>
            <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                    <Label>NIP (Opsional)</Label>
                    <Input type="text" placeholder="198..." bind:value={formData.nip} />
                </div>
                <div class="space-y-2">
                    <Label>Nomor HP (WhatsApp) <span class="text-red-500">*</span></Label>
                    <Input type="text" placeholder="08..." bind:value={formData.nomorHp} />
                </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                    <Label>Pangkat</Label>
                    <Input type="text" placeholder="Contoh: Pembina Utama" bind:value={formData.pangkat} />
                </div>
                <div class="space-y-2">
                    <Label>Golongan</Label>
                    <Input type="text" placeholder="Contoh: IV/e" bind:value={formData.golongan} />
                </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                    <Label>Jabatan</Label>
                    <Input type="text" placeholder="Contoh: Analis Protokol" bind:value={formData.jabatan} />
                </div>
                <div class="space-y-2">
                    <Label>Tingkat Biaya</Label>
                    <Input type="text" placeholder="Contoh: B, C, D" bind:value={formData.tingkatBiaya} />
                </div>
            </div>
            <div class="space-y-2">
                <Label>Username <span class="text-red-500">*</span></Label>
                <Input type="text" placeholder="Contoh: budi" bind:value={formData.email} />
            </div>
            <div class="space-y-2">
                <Label>Password {#if editingId}(Kosongkan jika tidak diubah){:else}<span class="text-red-500">*</span>{/if}</Label>
                <Input type="text" placeholder="Minimal 6 karakter" bind:value={formData.password} />
            </div>
            <div class="space-y-2">
                <Label>Role (Hak Akses) <span class="text-red-500">*</span></Label>
                <div class="relative w-full">
                    <Select bind:value={formData.role} class="bg-white border-slate-200">
                        <option value="protokol">Protokol (Staf Pengaju)</option>
                        <option value="super_admin">Super Admin</option>
                        <option value="kasubag">Kasubag (Approval)</option>
                    </Select>
                </div>
            </div>
        </div>

        <DialogFooter class="border-t border-slate-100 pt-4 flex justify-end gap-2">
            <Button variant="outline" on:click={() => isModalOpen = false}>Batal</Button>
            <Button class="bg-blue-600 hover:bg-blue-700 text-white" on:click={confirmSubmit}>Simpan</Button>
        </DialogFooter>
    </Dialog>

    <ConfirmationModal
        bind:open={isConfirmOpen}
        title={editingId ? 'Simpan Perubahan User' : 'Tambah User Baru'}
        description="Apakah Anda yakin data user ini sudah benar?"
        confirmText="Ya, Simpan"
        onConfirm={processSubmit}
    />
    
    <ConfirmationModal
        bind:open={isDeleteConfirmOpen}
        title="Hapus User"
        description="Apakah Anda yakin ingin menghapus user ini? Tindakan ini tidak dapat dibatalkan."
        confirmText="Ya, Hapus"
        onConfirm={processDelete}
        variant="destructive"
    />

    <AlertModal
        bind:open={isAlertOpen}
        title={alertTitle}
        description={alertDescription}
    />
</div>
