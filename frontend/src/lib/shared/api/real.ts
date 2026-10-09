import type {
    ApiClient,
    DalkotListResponse,
    DalkotMasterData,
    DalkotMonthlySPJRecapResponse,
    DalkotPeriodParams,
    DalkotRecapParams,
    DalkotRecapResponse,
    DalkotRecord,
    TravelRecord,
    User
} from './types';

// Hardcode BASE_URL to always use the proxy to prevent any CORS or Direct-Backend 0 byte body issues
const BASE_URL = '/api/v1';

export class RealApiClient implements ApiClient {
    private token: string | null = null;
    private unauthorizedHandler: ((reason?: string) => void) | null = null;

    constructor() {
        if (typeof window !== 'undefined') {
            this.token = localStorage.getItem('auth_token');
        }
    }

    setUnauthorizedHandler(handler: (reason?: string) => void) {
        this.unauthorizedHandler = handler;
    }

    private async request<T>(endpoint: string, options: RequestInit = {}, customFetch?: typeof fetch): Promise<T> {
        const headers: HeadersInit = {
            'Content-Type': 'application/json',
            ...(this.token ? { 'Authorization': `Bearer ${this.token}` } : {})
        };

        const fetcher = customFetch || (typeof window !== 'undefined' ? window.fetch : fetch);
        const response = await fetcher(`${BASE_URL}${endpoint}`, {
            ...options,
            headers: {
                ...headers,
                ...options.headers
            }
        });

        if (response.status === 204) {
            return null as unknown as T;
        }

        if (!response.ok) {
            const errorText = await response.text();
            let errorMessage = `API Error: ${response.status}`;
            try {
                const errorJson = JSON.parse(errorText);
                if (errorJson.message) errorMessage = errorJson.message;
            } catch (e) {
                // ignore
            }

            if (response.status === 401) {
                if (this.unauthorizedHandler && endpoint !== '/auth/login') {
                    this.unauthorizedHandler(errorMessage);
                }
                throw new Error(errorMessage);
            }
            
            throw new Error(errorMessage);
        }

        return response.json();
    }

    private async multipartRequest<T>(
        endpoint: string,
        method: 'POST' | 'PUT',
        formData: FormData
    ): Promise<T> {
        const headers: HeadersInit = {
            ...(this.token
                ? {
                      Authorization: `Bearer ${this.token}`
                  }
                : {})
        };

        const response = await fetch(
            `${BASE_URL}${endpoint}`,
            {
                method,
                headers,
                body: formData
            }
        );

        if (response.status === 204) {
            return null as T;
        }

        if (!response.ok) {
            const errorText = await response.text();

            let errorMessage =
                `API Error: ${response.status}`;

            try {
                const errorJson = JSON.parse(errorText);

                if (errorJson.message) {
                    errorMessage = errorJson.message;
                }
            } catch {
                if (errorText) {
                    errorMessage = errorText;
                }
            }

            if (
                response.status === 401 &&
                this.unauthorizedHandler
            ) {
                this.unauthorizedHandler(
                    errorMessage
                );
            }

            throw new Error(errorMessage);
        }

        return response.json();
    }

