<template>
    <div class="script-editor-wrap">
        <!-- 左侧章节列表 -->
        <div class="chapter-sidebar">
            <div
                v-for="ch in chapters"
                :key="ch.id"
                class="chapter-item"
                :class="{ active: selectedChapterId === ch.id }"
                @click="selectChapter(ch)"
            >
                <div class="chapter-title">{{ ch.title || `第${ch.chapter_num}章` }}</div>
                <div class="chapter-meta">{{ ch.word_count }} 字</div>
            </div>
            <a-empty v-if="chapters.length === 0" description="暂无章节" />
        </div>

        <!-- 右侧片段编辑区 -->
        <div class="segment-main">
            <div v-if="!selectedChapterId" class="empty-placeholder">请选择章节</div>
            <template v-else>
                <div class="segment-header">
                    <h3>{{ currentChapter?.title }}</h3>
                    <a-space>
                        <a-button
                            v-if="hasPermission('tts:synthesize')"
                            type="primary"
                            size="small"
                            :loading="batchGenerating"
                            :disabled="generatableCount === 0 && !batchGenerating"
                            @click="handleBatchGenerate"
                        >
                            <template #icon><icon-sound /></template>
                            <template v-if="batchGenerating && batchProgress.total > 0">
                                生成中 {{ batchProgress.completed }}/{{ batchProgress.total }}
                            </template>
                            <template v-else>
                                批量生成 ({{ generatableCount }})
                            </template>
                        </a-button>
                        <a-button
                            v-if="hasPermission('chapter:split')"
                            type="outline"
                            size="small"
                            :loading="splitting"
                            @click="handleSplit"
                        >
                            智能切割
                        </a-button>
                        <a-button
                            v-if="hasPermission('chapter:align')"
                            type="outline"
                            size="small"
                            :loading="aligning"
                            @click="handleAlign"
                        >
                            精准填充
                        </a-button>
                    </a-space>
                </div>

                <div v-if="loadingSegments" class="loading-area">
                    <a-spin />
                </div>

                <div v-else class="segment-list">
                    <div v-for="seg in segments" :key="seg.id" class="segment-card" :class="segmentBorderClass(seg)">
                        <div class="segment-row">
                            <div class="segment-num">#{{ seg.segment_num }}</div>
                            <div class="segment-body">
                                <div class="segment-controls">
                                    <a-select
                                        v-model="seg.segment_type"
                                        size="mini"
                                        style="width: 100px"
                                        @change="() => saveSegment(seg)"
                                    >
                                        <a-option value="dialogue">对白</a-option>
                                        <a-option value="narration">旁白</a-option>
                                        <a-option value="monologue">独白</a-option>
                                        <a-option value="description">描述</a-option>
                                    </a-select>

                                    <a-select
                                        :model-value="seg.character_id ?? undefined"
                                        size="mini"
                                        style="width: 120px"
                                        :placeholder="seg.segment_type === 'narration' || seg.segment_type === 'description' ? '旁白角色' : '说话人'"
                                        allow-clear
                                        @update:model-value="(v: any) => { seg.character_id = v ?? null; saveSegment(seg); }"
                                    >
                                        <a-option v-for="c in characterOptions" :key="c.id" :value="c.id">
                                            {{ c.name }}
                                        </a-option>
                                    </a-select>

                                    <a-select
                                        v-model="seg.emotion_type"
                                        size="mini"
                                        style="width: 100px"
                                        placeholder="情绪"
                                        allow-clear
                                        @change="() => saveSegment(seg)"
                                    >
                                        <a-option value="neutral">平静</a-option>
                                        <a-option value="happy">开心</a-option>
                                        <a-option value="sad">悲伤</a-option>
                                        <a-option value="angry">愤怒</a-option>
                                        <a-option value="fear">恐惧</a-option>
                                        <a-option value="surprise">惊讶</a-option>
                                        <a-option value="disgust">厌恶</a-option>
                                    </a-select>

                                    <a-select
                                        v-model="seg.emotion_strength"
                                        size="mini"
                                        style="width: 80px"
                                        placeholder="强度"
                                        @change="() => saveSegment(seg)"
                                    >
                                        <a-option value="light">轻</a-option>
                                        <a-option value="medium">中</a-option>
                                        <a-option value="strong">强</a-option>
                                    </a-select>

                                    <a-tag size="small" :color="statusColor(seg.status)">
                                        {{ statusLabel(seg.status) }}
                                    </a-tag>
                                    <span class="version-label">v{{ seg.version }}</span>

                                    <a-tooltip :content="disableReason(seg)" :disabled="!disableReason(seg)" mini>
                                        <a-button
                                            type="outline"
                                            size="mini"
                                            status="normal"
                                            :loading="synthesizingId === seg.id"
                                            :disabled="!canGenerate(seg)"
                                            @click="handleGenerate(seg)"
                                        >
                                            <template #icon><icon-sound /></template>
                                            生成
                                        </a-button>
                                    </a-tooltip>
                                    <a-button
                                        v-if="seg.clip_id"
                                        type="text"
                                        size="mini"
                                        :class="{ 'playing-btn': playingId === seg.id }"
                                        @click="togglePlayAudio(seg)"
                                    >
                                        <template #icon>
                                            <icon-pause v-if="playingId === seg.id" />
                                            <icon-play-arrow v-else />
                                        </template>
                                    </a-button>
                                </div>

                                <a-textarea
                                    v-model="seg.content"
                                    :auto-size="{ minRows: 1, maxRows: 6 }"
                                    class="segment-textarea"
                                    @blur="() => saveSegment(seg)"
                                />

                                <div
                                    v-if="seg.original_content && seg.original_content !== seg.content"
                                    class="original-text"
                                >
                                    原文：{{ seg.original_content }}
                                </div>
                            </div>
                        </div>
                    </div>
                    <a-empty v-if="segments.length === 0" description="暂无脚本片段" />
                </div>
            </template>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { Message } from "@arco-design/web-vue";
