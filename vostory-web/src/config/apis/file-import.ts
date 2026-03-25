import request from "@/packages/request";

export type FileImportResponse = {
    project_id: number;
    file_name: string;
    file_size: number;
    source_type: string;
    source_file_url: string;
};

export type FileParseResponse = {
    project_id: number;
    total_chapters: number;
    total_words: number;
    message: string;
};

export function uploadSourceFile(projectId: number, file: File): Promise<FileImportResponse> {
    const formData = new FormData();
    formData.append("project_id", String(projectId));
    formData.append("file", file);
    return request({
        url: "/api/v1/project/import/upload",
        method: "post",
        data: formData,
        headers: { "Content-Type": "multipart/form-data" }
    });
}

export function parseSourceFile(projectId: number): Promise<FileParseResponse> {
    return request({
        url: `/api/v1/project/import/${projectId}/parse`,
        method: "post"
    });
}
