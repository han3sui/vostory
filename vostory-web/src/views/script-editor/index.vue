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
            <div v-if="!selectedChapterId" class="empty-placeholder">
                请选择章节
            </div>
            <template v-else>
                <div class="segment-header">
                    <h3>{{ currentChapter?.title }}</h3>
                    <a-space>
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
                    <div
                        v-for="seg in segments"
                        :key="seg.id"
                        class="segment-card"
                        :class="segmentBorderClass(seg)"
                    >
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
                                        v-if="seg.segment_type === 'dialogue' || seg.segment_type === 'monologue'"
                                        :model-value="seg.character_id ?? undefined"
                                        size="mini"
                                        style="width: 120px"
                                        placeholder="说话人"
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
import {
    getSegmentsByChapter,
    updateScriptSegment,
    ScriptSegmentDetailType
} from "@/config/apis/script-segment";
import { getCharactersByProject, CharacterOptionType } from "@/config/apis/character";
import { hasPermission } from "@/views/utils";
import request from "@/packages/request";

const props = defineProps<{ projectId: number }>();

const selectedChapterId = ref<number>();
const currentChapter = ref<any>(null);
const chapters = ref<any[]>([]);
const segments = ref<ScriptSegmentDetailType[]>([]);
const characterOptions = ref<CharacterOptionType[]>([]);
const loadingSegments = ref(false);
const aligning = ref(false);

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
    const map: Record<string, string> = { raw: "gray", edited: "blue", generated: "green" };
    return map[status] || "gray";
}

function statusLabel(status: string) {
    const map: Record<string, string> = { raw: "原始", edited: "已编辑", generated: "已生成" };
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
