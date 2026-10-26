<template>
    <arco-modal v-model:visible="localVisible" :config="modalConfig">
        <a-tabs v-model:active-key="mode" :disabled="running">
            <a-tab-pane key="generate" title="自动生成音色">
                <a-typography-paragraph type="secondary">
                    为未绑定声音的角色逐个生成基础音色：LLM 根据角色信息撰写音色描述 → Breeze-TTS-2 Voice Design
                    生成参考音频 → 保存为声音配置并绑定。每个角色需要几秒到十几秒，可随时停止。
                </a-typography-paragraph>
                <a-space style="margin-bottom: 12px">
                    <span>范围：</span>
                    <a-checkbox-group v-model="scopes" :disabled="running">
                        <a-checkbox value="main">主角</a-checkbox>
                        <a-checkbox value="supporting">配角</a-checkbox>
                        <a-checkbox value="pool">路人池</a-checkbox>
                        <a-checkbox value="minor">其他龙套</a-checkbox>
                    </a-checkbox-group>
                </a-space>
                <a-spin :loading="loadingCandidates" style="width: 100%">
                    <a-empty v-if="items.length === 0" description="没有待生成音色的角色" />
                    <div v-else class="item-list">
                        <div v-for="item in items" :key="item.character.id" class="item-row">
                            <span class="item-name">{{ item.character.name }}</span>
                            <a-tag size="small" :color="stateColor(item.state)">{{ stateLabel(item.state) }}</a-tag>
                            <span class="item-detail">{{ item.detail }}</span>
                        </div>
                    </div>
                </a-spin>
                <div class="footer">
                    <a-typography-text type="secondary">
                        共 {{ items.length }} 个，成功 {{ countBy("done") }}，失败 {{ countBy("failed") }}
                    </a-typography-text>
                    <a-space>
                        <a-button v-if="running" status="danger" @click="stopRequested = true">停止</a-button>
                        <a-button v-else type="primary" :disabled="pendingItems.length === 0" @click="runGenerate">
                            开始生成（{{ pendingItems.length }}）
                        </a-button>
                    </a-space>
                </div>
            </a-tab-pane>
            <a-tab-pane key="match" title="匹配已有声音">
                <a-typography-paragraph type="secondary">
                    使用 LLM
                    根据角色描述和声音配置描述，把未绑定声音的角色匹配到项目里已有的声音配置上，已绑定声音的角色会自动跳过。
                </a-typography-paragraph>
                <div class="footer">
                    <span />
                    <a-button type="primary" :loading="matching" @click="runMatch">开始匹配</a-button>
                </div>
            </a-tab-pane>
        </a-tabs>
    </arco-modal>
</template>
<script lang="ts" setup>
import { Message } from "@arco-design/web-vue";
import { ArcoModal } from "@easyfe/admin-component";
import { getCharacterList, autoVoiceDesign, voiceMatchCharacters, CharacterDetailType } from "@/config/apis/character";

type ItemState = "pending" | "running" | "done" | "failed";
type Item = { character: CharacterDetailType; state: ItemState; detail: string };

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

const mode = ref<"generate" | "match">("generate");
const scopes = ref<string[]>(["main", "supporting", "pool"]);
const candidates = ref<CharacterDetailType[]>([]);
const results = reactive(new Map<number, { state: ItemState; detail: string }>());
const loadingCandidates = ref(false);
const running = ref(false);
const stopRequested = ref(false);
const matching = ref(false);

const modalConfig = computed(() => ({
    title: "配置角色声音",
    width: "720px",
    top: "8vh",
    closable: !running.value,
    footer: false,
    onOpen
}));

function scopeOf(c: CharacterDetailType) {
    if (c.name.startsWith(POOL_PREFIX)) return "pool";
    if (c.level === "main" || c.level === "supporting") return c.level;
    // 旁白归入配角范围，保证默认会为旁白生成音色
    if (c.name === "旁白") return "supporting";
    return "minor";
}

const items = computed<Item[]>(() =>
    candidates.value
        .filter((c) => scopes.value.includes(scopeOf(c)))
        .map((c) => {
            const r = results.get(c.id);
            return { character: c, state: r?.state || "pending", detail: r?.detail || "" };
        })
);
const pendingItems = computed(() => items.value.filter((i) => i.state === "pending" || i.state === "failed"));

function countBy(state: ItemState) {
    return items.value.filter((i) => i.state === state).length;
}

function stateLabel(s: ItemState) {
    return { pending: "待生成", running: "生成中", done: "已完成", failed: "失败" }[s];
}
function stateColor(s: ItemState) {
    return { pending: "gray", running: "arcoblue", done: "green", failed: "red" }[s];
}

async function onOpen() {
    results.clear();
    stopRequested.value = false;
    loadingCandidates.value = true;
    try {
        const res = await getCharacterList({ project_id: props.projectId, page: 1, size: 1000 });
        candidates.value = (res.data || []).filter((c) => !c.voice_profile_id && c.status === "0");
    } finally {
        loadingCandidates.value = false;
    }
}

async function runGenerate() {
    running.value = true;
    stopRequested.value = false;
    let changed = false;
    try {
        for (const item of pendingItems.value) {
            if (stopRequested.value) break;
            const id = item.character.id;
            results.set(id, { state: "running", detail: "" });
            try {
                const res = await autoVoiceDesign(id);
                results.set(id, { state: "done", detail: res.instruction });
                changed = true;
            } catch (e: any) {
                results.set(id, { state: "failed", detail: e?.message || "生成失败" });
            }
        }
    } finally {
        running.value = false;
        if (changed) emit("changed");
    }
    Message.info(`音色生成结束：成功 ${countBy("done")} 个，失败 ${countBy("failed")} 个`);
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
        await onOpen();
    } finally {
        matching.value = false;
    }
}
</script>
<style lang="scss" scoped>
.item-list {
    max-height: 45vh;
    overflow-y: auto;
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
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
    align-items: center;
    justify-content: space-between;
    margin-top: 16px;
}
</style>
