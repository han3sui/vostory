import request from "@/packages/request";
import { RequestOption } from "@arco-design/web-vue";
import storage from "@/utils/tools/storage";
import envHelper from "@/utils/helper/env";

export type UploadResult = {
    path: string;
    filename: string;
};

/**
 * 从存储路径中提取原始文件名。
 * 存储格式为 "storage/reference-audio/1774443473329_原始文件名.wav"，
 * 需去掉时间戳前缀。
 */
export function extractFilenameFromPath(storedPath: string): string {
    if (!storedPath) return "";
    const base = storedPath.split(/[/\\]/).pop() || "";
    const idx = base.indexOf("_");
    return idx > 0 ? base.substring(idx + 1) : base;
}

/**
 * 从 formHelper.upload 提交的值中提取存储路径字符串。
 * upload 组件提交时值可能是 FileItem[]、字符串或 undefined。
 */
export function extractUploadUrl(value: any): string {
    if (!value) return "";
    if (typeof value === "string") return value;
    if (Array.isArray(value)) {
        const item = value[0];
        if (!item) return "";
        if (item.url) return item.url;
        if (typeof item.response === "string") return item.response;
        if (item.response?.path) return item.response.path;
        if (item.response?.url) return item.response.url;
        return "";
    }
    return "";
}

/**
 * 将后端存储路径转换为 FileItem[] 供 formHelper.upload 回显。
 * 编辑时调用，让 upload 组件显示已上传的文件名。
 */
export function pathToFileList(storedPath: string): any[] {
    if (!storedPath) return [];
    return [
        {
            uid: storedPath,
            name: extractFilenameFromPath(storedPath),
            url: storedPath,
            status: "done",
            percent: 1
        }
    ];
}

export type ReferenceAudioSource = "voice-asset" | "voice-profile" | "voice-emotion";

/**
 * 获取参考音频流播放 URL。
 */
export function getReferenceAudioStreamURL(source: ReferenceAudioSource, id: number): string {
    const base = envHelper.get("VITE_APP_API_URL") || "";
    return `${base}/api/v1/common/reference-audio/stream?source=${source}&id=${id}`;
}

/**
 * 通过流接口播放参考音频，返回 Blob URL。
 */
export async function fetchReferenceAudioBlob(source: ReferenceAudioSource, id: number): Promise<string> {
    const url = getReferenceAudioStreamURL(source, id);
    const token = storage.getToken();
    const resp = await fetch(url, {
        headers: { Authorization: `Bearer ${token}` }
    });
    if (!resp.ok) throw new Error("获取音频失败");
    const blob = await resp.blob();
    return URL.createObjectURL(blob);
}

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
            fileItem.url = res.path;
            fileItem.name = res.filename;
            onSuccess(res.path);
            return res;
        })
        .catch((err) => {
            onError(err);
            throw err;
        });
}