import { Modal } from "@arco-design/web-vue";
import { IconSound, IconPlayArrow, IconPause } from "@arco-design/web-vue/es/icon";
import {
    getSegmentsByChapter,
    updateScriptSegment,
    splitChapter,
    ScriptSegmentDetailType
} from "@/config/apis/script-segment";
import { getCharactersByProject, CharacterOptionType } from "@/config/apis/character";
import { synthesizeSegment, batchGenerate, getTaskProgress, getTTSStreamURL } from "@/config/apis/tts";
import { hasPermission } from "@/views/utils";
import request from "@/packages/request";
import storage from "@/utils/tools/storage";

const props = defineProps<{ projectId: number }>();

const selectedChapterId = ref<number>();
const currentChapter = ref<any>(null);
const chapters = ref<any[]>([]);
const segments = ref<ScriptSegmentDetailType[]>([]);
const characterOptions = ref<CharacterOptionType[]>([]);
const loadingSegments = ref(false);
const aligning = ref(false);
const splitting = ref(false);
const synthesizingId = ref<number | null>(null);
const batchGenerating = ref(false);
const batchProgress = ref({ total: 0, completed: 0, status: "" });
let batchPollTimer: ReturnType<typeof setInterval> | null = null;
const playingId = ref<number | null>(null);
let currentAudioEl: HTMLAudioElement | null = null;
let currentBlobURL: string | null = null;

const generatableCount = computed(() => segments.value.filter((s) => canGenerate(s)).length);

function canGenerate(seg: ScriptSegmentDetailType): boolean {
    return !disableReason(seg);
}

function disableReason(seg: ScriptSegmentDetailType): string {
    if (!seg.content?.trim()) return "片段内容为空";
    if (!seg.character_id) return "未指定说话人角色";
    if (seg.status === "processing") return "正在生成中";
    return "";
}

async function loadChapters() {
    chapters.value = [];
    segments.value = [];
    selectedChapterId.value = undefined;
    currentChapter.value = null;
    if (!props.projectId) return;

    const res: any = await request({
        url: `/api/v1/common/chapter/project/${props.projectId}`
    });
    chapters.value = res || [];

    characterOptions.value = await getCharactersByProject(props.projectId);
}

onMounted(loadChapters);
watch(() => props.projectId, loadChapters);

async function selectChapter(ch: any) {
    selectedChapterId.value = ch.id;
    currentChapter.value = ch;
    loadingSegments.value = true;
    try {
        segments.value = await getSegmentsByChapter(ch.id);
    } finally {
        loadingSegments.value = false;
    }
}

let saveTimer: ReturnType<typeof setTimeout> | null = null;
async function saveSegment(seg: ScriptSegmentDetailType) {
    if (saveTimer) clearTimeout(saveTimer);
    saveTimer = setTimeout(async () => {
        try {
            await updateScriptSegment({
                id: seg.id,
                segment_type: seg.segment_type,
                content: seg.content,
                character_id: seg.character_id,
                emotion_type: seg.emotion_type,
                emotion_strength: seg.emotion_strength,
                status: "edited"
            });
        } catch {
            Message.error("保存失败");
        }
    }, 500);
}

