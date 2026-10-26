<template>
    <a-modal
        :visible="visible"
        title="泛称角色归入路人池"
        width="680px"
        title-align="start"
        :align-center="false"
        top="10vh"
        :mask-closable="false"
        ok-text="确认归入"
        :ok-loading="submitting"
        :ok-button-props="{ disabled: loading || !preview || preview.merged_count === 0 }"
        @update:visible="(v: boolean) => emit('update:visible', v)"
        @open="loadPreview"
        @before-ok="handleOk"
    >
        <a-typography-paragraph type="secondary">
            将「老者」「店小二」「众人」等没有名字的泛称角色，按性别与年龄段合并到路人池角色，共享路人音色。主角级角色不会被处理。
        </a-typography-paragraph>
        <a-spin :loading="loading" style="width: 100%">
            <a-empty v-if="preview && preview.groups.length === 0" description="没有识别到需要归入路人池的泛称角色" />
            <template v-else-if="preview">
                <a-alert v-if="preview.skipped_names?.length" type="warning" style="margin-bottom: 12px">
                    以下角色有片段正在排队或生成，本次跳过：{{ preview.skipped_names.join("、") }}
                </a-alert>
                <a-typography-text>共 {{ preview.merged_count }} 个角色将被归入路人池：</a-typography-text>
                <a-descriptions :column="1" bordered size="small" style="margin-top: 8px">
                    <a-descriptions-item v-for="g in preview.groups" :key="g.pool_name">
                        <template #label>
                            {{ g.pool_name }}
                            <a-tag v-if="!g.pool_id" size="small" color="green" style="margin-left: 4px">新建</a-tag>
                        </template>
                        <a-space wrap>
                            <a-tag v-for="n in g.source_names" :key="n" size="small">{{ n }}</a-tag>
                        </a-space>
                    </a-descriptions-item>
                </a-descriptions>
            </template>
        </a-spin>
    </a-modal>
</template>
<script lang="ts" setup>
import { mergeGenericCharacters, CharacterMergeGenericResult } from "@/config/apis/character";

const props = defineProps<{
    projectId: number;
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: "update:visible", v: boolean): void;
    (e: "success", res: CharacterMergeGenericResult): void;
}>();

const loading = ref(false);
const submitting = ref(false);
const preview = ref<CharacterMergeGenericResult | null>(null);

async function loadPreview() {
    preview.value = null;
    loading.value = true;
    try {
        preview.value = await mergeGenericCharacters(props.projectId, true);
    } finally {
        loading.value = false;
    }
}

async function handleOk(done: (closed: boolean) => void) {
    submitting.value = true;
    try {
        const res = await mergeGenericCharacters(props.projectId, false);
        emit("success", res);
        done(true);
    } catch {
        done(false);
    } finally {
        submitting.value = false;
    }
}
</script>
