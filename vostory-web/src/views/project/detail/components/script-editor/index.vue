<template>
    <div class="script-editor-wrap">
        <!-- 左侧章节列表 -->
        <div class="chapter-sidebar">
            <div class="chapter-batch-bar">
                <div class="chapter-batch-top">
                    <a-checkbox
                        v-if="chapters.length > 0 && hasPermission('chapter:split')"
                        :model-value="checkedChapterIds.size === chapters.length && chapters.length > 0"
                        :indeterminate="checkedChapterIds.size > 0 && checkedChapterIds.size < chapters.length"
                        @change="toggleAllChapters"
                    >
                        全选
                    </a-checkbox>
                    <a-space v-if="chapters.length > 0" size="mini">
                        <a-tooltip v-if="hasPermission('chapter:split')" content="选中所有未切割的章节">
                            <a-button
                                size="mini"
                                type="outline"
                                :disabled="unsplitChapterIds.length === 0"
                                @click="selectUnsplitChapters"
                            >
                                <template #icon><icon-filter /></template>
                                {{ unsplitChapterIds.length }}
                            </a-button>
                        </a-tooltip>
                        <a-tooltip content="刷新章节列表">
                            <a-button size="mini" :loading="refreshingChapters" @click="handleRefreshChapters">
                                <template #icon><icon-refresh /></template>
                            </a-button>
                        </a-tooltip>
                    </a-space>
                </div>
                <transition name="batch-slide">
                    <div
                        v-if="checkedChapterIds.size > 0 && hasPermission('chapter:split')"
                        class="chapter-batch-actions"
                    >
                        <span class="batch-hint">已选 {{ checkedChapterIds.size }} 章</span>
                        <span class="batch-spacer" />
                        <a-button size="mini" type="text" @click="clearCheckedChapters">清空</a-button>
                        <a-button type="primary" size="mini" :loading="batchSplitting" @click="handleBatchSplit">
                            批量切割
                        </a-button>
                    </div>
                </transition>
            </div>
            <div class="chapter-list">
                <div
                    v-for="ch in chapters"
                    :key="ch.id"
                    class="chapter-item"
                    :class="{ active: selectedChapterId === ch.id, splitting: splittingChapterIds.has(ch.id) }"
                    @click="selectChapter(ch)"
                >
                    <a-checkbox
                        v-if="hasPermission('chapter:split')"
                        :model-value="checkedChapterIds.has(ch.id)"
                        class="chapter-checkbox"
                        @change="
                            (v: boolean | (string | boolean | number)[]) => toggleChapterCheck(ch.id, v as boolean)
                        "
                        @click.stop
                    />
                    <div class="chapter-info">
                        <div class="chapter-title">{{ ch.title || `第${ch.chapter_num}章` }}</div>
                        <div class="chapter-meta">
                            {{ ch.word_count }} 字
                            <a-tag v-if="splittingChapterIds.has(ch.id)" size="small" color="orange">切割中</a-tag>
                            <a-tag v-else-if="ch.segment_count > 0" size="small" color="green">已切割</a-tag>
                            <a-tag v-else size="small" color="gray">未切割</a-tag>
                        </div>
                    </div>
                </div>
                <a-empty v-if="chapters.length === 0" description="暂无章节" />
            </div>
        </div>

        <!-- 右侧片段编辑区 -->
        <div class="segment-main">
            <div v-if="!selectedChapterId" class="empty-placeholder">请选择章节</div>
            <template v-else>
                <div class="segment-header">
                    <h3>{{ currentChapter?.title }}</h3>
                    <a-space wrap>
                        <a-button
                            v-if="hasPermission('tts:synthesize')"
                            type="primary"
                            size="small"
                            :disabled="segmentStats.generatable === 0"
                            @click="handleBatchGenerate"
                        >
                            <template #icon><icon-sound /></template>
                            批量生成 ({{ segmentStats.generatable }})
                        </a-button>
                        <a-popconfirm
                            v-if="segmentStats.queued > 0"
                            content="确认取消当前章节所有排队中的片段？"
                            @ok="handleCancelQueue"
                        >
                            <a-button type="outline" size="small" status="danger">
                                取消队列 ({{ segmentStats.queued }})
                            </a-button>
                        </a-popconfirm>
                        <a-button
                            v-if="segmentStats.generated > 0"
                            type="outline"
                            size="small"
                            @click="handleBatchLock"
                        >
                            <template #icon><icon-lock /></template>
                            全部锁定 ({{ segmentStats.generated }})
                        </a-button>
                        <a-button v-if="segmentStats.locked > 0" type="outline" size="small" @click="handleBatchUnlock">
                            <template #icon><icon-unlock /></template>
                            全部解锁 ({{ segmentStats.locked }})
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

                <div ref="segmentScrollRef" class="segment-scroll">
                    <div class="segment-list">
                        <div
                            v-for="seg in segments"
                            :key="seg.id"
                            :data-seg-id="seg.id"
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

                                        <a-button size="mini" type="text" @click="handleSelectCharacter(seg)">
                                            <span v-if="seg.character_id && characterMap[seg.character_id]">
                                                {{ characterMap[seg.character_id] }}
                                            </span>
                                            <span v-else class="character-placeholder">
                                                {{
                                                    seg.segment_type === "narration" ||
                                                    seg.segment_type === "description"
                                                        ? "旁白角色"
                                                        : "说话人"
                                                }}
                                            </span>
                                        </a-button>

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

                                        <a-tooltip
                                            :content="seg.error_message"
                                            :disabled="seg.status !== 'failed' || !seg.error_message"
                                            position="top"
                                        >
                                            <a-tag
                                                size="small"
                                                :color="statusColor(seg.status)"
                                                style="cursor: default"
                                            >
                                                {{ statusLabel(seg.status) }}
                                            </a-tag>
                                        </a-tooltip>
                                        <span class="version-label">v{{ seg.version }}</span>

                                        <a-tooltip :content="disableReason(seg)" :disabled="!disableReason(seg)" mini>
                                            <a-button
                                                type="outline"
                                                size="mini"
                                                status="normal"
                                                :loading="synthesizingIds.has(seg.id)"
                                                :disabled="!canGenerate(seg)"
                                                @click="handleGenerate(seg)"
                                            >
                                                <template #icon><icon-sound /></template>
                                                生成
                                            </a-button>
                                        </a-tooltip>
                                        <a-button
                                            v-if="seg.status === 'generated'"
                                            type="text"
                                            size="mini"
                                            @click="handleLock(seg)"
                                        >
                                            <template #icon><icon-lock /></template>
                                        </a-button>
                                        <a-button
                                            v-if="seg.status === 'locked'"
                                            type="text"
                                            size="mini"
                                            @click="handleUnlock(seg)"
                                        >
                                            <template #icon><icon-unlock /></template>
                                        </a-button>
                                        <a-trigger
                                            v-if="seg.clip_id"
                                            trigger="hover"
                                            position="bottom"
                                            :unmount-on-close="false"
                                        >
                                            <a-button
                                                type="text"
                                                size="mini"
                                                :class="{
                                                    'playing-btn':
                                                        playingId === seg.id || continuousPlayingFromId === seg.id
                                                }"
                                                @click="togglePlayAudio(seg)"
                                            >
                                                <template #icon>
                                                    <icon-pause v-if="playingId === seg.id" />
                                                    <icon-play-arrow v-else />
                                                </template>
                                            </a-button>
                                            <template #content>
                                                <div class="play-popover">
                                                    <div class="play-popover-item" @click="togglePlayAudio(seg)">
                                                        <icon-play-arrow />
                                                        <span>播放当前</span>
                                                    </div>
                                                    <div class="play-popover-item" @click="toggleContinuousPlay(seg)">
                                                        <icon-drag-dot-vertical />
                                                        <span>{{
                                                            continuousPlayingFromId === seg.id ? "停止连播" : "连续播放"
                                                        }}</span>
                                                    </div>
                                                </div>
                                            </template>
                                        </a-trigger>
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
                </div>
            </template>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { Message, Modal } from "@arco-design/web-vue";
