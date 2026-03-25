import request from "@/packages/request";

export type VoiceEmotionDetailType = {
    id: number;
    voice_profile_id: number;
    emotion_type: string;
    emotion_strength: string;
    reference_audio_url: string;
    reference_text: string;
    created_at: string;
    updated_at: string;
};

export type VoiceEmotionListParams = {
    page?: number;
    size?: number;
    voice_profile_id?: number;
    emotion_type?: string;
    emotion_strength?: string;
};

export function getVoiceEmotionList(params?: VoiceEmotionListParams): Promise<{
    data: VoiceEmotionDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/voice-emotion/list", params });
}

export function getVoiceEmotion(id: number): Promise<VoiceEmotionDetailType> {
    return request({ url: `/api/v1/voice-emotion/${id}` });
}

export function addVoiceEmotion(data: Partial<VoiceEmotionDetailType>) {
    return request({ url: "/api/v1/voice-emotion", method: "post", data });
}

export function updateVoiceEmotion(data: Partial<VoiceEmotionDetailType>) {
    return request({ url: `/api/v1/voice-emotion/${data.id}`, method: "put", data });
}

export function deleteVoiceEmotion(id: number) {
    return request({ url: `/api/v1/voice-emotion/${id}`, method: "delete" });
}

export function getVoiceEmotionsByProfile(voiceProfileId: number): Promise<VoiceEmotionDetailType[]> {
    return request({ url: `/api/v1/common/voice-emotion/profile/${voiceProfileId}` });
}
