<template>
    <arco-modal v-model:visible="localVisible" :config="modalConfig">
        <div class="auto-voice-body">
            <a-tabs v-model:active-key="mode" class="auto-voice-tabs">
                <a-tab-pane key="generate" title="自动生成音色">
                    <div class="pane">
                        <p class="hint">
                            为未绑定声音的角色生成基础音色，任务在后台逐个执行，关闭弹窗不影响进度。
                        </p>

                        <template v-if="view === 'select'">
                            <div class="filters">
                                <div class="scope-row">
                                    <span>范围：</span>
                                    <a-checkbox-group v-model="scopes">
                                        <a-checkbox value="main">主角（{{ scopeCounts.main }}）</a-checkbox>
                                        <a-checkbox value="supporting">配角（{{ scopeCounts.supporting }}）</a-checkbox>
                                        <a-checkbox value="pool">路人池（{{ scopeCounts.pool }}）</a-checkbox>
                                        <a-checkbox value="minor">其他龙套（{{ scopeCounts.minor }}）</a-checkbox>
                                    </a-checkbox-group>
                                </div>
                                <a-checkbox v-model="confirmImportant">
                                    主角、配角、旁白先生成草稿，试听确认后再保存（其余角色直接保存）
                                </a-checkbox>
                            </div>
                            <div class="list-wrap">
                                <a-spin v-if="loadingCandidates" class="list-loading" />
                                <a-empty
                                    v-else-if="selectedCandidates.length === 0"
                                    description="没有待生成音色的角色"
                                />
                                <a-list
                                    v-else
                                    class="item-list"
                                    size="small"
                                    :data="selectedCandidates"
                                    :bordered="false"
                                    :split="false"
                                    :virtual-list-props="candidateListProps"
                                >
                                    <template #item="{ item }">
                                        <div class="item-row">
                                            <span class="item-name">{{ item.name }}</span>
                                            <a-tag size="small">{{ levelLabel(item.level) }}</a-tag>
                                            <a-tag v-if="willConfirm(item)" size="small" color="orangered">
                                                需试听确认
                                            </a-tag>
                                        </div>
                                    </template>
                                </a-list>
                            </div>
                            <div class="footer">
                                <a-typography-text type="secondary">
                                    共 {{ selectedCandidates.length }} 个角色
                                </a-typography-text>
                                <a-button
                                    type="primary"
                                    :loading="starting"
                                    :disabled="selectedCandidates.length === 0"
                                    @click="startBatch"
                                >
                                    开始生成（{{ selectedCandidates.length }}）
                                </a-button>
                            </div>
                        </template>

                        <template v-else-if="batch">
                            <a-space wrap>
                                <a-tag v-for="s in summary" :key="s.label" :color="s.color">
                                    {{ s.label }} {{ s.count }}
                                </a-tag>
                            </a-space>
                            <div class="list-wrap">
                                <a-list
                                    class="item-list"
                                    size="small"
                                    :data="batch.items"
                                    :bordered="false"
                                    :split="false"
                                    :virtual-list-props="batchListProps"
                                >
                                    <template #item="{ item }">
                                        <div class="item-row">
                                            <span class="item-name">
                                                {{ item.character_name || `角色 ${item.character_id}` }}
                                            </span>
                                            <a-tag size="small" :color="statusColor(item.status)">
                                                {{ statusLabel(item.status) }}
                                            </a-tag>
                                            <a-tooltip :content="itemDetail(item)" :disabled="!itemDetail(item)">
                                                <span class="item-detail">{{ itemDetail(item) }}</span>
                                            </a-tooltip>
                                            <a-space v-if="item.status === 'draft'" size="mini">
                                                <a-button
                                                    size="mini"
                                                    :loading="audioLoadingId === item.item_id"
                                                    @click="togglePlay(item)"
                                                >
                                                    {{ playingId === item.item_id ? "停止" : "试听" }}
                                                </a-button>
                                                <a-button
                                                    size="mini"
                                                    type="primary"
                                                    :loading="isActing(item.item_id)"
                                                    @click="apply(item)"
                                                >
                                                    保存
                                                </a-button>
                                                <a-button
                                                    size="mini"
                                                    :loading="isActing(item.item_id)"
                                                    @click="regenerate(item)"
                                                >
                                                    重新生成
                                                </a-button>
                                                <a-button size="mini" status="danger" @click="discard(item)">
                                                    放弃
                                                </a-button>
                                            </a-space>
                                            <a-button
                                                v-else-if="item.status === 'failed'"
                                                size="mini"
                                                :loading="isActing(item.item_id)"
                                                @click="regenerate(item)"
                                            >
                                                重试
                                            </a-button>
                                        </div>
                                    </template>
                                </a-list>
                            </div>
                            <div class="footer">
                                <a-typography-text type="secondary">
                                    {{ batch.running ? "后台生成中，每个角色需要几秒到十几秒" : "本批次已结束" }}
                                </a-typography-text>
                                <a-space>
                                    <a-button
                                        v-if="batch.running"
                                        status="danger"
                                        :loading="cancelling"
                                        @click="cancelBatch"
                                    >
                                        停止
                                    </a-button>
                                    <a-button v-else @click="backToSelect">选择其他角色</a-button>
                                </a-space>
                            </div>
                        </template>
                    </div>
                </a-tab-pane>
                <a-tab-pane key="match" title="匹配已有声音">
                    <div class="pane">
                        <p class="hint">
                            使用 LLM
                            把未绑定声音的角色匹配到项目里已有的声音配置上，已绑定声音的角色会自动跳过。
                        </p>
                        <div class="footer">
                            <span />
                            <a-button type="primary" :loading="matching" @click="runMatch">开始匹配</a-button>
                        </div>
                    </div>
                </a-tab-pane>
            </a-tabs>
        </div>
    </arco-modal>