import { ArcoModalTableShow, formHelper, tableHelper } from "@easyfe/admin-component";
import {
    IconSound,
    IconPlayArrow,
    IconPause,
    IconLock,
    IconUnlock,
    IconDragDotVertical,
    IconFilter
} from "@arco-design/web-vue/es/icon";
import {
    getSegmentsByChapter,
    updateScriptSegment,
    splitChapter,
    batchSplitChapters,
    ScriptSegmentDetailType
} from "@/config/apis/script-segment";
import {
    getCharactersByProject,
    getCharacterList,
    CharacterOptionType,
    CharacterDetailType
} from "@/config/apis/character";
import {
    synthesizeSegment,
    batchGenerate,
    getActiveTasksByProject,
    getTTSStreamURL,
    TTSSegmentEvent,
    lockSegment,
    unlockSegment,
    batchLockChapter,
    batchUnlockChapter,
    cancelChapterQueue
} from "@/config/apis/tts";
import { hasPermission, PageTableConfig } from "@/views/utils";
import request from "@/packages/request";
import storage from "@/utils/tools/storage";

const props = defineProps<{ projectId: number }>();

const onTTSEvent = inject<(handler: (evt: TTSSegmentEvent) => void) => () => void>("onTTSEvent");

const segmentScrollRef = ref<HTMLElement>();
const selectedChapterId = ref<number>();
const currentChapter = ref<any>(null);
const chapters = ref<any[]>([]);
const segments = ref<ScriptSegmentDetailType[]>([]);
const characterOptions = ref<CharacterOptionType[]>([]);
const aligning = ref(false);
const splitting = ref(false);
const batchSplitting = ref(false);
const refreshingChapters = ref(false);
const checkedChapterIds = ref<Set<number>>(new Set());
const splittingChapterIds = ref<Set<number>>(new Set());
const synthesizingIds = ref<Set<number>>(new Set());
const segmentById = ref<Map<number, ScriptSegmentDetailType>>(new Map());
const playingId = ref<number | null>(null);
const continuousPlayingFromId = ref<number | null>(null);
let currentAudioEl: HTMLAudioElement | null = null;
let currentBlobURL: string | null = null;