    // --- Auth ---
    async login(email: string, password?: string): Promise<{ user: User; token?: string; require_password_change?: boolean }> {
        const res = await this.request<{ user: User; token: string; require_password_change: boolean }>('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });
        
        this.token = res.token;
        if (typeof window !== 'undefined') localStorage.setItem('auth_token', res.token);
        
        return res;
    }

    async register(email: string, password: string, name: string, role?: string): Promise<User> {
        return this.request<User>('/auth/register', {
            method: 'POST',
            body: JSON.stringify({ email, password, name, role })
        });
    }

    async logout(): Promise<void> {
        this.token = null;
        if (typeof window !== 'undefined') localStorage.removeItem('auth_token');
        // Optional: Call backend logout endpoint if it exists
    }

    async getCurrentUser(): Promise<User | null> {
        if (!this.token) return null;
        try {
            return await this.request<User>('/auth/me');
        } catch (e) {
            return null;
        }
    }

    async changePassword(newPassword: string): Promise<void> {
        return this.request<void>('/auth/change-password', {
            method: 'PUT',
            body: JSON.stringify({ new_password: newPassword })
        });
    }

    // --- User Management ---

    async getUsers(filters?: { role?: string; search?: string }): Promise<User[]> {
        let query = '';
        if (filters) {
            const params = new URLSearchParams();
            if (filters.role) params.append('role', filters.role);
            if (filters.search) params.append('search', filters.search);
            query = '?' + params.toString();
        }
        return this.request<User[]>(`/users${query}`);
    }

    async createUser(user: Omit<User, 'id'>): Promise<User> {
        return this.request<User>('/users', {
            method: 'POST',
            body: JSON.stringify(user)
        });
    }

    async updateUser(id: string | number, user: Partial<User>): Promise<User> {
        return this.request<User>(`/users/${id}`, {
            method: 'PUT',
            body: JSON.stringify(user)
        });
    }

    async deleteUser(id: string | number): Promise<void> {
        return this.request<void>(`/users/${id}`, {
            method: 'DELETE'
        });
    }

    // --- Files ---
    async uploadFile(file: File): Promise<{path: string}> {
        const formData = new FormData();
        const filename = file.name || 'uploaded_file.pdf';
        formData.append('file', file, filename);

        const headers: HeadersInit = {
            ...(this.token ? { 'Authorization': `Bearer ${this.token}` } : {})
        };

        const response = await fetch(`${BASE_URL}/upload`, {
            method: 'POST',
            headers,
            body: formData
        });

        if (!response.ok) {
            if (response.status === 401 && this.unauthorizedHandler) {
                this.unauthorizedHandler();
            }
            throw new Error(`Upload failed: ${response.statusText}`);
        }

        return response.json();
    }

    // --- Records ---
    async getDashboardSummary(customFetch?: typeof fetch): Promise<import('./types').DashboardSummary> {
        return this.request<import('./types').DashboardSummary>('/dashboard/summary', {}, customFetch);
    }

    async getRecords(filters?: Record<string, any>): Promise<TravelRecord[]> {
        const query = filters ? '?' + new URLSearchParams(filters).toString() : '';
        return this.request<TravelRecord[]>(`/records${query}`);
    }

    async getPaginatedRecords(params: import('./types').PaginatedParams, customFetch?: typeof fetch): Promise<import('./types').PaginatedResponse> {
        // Remove undefined values
        const cleanParams = Object.fromEntries(Object.entries(params).filter(([_, v]) => v !== undefined && v !== ''));
        const query = new URLSearchParams(cleanParams as any).toString();
        return this.request<import('./types').PaginatedResponse>(`/records/paginated?${query}`, {}, customFetch);
    }

    async getRecordById(id: string): Promise<TravelRecord | null> {
        return this.request<TravelRecord>(`/records/${id}`);
    }

    private async fetchDocument(id: string, type: string, format: 'pdf' | 'docx' = 'pdf'): Promise<Blob> {
        const headers: HeadersInit = {
            ...(this.token ? { 'Authorization': `Bearer ${this.token}` } : {})
        };

        const formatPath = format === 'pdf' ? 'stream' : 'docx';
        const response = await fetch(`${BASE_URL}/records/${id}/${type}-${formatPath}`, {
            method: 'GET',
            headers
        });


        if (!response.ok) {
            if (response.status === 401 && this.unauthorizedHandler) {
                this.unauthorizedHandler();
            }
            const errorText = await response.text();
            let errorMessage = response.statusText;
            try {
                const errorJson = JSON.parse(errorText);
                if (errorJson.message) errorMessage = errorJson.message;
            } catch (e) {
                if (errorText) errorMessage = errorText;
            }
            throw new Error(`Export ${type.toUpperCase()} ${format.toUpperCase()} failed: ${errorMessage}`);
        }

        const buffer = await response.arrayBuffer();
        if (buffer.byteLength === 0) {
            throw new Error(`Data dokumen ${format.toUpperCase()} kosong (0 byte) diterima dari server untuk tipe ${type}.`);
        }
        
        const mimeType = format === 'pdf' 
            ? 'application/pdf' 
            : 'application/vnd.openxmlformats-officedocument.wordprocessingml.document';
            
        return new Blob([buffer], { type: mimeType });
    }

    async exportSpdPdf(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'spd', 'pdf');
    }

