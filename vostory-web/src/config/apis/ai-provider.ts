import request from "@/packages/request";

export type AIProviderType = {
    id: number;
    name: string;
    api_base_url: string;
    api_key: string;
    max_concurrency: number;
    sort_order: number;
    status: string;
    remark: string;
    created_at?: string;
    updated_at?: string;
};

export type AIProviderTestResult = {
    success: boolean;
    message: string;
    duration: number;
};

export function getProviderList(params?: Record<string, any>): Promise<{
    data: AIProviderType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/ai/provider/list", method: "get", params });
}

export function getProvider(id: number): Promise<AIProviderType> {
    return request({ url: `/api/v1/ai/provider/${id}`, method: "get" });
}

export function addProvider(data: Partial<AIProviderType>) {
    return request({ url: "/api/v1/ai/provider", method: "post", data });
}

export function updateProvider(data: Partial<AIProviderType>) {
    return request({ url: `/api/v1/ai/provider/${data.id}`, method: "put", data });
}

export function deleteProvider(id: number) {
    return request({ url: `/api/v1/ai/provider/${id}`, method: "delete" });
}

export function enableProvider(id: number) {
    return request({ url: `/api/v1/ai/provider/${id}/enable`, method: "put" });
}

export function disableProvider(id: number) {
    return request({ url: `/api/v1/ai/provider/${id}/disable`, method: "put" });
}

export function testProvider(data: {
    id?: number;
    api_base_url: string;
    api_key?: string;
    model: string;
}): Promise<AIProviderTestResult> {
    return request({ url: "/api/v1/ai/provider/test", method: "post", data });
}