</template>
<script lang="ts" setup>
import { Message } from "@arco-design/web-vue";
import { ArcoModal } from "@easyfe/admin-component";
import {
    getCharacterList,
    voiceMatchCharacters,
    startVoiceDesignBatch,
    getVoiceDesignBatch,
    cancelVoiceDesignBatch,
    applyVoiceDesignItem,
    regenerateVoiceDesignItem,
    discardVoiceDesignItem,
    fetchVoiceDesignItemAudio,
    CharacterDetailType,
    VoiceDesignBatchItem,
    VoiceDesignBatchItemStatus,
    VoiceDesignBatchResult
} from "@/config/apis/character";

const props = defineProps<{
    projectId: number;
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: "update:visible", v: boolean): void;
    (e: "changed"): void;
}>();

const localVisible = computed({
    get: () => props.visible,
    set: (v: boolean) => emit("update:visible", v)
});

const POOL_PREFIX = "路人-";
const NARRATOR_NAME = "旁白";
const POLL_INTERVAL = 2000;
const PAGE_SIZE = 1000;
const SCOPE_RANK: Record<VoiceScope, number> = { main: 0, supporting: 1, pool: 2, minor: 3 };

type VoiceScope = "main" | "supporting" | "pool" | "minor";

const STATUS_META: Record<VoiceDesignBatchItemStatus, { label: string; color: string }> = {
    queued: { label: "排队中", color: "gray" },
    running: { label: "生成中", color: "arcoblue" },
    draft: { label: "待确认", color: "orangered" },
    applied: { label: "已保存", color: "green" },
    skipped: { label: "已跳过", color: "gray" },
    failed: { label: "失败", color: "red" },
    discarded: { label: "已放弃", color: "gray" },
    cancelled: { label: "已取消", color: "gray" }
};
const LEVEL_LABELS: Record<string, string> = { main: "主角", supporting: "配角", minor: "龙套" };

const mode = ref<"generate" | "match">("generate");
const view = ref<"select" | "batch">("select");
const scopes = ref<string[]>(["main", "supporting", "pool"]);
const confirmImportant = ref(true);
const candidates = ref<CharacterDetailType[]>([]);
const batch = ref<VoiceDesignBatchResult | null>(null);
const loadingCandidates = ref(false);
const starting = ref(false);
const cancelling = ref(false);
const matching = ref(false);
const actingIds = ref<Set<number>>(new Set());

function isActing(id: number) {
    return actingIds.value.has(id);
}

function beginAct(id: number) {
    const next = new Set(actingIds.value);
    next.add(id);
    actingIds.value = next;
}

function endAct(id: number) {
    const next = new Set(actingIds.value);
    next.delete(id);
    actingIds.value = next;
}

const modalConfig = computed(() => ({
    title: "配置角色声音",
    width: "760px",
    top: "8vh",
    footer: false,
    onOpen,
    onClose
}));

function scopeOf(c: CharacterDetailType): VoiceScope {
    if (c.name.startsWith(POOL_PREFIX)) return "pool";
    if (c.level === "main" || c.level === "supporting") return c.level;
    // 旁白归入配角范围，保证默认会为旁白生成音色
    if (c.name === NARRATOR_NAME) return "supporting";
    return "minor";
}