    async exportSpdDocx(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'spd', 'docx');
    }

    async exportLaporanPdf(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'laporan', 'pdf');
    }

    async exportLaporanDocx(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'laporan', 'docx');
    }

    async exportRincianPdf(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'rincian', 'pdf');
    }

    async exportRincianDocx(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'rincian', 'docx');
    }

    async createRecord(record: any): Promise<TravelRecord[]> {
         const employees = record.employees || [record.employee];
         const createdRecords: TravelRecord[] = [];

         const safeDate = (d: any) => {
             if (!d) return undefined;
             const date = new Date(d);
             return isNaN(date.getTime()) ? undefined : date.toISOString();
         };

         // Map locations to have ISO dates
         const locations = (record.locations || []).map((loc: any) => ({
             ...loc,
             startDate: safeDate(loc.startDate),
             endDate: safeDate(loc.endDate)
         }));

         const employeeIds = employees.map((e: any) => e.id);
         const payload = {
             ...record,
             employeeIds: employeeIds,
             employee: undefined,
             employees: undefined,
             startDate: safeDate(record.startDate),
             endDate: safeDate(record.endDate),
             locations: locations
         };

         // The backend CreateRecordsBulk now returns the created records.
         try {
             const res = await this.request<TravelRecord[]>('/records', {
                 method: 'POST',
                 body: JSON.stringify(payload)
             });
             
             // Attach employee data manually from the original selection so the UI has all info immediately
             if (Array.isArray(res)) {
                 return res.map(r => {
                     const emp = employees.find((e: any) => e.id === r.employeeId);
                     return { ...r, employee: emp };
                 });
             }
             return res as any;
         } catch (e: any) {
             throw e;
         }
    }
    async updateRecord(id: string, record: Partial<TravelRecord>): Promise<TravelRecord> {
        const safeDate = (d: any) => {
            if (!d) return undefined;
            const date = new Date(d);
            return isNaN(date.getTime()) ? undefined : date.toISOString();
        };

        const payload: any = {
            ...record,
        };
        if (record.startDate !== undefined) payload.startDate = safeDate(record.startDate);
        if (record.endDate !== undefined) payload.endDate = safeDate(record.endDate);
        if (record.locations !== undefined) {
            payload.locations = record.locations.map((loc: any) => ({
                ...loc,
                startDate: safeDate(loc.startDate),
                endDate: safeDate(loc.endDate)
            }));
        }

        return this.request<TravelRecord>(`/records/${id}`, {
            method: 'PUT',
            body: JSON.stringify(payload)
        });
    }

    deleteRecord(id: string): Promise<void> {
        return this.request<void>(`/records/${id}`, {
            method: 'DELETE'
        });
    }

    deleteRecordsBySpd(spd: string): Promise<void> {
        return this.request<void>(`/records/spd/${spd}`, {
            method: 'DELETE'
        });
    }

    async importExcel(file: File): Promise<import('./types').ImportResult> {
        const formData = new FormData();
        formData.append('file', file);

        const headers: HeadersInit = {
            ...(this.token ? { 'Authorization': `Bearer ${this.token}` } : {})
        };

        const response = await fetch(`${BASE_URL}/records/import`, {
            method: 'POST',
            headers,
            body: formData
        });

        if (!response.ok) {
            if (response.status === 401 && this.unauthorizedHandler) {
                this.unauthorizedHandler();
            }
            const errorText = await response.text();
            let errorMessage = `Import gagal: ${response.status}`;
            try {
                const errorJson = JSON.parse(errorText);
                if (errorJson.message) errorMessage = errorJson.message;
            } catch (e) { /* ignore */ }
            throw new Error(errorMessage);
        }

        return response.json();
    }

    // Master Data
    getProvinces(): Promise<any[]> {
        return this.request<any[]>('/master/provinces');
    }

    getSBMRates(): Promise<any[]> {
        return this.request<any[]>('/master/sbm-rates');
    }

    getSettings(): Promise<any> {
        return this.request<any>('/master/settings');
    }

    updateSettings(settings: any): Promise<void> {
        return this.request<void>('/master/settings', {
            method: 'PUT',
            body: JSON.stringify(settings)
        });
    }

    // GUP

