import request from "@/packages/request";

export type PronunciationDictDetailType = {
    id: number;
    workspace_id: number;
    workspace_name: string;
    project_id: number | null;
    project_name: string;
    word: string;
    phoneme: string;
    remark: string;
    created_at: string;
    updated_at: string;
};

export type PronunciationDictListParams = {
    page?: number;
    size?: number;
    workspace_id?: number;
    project_id?: number;
    word?: string;
};

export function getPronunciationDictList(params?: PronunciationDictListParams): Promise<{
    data: PronunciationDictDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/pronunciation-dict/list", params });
}

export function getPronunciationDict(id: number): Promise<PronunciationDictDetailType> {
    return request({ url: `/api/v1/pronunciation-dict/${id}` });
}

export function addPronunciationDict(data: Partial<PronunciationDictDetailType>) {
    return request({ url: "/api/v1/pronunciation-dict", method: "post", data });
}

export function updatePronunciationDict(data: Partial<PronunciationDictDetailType>) {
    return request({ url: `/api/v1/pronunciation-dict/${data.id}`, method: "put", data });
}

export function deletePronunciationDict(id: number) {
    return request({ url: `/api/v1/pronunciation-dict/${id}`, method: "delete" });
}

export function getEffectivePronunciationDict(workspaceId: number, projectId: number): Promise<PronunciationDictDetailType[]> {
    return request({ url: `/api/v1/common/pronunciation-dict/${workspaceId}/${projectId}` });
}
