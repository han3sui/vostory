import request from "@/packages/request";

export type TTSSynthesizeResult = {
    clip_id: number;
    audio_url: string;
    duration: number;
    version: number;
};

export function synthesizeSegment(segmentId: number): Promise<TTSSynthesizeResult> {
    return request({ url: `/api/v1/tts/synthesize/${segmentId}`, method: "post", timeout: 0 });
}

export function getSegmentAudio(segmentId: number): Promise<TTSSynthesizeResult> {
    return request({ url: `/api/v1/tts/audio/${segmentId}` });
}
