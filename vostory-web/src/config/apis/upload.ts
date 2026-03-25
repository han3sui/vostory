import request from "@/packages/request";
import { RequestOption } from "@arco-design/web-vue";

export type UploadResult = {
    url: string;
};

export function uploadReferenceAudio(option: RequestOption): any {
    const { onProgress, onError, onSuccess, fileItem } = option;
    const data = new FormData();
    data.append("file", fileItem.file as File);
    return request({
        url: "/api/v1/common/upload/reference-audio",
        method: "POST",
        headers: { "Content-Type": "multipart/form-data" },
        enableCancel: false,
        timeout: 0,
        data,
        onUploadProgress: (progressEvent) => {
            onProgress(progressEvent.loaded / progressEvent.total, progressEvent);
        }
    })
        .then((res: any) => {
            onSuccess(res);
            return res;
        })
        .catch((err) => {
            onError(err);
            throw err;
        });
}
