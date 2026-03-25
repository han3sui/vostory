import request from "@/packages/request";

// ============= LLM 提供商相关接口 =============

export type LLMProviderDetailType = {
    id: number;
    name: string;
    provider_type: string;
    api_base_url: string;
    api_key: string;
    model_list: string[];
    default_model: string;
    custom_params: Record<string, any>;
    sort_order: number;
    status: string;
    created_at: string;
    updated_at: string;
};

export type LLMProviderOptionType = {
    id: number;
    name: string;
    provider_type: string;
    model_list: string[];
    default_model: string;
};

export type LLMProviderTestResult = {
    success: boolean;
    message: string;
    models?: string[];
    duration: number;
};

export function getLLMProviderList(params?: Record<string, any>): Promise<{
    data: LLMProviderDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/ai/llm-provider/list", params });
}

export function getLLMProvider(id: number): Promise<LLMProviderDetailType> {
    return request({ url: `/api/v1/ai/llm-provider/${id}` });
}

export function addLLMProvider(data: Partial<LLMProviderDetailType>) {
    return request({ url: "/api/v1/ai/llm-provider", method: "post", data });
}

export function updateLLMProvider(data: Partial<LLMProviderDetailType>) {
    return request({ url: `/api/v1/ai/llm-provider/${data.id}`, method: "put", data });
}

export function deleteLLMProvider(id: number) {
    return request({ url: `/api/v1/ai/llm-provider/${id}`, method: "delete" });
}

export function enableLLMProvider(id: number) {
    return request({ url: `/api/v1/ai/llm-provider/${id}/enable`, method: "put" });
}

export function disableLLMProvider(id: number) {
    return request({ url: `/api/v1/ai/llm-provider/${id}/disable`, method: "put" });
}

export function testLLMProvider(data: {
    provider_type: string;
    api_base_url: string;
    api_key?: string;
    model?: string;
    custom_params?: Record<string, any>;
}): Promise<LLMProviderTestResult> {
    return request({ url: "/api/v1/ai/llm-provider/test", method: "post", data });
}

// ============= TTS 提供商相关接口 =============

export type TTSProviderDetailType = {
    id: number;
    name: string;
    provider_type: string;
    api_base_url: string;
    api_key: string;
    supported_features: string[];
    custom_params: Record<string, any>;
    sort_order: number;
    status: string;
    created_at: string;
    updated_at: string;
};

export type TTSProviderOptionType = {
    id: number;
    name: string;
    provider_type: string;
    supported_features: string[];
};

export type TTSProviderTestResult = {
    success: boolean;
    message: string;
    duration: number;
};

export function getTTSProviderList(params?: Record<string, any>): Promise<{
    data: TTSProviderDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/ai/tts-provider/list", params });
}

export function getTTSProvider(id: number): Promise<TTSProviderDetailType> {
    return request({ url: `/api/v1/ai/tts-provider/${id}` });
}

export function addTTSProvider(data: Partial<TTSProviderDetailType>) {
    return request({ url: "/api/v1/ai/tts-provider", method: "post", data });
}

export function updateTTSProvider(data: Partial<TTSProviderDetailType>) {
    return request({ url: `/api/v1/ai/tts-provider/${data.id}`, method: "put", data });
}

export function deleteTTSProvider(id: number) {
    return request({ url: `/api/v1/ai/tts-provider/${id}`, method: "delete" });
}

export function enableTTSProvider(id: number) {
    return request({ url: `/api/v1/ai/tts-provider/${id}/enable`, method: "put" });
}

export function disableTTSProvider(id: number) {
    return request({ url: `/api/v1/ai/tts-provider/${id}/disable`, method: "put" });
}

export function testTTSProvider(data: {
    provider_type: string;
    api_base_url: string;
    api_key?: string;
    custom_params?: Record<string, any>;
}): Promise<TTSProviderTestResult> {
    return request({ url: "/api/v1/ai/tts-provider/test", method: "post", data, timeout: 0 });
}

// ============= Prompt 模板相关接口 =============

export type PromptTemplateDetailType = {
    id: number;
    name: string;
    template_type: string;
    content: string;
    description: string;
    is_system: string;
    version: number;
    sort_order: number;
    status: string;
    created_at: string;
    updated_at: string;
};

export type PromptTemplateOptionType = {
    id: number;
    name: string;
    template_type: string;
};

export function getPromptTemplateList(params?: Record<string, any>): Promise<{
    data: PromptTemplateDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/ai/prompt-template/list", params });
}

export function getPromptTemplate(id: number): Promise<PromptTemplateDetailType> {
    return request({ url: `/api/v1/ai/prompt-template/${id}` });
}

export function addPromptTemplate(data: Partial<PromptTemplateDetailType>) {
    return request({ url: "/api/v1/ai/prompt-template", method: "post", data });
}

export function updatePromptTemplate(data: Partial<PromptTemplateDetailType>) {
    return request({ url: `/api/v1/ai/prompt-template/${data.id}`, method: "put", data });
}

export function deletePromptTemplate(id: number) {
    return request({ url: `/api/v1/ai/prompt-template/${id}`, method: "delete" });
}

export function enablePromptTemplate(id: number) {
    return request({ url: `/api/v1/ai/prompt-template/${id}/enable`, method: "put" });
}

export function disablePromptTemplate(id: number) {
    return request({ url: `/api/v1/ai/prompt-template/${id}/disable`, method: "put" });
}

export function getPromptTemplatesByType(type: string): Promise<PromptTemplateOptionType[]> {
    return request({ url: `/api/v1/common/prompt-template/type/${type}` });
}
