// Shared Types for API Interaction

export interface User {
    id: string | number;
    email: string;
    name: string;
    role: 'super_admin' | 'kasubag' | 'protokol';
    token?: string; // JWT for real API
    nip?: string;
    pangkat?: string;
    golongan?: string;
    jabatan?: string;
    tingkatBiaya?: string;
    nomorHp?: string;
}

export interface TravelCost {
    ticketGo: number;
    ticketBack: number;
    dailyAllowanceDays?: number;
    dailyAllowanceRate?: number;
    hotelDays?: number;
    hotelRate?: number;
    localTransport?: number;
    regionalTransport?: number;
    transportMode: string;
    transportAmount?: number;
    otherCost?: number;
    otherCostDesc?: string;
    additionalCosts?: { name: string; amount: number; file?: any }[];
    boardingPassFiles?: any[];
    details?: TravelCost[]; // Per-location cost breakdown
}

export interface TravelReport {
    text: string;
    submittedAt?: string;
    files: any[];
    sppdFile?: any;
    suratTugasFile?: any;
    tanggalMerah?: string | string[];
    lastDraftSavedAt?: string;
}

export interface TravelLocation {
    id: string;
    travelRecordId: string;
    location: string;
    province: string;
    startDate: string;
    endDate: string;
}

export interface TravelRecord {
    id: string;
    spd: string;
    employee: User;
    employeeId?: string;
    creator?: User;
    creatorId?: string;
    email: string; // Creator email
    startDate: string;
    endDate: string;
    createdAt: string;
    suratTugasPath?: string;
    suratTugasNumber?: string;
    suratTugasDate?: string;
    isViewed?: boolean;
    location: string;
    province: string;
    locations?: TravelLocation[];
    type?: string;
    purpose: string;
    stakeholder: string;
    agenda: string;
    status: 'Draft' | 'Submitted' | 'Approved' | 'Rejected';
    reportStatus: 'Pending' | 'Completed' | 'Draft';
    budgetSource?: 'GUP' | 'LS' | '';
    financeSourceId?: string;
    totalCost: number;
    costs?: TravelCost;
    reportData?: TravelReport | null;
}

export interface ImportDetail {
    row: number;
    spjId: string;
    name: string;
    status: 'imported' | 'skipped_duplicate' | 'skipped_thr' | 'skipped_no_user' | 'failed';
    message?: string;
}

export interface ImportResult {
    totalRows: number;
    imported: number;
    skipped: number;
    failed: number;
    details: ImportDetail[];
}

export interface DashboardBudget {
    year: number;
    month: number;
    total: number;
}

export interface DashboardSummary {
    totalTrips: number;
    activeTrips: number;
    statusCompleted: number;
    statusInProgress: number;
    statusAssigned: number;
    statusRejected: number;
    reportCompleted: number;
    reportPending: number;
    recentRecords: TravelRecord[];
    budgets: DashboardBudget[];
}

export interface PaginatedParams {
    cursor?: string;
    limit?: number;
    search?: string;
    status?: string;
    report_status?: string;
    payment_status?: string;
    sort_by?: string;
    start_date?: string; // ISO string
    end_date?: string;   // ISO string
    user_id?: string;
}

export interface PaginatedResponse {
    data: TravelRecord[];
    totalItems: number;
    totalRecords: number;
    nextCursor: string;
    limit: number;
}

export interface ApiClient {
    // Auth
    login(email: string, password?: string): Promise<{ user: User; token?: string; require_password_change?: boolean }>;
    register(email: string, password: string, name: string, role?: string): Promise<User>;
    logout(): Promise<void>;
    getCurrentUser(): Promise<User | null>;
    changePassword(newPassword: string): Promise<void>;

    // User Management (Admin)
    getUsers(filters?: { role?: string; search?: string }): Promise<User[]>;
    createUser(user: Omit<User, 'id'>): Promise<User>;
    updateUser(id: string | number, user: Partial<User>): Promise<User>;
    deleteUser(id: string | number): Promise<void>;

    // Files
    uploadFile(file: File): Promise<{path: string}>;