async function handleAlign() {
    if (!selectedChapterId.value) return;
    aligning.value = true;
    try {
        const res: any = await request({
            url: `/api/v1/chapter/${selectedChapterId.value}/align`,
            method: "post"
        });
        Message.success(`精准填充完成，对齐了 ${res.aligned_count} 个片段`);
        segments.value = await getSegmentsByChapter(selectedChapterId.value);
    } catch {
        Message.error("精准填充失败");
    } finally {
        aligning.value = false;
    }
}

async function handleSplit() {
    if (!selectedChapterId.value) return;

    const doSplit = async () => {
        splitting.value = true;
        try {
            const res = await splitChapter(selectedChapterId.value!);
            let msg = `智能切割完成：${res.scene_count} 个场景，${res.segment_count} 个片段`;
            if (res.new_characters > 0) {
                msg += `，自动发现 ${res.new_characters} 个新角色`;
            }
            Message.success(msg);
            segments.value = await getSegmentsByChapter(selectedChapterId.value!);
            characterOptions.value = await getCharactersByProject(props.projectId);
        } finally {
            splitting.value = false;
        }
    };

    if (segments.value.length > 0) {
        Modal.warning({
            title: "确认重新切割",
            content: `当前章节已有 ${segments.value.length} 个片段，重新切割将覆盖现有数据，是否继续？`,
            okText: "确认切割",
            cancelText: "取消",
            hideCancel: false,
            onOk: doSplit
        });
    } else {
        doSplit();
    }
}

async function handleGenerate(seg: ScriptSegmentDetailType) {
    if (!canGenerate(seg)) return;

    synthesizingId.value = seg.id;
    seg.status = "processing";
    try {
        const result = await synthesizeSegment(seg.id);
        seg.clip_id = result.clip_id;
        seg.has_audio = true;
        seg.status = "generated";
        Message.success(`#${seg.segment_num} 生成完成`);
    } catch {
        seg.status = "failed";
    } finally {
        synthesizingId.value = null;
    }
}

async function handleBatchGenerate() {
    if (!selectedChapterId.value) return;
    const todo = segments.value.filter((s) => canGenerate(s));
    if (!todo.length) {
        Message.info("没有可生成的片段");
        return;
    }

    Modal.confirm({
        title: "批量生成配音",
        content: `将为 ${todo.length} 个片段生成配音，已绑定角色和声音的片段才会被处理。是否继续？`,
        okText: "确认生成",
        cancelText: "取消",
        onOk: async () => {
            batchGenerating.value = true;
            batchProgress.value = { total: 0, completed: 0, status: "pending" };
            todo.forEach((seg) => (seg.status = "processing"));

            try {
                const result = await batchGenerate(selectedChapterId.value!);
                batchProgress.value = { total: result.total_count, completed: 0, status: "running" };
                startBatchPolling(result.task_id);
            } catch {
                batchGenerating.value = false;
                todo.forEach((seg) => (seg.status = "failed"));
                Message.error("批量生成启动失败");
            }
        }
    });
}

function startBatchPolling(taskId: number) {
    if (batchPollTimer) clearInterval(batchPollTimer);

    batchPollTimer = setInterval(async () => {
        try {
            const progress = await getTaskProgress(taskId);
            batchProgress.value = {
                total: progress.total_count,
                completed: progress.completed_count,
                status: progress.status
            };

            if (progress.status === "completed" || progress.status === "failed") {
                clearInterval(batchPollTimer!);
                batchPollTimer = null;
                batchGenerating.value = false;

                // 刷新片段列表以获取最新状态和 clip_id
                if (selectedChapterId.value) {
                    segments.value = await getSegmentsByChapter(selectedChapterId.value);
                }

                if (progress.error_message) {
                    Message.warning(`批量生成完成：${progress.error_message}`);
                } else {
                    Message.success(`批量生成完成：${progress.completed_count} 个片段`);
                }
            }
        } catch {
            // 轮询失败不中断，继续等待
        }
    }, 2000);
}

onUnmounted(() => {
    if (batchPollTimer) {
        clearInterval(batchPollTimer);
        batchPollTimer = null;
    }
    stopAudio();
});

