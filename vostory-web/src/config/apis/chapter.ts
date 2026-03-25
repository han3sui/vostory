import request from "@/packages/request";

export type ChapterDetailType = {
    id: number;
    project_id: number;
    title: string;
    chapter_num: number;
    content: string;
    word_count: number;
    status: string;
    remark: string;
    created_at: string;
    updated_at: string;
};

export type ChapterListParams = {
    page?: number;
    size?: number;
    project_id?: number;
    title?: string;
    status?: string;
};

export function getChapterList(params?: ChapterListParams): Promise<{
    data: ChapterDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/chapter/list", params });
}

export function getChapter(id: number): Promise<ChapterDetailType> {
    return request({ url: `/api/v1/chapter/${id}`, loading: true });
}

export function addChapter(data: Partial<ChapterDetailType>) {
    return request({ url: "/api/v1/chapter", method: "post", data });
}

export function updateChapter(data: Partial<ChapterDetailType>) {
    return request({ url: `/api/v1/chapter/${data.id}`, method: "put", data });
}

export function deleteChapter(id: number) {
    return request({ url: `/api/v1/chapter/${id}`, method: "delete" });
}

export function getChaptersByProject(projectId: number): Promise<ChapterDetailType[]> {
    return request({ url: `/api/v1/common/chapter/project/${projectId}` });
}
