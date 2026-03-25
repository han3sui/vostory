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
