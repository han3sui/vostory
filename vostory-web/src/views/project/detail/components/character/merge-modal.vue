<template>
    <a-modal
        :visible="visible"
        title="合并角色"
        width="640px"
        title-align="start"
        :align-center="false"
        top="10vh"
        :mask-closable="false"
        ok-text="确认合并"
        :ok-loading="submitting"
        :ok-button-props="{ disabled: !canSubmit }"
        @update:visible="(v: boolean) => emit('update:visible', v)"
        @open="onOpen"
        @before-ok="handleOk"
    >
        <a-alert type="info">
            被合并角色的台词会迁移到目标角色，其名称和别名并入目标角色的别名，之后切分章节时会自动识别为目标角色。被合并角色将从列表中移除。
        </a-alert>
        <div style="margin-top: 8px">
            <a-link @click="emit('open-generic')">「老者」「众人」等泛称角色？一键归入路人池</a-link>
        </div>
        <a-form :model="form" layout="vertical" style="margin-top: 16px">
            <a-form-item label="被合并的角色（重复的称呼）" required>
                <a-select
                    v-model="form.sourceIds"
                    multiple
                    allow-search
                    allow-clear
                    :filter-option="false"
                    :loading="searching"
                    placeholder="搜索角色名称，可多选"
                    @search="onSearch"
                    @popup-visible-change="onPopupVisibleChange"
                >
                    <a-option v-for="c in sourceOptions" :key="c.id" :value="c.id" :label="c.name">
                        <character-option-label :character="c" />
                    </a-option>
                </a-select>
            </a-form-item>
            <a-form-item label="合并到（保留的角色）" required>
                <a-select
                    v-model="form.targetId"
                    allow-search
                    allow-clear
                    :filter-option="false"
                    :loading="searching"
                    placeholder="搜索角色名称"
                    @search="onSearch"
                    @popup-visible-change="onPopupVisibleChange"
                >
                    <a-option v-for="c in targetOptions" :key="c.id" :value="c.id" :label="c.name">
                        <character-option-label :character="c" />
                    </a-option>
                </a-select>
            </a-form-item>
        </a-form>
        <div v-if="targetCharacter && form.sourceIds.length > 0">
            <a-typography-text type="secondary">合并后「{{ targetCharacter.name }}」的别名：</a-typography-text>
            <a-space wrap style="margin-top: 8px">
                <a-tag v-for="a in mergedAliases" :key="a" size="small">{{ a }}</a-tag>
            </a-space>
        </div>
    </a-modal>
</template>
<script lang="ts" setup>
import type { PropType } from "vue";
import { Message } from "@arco-design/web-vue";
import { debounce } from "lodash-es";
import { getCharacterList, mergeCharacters, CharacterDetailType, CharacterMergeResult } from "@/config/apis/character";

const props = defineProps<{
    projectId: number;
    visible: boolean;
    presetSource?: CharacterDetailType | null;
}>();

const emit = defineEmits<{
    (e: "update:visible", v: boolean): void;
    (e: "success", res: CharacterMergeResult, target: CharacterDetailType): void;
    (e: "open-generic"): void;
}>();

const LEVEL_LABELS: Record<string, string> = { main: "主角", supporting: "配角", minor: "龙套" };
const GENDER_LABELS: Record<string, string> = { male: "男", female: "女" };
const NARRATOR_NAME = "旁白";

const CharacterOptionLabel = defineComponent({
    props: { character: { type: Object as PropType<CharacterDetailType>, required: true } },
    setup(p) {
        return () => {
            const c = p.character;
            const meta = [LEVEL_LABELS[c.level] || c.level, GENDER_LABELS[c.gender]].filter(Boolean).join("·");
            const aliases = (c.aliases || []).join("、");
            return h("span", null, [
                c.name,
                h("span", { style: "color: var(--color-text-3); margin-left: 8px; font-size: 12px" }, [
                    meta,
                    aliases ? `｜别名：${aliases}` : ""
                ])
            ]);
        };
    }
});

const form = reactive<{ sourceIds: number[]; targetId: number | undefined }>({ sourceIds: [], targetId: undefined });
const submitting = ref(false);
const searching = ref(false);
const searchResults = ref<CharacterDetailType[]>([]);
// 已选中的角色需要始终在 options 中，否则搜索后选中项会丢失名称
const characterCache = reactive(new Map<number, CharacterDetailType>());

const selectedSources = computed(
    () => form.sourceIds.map((id) => characterCache.get(id)).filter(Boolean) as CharacterDetailType[]
);
const targetCharacter = computed(() => (form.targetId ? characterCache.get(form.targetId) : undefined));

function withSelected(list: CharacterDetailType[], selected: CharacterDetailType[]) {
    const ids = new Set(list.map((c) => c.id));
    return [...selected.filter((c) => !ids.has(c.id)), ...list];
}

const sourceOptions = computed(() =>
    withSelected(searchResults.value, selectedSources.value).filter(
        (c) => c.id !== form.targetId && c.name !== NARRATOR_NAME
    )
);
const targetOptions = computed(() =>
    withSelected(searchResults.value, targetCharacter.value ? [targetCharacter.value] : []).filter(
        (c) => !form.sourceIds.includes(c.id)
    )
);

const canSubmit = computed(() => !!form.targetId && form.sourceIds.length > 0);

const mergedAliases = computed(() => {
    const target = targetCharacter.value;
    if (!target) return [];
    const seen = new Set([target.name.toLowerCase()]);
    const result: string[] = [];
    const add = (name: string) => {
        const v = name.trim();
        if (!v || seen.has(v.toLowerCase())) return;
        seen.add(v.toLowerCase());
        result.push(v);
    };
    (target.aliases || []).forEach(add);
    selectedSources.value.forEach((c) => {
        add(c.name);
        (c.aliases || []).forEach(add);
    });
    return result;
});

let searchSeq = 0;
async function search(keyword: string) {
    const seq = ++searchSeq;
    searching.value = true;
    try {
        const res = await getCharacterList({ project_id: props.projectId, name: keyword, page: 1, size: 50 });
        // 只采用最后一次请求的结果，避免先发后到的响应覆盖当前搜索
        if (seq !== searchSeq) return;
        searchResults.value = res.data || [];
        searchResults.value.forEach((c) => characterCache.set(c.id, c));
    } finally {
        if (seq === searchSeq) searching.value = false;
    }
}
const onSearch = debounce((keyword: string) => search(keyword), 300);

// 两个下拉共用搜索结果，打开下拉时重新拉取，避免沿用另一个下拉的过滤结果
function onPopupVisibleChange(visible: boolean) {
    if (visible) {
        onSearch.cancel();
        search("");
    }
}

function onOpen() {
    form.sourceIds = [];
    form.targetId = undefined;
    characterCache.clear();
    if (props.presetSource) {
        characterCache.set(props.presetSource.id, props.presetSource);
        form.sourceIds = [props.presetSource.id];
    }
    search("");
}

async function handleOk(done: (closed: boolean) => void) {
    const target = targetCharacter.value;
    if (!target || form.sourceIds.length === 0) {
        Message.warning("请选择被合并的角色和目标角色");
        done(false);
        return;
    }
    submitting.value = true;
    try {
        const res = await mergeCharacters({ target_id: target.id, source_ids: form.sourceIds });
        emit("success", res, target);
        done(true);
    } catch {
        done(false);
    } finally {
        submitting.value = false;
    }
}
</script>
