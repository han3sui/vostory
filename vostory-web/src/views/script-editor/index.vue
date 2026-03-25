<template>
    <div class="flex" style="height: calc(100vh - 280px)">
        <!-- 左侧章节列表 -->
        <div class="w-60 border-r border-gray-200 overflow-y-auto p-3">
            <div
                v-for="ch in chapters"
                :key="ch.chapter_id"
                class="px-3 py-2 rounded cursor-pointer mb-1 text-sm"
                :class="selectedChapterId === ch.chapter_id ? 'bg-blue-50 text-blue-600 font-medium' : 'hover:bg-gray-50'"
                @click="selectChapter(ch)"
            >
                <div class="truncate">{{ ch.title || `第${ch.chapter_num}章` }}</div>
                <div class="text-xs text-gray-400 mt-0.5">{{ ch.word_count }} 字</div>
            </div>
            <a-empty v-if="chapters.length === 0" description="暂无章节" />
        </div>

        <!-- 右侧片段编辑区 -->
        <div class="flex-1 overflow-y-auto p-4">
            <div v-if="!selectedChapterId" class="flex items-center justify-center h-full text-gray-400">
                请选择章节
            </div>
            <template v-else>
                <div class="flex items-center justify-between mb-4">
                    <h3 class="text-lg font-medium">{{ currentChapter?.title }}</h3>
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

                <div v-if="loadingSegments" class="flex justify-center py-10">
                    <a-spin />
                </div>

                <div v-else class="space-y-2">
                    <div
                        v-for="seg in segments"
                        :key="seg.id"
                        class="border rounded-lg p-3 hover:shadow-sm transition-shadow"
                        :class="segmentBorderClass(seg)"
                    >
                        <div class="flex items-start gap-3">
                            <div class="flex-shrink-0 w-8 text-center text-xs text-gray-400 pt-1">
                                #{{ seg.segment_num }}
                            </div>
                            <div class="flex-1 min-w-0">
                                <div class="flex items-center gap-2 mb-2 flex-wrap">
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
                                        v-model="seg.character_id"
                                        size="mini"
                                        style="width: 120px"
                                        placeholder="说话人"
                                        allow-clear
                                        @change="() => saveSegment(seg)"
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
                                    <span class="text-xs text-gray-400">v{{ seg.version }}</span>
                                </div>

                                <a-textarea
                                    v-model="seg.content"
                                    :auto-size="{ minRows: 1, maxRows: 6 }"
                                    class="!border-0 !bg-gray-50 !rounded"
                                    @blur="() => saveSegment(seg)"
                                />

                                <div
                                    v-if="seg.original_content && seg.original_content !== seg.content"
                                    class="mt-1 text-xs text-gray-400 bg-yellow-50 p-1 rounded"
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
    selectedChapterId.value = ch.chapter_id;
    currentChapter.value = ch;
    loadingSegments.value = true;
    try {
        segments.value = await getSegmentsByChapter(ch.chapter_id);
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
        dialogue: "border-blue-200",
        narration: "border-gray-200",
        monologue: "border-purple-200",
        description: "border-green-200"
    };
    return map[seg.segment_type] || "border-gray-200";
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
<style lang="scss" scoped></style>