function setSegments(next: ScriptSegmentDetailType[]) {
    segments.value = next;
    segmentById.value = new Map(next.map((s) => [s.id, s]));
}

const characterMap = computed(() => {
    const map: Record<number, string> = {};
    characterOptions.value.forEach((c) => {
        map[c.id] = c.name;
    });
    return map;
});

const segmentStats = computed(() => {
    let generatable = 0;
    let queued = 0;
    let generated = 0;
    let locked = 0;
    for (const seg of segments.value) {
        switch (seg.status) {
            case "queued":
                queued++;
                break;
            case "generated":
                generated++;
                break;
            case "locked":
                locked++;
                break;
        }
        if (
            seg.content?.trim() &&
            seg.character_id &&
            seg.status !== "queued" &&
            seg.status !== "processing" &&
            seg.status !== "locked"
        ) {
            generatable++;
        }
    }
    return { generatable, queued, generated, locked };
});

const unsplitChapterIds = computed<number[]>(() =>
    chapters.value
        .filter((ch: any) => Number(ch?.segment_count || 0) <= 0 && !splittingChapterIds.value.has(ch.id))
        .map((ch: any) => ch.id)
);

function canGenerate(seg: ScriptSegmentDetailType): boolean {
    return !disableReason(seg);
}

function disableReason(seg: ScriptSegmentDetailType): string {
    if (!seg.content?.trim()) return "片段内容为空";
    if (!seg.character_id) return "未指定说话人角色";
    if (seg.status === "queued") return "已在队列中等待生成";
    if (seg.status === "processing") return "正在生成中";
    if (seg.status === "locked") return "片段已锁定";
    return "";
}

