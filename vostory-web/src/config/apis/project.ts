import request from "@/packages/request";

// ============= 项目相关接口 =============

export type ProjectDetailType = {
    id: number;
    workspace_id: number;
    workspace_name: string;
    name: string;
    description: string;
    cover_url: string;
    source_type: string;
    source_file_url: string;
    status: string;
    llm_provider_id: number | null;
    llm_provider_name: string;
    tts_provider_id: number | null;
    tts_provider_name: string;
    prompt_template_ids: Record<string, number>;
    total_chapters: number;
    total_characters: number;
    remark: string;
    created_by: string;
    created_at: string;
    updated_at: string;
};

export type ProjectOptionType = {
    id: number;
    name: string;
};

export type ProjectListParams = {
    page?: number;
    size?: number;
    workspace_id?: number;
    name?: string;
    status?: string;
    source_type?: string;
};

export type ProjectCreateParams = {
    workspace_id: number;
    name: string;
    description?: string;
    cover_url?: string;
    llm_provider_id?: number | null;
    tts_provider_id?: number | null;
    prompt_template_ids?: Record<string, number>;
    remark?: string;
};

export type ProjectUpdateParams = {
    id: number;
    name: string;
    description?: string;
    cover_url?: string;
    llm_provider_id?: number | null;
    tts_provider_id?: number | null;
    prompt_template_ids?: Record<string, number>;
    remark?: string;
};

export function getProjectList(params?: ProjectListParams): Promise<{
    data: ProjectDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/project/list", params });
}

export function getProject(id: number): Promise<ProjectDetailType> {
    return request({ url: `/api/v1/project/${id}` });
}

export function addProject(data: ProjectCreateParams) {
    return request({ url: "/api/v1/project", method: "post", data });
}

export function updateProject(data: ProjectUpdateParams) {
    return request({ url: `/api/v1/project/${data.id}`, method: "put", data });
}

export function deleteProject(id: number) {
    return request({ url: `/api/v1/project/${id}`, method: "delete" });
}

export function getProjectsByWorkspace(workspaceId: number): Promise<ProjectOptionType[]> {
    return request({ url: `/api/v1/common/project/workspace/${workspaceId}` });
}