async getGupTransactions(): Promise<
    import('./types').GupTransaction[]
> {
    return this.request<
        import('./types').GupTransaction[]
    >('/gup');
}

async getGupTransactionById(
    id: string
): Promise<import('./types').GupTransaction> {
    return this.request<
        import('./types').GupTransaction
    >(`/gup/${encodeURIComponent(id)}`);
}

async getGupMaster(
    year: number = new Date().getFullYear()
): Promise<import('./types').GupMasterData> {
    return this.request<
        import('./types').GupMasterData
    >(`/gup/master?year=${encodeURIComponent(year)}`);
}

async createGup(
    formData: FormData
): Promise<import('./types').GupTransaction> {
    return this.multipartRequest<
        import('./types').GupTransaction
    >(
        '/gup',
        'POST',
        formData
    );
}

async updateGup(
    id: string,
    formData: FormData
): Promise<import('./types').GupTransaction> {
    return this.multipartRequest<
        import('./types').GupTransaction
    >(
        `/gup/${encodeURIComponent(id)}`,
        'PUT',
        formData
    );
}

async deleteGup(
    id: string
): Promise<void> {
    await this.request<{ message: string }>(
        `/gup/${encodeURIComponent(id)}`,
        {
            method: 'DELETE'
        }
    );
}

async getGupReport(
    year: number = new Date().getFullYear()
): Promise<any> {
    return this.request<any>(
        `/gup-report?year=${encodeURIComponent(
            String(year)
        )}`
    );
}

async updateGupReportBudget(payload: {
    year: number;
    procurementTypeId: string;
    amount: number;
}): Promise<any> {
    return this.request<any>(
        '/gup-report/budget',
        {
            method: 'PUT',
            body: JSON.stringify(payload)
        }
    );
}

async getGupLs(
    year: number = new Date().getFullYear()
): Promise<any> {
    return this.request<any>(
        `/gup-ls?year=${encodeURIComponent(String(year))}`
    );
}

async updateGupLs(payload: {
    year: number;
    fundingSourceId: string;
    amount: number;
}): Promise<any> {
    return this.request<any>(
        '/gup-ls',
        {
            method: 'PUT',
            body: JSON.stringify(payload)
        }
    );
}

async getFinanceSources(
    year: number = new Date().getFullYear()
): Promise<import('./types').FinanceDashboard> {
    return this.request<import('./types').FinanceDashboard>(
        `/finance-sources?year=${encodeURIComponent(String(year))}`
    );
}

async getFinanceSourceOptions(
    year: number,
    budgetType: 'GUP' | 'LS',
    spdNumber: string = ''
): Promise<import('./types').FinanceSourceOption[]> {
    const params = new URLSearchParams();
    params.set('year', String(year));
    params.set('budgetType', budgetType);
    if (spdNumber) {
        params.set('spdNumber', spdNumber);
    }

    return this.request<import('./types').FinanceSourceOption[]>(
        `/finance-sources/options?${params.toString()}`
    );
}

async createFinanceSource(
    payload: import('./types').SaveFinanceSourcePayload
): Promise<import('./types').FinanceSource> {
    return this.request<import('./types').FinanceSource>(
        '/finance-sources',
        {
            method: 'POST',
            body: JSON.stringify(payload)
        }
    );
}

async updateFinanceSource(
    id: string,
    payload: import('./types').SaveFinanceSourcePayload
): Promise<import('./types').FinanceSource> {
    return this.request<import('./types').FinanceSource>(
        `/finance-sources/${encodeURIComponent(id)}`,
        {
            method: 'PUT',
            body: JSON.stringify(payload)
        }
    );
}

async deleteFinanceSource(id: string): Promise<void> {
    await this.request<{ message: string }>(
        `/finance-sources/${encodeURIComponent(id)}`,
        { method: 'DELETE' }
    );
}