async function loadChapters() {
    chapters.value = [];
    setSegments([]);
    selectedChapterId.value = undefined;
    currentChapter.value = null;
    if (!props.projectId) return;

    const res: any = await request({
        url: `/api/v1/common/chapter/project/${props.projectId}`
    });
    chapters.value = res || [];

    characterOptions.value = await getCharactersByProject(props.projectId);
    await syncSplittingChapterIds();
}

async function syncSplittingChapterIds() {
    if (!props.projectId) return;
    try {
        const tasks = await getActiveTasksByProject(props.projectId);
        const ids = new Set<number>();
        for (const task of tasks) {
            if (task.task_type === "chapter_split" && task.segment_ids) {
                const processed = task.completed_count + task.failed_count;
                task.segment_ids.slice(processed).forEach((id) => ids.add(id));
            }
        }
        splittingChapterIds.value = ids;
    } catch {
        // 查询失败不影响主流程
    }
}

async function handleRefreshChapters() {
    if (!props.projectId) return;
    refreshingChapters.value = true;
    try {
        const prevSelectedId = selectedChapterId.value;
        const res: any = await request({
            url: `/api/v1/common/chapter/project/${props.projectId}`
        });
        chapters.value = res || [];
        characterOptions.value = await getCharactersByProject(props.projectId);
        if (prevSelectedId && chapters.value.some((ch: any) => ch.id === prevSelectedId)) {
            setSegments(await getSegmentsByChapter(prevSelectedId));
        }
        await syncSplittingChapterIds();
    } finally {
        refreshingChapters.value = false;
    }
}

onMounted(loadChapters);
watch(() => props.projectId, loadChapters);

