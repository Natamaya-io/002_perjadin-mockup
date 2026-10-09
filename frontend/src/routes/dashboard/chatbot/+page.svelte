<script lang="ts">
    import { tick } from 'svelte';
    import { api } from '$lib/shared/api';
    import type {
        ChatbotGupAnalysis,
        ChatbotInsideCityAnalysis,
        ChatbotMessageResponse,
        ChatbotOutsideCityAnalysis
    } from '$lib/shared/api/types';

    type ConversationItem = {
        id: number;
        role: 'user' | 'assistant';
        text: string;
        response?: ChatbotMessageResponse;
        isError?: boolean;
    };

    let sequence = 1;
    let input = '';
    let loading = false;
    let chatContainer: HTMLDivElement;

    let messages: ConversationItem[] = [
        {
            id: sequence++,
            role: 'assistant',
            text: 'Halo. Ketik perintah untuk memulai analisis. Gunakan #Menu-GUP, #Menu-Luar_Kota, atau #Menu-Dalam_Kota untuk melihat perintah yang tersedia.'
        }
    ];

    const initialCommands = ['#Menu-GUP', '#Menu-Luar_Kota', '#Menu-Dalam_Kota'];

    function formatRupiah(value: number | null | undefined) {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            maximumFractionDigits: 0
        }).format(Number(value || 0));
    }

    function formatNumber(value: number | null | undefined) {
        return new Intl.NumberFormat('id-ID').format(Number(value || 0));
    }

    function formatDate(value: string) {
        if (!value) return '-';
        return new Intl.DateTimeFormat('id-ID', {
            day: '2-digit',
            month: 'short',
            year: 'numeric'
        }).format(new Date(`${value}T00:00:00`));
    }

    function asGup(data: unknown) {
        return data as ChatbotGupAnalysis;
    }

    function asOutside(data: unknown) {
        return data as ChatbotOutsideCityAnalysis;
    }

    function asInside(data: unknown) {
        return data as ChatbotInsideCityAnalysis;
    }

    function chartPercent(value: number, data: ChatbotGupAnalysis) {
        const max = Math.max(
            1,
            ...data.monthly.flatMap((row) => [row.income, row.realization, Math.max(0, row.balance)])
        );
        if (value <= 0) return 2;
        return Math.max(4, Math.round((value / max) * 100));
    }

    async function scrollToBottom() {
        await tick();
        chatContainer?.scrollTo({ top: chatContainer.scrollHeight, behavior: 'smooth' });
    }

    async function sendCommand(command?: string) {
        if (loading) return;
        const text = (command ?? input).trim();
        if (!text) return;

        input = '';
        messages = [...messages, { id: sequence++, role: 'user', text }];
        loading = true;
        await scrollToBottom();

        try {
            const response = await api.sendChatbotMessage(text);
            messages = [
                ...messages,
                {
                    id: sequence++,
                    role: 'assistant',
                    text: response.message,
                    response
                }
            ];
        } catch (error: any) {
            messages = [
                ...messages,
                {
                    id: sequence++,
                    role: 'assistant',
                    text: error?.message || 'Perintah gagal diproses.',
                    isError: true
                }
            ];
        } finally {
            loading = false;
            await scrollToBottom();
        }
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            sendCommand();
        }
    }

    function clearChat() {
        messages = [
            {
                id: sequence++,
                role: 'assistant',
                text: 'Percakapan dibersihkan. Ketik perintah baru untuk memulai analisis.'
            }
        ];
        input = '';
    }
</script>

<svelte:head>
    <title>Chatbot Analisis Keuangan | Perjadin</title>
</svelte:head>

