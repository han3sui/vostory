import request from "@/packages/request";

export type AIModelType = {
    id: number;
    provider_id: number;
    provider_name?: string;
    model_name: string;
    sort_order: number;
    status: string;
    remark: string;
    created_at?: string;
    updated_at?: string;
};

export function getModelList(params?: Record<string, any>): Promise<{
    data: AIModelType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/ai/model/list", method: "get", params });
}

export function getModel(id: number): Promise<AIModelType> {
    return request({ url: `/api/v1/ai/model/${id}`, method: "get" });
}

export function addModel(data: Partial<AIModelType>) {
    return request({ url: "/api/v1/ai/model", method: "post", data });
}

export function updateModel(data: Partial<AIModelType>) {
    return request({ url: `/api/v1/ai/model/${data.id}`, method: "put", data });
}

export function deleteModel(id: number) {
    return request({ url: `/api/v1/ai/model/${id}`, method: "delete" });
}

export function enableModel(id: number) {
    return request({ url: `/api/v1/ai/model/${id}/enable`, method: "put" });
}

export function disableModel(id: number) {
    return request({ url: `/api/v1/ai/model/${id}/disable`, method: "put" });
}
