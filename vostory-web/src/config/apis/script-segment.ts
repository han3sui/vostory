import request from "@/packages/request";

export type ScriptSegmentDetailType = {
    id: number;
    scene_id: number;
    chapter_id: number;
    segment_num: number;
    segment_type: string;
    content: string;
    original_content: string;
    character_id: number | null;
    character_name: string;
    emotion_type: string;
    emotion_strength: string;
    status: string;
    error_message: string;
    version: number;
    has_audio: boolean;
    audio_url: string;
    clip_id: number | null;
    created_at: string;
    updated_at: string;
};

export type ScriptSegmentListParams = {
    page?: number;
    size?: number;
    chapter_id?: number;
    scene_id?: number;
    segment_type?: string;
    character_id?: number;
    status?: string;
};

export function getScriptSegmentList(params?: ScriptSegmentListParams): Promise<{
    data: ScriptSegmentDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({ url: "/api/v1/script-segment/list", params });
}

export function getScriptSegment(id: number): Promise<ScriptSegmentDetailType> {
    return request({ url: `/api/v1/script-segment/${id}` });
}

export function addScriptSegment(data: Partial<ScriptSegmentDetailType>) {
    return request({ url: "/api/v1/script-segment", method: "post", data });
}

export function updateScriptSegment(data: Partial<ScriptSegmentDetailType>) {
    return request({ url: `/api/v1/script-segment/${data.id}`, method: "put", data });
}

export function deleteScriptSegment(id: number) {
    return request({ url: `/api/v1/script-segment/${id}`, method: "delete" });
}

export function getSegmentsByChapter(chapterId: number): Promise<ScriptSegmentDetailType[]> {
    return request({ url: `/api/v1/common/script-segment/chapter/${chapterId}` });
}

export type ChapterSplitResult = {
    scene_count: number;
    segment_count: number;
    new_characters: number;
    input_tokens: number;
    output_tokens: number;
};

export function splitChapter(chapterId: number): Promise<ChapterSplitResult> {
    return request({ url: `/api/v1/chapter/${chapterId}/split`, method: "post", timeout: 0 });
}