<div class="min-h-[calc(100vh-4rem)] bg-slate-100 p-3 md:p-5">
    <section class="mx-auto flex h-[calc(100vh-7rem)] max-w-6xl flex-col overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <header class="flex items-center justify-between border-b border-slate-200 bg-slate-950 px-4 py-4 text-white md:px-6">
            <div>
                <div class="text-xs font-bold uppercase tracking-[0.18em] text-cyan-300">Perjadin Assistant</div>
                <h1 class="mt-1 text-lg font-black md:text-xl">Chatbot Analisis Keuangan</h1>
                <p class="mt-1 text-xs text-slate-400">Data baru dianalisis setelah Anda mengirim perintah.</p>
            </div>
            <button
                type="button"
                class="rounded-xl border border-slate-700 px-3 py-2 text-xs font-bold text-slate-300 hover:border-slate-500 hover:text-white"
                onclick={clearChat}
            >
                Bersihkan Chat
            </button>
        </header>

        <div bind:this={chatContainer} class="flex-1 overflow-y-auto bg-slate-50 px-3 py-5 md:px-6">
            <div class="mx-auto max-w-5xl space-y-5">
                {#each messages as item (item.id)}
                    <div class:flex-row-reverse={item.role === 'user'} class="flex items-start gap-3">
                        <div
                            class={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-xs font-black ${item.role === 'user' ? 'bg-slate-900 text-white' : 'bg-cyan-100 text-cyan-800'}`}
                        >
                            {item.role === 'user' ? 'U' : 'AI'}
                        </div>

                        <div class={`min-w-0 max-w-[92%] ${item.role === 'user' ? 'text-right' : ''}`}>
                            <div
                                class={`inline-block rounded-2xl px-4 py-3 text-left text-sm leading-6 ${item.role === 'user' ? 'bg-slate-900 text-white' : item.isError ? 'border border-red-200 bg-red-50 text-red-800' : 'border border-slate-200 bg-white text-slate-800 shadow-sm'}`}
                            >
                                {item.text}
                            </div>

                            {#if item.response?.suggestions?.length}
                                <div class="mt-3 flex flex-wrap gap-2">
                                    {#each item.response.suggestions as suggestion}
                                        <button
                                            type="button"
                                            class="rounded-full border border-cyan-200 bg-cyan-50 px-3 py-2 text-left text-xs font-bold text-cyan-800 transition hover:border-cyan-400 hover:bg-cyan-100"
                                            onclick={() => sendCommand(suggestion)}
                                        >
                                            {suggestion}
                                        </button>
                                    {/each}
                                </div>
                            {/if}

                            {#if item.response?.data}
                                {@const intent = item.response.intent}

                                {#if intent === 'gup_analysis'}
                                    {@const data = asGup(item.response.data)}
                                    <div class="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                                        <div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Total LS</div><div class="mt-1 text-lg font-black text-slate-900">{formatRupiah(data.summary.totalLs)}</div></div>
                                        <div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Realisasi</div><div class="mt-1 text-lg font-black text-slate-900">{formatRupiah(data.summary.totalRealization)}</div></div>
                                        <div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Balance</div><div class="mt-1 text-lg font-black text-slate-900">{formatRupiah(data.summary.balance)}</div></div>
                                        <div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Persentase Realisasi</div><div class="mt-1 text-lg font-black text-slate-900">{data.summary.realizationPercentage.toLocaleString('id-ID')}%</div></div>
                                        <div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Jumlah Transaksi</div><div class="mt-1 text-lg font-black text-slate-900">{formatNumber(data.summary.transactionCount)}</div></div>
                                        <div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Total Pajak</div><div class="mt-1 text-lg font-black text-slate-900">{formatRupiah(data.summary.totalTax)}</div></div>
                                    </div>
                                {:else if intent === 'gup_ranking'}
                                    {@const data = asGup(item.response.data)}
                                    <div class="mt-3 grid gap-4 xl:grid-cols-2">
                                        <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white">
                                            <div class="border-b border-slate-200 px-4 py-3 text-sm font-black text-slate-900">3 Pengeluaran Terbesar</div>
                                            <div class="overflow-x-auto">
                                                <table class="min-w-full text-left text-xs">
                                                    <thead class="bg-slate-50 text-slate-500"><tr><th class="px-3 py-2">Rank</th><th class="px-3 py-2">Uraian</th><th class="px-3 py-2">Akun</th><th class="px-3 py-2 text-right">Realisasi</th></tr></thead>
                                                    <tbody>
                                                        {#each data.topExpenses as row}
                                                            <tr class="border-t border-slate-100"><td class="px-3 py-3 font-black">{row.rank}</td><td class="px-3 py-3"><div class="font-bold text-slate-800">{row.paymentDescription || '-'}</div><div class="text-slate-500">{row.recipient || '-'} • {formatDate(row.receiptDate)}</div></td><td class="px-3 py-3">{row.accountCode}</td><td class="px-3 py-3 text-right font-black">{formatRupiah(row.paidAmount)}</td></tr>
                                                        {/each}
                                                    </tbody>
                                                </table>
                                            </div>
                                        </div>
                                        <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white">
                                            <div class="border-b border-slate-200 px-4 py-3 text-sm font-black text-slate-900">3 Pengeluaran Terkecil</div>
                                            <div class="overflow-x-auto">
                                                <table class="min-w-full text-left text-xs">
                                                    <thead class="bg-slate-50 text-slate-500"><tr><th class="px-3 py-2">Rank</th><th class="px-3 py-2">Uraian</th><th class="px-3 py-2">Akun</th><th class="px-3 py-2 text-right">Realisasi</th></tr></thead>
                                                    <tbody>
                                                        {#each data.lowestExpenses as row}
                                                            <tr class="border-t border-slate-100"><td class="px-3 py-3 font-black">{row.rank}</td><td class="px-3 py-3"><div class="font-bold text-slate-800">{row.paymentDescription || '-'}</div><div class="text-slate-500">{row.recipient || '-'} • {formatDate(row.receiptDate)}</div></td><td class="px-3 py-3">{row.accountCode}</td><td class="px-3 py-3 text-right font-black">{formatRupiah(row.paidAmount)}</td></tr>
                                                        {/each}
                                                    </tbody>
                                                </table>
                                            </div>
                                        </div>
                                    </div>
                                {:else if intent === 'gup_ls'}
                                    {@const data = asGup(item.response.data)}
                                    <div class="mt-3 overflow-hidden rounded-2xl border border-slate-200 bg-white">
                                        <div class="overflow-x-auto">
                                            <table class="min-w-full text-left text-xs">
                                                <thead class="bg-slate-50 text-slate-500"><tr><th class="px-3 py-3">Bulan</th><th class="px-3 py-3 text-right">Pemasukan</th><th class="px-3 py-3 text-right">Realisasi</th><th class="px-3 py-3 text-right">Balance</th><th class="px-3 py-3 text-right">Transaksi</th><th class="px-3 py-3">Status</th></tr></thead>
                                                <tbody>
                                                    {#each data.monthly as row}
                                                        <tr class="border-t border-slate-100"><td class="px-3 py-3 font-bold">{row.monthName}</td><td class="px-3 py-3 text-right">{formatRupiah(row.income)}</td><td class="px-3 py-3 text-right">{formatRupiah(row.realization)}</td><td class="px-3 py-3 text-right font-bold">{formatRupiah(row.balance)}</td><td class="px-3 py-3 text-right">{formatNumber(row.transactionCount)}</td><td class="px-3 py-3">{row.status}</td></tr>
                                                    {/each}
                                                </tbody>
                                            </table>
                                        </div>
                                    </div>
                                {:else if intent === 'gup_account'}
                                    {@const data = asGup(item.response.data)}
                                    <div class="mt-3 overflow-hidden rounded-2xl border border-slate-200 bg-white">
                                        {#if data.largestAccount}
                                            <div class="border-b border-slate-200 bg-cyan-50 px-4 py-4"><div class="text-xs font-bold uppercase tracking-wider text-cyan-700">Akun Realisasi Terbesar</div><div class="mt-1 text-lg font-black text-slate-900">{data.largestAccount.accountCode} • {formatRupiah(data.largestAccount.paidAmount)}</div><div class="mt-1 text-xs text-slate-600">{data.largestAccount.percentage.toLocaleString('id-ID')}% dari total realisasi • {formatNumber(data.largestAccount.transactionCount)} transaksi</div></div>
                                        {/if}
                                        <div class="overflow-x-auto"><table class="min-w-full text-left text-xs"><thead class="bg-slate-50 text-slate-500"><tr><th class="px-3 py-3">Rank</th><th class="px-3 py-3">Akun</th><th class="px-3 py-3">MAK</th><th class="px-3 py-3 text-right">Realisasi</th><th class="px-3 py-3 text-right">Kontribusi</th></tr></thead><tbody>{#each data.accountRanking as row}<tr class="border-t border-slate-100"><td class="px-3 py-3 font-black">{row.rank}</td><td class="px-3 py-3 font-bold">{row.accountCode}</td><td class="px-3 py-3">{row.mak || '-'}</td><td class="px-3 py-3 text-right font-bold">{formatRupiah(row.paidAmount)}</td><td class="px-3 py-3 text-right">{row.percentage.toLocaleString('id-ID')}%</td></tr>{/each}</tbody></table></div>
                                    </div>
                                {:else if intent === 'gup_chart'}
                                    {@const data = asGup(item.response.data)}
                                    <div class="mt-3 rounded-2xl border border-slate-200 bg-white p-4">
                                        <div class="mb-4 flex flex-wrap gap-4 text-xs font-bold text-slate-600"><span>■ Pemasukan</span><span>▥ Realisasi</span><span>▤ Balance</span></div>
                                        <div class="flex min-w-[650px] items-end gap-3 overflow-x-auto border-b border-slate-300 pb-2 pt-5">
                                            {#each data.monthly as row}
                                                <div class="flex min-w-12 flex-1 flex-col items-center gap-2">
                                                    <div class="flex h-52 w-full items-end justify-center gap-1">
                                                        <div class="w-2.5 rounded-t bg-slate-900" style={`height:${chartPercent(row.income, data)}%`} title={`Pemasukan ${formatRupiah(row.income)}`}></div>
                                                        <div class="w-2.5 rounded-t bg-cyan-500" style={`height:${chartPercent(row.realization, data)}%`} title={`Realisasi ${formatRupiah(row.realization)}`}></div>
                                                        <div class="w-2.5 rounded-t bg-emerald-500" style={`height:${chartPercent(Math.max(0, row.balance), data)}%`} title={`Balance ${formatRupiah(row.balance)}`}></div>
                                                    </div>
                                                    <div class="text-[10px] font-bold text-slate-500">{row.monthName.slice(0, 3)}</div>
                                                </div>
                                            {/each}
                                        </div>
                                    </div>
                                {:else if intent === 'luar_kota_analysis'}
                                    {@const data = asOutside(item.response.data)}
                                    <div class="mt-3 space-y-3">
                                        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"><div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Total Perjalanan</div><div class="mt-1 text-xl font-black">{formatNumber(data.summary.totalTrips)}</div></div><div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Telah Dilaksanakan</div><div class="mt-1 text-xl font-black">{formatNumber(data.summary.executedTrips)}</div></div><div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Biaya Terlaksana</div><div class="mt-1 text-xl font-black">{formatRupiah(data.summary.executedCost)}</div></div><div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">User Protokol</div><div class="mt-1 text-xl font-black">{formatNumber(data.summary.protocolUsers)}</div></div></div>
                                        <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white"><div class="overflow-x-auto"><table class="min-w-full text-left text-xs"><thead class="bg-slate-50 text-slate-500"><tr><th class="px-3 py-3">Rank</th><th class="px-3 py-3">User Protokol</th><th class="px-3 py-3 text-right">Perjalanan</th><th class="px-3 py-3 text-right">Terlaksana</th><th class="px-3 py-3 text-right">Biaya</th><th class="px-3 py-3 text-right">Assigned</th><th class="px-3 py-3 text-right">In Progress</th><th class="px-3 py-3 text-right">Completed</th><th class="px-3 py-3 text-right">Rejected</th></tr></thead><tbody>{#each data.users as row}<tr class="border-t border-slate-100"><td class="px-3 py-3 font-black">{row.rank}</td><td class="px-3 py-3"><div class="font-bold">{row.name}</div><div class="text-slate-500">{row.position || row.nip || '-'}</div></td><td class="px-3 py-3 text-right">{formatNumber(row.totalTrips)}</td><td class="px-3 py-3 text-right">{formatNumber(row.executedTrips)}</td><td class="px-3 py-3 text-right font-bold">{formatRupiah(row.executedCost)}</td><td class="px-3 py-3 text-right">{row.statuses.assigned}</td><td class="px-3 py-3 text-right">{row.statuses.inProgress}</td><td class="px-3 py-3 text-right">{row.statuses.completed}</td><td class="px-3 py-3 text-right">{row.statuses.rejected}</td></tr>{/each}</tbody></table></div></div>
                                    </div>
                                {:else if intent === 'dalam_kota_analysis'}
                                    {@const data = asInside(item.response.data)}
                                    <div class="mt-3 space-y-3">
                                        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"><div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Total Kegiatan</div><div class="mt-1 text-xl font-black">{formatNumber(data.summary.totalActivities)}</div></div><div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Terlaksana</div><div class="mt-1 text-xl font-black">{formatNumber(data.summary.executedActivities)}</div></div><div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Total SPJ</div><div class="mt-1 text-xl font-black">{formatRupiah(data.summary.totalSpj)}</div></div><div class="rounded-2xl border border-slate-200 bg-white p-4"><div class="text-xs font-bold text-slate-500">Total RIIL</div><div class="mt-1 text-xl font-black">{formatRupiah(data.summary.totalRiil)}</div></div></div>
                                        <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white"><div class="overflow-x-auto"><table class="min-w-full text-left text-xs"><thead class="bg-slate-50 text-slate-500"><tr><th class="px-3 py-3">Rank</th><th class="px-3 py-3">User Protokol</th><th class="px-3 py-3 text-right">Kegiatan</th><th class="px-3 py-3 text-right">SPJ</th><th class="px-3 py-3 text-right">RIIL</th><th class="px-3 py-3 text-right">Balance</th><th class="px-3 py-3 text-right">Draft</th><th class="px-3 py-3 text-right">Pending</th><th class="px-3 py-3 text-right">Sukses</th></tr></thead><tbody>{#each data.users as row}<tr class="border-t border-slate-100"><td class="px-3 py-3 font-black">{row.rank}</td><td class="px-3 py-3"><div class="font-bold">{row.name}</div><div class="text-slate-500">{row.position || row.nip || '-'}</div></td><td class="px-3 py-3 text-right">{formatNumber(row.totalActivities)}</td><td class="px-3 py-3 text-right">{formatRupiah(row.totalSpj)}</td><td class="px-3 py-3 text-right font-bold">{formatRupiah(row.totalRiil)}</td><td class="px-3 py-3 text-right">{formatRupiah(row.balance)}</td><td class="px-3 py-3 text-right">{row.statuses.draft}</td><td class="px-3 py-3 text-right">{row.statuses.pending}</td><td class="px-3 py-3 text-right">{row.statuses.success}</td></tr>{/each}</tbody></table></div></div>
                                    </div>
                                {/if}
                            {/if}
                        </div>
                    </div>
                {/each}

                {#if messages.length === 1}
                    <div class="ml-12 flex flex-wrap gap-2">
                        {#each initialCommands as command}
                            <button type="button" class="rounded-full border border-slate-300 bg-white px-3 py-2 text-xs font-bold text-slate-700 hover:border-cyan-400 hover:text-cyan-700" onclick={() => sendCommand(command)}>{command}</button>
                        {/each}
                    </div>
                {/if}

                {#if loading}
                    <div class="flex items-start gap-3">
                        <div class="flex h-9 w-9 items-center justify-center rounded-full bg-cyan-100 text-xs font-black text-cyan-800">AI</div>
                        <div class="rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-500 shadow-sm">
                            <span class="animate-pulse">Menganalisis perintah dan data...</span>
                        </div>
                    </div>
                {/if}
            </div>
        </div>

        <footer class="border-t border-slate-200 bg-white p-3 md:p-4">
            <div class="mx-auto flex max-w-5xl items-end gap-2 rounded-2xl border border-slate-300 bg-white p-2 focus-within:border-cyan-500 focus-within:ring-2 focus-within:ring-cyan-100">
                <textarea
                    rows="1"
                    bind:value={input}
                    onkeydown={handleKeydown}
                    disabled={loading}
                    placeholder="Ketik perintah, contoh: Analisa Keuangan GUP Agustus 2026"
                    class="max-h-32 min-h-11 flex-1 resize-none border-0 bg-transparent px-3 py-2.5 text-sm text-slate-900 outline-none placeholder:text-slate-400 disabled:opacity-60"
                ></textarea>
                <button
                    type="button"
                    onclick={() => sendCommand()}
                    disabled={loading || !input.trim()}
                    class="h-11 rounded-xl bg-slate-950 px-5 text-sm font-black text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-40"
                >
                    Kirim
                </button>
            </div>
            <div class="mx-auto mt-2 max-w-5xl px-1 text-[11px] text-slate-400">Enter untuk mengirim. Shift + Enter untuk baris baru.</div>
        </footer>
    </section>
</div>
