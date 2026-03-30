import request from "@/packages/request";

export interface LicenseStatusResponse {
    activated: boolean;
    license_code?: string;
    product_code?: string;
    license_type?: string;
    expires_at?: string;
    features?: string;
    mode?: string;
    fingerprint?: string;
    hostname?: string;
    reason?: string;
}

export function getLicenseStatus(): Promise<LicenseStatusResponse> {
    return request({
        url: "/api/v1/license/status",
        notify: false
    });
}

export function activateOnline(licenseCode: string): Promise<LicenseStatusResponse> {
    return request({
        url: "/api/v1/license/activate/online",
        method: "POST",
        data: { license_code: licenseCode }
    });
}

export function activateOffline(licenseFileContent: string, publicKey: string): Promise<LicenseStatusResponse> {
    return request({
        url: "/api/v1/license/activate/offline",
        method: "POST",
        data: { license_file_content: licenseFileContent, public_key: publicKey }
    });
}

export function deactivateLicense(): Promise<void> {
    return request({
        url: "/api/v1/license/deactivate",
        method: "POST"
    });
}