    // Records
    getDashboardSummary(): Promise<DashboardSummary>;
    getRecords(filters?: Record<string, any>): Promise<TravelRecord[]>;
    getPaginatedRecords(params: PaginatedParams): Promise<PaginatedResponse>;
    getRecordById(id: string): Promise<TravelRecord | null>;
    exportSpdPdf(id: string): Promise<Blob>;
    exportSpdDocx(id: string): Promise<Blob>;
    exportLaporanPdf(id: string): Promise<Blob>;
    exportLaporanDocx(id: string): Promise<Blob>;
    exportRincianPdf(id: string): Promise<Blob>;
    exportRincianDocx(id: string): Promise<Blob>;
    createRecord(record: Omit<TravelRecord, 'id' | 'spd' | 'status' | 'reportStatus'>): Promise<TravelRecord[]>; // Returns array because one request can create multiple records (bulk)
    updateRecord(id: string, record: Partial<TravelRecord>): Promise<TravelRecord>;
    deleteRecord(id: string): Promise<void>;
    deleteRecordsBySpd(spd: string): Promise<void>;

    // Import
    importExcel(file: File): Promise<ImportResult>;

    // Master Data
    getProvinces(): Promise<any[]>;
    getSBMRates(): Promise<any[]>;
    getSettings(): Promise<any>;
    updateSettings(settings: any): Promise<void>;

        // GUP
    getGupTransactions(): Promise<GupTransaction[]>;

    getGupTransactionById(
        id: string
    ): Promise<GupTransaction>;

    getGupMaster(
        year?: number
    ): Promise<GupMasterData>;

    createGup(
        formData: FormData
    ): Promise<GupTransaction>;

    updateGup(
        id: string,
        formData: FormData
    ): Promise<GupTransaction>;

    deleteGup(
        id: string
    ): Promise<void>;

    // Keuangan - Sumber Dana
    getFinanceSources(year?: number): Promise<FinanceDashboard>;
    getFinanceSourceOptions(year: number, budgetType: 'GUP' | 'LS', spdNumber?: string): Promise<FinanceSourceOption[]>;
    createFinanceSource(payload: SaveFinanceSourcePayload): Promise<FinanceSource>;
    updateFinanceSource(id: string, payload: SaveFinanceSourcePayload): Promise<FinanceSource>;
    deleteFinanceSource(id: string): Promise<void>;
    createFinanceAllocation(payload: SaveFinanceAllocationPayload): Promise<FinanceAllocation>;
    deleteFinanceAllocation(id: string): Promise<void>;

    // Dalkot
    getDalkotMaster(): Promise<DalkotMasterData>;
    getDalkotList(): Promise<DalkotListResponse>;
    getDalkotRecap(params: DalkotRecapParams): Promise<DalkotRecapResponse>;
    getDalkotMonthlySPJRecap(params: DalkotPeriodParams): Promise<DalkotMonthlySPJRecapResponse>;
    getDalkotById(id: string): Promise<DalkotRecord>;
    createDalkot(formData: FormData): Promise<DalkotRecord>;
    updateDalkot(id: string, formData: FormData): Promise<DalkotRecord>;
    deleteDalkot(id: string): Promise<void>;

    // Chatbot Analytics
    getChatbotGup(params?: { year?: number; month?: number }): Promise<ChatbotGupAnalysis>;
    getChatbotOutsideCity(params?: { year?: number; month?: number }): Promise<ChatbotOutsideCityAnalysis>;
    getChatbotInsideCity(params?: { year?: number; month?: number }): Promise<ChatbotInsideCityAnalysis>;
    sendChatbotMessage(message: string): Promise<ChatbotMessageResponse>;
}

export interface GupProcurementType {
    id: string;
    name: string;
    accountCode: string;
    mak: string;
}

export interface GupFundingSource {
    id: string;
    year: number;
    monthNumber: number;
    monthName: string;
    gupLabel: string;
}

export interface GupMasterData {
    procurementTypes: GupProcurementType[];
    fundingSources: GupFundingSource[];
}

export interface GupTransaction {
    id: string;
    businessId: string;
    paymentDescription: string;
    procurementTypeId: string;
    procurementTypeName: string;
    accountCode: string;
    mak: string;
    fundingSourceId?: string | null;
    fundingSourceLabel: string;
    valueAmount: number;
    paidAmount: number;
    taxAmount: number;
    differenceAmount: number;
    receiptDate: string;
    recipient: string;
    pum: string;
    documentOriginalName?: string;
    documentMime?: string;
    documentSize?: number;
    documentUrl?: string;
    createdAt: string;
    updatedAt: string;
}