async function selectChapter(ch: any) {
    selectedChapterId.value = ch.id;
    currentChapter.value = ch;
    setSegments(await getSegmentsByChapter(ch.id));
    segmentScrollRef.value?.scrollTo({ top: 0 });
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

const LEVELS = [
    { label: "主角", value: "main" },
    { label: "配角", value: "supporting" },
    { label: "龙套", value: "minor" }
];
const GENDERS = [
    { label: "男", value: "male" },
    { label: "女", value: "female" },
    { label: "未知", value: "unknown" }
];

function handleSelectCharacter(seg: ScriptSegmentDetailType) {
    const filterData = ref<Record<string, any>>({});
    ArcoModalTableShow({
        modalConfig: {
            title: "选择角色",
            width: "1200px"
        },
        defaultSelected: seg.character_id ? [seg.character_id] : [],
        tableConfig: {
            tableConfig: {
                arcoProps: {
                    rowKey: "id",
                    rowSelection: {
                        type: "radio",
                        showCheckedAll: false
                    }
                },
                ...PageTableConfig,
                showRefresh: false,
                maxHeight: "40vh",
                columns: [
                    tableHelper.default("角色名称", "name"),
                    tableHelper.status("性别", "gender", (item: any) => {
                        const found = GENDERS.find((g) => g.value === item.gender);
                        return { text: found?.label || item.gender, status: "normal" };
                    }),
                    tableHelper.status("层级", "level", (item: any) => {
                        const found = LEVELS.find((l) => l.value === item.level);
                        return { text: found?.label || item.level, status: "normal" };
                    }),
                    tableHelper.default("描述", "description"),
                    tableHelper.default("声音配置", "voice_profile_name")
                ]
            },
            filterConfig: [
                formHelper.input("角色名称", "name", { span: 8, debounce: 500 }),
                formHelper.select("层级", "level", LEVELS, { span: 8 }),
                formHelper.select("性别", "gender", GENDERS, { span: 8 })
            ],
            filterData: filterData.value,
            req: {
                fn: getCharacterList,
                params: { project_id: props.projectId, status: "0", ...filterData.value }
            }
        },
        ok: async (selected: CharacterDetailType[]) => {
            if (selected.length > 0) {
                const char = selected[0];
                seg.character_id = char.id;
                if (!characterMap.value[char.id]) {
                    characterOptions.value = [...characterOptions.value, { id: char.id, name: char.name }];
                }
            } else {
                seg.character_id = null;
            }
            saveSegment(seg);
        }
    });
}

async function handleAlign() {
    if (!selectedChapterId.value) return;
    aligning.value = true;
    const res: any = await request({
        url: `/api/v1/chapter/${selectedChapterId.value}/align`,
        method: "post"
    });
    Message.success(`精准填充完成，对齐了 ${res.aligned_count} 个片段`);
    setSegments(await getSegmentsByChapter(selectedChapterId.value));
    aligning.value = false;
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
            setSegments(await getSegmentsByChapter(selectedChapterId.value!));
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

function toggleAllChapters(checked: boolean | (string | boolean | number)[]) {
    if (checked) {
        checkedChapterIds.value = new Set(chapters.value.map((ch: any) => ch.id));
    } else {
        checkedChapterIds.value = new Set();
    }
}

function selectUnsplitChapters() {
    checkedChapterIds.value = new Set(unsplitChapterIds.value);
}

function clearCheckedChapters() {
    checkedChapterIds.value = new Set();
}

function toggleChapterCheck(chapterId: number, checked: boolean) {
    const next = new Set(checkedChapterIds.value);
    if (checked) {
        next.add(chapterId);
    } else {
        next.delete(chapterId);
    }
    checkedChapterIds.value = next;
}

async function handleBatchSplit() {
    if (checkedChapterIds.value.size === 0) return;
    const checkedArr = [...checkedChapterIds.value];

    Modal.confirm({
        title: "批量智能切割",
        content: `将对 ${checkedArr.length} 个章节依次进行智能切割，已有片段的章节将被覆盖。是否继续？`,
        okText: "确认切割",
        cancelText: "取消",
        onOk: async () => {
            batchSplitting.value = true;
            try {
                const res = await batchSplitChapters(props.projectId, checkedArr);
                const ids = new Set(splittingChapterIds.value);
                checkedArr.forEach((id) => ids.add(id));
                splittingChapterIds.value = ids;
                checkedChapterIds.value = new Set();
                Message.success(`已提交 ${res.total} 个章节到切割队列`);
            } catch (e: any) {
                const msg = e?.response?.data?.message || e?.message || "批量切割提交失败";
                Message.error(msg);
            } finally {
                batchSplitting.value = false;
            }
        }
    });
}

function handleChapterSplitEvent(evt: any) {
    if (evt.type !== "chapter_split_done") return;

    const ids = new Set(splittingChapterIds.value);
    ids.delete(evt.chapter_id);
    splittingChapterIds.value = ids;

    if (evt.status === "completed") {
        Message.success(
            `章节「${evt.chapter_title || evt.chapter_id}」切割完成：${evt.scene_count} 场景，${evt.segment_count} 片段`
        );
        const ch = chapters.value.find((c: any) => c.id === evt.chapter_id);
        if (ch) {
            ch.segment_count = evt.segment_count || 1;
        }
    } else {
        Message.warning(`章节「${evt.chapter_title || evt.chapter_id}」切割失败：${evt.error_message || "未知错误"}`);
    }

    if (evt.chapter_id === selectedChapterId.value) {
        getSegmentsByChapter(evt.chapter_id).then((segs) => {
            setSegments(segs);
        });
        getCharactersByProject(props.projectId).then((chars) => {
            characterOptions.value = chars;
        });
    }

    if (evt.task_done) {
        if (evt.failed > 0) {
            Message.warning(`批量切割完成：${evt.completed} 成功，${evt.failed} 失败`);
        } else if (evt.total > 1) {
            Message.success(`批量切割全部完成：${evt.completed} 个章节`);
        }
    }
}

async function handleGenerate(seg: ScriptSegmentDetailType) {
    if (!canGenerate(seg)) return;

    synthesizingIds.value = new Set([...synthesizingIds.value, seg.id]);
    seg.status = "queued";
    try {
        await synthesizeSegment(seg.id);
    } catch {
        seg.status = "failed";
    } finally {
        const next = new Set(synthesizingIds.value);
        next.delete(seg.id);
        synthesizingIds.value = next;
    }
}

function handleSegmentEvent(evt: TTSSegmentEvent) {
    if (evt.chapter_id !== selectedChapterId.value) return;

    const seg = segmentById.value.get(evt.segment_id);
    if (seg) {
        seg.status = evt.status;
        seg.error_message = evt.error_message || "";
        if (evt.status === "generated" && evt.clip_id) {
            seg.clip_id = evt.clip_id;
            seg.has_audio = true;
        }
    }

    if (synthesizingIds.value.has(evt.segment_id) && (evt.status === "generated" || evt.status === "failed")) {
        const next = new Set(synthesizingIds.value);
        next.delete(evt.segment_id);
        synthesizingIds.value = next;
    }

    if (evt.task_done) {
        if (evt.failed > 0) {
            Message.warning(`批量生成完成：${evt.completed} 成功，${evt.failed} 失败`);
        } else if (evt.total > 1) {
            Message.success(`批量生成完成：${evt.completed} 个片段`);
        }
    }
}

let unsubscribeTTSEvent: (() => void) | null = null;
let unsubscribeLLMEvent: (() => void) | null = null;

onMounted(() => {
    if (onTTSEvent) {
        unsubscribeTTSEvent = onTTSEvent(handleSegmentEvent);
        unsubscribeLLMEvent = onTTSEvent(handleChapterSplitEvent);
    }
});

async function handleBatchGenerate() {
    if (!selectedChapterId.value) return;
    const todo = segments.value.filter((s) => canGenerate(s));
    if (!todo.length) {
        Message.info("没有可生成的片段");
        return;
    }

    Modal.confirm({
        title: "批量生成配音",
        content: `将为 ${todo.length} 个片段生成配音（已锁定的片段不会被覆盖）。是否继续？`,
        okText: "确认生成",
        cancelText: "取消",
        onOk: async () => {
            todo.forEach((seg) => (seg.status = "queued"));
            try {
                await batchGenerate(selectedChapterId.value!);
            } catch (e: any) {
                todo.forEach((seg) => {
                    if (seg.status === "queued") seg.status = "failed";
                });
                const msg = e?.response?.data?.message || e?.message || "";
                if (msg.includes("已有正在运行的生成任务")) {
                    Message.warning("该章节已有正在运行的批量生成任务，请等待完成");
                } else {
                    Message.error("批量生成启动失败");
                }
            }
        }
    });
}

async function handleLock(seg: ScriptSegmentDetailType) {
    await lockSegment(seg.id);
    seg.status = "locked";
}

async function handleUnlock(seg: ScriptSegmentDetailType) {
    await unlockSegment(seg.id);
    seg.status = "generated";
}

async function handleBatchLock() {
    if (!selectedChapterId.value) return;
    const res = await batchLockChapter(selectedChapterId.value);
    segments.value.forEach((s) => {
        if (s.status === "generated") s.status = "locked";
    });
    Message.success(`已锁定 ${res.affected_count} 个片段`);
}

async function handleBatchUnlock() {
    if (!selectedChapterId.value) return;
    const res = await batchUnlockChapter(selectedChapterId.value);
    segments.value.forEach((s) => {
        if (s.status === "locked") s.status = "generated";
    });
    Message.success(`已解锁 ${res.affected_count} 个片段`);
}

async function handleCancelQueue() {
    if (!selectedChapterId.value) return;
    const res = await cancelChapterQueue(selectedChapterId.value);
    segments.value.forEach((s) => {
        if (s.status === "queued") s.status = "cancelled";
    });
    Message.success(`已取消 ${res.cancelled_count} 个排队片段`);
}

onUnmounted(() => {
    if (unsubscribeTTSEvent) unsubscribeTTSEvent();
    if (unsubscribeLLMEvent) unsubscribeLLMEvent();
    segmentById.value.clear();
    stopAudio();
});

async function fetchAudioBlob(clipId: number): Promise<string | null> {
    try {
        const resp = await fetch(getTTSStreamURL(clipId), {
            headers: { Authorization: `Bearer ${storage.getToken()}` }
        });
        if (!resp.ok) return null;
        const blob = await resp.blob();
        return URL.createObjectURL(blob);
    } catch {
        return null;
    }
}

function playBlobURL(blobURL: string): Promise<boolean> {
    currentBlobURL = blobURL;
    const audio = new Audio(blobURL);
    currentAudioEl = audio;

    return new Promise<boolean>((resolve) => {
        audio.addEventListener("ended", () => {
            cleanupAudioResource();
            resolve(true);
        });
        audio.addEventListener("error", () => {
            Message.warning("音频播放失败");
            cleanupAudioResource();
            resolve(false);
        });
        audio.play().catch(() => {
            Message.warning("音频播放失败");
            cleanupAudioResource();
            resolve(false);
        });
    });
}

async function playSegmentAudio(seg: ScriptSegmentDetailType): Promise<boolean> {
    if (!seg.clip_id) return false;

    playingId.value = seg.id;
    const blobURL = await fetchAudioBlob(seg.clip_id);
    if (!blobURL) {
        Message.warning("音频播放失败");
        return false;
    }
    return playBlobURL(blobURL);
}

async function togglePlayAudio(seg: ScriptSegmentDetailType) {
    if (playingId.value === seg.id) {
        stopAudio();
        return;
    }
    stopAudio();
    await playSegmentAudio(seg);
    playingId.value = null;
}

const prefetchCache = new Map<number, Promise<string | null>>();

function prefetchNext(segs: ScriptSegmentDetailType[], fromIdx: number) {
    for (let j = fromIdx + 1; j < segs.length; j++) {
        const next = segs[j];
        if (!next.clip_id) continue;
        if (!prefetchCache.has(next.clip_id)) {
            prefetchCache.set(next.clip_id, fetchAudioBlob(next.clip_id));
        }
        break;
    }
}

function scrollToSegment(segId: number) {
    const el = document.querySelector(`[data-seg-id="${segId}"]`);
    if (el) {
        el.scrollIntoView({ behavior: "smooth", block: "center" });
    }
}

function clearPrefetchCache() {
    prefetchCache.forEach((promise) => {
        promise.then((url) => {
            if (url) URL.revokeObjectURL(url);
        });
    });
    prefetchCache.clear();
}

async function toggleContinuousPlay(seg: ScriptSegmentDetailType) {
    if (continuousPlayingFromId.value === seg.id) {
        stopAudio();
        return;
    }

    stopAudio();
    clearPrefetchCache();
    continuousPlayingFromId.value = seg.id;

    const startIdx = segments.value.findIndex((s) => s.id === seg.id);
    if (startIdx === -1) return;

    for (let i = startIdx; i < segments.value.length; i++) {
        if (continuousPlayingFromId.value !== seg.id) break;

        const current = segments.value[i];
        if (!current.clip_id) continue;

        prefetchNext(segments.value, i);

        playingId.value = current.id;
        scrollToSegment(current.id);

        let blobURL: string | null;
        const cached = prefetchCache.get(current.clip_id);
        if (cached) {
            blobURL = await cached;
            prefetchCache.delete(current.clip_id);
        } else {
            blobURL = await fetchAudioBlob(current.clip_id);
        }

        if (continuousPlayingFromId.value !== seg.id) {
            if (blobURL) URL.revokeObjectURL(blobURL);
            break;
        }

        if (!blobURL) {
            Message.warning(`片段 #${current.segment_num} 音频加载失败，连播已停止`);
            break;
        }

        const ok = await playBlobURL(blobURL);
        if (!ok) {
            Message.warning(`片段 #${current.segment_num} 播放失败，连播已停止`);
            break;
        }
        if (continuousPlayingFromId.value !== seg.id) break;
    }

    clearPrefetchCache();
    if (continuousPlayingFromId.value === seg.id) {
        continuousPlayingFromId.value = null;
    }
    playingId.value = null;
}

function cleanupAudioResource() {
    if (currentAudioEl) {
        currentAudioEl.pause();
        currentAudioEl = null;
    }
    if (currentBlobURL) {
        URL.revokeObjectURL(currentBlobURL);
        currentBlobURL = null;
    }
}

function stopAudio() {
    cleanupAudioResource();
    playingId.value = null;
    continuousPlayingFromId.value = null;
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
        queued: "cyan",
        processing: "orangered",
        generated: "green",
        failed: "red",
        locked: "purple",
        cancelled: "orangered"
    };
    return map[status] || "gray";
}

function statusLabel(status: string) {
    const map: Record<string, string> = {
        raw: "原始",
        edited: "已编辑",
        queued: "队列中",
        processing: "生成中",
        generated: "已生成",
        failed: "生成失败",
        locked: "已锁定",
        cancelled: "已取消"
    };
    return map[status] || status;
}
</script>
<style lang="scss" scoped>
.script-editor-wrap {
    display: flex;
    height: calc(100vh - 310px);
    min-height: 400px;
}

.chapter-sidebar {
    width: 240px;
    flex-shrink: 0;
    border-right: 1px solid var(--color-border);
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.chapter-list {
    flex: 1;
    overflow-y: auto;
    padding: 0 12px 12px;
}

.chapter-batch-bar {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    border-bottom: 1px solid var(--color-border);
    flex-shrink: 0;
    background-color: var(--color-bg-2);
}

.chapter-batch-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    min-height: 24px;
}

.chapter-batch-actions {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    background-color: rgb(var(--primary-1));
    border-top: 1px solid var(--color-border-2);
}

.batch-hint {
    font-size: 12px;
    color: rgb(var(--primary-6));
    font-weight: 500;
    white-space: nowrap;
}

.batch-spacer {
    flex: 1;
}

.batch-slide-enter-active,
.batch-slide-leave-active {
    transition: all 0.2s ease;
    overflow: hidden;
}

.batch-slide-enter-from,
.batch-slide-leave-to {
    max-height: 0;
    padding-top: 0;
    padding-bottom: 0;
    opacity: 0;
}

.batch-slide-enter-to,
.batch-slide-leave-from {
    max-height: 40px;
    opacity: 1;
}

.chapter-item {
    display: flex;
    align-items: center;
    gap: 8px;
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

    &.splitting {
        border-left: 3px solid rgb(var(--orange-6));
    }
}

.chapter-checkbox {
    flex-shrink: 0;
}

.chapter-info {
    flex: 1;
    min-width: 0;
}

.chapter-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.chapter-meta {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    color: var(--color-text-3);
    margin-top: 2px;
}

.segment-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
}

.segment-scroll {
    flex: 1;
    overflow-y: auto;
    padding: 0 16px 16px;
}

.empty-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    flex: 1;
    color: var(--color-text-3);
}

.segment-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0px 16px 16px 12px;
    flex-shrink: 0;
    background-color: var(--color-bg-2);

    h3 {
        font-size: 16px;
        font-weight: 500;
        margin: 0;
    }
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

.play-popover {
    padding: 4px 0;
    min-width: 120px;
}

.play-popover-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    font-size: 13px;
    cursor: pointer;
    color: var(--color-text-1);
    transition: background-color 0.15s;

    &:hover {
        background-color: var(--color-fill-2);
    }
}

.character-placeholder {
    color: var(--color-text-3);
}
</style>
