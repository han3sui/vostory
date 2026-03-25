import request from "@/packages/request";

export type LLMLogDetailType = {
    id: number;
    project_id: number | null;
    project_name: string;
    provider_id: number;
    provider_name: string;
    template_id: number | null;
    template_name: string;
    model_name: string;
    input_tokens: number;
    output_tokens: number;
    input_summary: string;
    output_summary: string;
    cost_time: number;
    status: number;
    error_message: string;
    created_at: string;
};

export type LLMLogListParams = {
    page?: number;
    size?: number;
    project_id?: number;
    provider_id?: number;
    model_name?: string;
    status?: number;
};

export function getLLMLogList(params?: LLMLogListParams): Promise<{
    data: LLMLogDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/ai/llm-log/list", params });
}

export function getLLMLog(id: number): Promise<LLMLogDetailType> {
    return request({ url: `/api/v1/ai/llm-log/${id}` });
}

export function deleteLLMLog(id: number) {
    return request({ url: `/api/v1/ai/llm-log/${id}`, method: "delete" });
}
