import request from "@/packages/request";

export type CharacterDetailType = {
    id: number;
    project_id: number;
    name: string;
    aliases: string[];
    gender: string;
    description: string;
    level: string;
    voice_profile_id: number | null;
    voice_profile_name?: string;
    sort_order: number;
    status: string;
    created_at: string;
    updated_at: string;
};

export type CharacterOptionType = {
    id: number;
    name: string;
};

export type CharacterListParams = {
    page?: number;
    size?: number;
    project_id?: number;
    name?: string;
    gender?: string;
    level?: string;
    status?: string;
};

export function getCharacterList(params?: CharacterListParams): Promise<{
    data: CharacterDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/character/list", params });
}

export function getCharacter(id: number): Promise<CharacterDetailType> {
    return request({ url: `/api/v1/character/${id}` });
}

export function addCharacter(data: Partial<CharacterDetailType>) {
    return request({ url: "/api/v1/character", method: "post", data });
}

export function updateCharacter(data: Partial<CharacterDetailType>) {
    return request({ url: `/api/v1/character/${data.id}`, method: "put", data });
}

export function deleteCharacter(id: number) {
    return request({ url: `/api/v1/character/${id}`, method: "delete" });
}

export function enableCharacter(id: number) {
    return request({ url: `/api/v1/character/${id}/enable`, method: "put" });
}

export function disableCharacter(id: number) {
    return request({ url: `/api/v1/character/${id}/disable`, method: "put" });
}

export function getCharactersByProject(projectId: number): Promise<CharacterOptionType[]> {
    return request({ url: `/api/v1/common/character/project/${projectId}` });
}

export type CharacterExtractResult = {
    extracted_count: number;
    new_count: number;
    skipped_count: number;
    input_tokens: number;
    output_tokens: number;
};

export function extractCharacters(projectId: number): Promise<CharacterExtractResult> {
    return request({ url: `/api/v1/character/extract/${projectId}`, method: "post" });
}

export function extractFromText(projectId: number, text: string): Promise<CharacterExtractResult> {
    return request({
        url: "/api/v1/character/extract-from-text",
        method: "post",
        data: { project_id: projectId, text },
        timeout: 0
    });
}

export type VoiceMatchResult = {
    matched_count: number;
    skipped_count: number;
    failed_count: number;
    details?: {
        character_id: number;
        character_name: string;
        voice_profile_id: number;
        voice_name: string;
        reason: string;
    }[];
    input_tokens: number;
    output_tokens: number;
};

export function voiceMatchCharacters(projectId: number): Promise<VoiceMatchResult> {
    return request({ url: `/api/v1/character/voice-match/${projectId}`, method: "post", timeout: 0 });
}

export type CharacterMergeResult = {
    target_id: number;
    merged_count: number;
    affected_segments: number;
    need_regenerate: number;
};

export function mergeCharacters(data: { target_id: number; source_ids: number[] }): Promise<CharacterMergeResult> {
    return request({ url: "/api/v1/character/merge", method: "post", data });
}

export type CharacterMergeGenericGroup = {
    pool_name: string;
    pool_id?: number;
    source_names: string[];
};

export type CharacterMergeGenericResult = {
    dry_run: boolean;
    merged_count: number;
    affected_segments: number;
    need_regenerate: number;
    groups: CharacterMergeGenericGroup[];
    skipped_names?: string[];
};

export function mergeGenericCharacters(projectId: number, dryRun: boolean): Promise<CharacterMergeGenericResult> {
    return request({
        url: `/api/v1/character/merge-generic/${projectId}`,
        method: "post",
        data: { dry_run: dryRun }
    });
}

export type CharacterMergeSuggestGroup = {
    rule: string;
    reason: string;
    target_kind: "character" | "pool" | string;
    target_id?: number;
    target_name: string;
    source_ids: number[];
    source_names?: string[];
};

export type CharacterMergeSuggestResult = {
    dry_run: boolean;
    merged_count: number;
    affected_segments: number;
    need_regenerate: number;
    groups: CharacterMergeSuggestGroup[];
    skipped_names?: string[];
};

export function suggestMergeCharacters(
    projectId: number,
    data: { dry_run: boolean; groups?: CharacterMergeSuggestGroup[] }
): Promise<CharacterMergeSuggestResult> {
    return request({
        url: `/api/v1/character/merge-suggest/${projectId}`,
        method: "post",
        data
    });
}

export type VoiceDesignSuggestResult = {
    instruction: string;
    sample_text: string;
    input_tokens: number;
    output_tokens: number;
};

export type VoiceDesignPreviewResult = {
    preview_path: string;
    seed: number;
    audio_base64: string;
};

export type VoiceDesignApplyResult = {
    voice_profile_id: number;
    voice_profile_name: string;
};

export type VoiceDesignAutoResult = VoiceDesignApplyResult & {
    character_id: number;
    instruction: string;
    sample_text: string;
    seed: number;
};

export function suggestVoiceDesign(characterId: number): Promise<VoiceDesignSuggestResult> {
    return request({ url: `/api/v1/character/${characterId}/voice-design/suggest`, method: "post", timeout: 0 });
}

export function previewVoiceDesign(
    characterId: number,
    data: { instruction: string; sample_text: string; seed?: number }
): Promise<VoiceDesignPreviewResult> {
    return request({
        url: `/api/v1/character/${characterId}/voice-design/preview`,
        method: "post",
        data,
        timeout: 0
    });
}

export function applyVoiceDesign(
    characterId: number,
    data: { preview_path: string; instruction: string; sample_text: string; seed: number; name?: string }
): Promise<VoiceDesignApplyResult> {
    return request({ url: `/api/v1/character/${characterId}/voice-design/apply`, method: "post", data });
}

export function autoVoiceDesign(characterId: number): Promise<VoiceDesignAutoResult> {
    return request({ url: `/api/v1/character/${characterId}/voice-design/auto`, method: "post", timeout: 0 });
}