// 与后端 isImportantForVoice 保持一致
function willConfirm(c: CharacterDetailType) {
    if (!confirmImportant.value || c.name.startsWith(POOL_PREFIX)) return false;
    return c.name === NARRATOR_NAME || c.level === "main" || c.level === "supporting";
}

function levelLabel(level: string) {
    return LEVEL_LABELS[level] || level;
}

function statusLabel(s: VoiceDesignBatchItemStatus) {
    return STATUS_META[s]?.label || s;
}

function statusColor(s: VoiceDesignBatchItemStatus) {
    return STATUS_META[s]?.color || "gray";
}

function itemDetail(item: VoiceDesignBatchItem) {
    return item.status === "failed" || item.status === "skipped" ? item.error_message : item.instruction;
}

const { height: windowHeight } = useWindowSize();
// top: 8vh + 标题/Tab/筛选项/底部按钮，列表只吃剩余高度
const listMaxHeight = computed(() => {
    const reserved = Math.round(windowHeight.value * 0.08) + 300;
    return Math.max(160, Math.min(240, windowHeight.value - reserved));
});
function listHeightOf(count: number) {
    if (count <= 0) return listMaxHeight.value;
    return Math.min(listMaxHeight.value, count * 36 + 4);
}
const candidateListProps = computed(() => ({ height: listHeightOf(selectedCandidates.value.length), itemKey: "id" }));
const batchListProps = computed(() => ({ height: listHeightOf(batch.value?.items.length || 0), itemKey: "item_id" }));

const selectedCandidates = computed(() => candidates.value.filter((c) => scopes.value.includes(scopeOf(c))));

const scopeCounts = computed(() => {
    const counts: Record<VoiceScope, number> = { main: 0, supporting: 0, pool: 0, minor: 0 };
    for (const c of candidates.value) counts[scopeOf(c)] += 1;
    return counts;
});

const summary = computed(() => {
    const items = batch.value?.items || [];
    const count = (statuses: VoiceDesignBatchItemStatus[]) => items.filter((i) => statuses.includes(i.status)).length;
    return [
        { label: "排队/生成中", count: count(["queued", "running"]), color: "arcoblue" },
        { label: "待确认", count: count(["draft"]), color: "orangered" },
        { label: "已保存", count: count(["applied"]), color: "green" },
        { label: "失败", count: count(["failed"]), color: "red" }
    ].filter((s) => s.count > 0);
});

// 列表接口单页最多 1000 条，翻页拉全量未绑定角色
async function loadCandidates() {
    loadingCandidates.value = true;
    try {
        const all: CharacterDetailType[] = [];
        let page = 1;
        let total = 0;
        do {
            const res = await getCharacterList({
                project_id: props.projectId,
                page,
                size: PAGE_SIZE,
                status: "0"
            });
            all.push(...(res.data || []));
            total = res.total || 0;
            page += 1;
        } while (all.length < total && page < 50);

        candidates.value = all
            .filter((c) => !c.voice_profile_id)
            .sort((a, b) => {
                const rank = SCOPE_RANK[scopeOf(a)] - SCOPE_RANK[scopeOf(b)];
                if (rank !== 0) return rank;
                return (a.sort_order || 0) - (b.sort_order || 0) || a.id - b.id;
            });
    } finally {
        loadingCandidates.value = false;
    }
}

let pollTimer: ReturnType<typeof setTimeout> | null = null;
let lastAppliedCount = 0;

function stopPolling() {
    if (pollTimer) {
        clearTimeout(pollTimer);
        pollTimer = null;
    }
}

async function refreshBatch() {
    const res = await getVoiceDesignBatch(props.projectId);
    batch.value = res;
    const applied = res.items.filter((i) => i.status === "applied").length;
    if (applied > lastAppliedCount) emit("changed");
    lastAppliedCount = applied;
    return res;
}

function schedulePoll() {
    stopPolling();
    if (!localVisible.value || !batch.value?.running) return;
    pollTimer = setTimeout(async () => {
        try {
            await refreshBatch();
        } finally {
            schedulePoll();
        }
    }, POLL_INTERVAL);
}

async function onOpen() {
    view.value = "select";
    batch.value = null;
    lastAppliedCount = 0;
    await Promise.all([loadCandidates(), refreshBatch()]);
    lastAppliedCount = batch.value?.items.filter((i) => i.status === "applied").length || 0;
    const hasDraft = batch.value?.items.some((i) => i.status === "draft");
    if (batch.value?.running || hasDraft) {
        view.value = "batch";
        schedulePoll();
    }
}

function onClose() {
    stopPolling();
    stopAudio();
}

onBeforeUnmount(onClose);

