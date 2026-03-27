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

export function synthesizeSegment(segmentId: number): Promise<BatchGenerateResult> {
    return request({ url: `/api/v1/tts/synthesize/${segmentId}`, method: "post" });
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

export function getActiveTask(chapterId: number): Promise<TaskProgressResult | null> {
    return request({ url: `/api/v1/tts/chapter/${chapterId}/active-task` });
}

export function getTTSStreamURL(clipId: number): string {
    const base = envHelper.get("VITE_APP_API_URL") || "";
    return `${base}/api/v1/tts/stream/${clipId}`;
}

export type TTSSegmentEvent = {
    type: string;
    task_id: number;
    chapter_id: number;
    chapter_title?: string;
    segment_id: number;
    status: string;
    error_message?: string;
    clip_id?: number;
    audio_url?: string;
    progress: number;
    completed: number;
    failed: number;
    total: number;
    task_done: boolean;
    task_status: string;
};

export function getProjectEventsURL(projectId: number): string {
    const base = envHelper.dev() ? envHelper.get("VITE_APP_API_URL") : "";
    return `${base}/api/v1/tts/project/${projectId}/events`;
}

export type ProjectTaskProgress = {
    task_id: number;
    chapter_id: number | null;
    chapter_title: string;
    task_type: string;
    status: string;
    progress: number;
    total_count: number;
    completed_count: number;
    failed_count: number;
    segment_ids?: number[];
};

export function getActiveTasksByProject(projectId: number): Promise<ProjectTaskProgress[]> {
    return request({ url: `/api/v1/tts/project/${projectId}/active-tasks` });
}

export function lockSegment(segmentId: number) {
    return request({ url: `/api/v1/tts/segment/${segmentId}/lock`, method: "put" });
}

export function unlockSegment(segmentId: number) {
    return request({ url: `/api/v1/tts/segment/${segmentId}/unlock`, method: "put" });
}

export function batchLockChapter(chapterId: number): Promise<{ affected_count: number }> {
    return request({ url: `/api/v1/tts/chapter/${chapterId}/lock`, method: "put" });
}

export function batchUnlockChapter(chapterId: number): Promise<{ affected_count: number }> {
    return request({ url: `/api/v1/tts/chapter/${chapterId}/unlock`, method: "put" });
}

export function cancelChapterQueue(chapterId: number): Promise<{ cancelled_count: number }> {
    return request({ url: `/api/v1/tts/chapter/${chapterId}/cancel`, method: "post" });
}

export function cancelProjectQueue(projectId: number): Promise<{ cancelled_count: number }> {
    return request({ url: `/api/v1/tts/project/${projectId}/cancel`, method: "post" });
}