async function togglePlayAudio(seg: ScriptSegmentDetailType) {
    if (playingId.value === seg.id) {
        stopAudio();
        return;
    }

    stopAudio();
    if (!seg.clip_id) return;

    playingId.value = seg.id;
    try {
        const resp = await fetch(getTTSStreamURL(seg.clip_id), {
            headers: { Authorization: `Bearer ${storage.getToken()}` }
        });
        if (!resp.ok) throw new Error("fetch failed");

        const blob = await resp.blob();
        const blobURL = URL.createObjectURL(blob);
        currentBlobURL = blobURL;

        const audio = new Audio(blobURL);
        currentAudioEl = audio;
        audio.addEventListener("ended", () => stopAudio());
        audio.addEventListener("error", () => {
            Message.warning("音频播放失败");
            stopAudio();
        });
        await audio.play();
    } catch {
        Message.warning("音频播放失败");
        stopAudio();
    }
}

function stopAudio() {
    if (currentAudioEl) {
        currentAudioEl.pause();
        currentAudioEl = null;
    }
    if (currentBlobURL) {
        URL.revokeObjectURL(currentBlobURL);
        currentBlobURL = null;
    }
    playingId.value = null;
}

function segmentBorderClass(seg: ScriptSegmentDetailType) {
    const map: Record<string, string> = {
        dialogue: "border-dialogue",
        narration: "border-narration",
        monologue: "border-monologue",
        description: "border-description"
    };
    return map[seg.segment_type] || "border-narration";
}

function statusColor(status: string) {
    const map: Record<string, string> = {
        raw: "gray",
        edited: "blue",
        processing: "orangered",
        generated: "green",
        failed: "red"
    };
    return map[status] || "gray";
}

function statusLabel(status: string) {
    const map: Record<string, string> = {
        raw: "原始",
        edited: "已编辑",
        processing: "生成中",
        generated: "已生成",
        failed: "生成失败"
    };
    return map[status] || status;
}
</script>
<style lang="scss" scoped>
.script-editor-wrap {
    display: flex;
    height: calc(100vh - 300px);
    min-height: 400px;
}

.chapter-sidebar {
    width: 240px;
    flex-shrink: 0;
    border-right: 1px solid var(--color-border);
    overflow-y: auto;
    padding: 12px;
}

.chapter-item {
    padding: 8px 12px;
    border-radius: 4px;
    cursor: pointer;
    margin-bottom: 4px;
    font-size: 13px;

    &:hover {
        background-color: var(--color-fill-2);
    }

    &.active {
        background-color: rgb(var(--primary-1));
        color: rgb(var(--primary-6));
        font-weight: 500;
    }
}

.chapter-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.chapter-meta {
    font-size: 12px;
    color: var(--color-text-3);
    margin-top: 2px;
}

.segment-main {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    min-width: 0;
}

.empty-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--color-text-3);
}

.segment-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;

    h3 {
        font-size: 16px;
        font-weight: 500;
        margin: 0;
    }
}

.loading-area {
    display: flex;
    justify-content: center;
    padding: 40px 0;
}

.segment-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.segment-card {
    border: 1px solid var(--color-border);
    border-radius: 8px;
    padding: 12px;
    transition: box-shadow 0.2s;

    &:hover {
        box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
    }

    &.border-dialogue {
        border-color: rgb(var(--primary-3));
    }
    &.border-narration {
        border-color: var(--color-border);
    }
    &.border-monologue {
        border-color: rgb(var(--purple-3));
    }
    &.border-description {
        border-color: rgb(var(--green-3));
    }
}

.segment-row {
    display: flex;
    align-items: flex-start;
    gap: 12px;
}

.segment-num {
    flex-shrink: 0;
    width: 32px;
    text-align: center;
    font-size: 12px;
    color: var(--color-text-3);
    padding-top: 4px;
}

.segment-body {
    flex: 1;
    min-width: 0;
}

.segment-controls {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
    flex-wrap: wrap;
}

.version-label {
    font-size: 12px;
    color: var(--color-text-3);
}

.playing-btn {
    color: rgb(var(--primary-6));
    animation: pulse 1s ease-in-out infinite;
}

@keyframes pulse {
    0%,
    100% {
        opacity: 1;
    }
    50% {
        opacity: 0.5;
    }
}

.segment-textarea {
    :deep(.arco-textarea) {
        border: none;
        background-color: var(--color-fill-1);
        border-radius: 4px;
    }
}

.original-text {
    margin-top: 4px;
    font-size: 12px;
    color: var(--color-text-3);
    background-color: rgb(var(--warning-1));
    padding: 4px;
    border-radius: 4px;
}
</style>
