import request from "@/packages/request";

export type VoiceProfileDetailType = {
    id: number;
    project_id: number;
    name: string;
    gender: string;
    description: string;
    voice_asset_id: number | null;
    reference_audio_url: string;
    reference_text: string;
    tts_provider_id: number | null;
    tts_provider_name: string;
    tts_params: Record<string, any>;
    status: string;
    created_at: string;
    updated_at: string;
};

export type VoiceProfileOptionType = {
    id: number;
    name: string;
};

export type VoiceProfileListParams = {
    page?: number;
    size?: number;
    project_id?: number;
    name?: string;
    status?: string;
};

export function getVoiceProfileList(params?: VoiceProfileListParams): Promise<{
    data: VoiceProfileDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/voice-profile/list", params });
}

export function getVoiceProfile(id: number): Promise<VoiceProfileDetailType> {
    return request({ url: `/api/v1/voice-profile/${id}` });
}

export function addVoiceProfile(data: Partial<VoiceProfileDetailType>) {
    return request({ url: "/api/v1/voice-profile", method: "post", data });
}

export function updateVoiceProfile(data: Partial<VoiceProfileDetailType>) {
    return request({ url: `/api/v1/voice-profile/${data.id}`, method: "put", data });
}

export function deleteVoiceProfile(id: number) {
    return request({ url: `/api/v1/voice-profile/${id}`, method: "delete" });
}

export function enableVoiceProfile(id: number) {
    return request({ url: `/api/v1/voice-profile/${id}/enable`, method: "put" });
}

export function disableVoiceProfile(id: number) {
    return request({ url: `/api/v1/voice-profile/${id}/disable`, method: "put" });
}

export function getVoiceProfilesByProject(projectId: number): Promise<VoiceProfileOptionType[]> {
    return request({ url: `/api/v1/common/voice-profile/project/${projectId}` });
}
