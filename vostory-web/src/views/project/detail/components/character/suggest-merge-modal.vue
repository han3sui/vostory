<template>
    <arco-modal v-model:visible="localVisible" :config="modalConfig" :ok="handleOk">
        <a-typography-paragraph type="secondary">
            按括号别名、不具名、甲乙丙编号和泛称扫描重复角色。请勾选要合并的分组，主角不会被当作被合并项。
        </a-typography-paragraph>
        <a-spin :loading="loading" style="width: 100%">
            <a-empty v-if="preview && preview.groups.length === 0" description="没有扫描到可自动合并的角色" />
            <template v-else-if="preview">
                <a-space style="margin-bottom: 12px">
                    <a-button size="mini" @click="selectAll">全选</a-button>
                    <a-button size="mini" @click="selectedKeys = []">取消全选</a-button>
                    <a-typography-text type="secondary">
                        已选 {{ selectedSourceCount }} / {{ preview.merged_count }} 个角色
                    </a-typography-text>
                </a-space>
                <a-checkbox-group v-model="selectedKeys" direction="vertical" style="width: 100%">
                    <div v-for="(g, idx) in preview.groups" :key="groupKey(g, idx)" class="suggest-group">
                        <a-checkbox :value="groupKey(g, idx)">
                            <span class="suggest-title">
                                合并到【{{ g.target_name }}】
                                <a-tag v-if="g.target_kind === 'pool'" size="small" color="orangered">路人池</a-tag>
                                <a-tag size="small">{{ ruleLabel(g.rule) }}</a-tag>
                                <a-typography-text type="secondary">{{ g.source_ids.length }} 个</a-typography-text>
                            </span>
                        </a-checkbox>
                        <div class="suggest-reason">{{ g.reason }}</div>
                        <a-space wrap class="suggest-names">
                            <a-tag v-for="n in visibleNames(g)" :key="n" size="small">{{ n }}</a-tag>
                            <a-tag v-if="hiddenNameCount(g) > 0" size="small" color="gray">
                                等 {{ hiddenNameCount(g) }} 个
                            </a-tag>
                        </a-space>
                    </div>
                </a-checkbox-group>
            </template>
        </a-spin>
    </arco-modal>
</template>
<script lang="ts" setup>
import { Message } from "@arco-design/web-vue";
import { ArcoModal } from "@easyfe/admin-component";
import {
    suggestMergeCharacters,
    CharacterMergeSuggestGroup,
    CharacterMergeSuggestResult
} from "@/config/apis/character";

const props = defineProps<{
    projectId: number;
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: "update:visible", v: boolean): void;
    (e: "success", res: CharacterMergeSuggestResult): void;
}>();

const localVisible = computed({
    get: () => props.visible,
    set: (v: boolean) => emit("update:visible", v)
});

const RULE_LABELS: Record<string, string> = {
    paren_alias: "括号别名",
    unnamed_to_named: "不具名对应",
    unnamed_pool: "不具名路人",
    numbered_pool: "编号路人",
    generic_pool: "泛称路人"
};

const loading = ref(false);
const preview = ref<CharacterMergeSuggestResult | null>(null);
const selectedKeys = ref<string[]>([]);

const modalConfig = computed(() => ({
    title: "智能合并角色",
    width: "760px",
    top: "8vh",
    okText: "确认合并勾选项",
    okButtonProps: { disabled: loading.value || selectedKeys.value.length === 0 },
    onOpen: loadPreview
}));

function groupKey(g: CharacterMergeSuggestGroup, idx: number) {
    return `${g.target_kind}:${g.target_id || g.target_name}:${idx}`;
}

function ruleLabel(rule: string) {
    return RULE_LABELS[rule] || rule;
}

function visibleNames(g: CharacterMergeSuggestGroup) {
    return (g.source_names || []).slice(0, 16);
}

function hiddenNameCount(g: CharacterMergeSuggestGroup) {
    return Math.max(0, (g.source_names || []).length - 16);
}

const selectedSourceCount = computed(() => {
    if (!preview.value) return 0;
    return preview.value.groups.reduce((sum, g, idx) => {
        return selectedKeys.value.includes(groupKey(g, idx)) ? sum + g.source_ids.length : sum;
    }, 0);
});

function selectAll() {
    if (!preview.value) return;
    selectedKeys.value = preview.value.groups.map((g, idx) => groupKey(g, idx));
}

async function loadPreview() {
    preview.value = null;
    selectedKeys.value = [];
    loading.value = true;
    try {
        preview.value = await suggestMergeCharacters(props.projectId, { dry_run: true });
        selectAll();
    } finally {
        loading.value = false;
    }
}

async function handleOk() {
    if (!preview.value) throw new Error("no preview");
    const groups = preview.value.groups.filter((g, idx) => selectedKeys.value.includes(groupKey(g, idx)));
    if (groups.length === 0) {
        Message.warning("请至少勾选一组");
        throw new Error("no selection");
    }
    const res = await suggestMergeCharacters(props.projectId, { dry_run: false, groups });
    emit("success", res);
}
</script>
<style lang="scss" scoped>
.suggest-group {
    width: 100%;
    padding: 8px 0 12px;
    border-bottom: 1px solid var(--color-border-2);
}
.suggest-title {
    display: inline-flex;
    align-items: center;
    gap: 6px;
}
.suggest-reason,
.suggest-names {
    margin-left: 24px;
    margin-top: 6px;
}
.suggest-reason {
    color: var(--color-text-3);
    font-size: 12px;
}
</style>