async function startBatch() {
    starting.value = true;
    try {
        const res = await startVoiceDesignBatch(props.projectId, {
            character_ids: selectedCandidates.value.map((c) => c.id),
            confirm_important: confirmImportant.value
        });
        batch.value = res;
        lastAppliedCount = 0;
        view.value = "batch";
        if (res.skipped?.length) Message.warning(`已跳过 ${res.skipped.length} 个角色：${res.skipped.join("；")}`);
        schedulePoll();
    } finally {
        starting.value = false;
    }
}

async function cancelBatch() {
    cancelling.value = true;
    try {
        const res = await cancelVoiceDesignBatch(props.projectId);
        Message.info(`已停止，取消 ${res.cancelled} 个排队中的角色`);
        await refreshBatch();
    } finally {
        cancelling.value = false;
    }
}

async function backToSelect() {
    stopAudio();
    view.value = "select";
    await loadCandidates();
}

async function apply(item: VoiceDesignBatchItem) {
    if (isActing(item.item_id)) return;
    beginAct(item.item_id);
    try {
        if (playingId.value === item.item_id) stopAudio();
        await applyVoiceDesignItem(item.item_id);
        Message.success(`已为「${item.character_name}」保存音色`);
        await refreshBatch();
    } finally {
        endAct(item.item_id);
    }
}

async function regenerate(item: VoiceDesignBatchItem) {
    if (isActing(item.item_id)) return;
    beginAct(item.item_id);
    try {
        if (playingId.value === item.item_id) stopAudio();
        await regenerateVoiceDesignItem(item.item_id);
        item.status = "queued";
        item.error_message = "";
        Message.success(`「${item.character_name}」已重新排队，将优先重试`);
        await refreshBatch();
        schedulePoll();
    } finally {
        endAct(item.item_id);
    }
}

async function discard(item: VoiceDesignBatchItem) {
    if (playingId.value === item.item_id) stopAudio();
    await discardVoiceDesignItem(item.item_id);
    await refreshBatch();
}

const playingId = ref<number | null>(null);
const audioLoadingId = ref<number | null>(null);
let audioEl: HTMLAudioElement | null = null;
let audioURL: string | null = null;

function stopAudio() {
    audioEl?.pause();
    audioEl = null;
    if (audioURL) URL.revokeObjectURL(audioURL);
    audioURL = null;
    playingId.value = null;
    audioLoadingId.value = null;
}

async function togglePlay(item: VoiceDesignBatchItem) {
    if (playingId.value === item.item_id) {
        stopAudio();
        return;
    }
    stopAudio();
    audioLoadingId.value = item.item_id;
    try {
        const url = await fetchVoiceDesignItemAudio(item.item_id);
        if (audioLoadingId.value !== item.item_id) {
            URL.revokeObjectURL(url);
            return;
        }
        audioURL = url;
        audioEl = new Audio(url);
        audioEl.addEventListener("ended", stopAudio);
        playingId.value = item.item_id;
        audioLoadingId.value = null;
        await audioEl.play();
    } catch {
        Message.warning("音频加载失败");
        stopAudio();
    }
}

async function runMatch() {
    matching.value = true;
    try {
        const res = await voiceMatchCharacters(props.projectId);
        const parts = [`匹配完成：成功 ${res.matched_count} 个`];
        if (res.skipped_count > 0) parts.push(`跳过 ${res.skipped_count} 个（已绑定）`);
        if (res.failed_count > 0) parts.push(`未匹配 ${res.failed_count} 个`);
        Message.success(parts.join("，"));
        emit("changed");
        await loadCandidates();
    } finally {
        matching.value = false;
    }
}
</script>
<style lang="scss" scoped>
.auto-voice-tabs {
    :deep(.arco-tabs-content) {
        padding-top: 8px;
    }
}

.pane {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.hint {
    flex-shrink: 0;
    margin: 0;
    color: var(--color-text-3);
    font-size: 13px;
    line-height: 1.5;
}

.filters {
    display: flex;
    flex-shrink: 0;
    flex-direction: column;
    gap: 8px;
}

.scope-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
}

.list-wrap {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
}

.list-loading {
    padding: 24px 0;
}

.item-list {
    width: 100%;
    height: 100%;

    :deep(.arco-list-content) {
        border: none;
    }
}

.item-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    border-bottom: 1px solid var(--color-border-1);

    &:last-child {
        border-bottom: none;
    }
}

.item-name {
    flex: 0 0 120px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.item-detail {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--color-text-3);
    font-size: 12px;
}

.footer {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: space-between;
    padding-top: 4px;
}
</style>
