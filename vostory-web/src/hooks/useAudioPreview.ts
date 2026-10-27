import { Message } from "@arco-design/web-vue";

/**
 * 列表试听：同一时刻只播一条，加载完成并真正开始播放后再切到「停止」。
 */
export function useAudioPreview() {
    const playingId = ref<number | null>(null);
    const loadingId = ref<number | null>(null);
    let audioEl: HTMLAudioElement | null = null;
    let blobURL: string | null = null;

    function stop() {
        if (audioEl) {
            audioEl.pause();
            audioEl = null;
        }
        if (blobURL) {
            URL.revokeObjectURL(blobURL);
            blobURL = null;
        }
        playingId.value = null;
        loadingId.value = null;
    }

    function isPlaying(id: number) {
        return playingId.value === id;
    }

    function isLoading(id: number) {
        return loadingId.value === id;
    }

    function buttonText(id: number) {
        if (isLoading(id)) return "加载中";
        if (isPlaying(id)) return "停止";
        return "试听";
    }

    async function toggle(id: number, load: () => Promise<string>) {
        // 正在播放或加载同一条时，再次点击取消
        if (isPlaying(id) || isLoading(id)) {
            stop();
            return;
        }

        stop();
        loadingId.value = id;

        try {
            const url = await load();
            if (loadingId.value !== id) {
                URL.revokeObjectURL(url);
                return;
            }

            blobURL = url;
            const audio = new Audio(url);
            audioEl = audio;

            audio.addEventListener("ended", () => stop());
            audio.addEventListener("error", () => {
                Message.warning("音频播放失败");
                stop();
            });

            // 等音频真正开始播放后再切到停止态
            await audio.play();
            if (loadingId.value !== id) return;
            playingId.value = id;
            loadingId.value = null;
        } catch {
            Message.warning("音频加载失败");
            stop();
        }
    }

    onUnmounted(stop);

    return { playingId, loadingId, toggle, stop, buttonText, isPlaying, isLoading };
}

export type AudioPreviewController = ReturnType<typeof useAudioPreview>;
