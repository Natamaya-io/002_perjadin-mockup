import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// [ANTI-OOM & SOTA 60FPS]: Blob Virtualization Engine untuk mendinginkan Main Thread DOM dari lag rendering Base64
const blobUrlCache = new Map<string, string>();
export function getBlobUrl(fileOrBase64: any): string {
    if (!fileOrBase64) return '';
    if (typeof fileOrBase64 === 'object' && fileOrBase64.path) {
        return '/uploads/' + fileOrBase64.path.replace(/^\/?uploads\//, '');
    }
    const dataStr = typeof fileOrBase64 === 'string' ? fileOrBase64 : fileOrBase64.data;
    if (!dataStr) return '';
    if (!dataStr.startsWith('data:') || dataStr.length < 500) return dataStr;
    
    if (blobUrlCache.has(dataStr)) return blobUrlCache.get(dataStr)!;
    
    try {
        const parts = dataStr.split(',');
        const mime = parts[0].match(/:(.*?);/)[1];
        const bstr = atob(parts[1]);
        let n = bstr.length;
        const u8arr = new Uint8Array(n);
        while (n--) u8arr[n] = bstr.charCodeAt(n);
        
        const blob = new Blob([u8arr], { type: mime });
        const url = URL.createObjectURL(blob);
        blobUrlCache.set(dataStr, url);
        return url;
    } catch(e) {
        return dataStr;
    }
}

// [ANTI-VRAM EXHAUSTION]: Client-Side Canvas Thumbnail Engine (Downsampling to 250px)
export async function generateThumbnailUrl(fileOrBase64: any): Promise<string> {
    const fullUrl = getBlobUrl(fileOrBase64);
    if (!fullUrl) return '';
    
    const type = typeof fileOrBase64 === 'object' ? (fileOrBase64.type || fileOrBase64.name || '') : '';
    if (type && !type.toLowerCase().match(/image|jpg|jpeg|png/i)) {
         return fullUrl;
    }

    try {
        const response = await fetch(fullUrl);
        const blob = await response.blob();
        
        // SOTA: Off-main-thread decoding (Zero UI Blocking)
        if (typeof createImageBitmap !== 'undefined') {
            const bitmap = await createImageBitmap(blob);
            const MAX_SIZE = 250;
            let width = bitmap.width;
            let height = bitmap.height;
            
            if (width <= MAX_SIZE && height <= MAX_SIZE) {
                bitmap.close();
                return fullUrl;
            }

            if (width > height) {
                height = Math.round(height * (MAX_SIZE / width));
                width = MAX_SIZE;
            } else {
                width = Math.round(width * (MAX_SIZE / height));
                height = MAX_SIZE;
            }

            const canvas = document.createElement('canvas');
            canvas.width = Math.max(1, width);
            canvas.height = Math.max(1, height);
            const ctx = canvas.getContext('2d');
            
            if (ctx) {
                ctx.drawImage(bitmap, 0, 0, width, height);
                bitmap.close(); // Instantly free the massive 48MB VRAM texture!
                
                return new Promise((resolve) => {
                    canvas.toBlob((thumbBlob) => {
                        if (thumbBlob) resolve(URL.createObjectURL(thumbBlob));
                        else resolve(fullUrl);
                    }, 'image/jpeg', 0.6);
                });
            }
        }
    } catch (e) {
        // Fallback silently if fetch or decode fails
    }
    
    return fullUrl;
}

export function getInitials(name: string) {
    if (!name) return '';
    const parts = name.split(' ').filter(Boolean);
    if (parts.length === 0) return '';
    if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

// Cache the NumberFormat instance to prevent severe memory leaks and GC pauses in rendering loops
export const idrFormatter = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
});

export function formatCurrency(amount: number) {
    if (typeof amount !== 'number') return idrFormatter.format(0);
    return idrFormatter.format(amount);
}

export function toTitleCase(str: string) {
    if (!str) return '';
    return str.replace(/\w\S*/g, (txt) => {
        const upper = txt.toUpperCase();
        if (upper === 'DKI' || upper === 'DI' || upper === 'NTB' || upper === 'NTT' || upper === 'RI') return upper;
        return txt.charAt(0).toUpperCase() + txt.slice(1).toLowerCase();
    });
}

export function formatLocations(record: any) {
    if (!record) return '-';
    if (record.locations && record.locations.length > 0) {
        return record.locations.map((loc: any) => `${toTitleCase(loc.location)}, ${toTitleCase(loc.province)}`).join(' & ');
    }
    return `${toTitleCase(record.location)}, ${toTitleCase(record.province)}`;
}

export function getStatusBadge(record: any) {
    if (!record) return { label: 'Assigned', class: 'bg-yellow-50 text-yellow-700 border-yellow-200' };

    // Group evaluation (if the record has employeesList, it represents the whole SPD group)
    if (record.employeesList && Array.isArray(record.employeesList) && record.employeesList.length > 0) {
        // Only Completed if ALL employees are Paid
        const allPaid = record.employeesList.every((emp: any) => emp.paymentStatus === 'Paid');
        if (allPaid) {
            return {
                label: 'Completed',
                class: 'bg-emerald-50 text-emerald-700 border-emerald-200'
            };
        }

        // Check if any is Rejected
        const anyRejected = record.employeesList.some((emp: any) => emp.status === 'Rejected');
        if (anyRejected) {
             return {
                 label: 'Rejected',
                 class: 'bg-red-50 text-red-700 border-red-200'
             };
        }

        // Otherwise it's In Progress if any is Submitted/Approved or Paid (but not all)
        const anyInProgress = record.employeesList.some((emp: any) => emp.status === 'Submitted' || emp.status === 'Approved' || emp.paymentStatus === 'Paid');
        if (anyInProgress) {
            return {
                label: 'In Progress',
                class: 'bg-orange-50 text-orange-700 border-orange-200'
            };
        }

        // Default to Assigned
        return { label: 'Assigned', class: 'bg-yellow-50 text-yellow-700 border-yellow-200' };
    }

    // Individual evaluation
    if (record.paymentStatus === 'Paid') {
        return {
            label: 'Completed',
            class: 'bg-emerald-50 text-emerald-700 border-emerald-200'
        };
    }

    if (record.status === 'Rejected') {
        return {
            label: 'Rejected',
            class: 'bg-red-50 text-red-700 border-red-200'
        };
    }

    if (record.status === 'Submitted' || record.status === 'Approved') {
        return {
            label: 'In Progress',
            class: 'bg-orange-50 text-orange-700 border-orange-200'
        };
    }

    return { label: 'Assigned', class: 'bg-yellow-50 text-yellow-700 border-yellow-200' };
}
import imageCompression from 'browser-image-compression';

export async function compressImage(file: File): Promise<File> {
    if (!file.type.startsWith('image/')) {
        return file;
    }
    const options = {
        maxSizeMB: 1,
        maxWidthOrHeight: 1920,
        useWebWorker: true,
        initialQuality: 0.8
    };
    try {
        const compressedBlob = await imageCompression(file, options);
        return new File([compressedBlob], file.name, {
            type: compressedBlob.type,
            lastModified: Date.now(),
        });
    } catch (error) {
        console.error('Error compressing image:', error);
        return file;
    }
}