export interface FinanceSource {
    id: string;
    year: number;
    code: string;
    name: string;
    amount: number;
    allocatedGup: number;
    allocatedLs: number;
    allocatedTotal: number;
    usedGup: number;
    usedLs: number;
    usedTotal: number;
    remainingGup: number;
    remainingLs: number;
    remainingTotal: number;
    available: number;
    notes: string;
    isActive: boolean;
    createdAt: string;
    updatedAt: string;
}

export interface FinanceAllocation {
    id: string;
    financeSourceId: string;
    sourceCode: string;
    sourceName: string;
    allocationType: 'GUP' | 'LS';
    amount: number;
    notes: string;
    createdAt: string;
    updatedAt: string;
}

export interface FinanceSummary {
    totalFunds: number;
    totalGup: number;
    totalLs: number;
    totalUsedGup: number;
    totalUsedLs: number;
    totalUsed: number;
    totalAllocated: number;
    totalAvailable: number;
    totalRemaining: number;
}

export interface FinanceSourceOption {
    id: string;
    year: number;
    code: string;
    name: string;
    budgetType: 'GUP' | 'LS';
    allocated: number;
    usedOther: number;
    available: number;
}

export interface FinanceDashboard {
    year: number;
    sources: FinanceSource[];
    allocations: FinanceAllocation[];
    summary: FinanceSummary;
}

export interface SaveFinanceSourcePayload {
    year: number;
    code: string;
    name: string;
    amount: number;
    notes: string;
    isActive: boolean;
}

export interface SaveFinanceAllocationPayload {
    financeSourceId: string;
    allocationType: 'GUP' | 'LS';
    amount: number;
    notes: string;
}

export type DalkotCategory =
    | 'Jam Kerja'
    | 'Overtime'
    | 'Hari Libur';

export type DalkotOfficerGroup =
    | 'SPJ'
    | 'RIIL';

export type DalkotStatus =
    | 'Draft'
    | 'Pending'
    | 'Sukses';

export interface DalkotMasterOption {
    value: string;
    label: string;
    amount?: number;
}

export interface DalkotUserOption {
    id: string;
    name: string;
    email: string;
    nip: string;
    position: string;
    grade: string;
    role: string;
    isActive: boolean;
}

export interface DalkotMasterData {
    users: DalkotUserOption[];
    categories: DalkotMasterOption[];
    officials: DalkotMasterOption[];
    locations: DalkotMasterOption[];
    spjAmount: number;
    jamKerjaAmount: number;
    overtimeAmount: number;
    hariLiburAmount: number;
}

export interface DalkotOfficer {
    id: string;
    dalkotId: string;
    employeeId?: string | null;
    officerGroup: DalkotOfficerGroup;
    name: string;
    nip: string;
    position: string;
    grade: string;
    amount: number;
    status: DalkotStatus;
    sortOrder: number;
    createdAt: string;
}

export interface DalkotRecord {
    id: string;
    businessId: string;
    serviceDate: string;
    category: DalkotCategory;
    accompaniedOfficial: string;
    dalkotType: 'SPJ_RILL' | 'KEBIJAKAN_PROTOKOL';
    assignmentLetterNumber: string;
    assignmentLetterDate?: string | null;
    activityName: string;
    location: string;
    reportContent: string;
    status: DalkotStatus;
    totalSPJAmount: number;
    totalRiilAmount: number;
    documentFilename?: string;
    documentOriginalName?: string;
    documentMime?: string;
    documentSize?: number;
    documentUrl?: string;
    officerCount: number;
    officers: DalkotOfficer[];
    createdAt: string;
    updatedAt: string;
}

export interface DalkotSummary {
    totalRecords: number;
    totalOfficers: number;
    totalSPJAmount: number;
    totalRiilAmount: number;
}

export interface DalkotListResponse {
    rows: DalkotRecord[];
    summary: DalkotSummary;
}

export interface DalkotPeriodParams {
    month: number;
    year: number;
}

export interface DalkotRecapParams extends DalkotPeriodParams {
    officerId?: string;
}

export interface DalkotRecapRow {
    businessId: string;
    serviceDate: string;
    accompaniedOfficial: string;
    activityName: string;
    officerName: string;
    category: DalkotCategory;
    dalkotType: 'SPJ_RILL' | 'KEBIJAKAN_PROTOKOL';
    nominal: number;
}

export interface DalkotRecapSummary {
    totalRows: number;
    totalNominal: number;
    totalSPJAmount: number;
    totalRiilAmount: number;
    balance: number;
    balanceMinus: boolean;
}