async createFinanceAllocation(
    payload: import('./types').SaveFinanceAllocationPayload
): Promise<import('./types').FinanceAllocation> {
    return this.request<import('./types').FinanceAllocation>(
        '/finance-sources/allocations',
        {
            method: 'POST',
            body: JSON.stringify(payload)
        }
    );
}

async deleteFinanceAllocation(id: string): Promise<void> {
    await this.request<{ message: string }>(
        `/finance-sources/allocations/${encodeURIComponent(id)}`,
        { method: 'DELETE' }
    );
}

    // --- Dalkot ---

    async getDalkotMaster(): Promise<DalkotMasterData> {
        return this.request<DalkotMasterData>(
            '/dalkot/master'
        );
    }

    async getDalkotList(): Promise<DalkotListResponse> {
        return this.request<DalkotListResponse>(
            '/dalkot'
        );
    }

    async getDalkotRecap(
        params: DalkotRecapParams
    ): Promise<DalkotRecapResponse> {
        const query = new URLSearchParams({
            month: String(params.month),
            year: String(params.year)
        });
        if (params.officerId) {
            query.set('officerId', params.officerId);
        }
        return this.request<DalkotRecapResponse>(
            `/dalkot/recap?${query.toString()}`
        );
    }

    async getDalkotMonthlySPJRecap(
        params: DalkotPeriodParams
    ): Promise<DalkotMonthlySPJRecapResponse> {
        const query = new URLSearchParams({
            month: String(params.month),
            year: String(params.year)
        });
        return this.request<DalkotMonthlySPJRecapResponse>(
            `/dalkot/recap/monthly-spj?${query.toString()}`
        );
    }

    async getDalkotById(
        id: string
    ): Promise<DalkotRecord> {
        return this.request<DalkotRecord>(
            `/dalkot/${encodeURIComponent(id)}`
        );
    }

    async createDalkot(
        formData: FormData
    ): Promise<DalkotRecord> {
        return this.multipartRequest<DalkotRecord>(
            '/dalkot',
            'POST',
            formData
        );
    }

    async updateDalkot(
        id: string,
        formData: FormData
    ): Promise<DalkotRecord> {
        return this.multipartRequest<DalkotRecord>(
            `/dalkot/${encodeURIComponent(id)}`,
            'PUT',
            formData
        );
    }

    async deleteDalkot(
        id: string
    ): Promise<void> {
        await this.request<{ message: string }>(
            `/dalkot/${encodeURIComponent(id)}`,
            {
                method: 'DELETE'
            }
        );
    }

    // --- Chatbot Financial Analytics ---

    private chatbotQuery(params?: { year?: number; month?: number }): string {
        const query = new URLSearchParams();
        if (params?.year) query.set('year', String(params.year));
        if (params?.month) query.set('month', String(params.month));
        const value = query.toString();
        return value ? `?${value}` : '';
    }

    async getChatbotGup(
        params?: { year?: number; month?: number }
    ): Promise<import('./types').ChatbotGupAnalysis> {
        return this.request<import('./types').ChatbotGupAnalysis>(
            `/chatbot/gup${this.chatbotQuery(params)}`
        );
    }

    async getChatbotOutsideCity(
        params?: { year?: number; month?: number }
    ): Promise<import('./types').ChatbotOutsideCityAnalysis> {
        return this.request<import('./types').ChatbotOutsideCityAnalysis>(
            `/chatbot/luar-kota${this.chatbotQuery(params)}`
        );
    }

    async getChatbotInsideCity(
        params?: { year?: number; month?: number }
    ): Promise<import('./types').ChatbotInsideCityAnalysis> {
        return this.request<import('./types').ChatbotInsideCityAnalysis>(
            `/chatbot/dalam-kota${this.chatbotQuery(params)}`
        );
    }

    async sendChatbotMessage(message: string): Promise<import('./types').ChatbotMessageResponse> {
        return this.request<import('./types').ChatbotMessageResponse>('/chatbot/message', {
            method: 'POST',
            body: JSON.stringify({ message })
        });
    }

}
