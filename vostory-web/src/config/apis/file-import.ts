import request from "@/packages/request";
import { RequestOption } from "@arco-design/web-vue";

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

export function uploadSourceFile(option: RequestOption, projectId: number): any {
    return new Promise((resolve, reject) => {
        const { onProgress, onError, onSuccess, fileItem } = option;
        const data = new FormData();
        data.append("project_id", String(projectId));
        data.append("file", fileItem.file as File);
        request({
            url: "/api/v1/project/import/upload",
            method: "POST",
            headers: { "Content-Type": "multipart/form-data" },
            enableCancel: false,
            timeout: 0,
            data,
            onUploadProgress: (progressEvent) => {
                onProgress(progressEvent.loaded / progressEvent.total, progressEvent);
            }
        })
            .then((res) => {
                onSuccess(res);
                resolve(res);
            })
            .catch((err) => {
                onError(err);
                reject(err);
            });
    });
}

export function parseSourceFile(projectId: number): Promise<FileParseResponse> {
    return request({
        url: `/api/v1/project/import/${projectId}/parse`,
        method: "post"
    });
}
