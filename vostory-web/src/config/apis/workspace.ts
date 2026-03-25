import request from "@/packages/request";

// ============= 工作空间相关接口 =============

export type WorkspaceDetailType = {
    id: number;
    name: string;
    description: string;
    owner_id: number;
    owner_name: string;
    status: string;
    created_at: string;
    updated_at: string;
};

export type WorkspaceOptionType = {
    id: number;
    name: string;
};

export type WorkspaceListParams = {
    page?: number;
    size?: number;
    name?: string;
    status?: string;
};

export type WorkspaceCreateParams = {
    name: string;
    description?: string;
    status: string;
};

export type WorkspaceUpdateParams = WorkspaceCreateParams & {
    id: number;
};

export function getWorkspaceList(params?: WorkspaceListParams): Promise<{
    data: WorkspaceDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/workspace/list", params });
}

export function getWorkspace(id: number): Promise<WorkspaceDetailType> {
    return request({ url: `/api/v1/workspace/${id}` });
}

export function addWorkspace(data: WorkspaceCreateParams) {
    return request({ url: "/api/v1/workspace", method: "post", data });
}

export function updateWorkspace(data: WorkspaceUpdateParams) {
    return request({ url: `/api/v1/workspace/${data.id}`, method: "put", data });
}

export function deleteWorkspace(id: number) {
    return request({ url: `/api/v1/workspace/${id}`, method: "delete" });
}

export function enableWorkspace(id: number) {
    return request({ url: `/api/v1/workspace/${id}/enable`, method: "put" });
}

export function disableWorkspace(id: number) {
    return request({ url: `/api/v1/workspace/${id}/disable`, method: "put" });
}

export function getWorkspaceOptions(): Promise<WorkspaceOptionType[]> {
    return request({ url: "/api/v1/common/workspace/options" });
}
