import request from "@/packages/request";

export type VoiceAssetDetailType = {
    id: number;
    name: string;
    gender: string;
    description: string;
    reference_audio_url: string;
    reference_text: string;
    tts_provider_id: number | null;
    tts_provider_name: string;
    tags: string[];
    status: string;
    created_at: string;
    updated_at: string;
};

export type VoiceAssetOptionType = {
    id: number;
    name: string;
    gender: string;
    tags: string[];
    reference_audio_url: string;
};

export type VoiceAssetListParams = {
    page?: number;
    size?: number;
    name?: string;
    gender?: string;
    status?: string;
};

export function getVoiceAssetList(params?: VoiceAssetListParams): Promise<{
    data: VoiceAssetDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/voice-asset/list", params });
}

export function getVoiceAsset(id: number): Promise<VoiceAssetDetailType> {
    return request({ url: `/api/v1/voice-asset/${id}` });
}

export function addVoiceAsset(data: Partial<VoiceAssetDetailType>) {
    return request({ url: "/api/v1/voice-asset", method: "post", data });
}

export function updateVoiceAsset(data: Partial<VoiceAssetDetailType>) {
    return request({ url: `/api/v1/voice-asset/${data.id}`, method: "put", data });
}

export function deleteVoiceAsset(id: number) {
    return request({ url: `/api/v1/voice-asset/${id}`, method: "delete" });
}

export function enableVoiceAsset(id: number) {
    return request({ url: `/api/v1/voice-asset/${id}/enable`, method: "put" });
}

export function disableVoiceAsset(id: number) {
    return request({ url: `/api/v1/voice-asset/${id}/disable`, method: "put" });
}

export function getVoiceAssetOptions(): Promise<VoiceAssetOptionType[]> {
    return request({ url: "/api/v1/common/voice-asset/options" });
}