export interface DalkotRecapResponse {
    month: number;
    year: number;
    officerId?: string | null;
    officerName: string;
    rows: DalkotRecapRow[];
    summary: DalkotRecapSummary;
}

export interface DalkotMonthlySPJRecapRow {
    officerId?: string | null;
    name: string;
    grade: string;
    dates: string[];
    rate: number;
    totalDays: number;
    totalAmount: number;
    assignmentLetterNumbers: string[];
    assignmentLetterDates: string[];
}

export interface DalkotMonthlySPJRecapSummary {
    totalOfficers: number;
    totalDays: number;
    totalAmount: number;
}

export interface DalkotMonthlySPJRecapResponse {
    month: number;
    year: number;
    account: string;
    title: string;
    rows: DalkotMonthlySPJRecapRow[];
    summary: DalkotMonthlySPJRecapSummary;
}


// ==========================================
// Chatbot Financial Analytics
// ==========================================

export interface ChatbotPeriod {
    year: number;
    month: number;
    label: string;
}

export interface ChatbotGupSummary {
    totalLs: number;
    totalRealization: number;
    balance: number;
    realizationPercentage: number;
    transactionCount: number;
    totalTax: number;
}

export interface ChatbotGupExpenseRank {
    rank: number;
    businessId: string;
    paymentDescription: string;
    procurementTypeName: string;
    accountCode: string;
    mak: string;
    fundingSourceLabel: string;
    recipient: string;
    receiptDate: string;
    paidAmount: number;
}

export interface ChatbotGupAccountRank {
    rank: number;
    accountCode: string;
    mak: string;
    paidAmount: number;
    transactionCount: number;
    percentage: number;
}

export interface ChatbotGupMonthlyRow {
    monthNumber: number;
    monthName: string;
    gupLabel: string;
    income: number;
    realization: number;
    balance: number;
    transactionCount: number;
    status: string;
}

export interface ChatbotGupAnalysis {
    period: ChatbotPeriod;
    summary: ChatbotGupSummary;
    topExpenses: ChatbotGupExpenseRank[];
    lowestExpenses: ChatbotGupExpenseRank[];
    accountRanking: ChatbotGupAccountRank[];
    largestAccount?: ChatbotGupAccountRank;
    monthly: ChatbotGupMonthlyRow[];
}

export interface ChatbotTripStatusCount {
    assigned: number;
    inProgress: number;
    completed: number;
    rejected: number;
}

export interface ChatbotOutsideCityUserRow {
    userId: string;
    name: string;
    email: string;
    nip: string;
    position: string;
    totalTrips: number;
    executedTrips: number;
    executedCost: number;
    statuses: ChatbotTripStatusCount;
    rank: number;
}

export interface ChatbotOutsideCityAnalysis {
    period: ChatbotPeriod;
    summary: {
        protocolUsers: number;
        totalTrips: number;
        executedTrips: number;
        executedCost: number;
        statuses: ChatbotTripStatusCount;
    };
    users: ChatbotOutsideCityUserRow[];
}

export interface ChatbotDalkotStatusCount {
    draft: number;
    pending: number;
    success: number;
}

export interface ChatbotInsideCityUserRow {
    userId: string;
    name: string;
    email: string;
    nip: string;
    position: string;
    totalActivities: number;
    executedActivities: number;
    totalSpj: number;
    totalRiil: number;
    balance: number;
    statuses: ChatbotDalkotStatusCount;
    rank: number;
}

export interface ChatbotInsideCityAnalysis {
    period: ChatbotPeriod;
    summary: {
        protocolUsers: number;
        totalActivities: number;
        executedActivities: number;
        totalSpj: number;
        totalRiil: number;
        balance: number;
        statuses: ChatbotDalkotStatusCount;
    };
    users: ChatbotInsideCityUserRow[];
}

export type ChatbotIntent =
    | 'menu_gup'
    | 'menu_luar_kota'
    | 'menu_dalam_kota'
    | 'gup_analysis'
    | 'gup_ranking'
    | 'gup_ls'
    | 'gup_account'
    | 'gup_chart'
    | 'luar_kota_analysis'
    | 'dalam_kota_analysis'
    | 'unknown'
    | 'empty';

export interface ChatbotMessageResponse {
    intent: ChatbotIntent | string;
    message: string;
    data?: ChatbotGupAnalysis | ChatbotOutsideCityAnalysis | ChatbotInsideCityAnalysis;
    suggestions?: string[];
}
