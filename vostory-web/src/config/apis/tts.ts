import request from "@/packages/request";
import envHelper from "@/utils/helper/env";

export type TTSSynthesizeResult = {
    clip_id: number;
    audio_url: string;
    duration: number;
    version: number;
};

export type BatchGenerateResult = {
    task_id: number;
    total_count: number;
    skipped_count: number;
};

export type TaskProgressResult = {
    task_id: number;
    status: string;
    progress: number;
    total_count: number;
    completed_count: number;
    failed_count: number;
    error_message?: string;
    started_at?: string;
    completed_at?: string;
};

export function synthesizeSegment(segmentId: number): Promise<TTSSynthesizeResult> {
    return request({ url: `/api/v1/tts/synthesize/${segmentId}`, method: "post", timeout: 0 });
}

export function getSegmentAudio(segmentId: number): Promise<TTSSynthesizeResult> {
    return request({ url: `/api/v1/tts/audio/${segmentId}` });
}

export function batchGenerate(chapterId: number): Promise<BatchGenerateResult> {
    return request({ url: "/api/v1/tts/batch-generate", method: "post", data: { chapter_id: chapterId } });
}

export function getTaskProgress(taskId: number): Promise<TaskProgressResult> {
    return request({ url: `/api/v1/tts/task/${taskId}` });
}

export function getTTSStreamURL(clipId: number): string {
    const base = envHelper.get("VITE_APP_API_URL") || "";
    return `${base}/api/v1/tts/stream/${clipId}`;
}
