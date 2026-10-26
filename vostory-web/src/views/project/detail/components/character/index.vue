<template>
    <div>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #switchSlot>
                <a-table-column title="启用">
                    <template #cell="{ record }">
                        <a-switch
                            :disabled="!hasPermission('character:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleToggle(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
            <template #levelSlot>
                <a-table-column title="层级">
                    <template #cell="{ record }">
                        <a-tag :color="levelColor(record.level)" size="small">
                            {{ levelLabel(record.level) }}
                        </a-tag>
                    </template>
                </a-table-column>
            </template>
            <template #aliasSlot>
                <a-table-column title="别名">
                    <template #cell="{ record }">
                        <a-space wrap>
                            <a-tag v-for="a in record.aliases || []" :key="a" size="small">{{ a }}</a-tag>
                        </a-space>
                    </template>
                </a-table-column>
            </template>
            <template #previewSlot>
                <a-table-column title="试听">
                    <template #cell="{ record }">
                        <a-button v-if="record.voice_profile_id" type="text" size="mini" @click="togglePreview(record)">
                            <template #icon>
                                <icon-play-arrow v-if="previewingId !== record.voice_profile_id" />
                                <icon-pause v-else />
                            </template>
                            {{ previewingId === record.voice_profile_id ? "停止" : "试听" }}
                        </a-button>
                        <span v-else style="color: var(--color-text-3)">-</span>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
        <a-modal
            v-model:visible="showSmartImportModal"
            title="智能录入角色"
            :ok-loading="smartImporting"
            ok-text="开始识别"
            :ok-button-props="{ disabled: !smartImportText.trim() }"
            width="680px"
            :align-center="false"
            top="10vh"
            title-align="start"
            :mask-closable="false"
            @before-ok="handleSmartImport"
        >
            <a-typography-paragraph>
                粘贴角色介绍文字，LLM
                将自动识别角色名称、性别、层级、性格描述等信息并录入角色库。已存在的角色会自动跳过。
            </a-typography-paragraph>
            <a-textarea
                v-model="smartImportText"
                placeholder="请粘贴角色介绍文字，例如：&#10;张三：男，主角，性格沉稳冷静，外表俊朗&#10;李四：女，配角，活泼开朗，是张三的青梅竹马"
                :auto-size="{ minRows: 8, maxRows: 16 }"
                allow-clear
            />
        </a-modal>
        <merge-modal
            v-model:visible="showMergeModal"
            :project-id="projectId"
            :preset-source="mergePresetSource"
            @success="onMergeSuccess"
            @open-generic="openGenericPool"
        />
        <voice-design-modal
            v-model:visible="showVoiceDesignModal"
            :character="voiceDesignCharacter"
            @success="onVoiceDesigned"
        />
        <auto-voice-modal v-model:visible="showAutoVoiceModal" :project-id="projectId" @changed="onAutoVoiceChanged" />
        <generic-pool-modal
            v-model:visible="showGenericPoolModal"
            :project-id="projectId"
            @success="onGenericPoolSuccess"
        />
    </div>
</template>
<script lang="ts" setup>
import { Message, Modal } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import {
    getCharacterList,
    addCharacter,
    updateCharacter,
    deleteCharacter,
    enableCharacter,
    disableCharacter,
    extractCharacters,
    extractFromText,
    CharacterDetailType
} from "@/config/apis/character";
import { getVoiceProfilesByProject, VoiceProfileOptionType } from "@/config/apis/voice-profile";
import { fetchReferenceAudioBlob } from "@/config/apis/upload";
import { regenerateByCharacter } from "@/config/apis/tts";
import { IconPlayArrow, IconPause } from "@arco-design/web-vue/es/icon";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";
import MergeModal from "./merge-modal.vue";
import GenericPoolModal from "./generic-pool-modal.vue";
import VoiceDesignModal from "./voice-design-modal.vue";
import AutoVoiceModal from "./auto-voice-modal.vue";
import type {
    CharacterMergeResult,
    CharacterMergeGenericResult,
    VoiceDesignApplyResult
} from "@/config/apis/character";

const props = defineProps<{ projectId: number }>();

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

function levelLabel(l: string) {
    return LEVELS.find((v) => v.value === l)?.label || l;
}
function levelColor(l: string) {
    const map: Record<string, string> = { main: "red", supporting: "blue", minor: "gray" };
    return map[l] || "gray";
}

const table = ref();
const filterData = ref<Record<string, any>>({});
const extracting = ref(false);
const showSmartImportModal = ref(false);
const smartImportText = ref("");
const smartImporting = ref(false);
const voiceProfileOptions = ref<VoiceProfileOptionType[]>([]);

async function loadVoiceProfiles() {
    if (!props.projectId) return;
    voiceProfileOptions.value = await getVoiceProfilesByProject(props.projectId);
}
onMounted(loadVoiceProfiles);
watch(() => props.projectId, loadVoiceProfiles);

const getFilterConfig = computed(() => {
    return [
        formHelper.input("角色名称", "name", { span: 5, debounce: 500 }),
        formHelper.select("层级", "level", LEVELS, { span: 5 }),
        formHelper.select("性别", "gender", GENDERS, { span: 5 })
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: { rowKey: "id" },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("角色名称", "name"),
            tableHelper.slot("aliasSlot"),
            tableHelper.status("性别", "gender", (item: any) => {
                const found = GENDERS.find((g) => g.value === item.gender);
                return { text: found?.label || item.gender, status: "normal" };
            }),
            tableHelper.slot("levelSlot"),
            tableHelper.default("描述", "description"),
            tableHelper.default("声音配置", "voice_profile_name"),
            tableHelper.slot("previewSlot"),
            tableHelper.slot("switchSlot"),
            tableHelper.date("更新时间", "updated_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns(
                "操作",
                [
                    {
                        label: "编辑",
                        if: () => hasPermission("character:edit"),
                        handler: onEdit
                    },
                    {
                        label: "设计音色",
                        if: () => hasPermission("character:voice_design"),
                        handler: (row: Record<string, any>) => openVoiceDesign(row as CharacterDetailType)
                    },
                    {
                        label: "合并到",
                        if: () => hasPermission("character:merge"),
                        handler: (row: Record<string, any>) => openMerge(row as CharacterDetailType)
                    },
                    {
                        label: "重新生成配音",
                        if: (row: Record<string, any>) => hasPermission("character:edit") && !!row.voice_profile_id,
                        handler: handleRegenerate
                    },
                    {
                        label: "删除",
                        status: "danger",
                        if: () => hasPermission("character:remove"),
                        handler(row: Record<string, any>) {
                            Modal.confirm({
                                title: "删除",
                                content: `确认删除角色【${row.name}】？`,
                                onBeforeOk: async () => {
                                    await deleteCharacter(row.id);
                                    table.value.refresh();
                                }
                            });
                        }
                    }
                ],
                {
                    width: 300
                }
            )
        ],
        trBtns: [
            {
                label: "合并角色",
                type: "outline",
                if: () => hasPermission("character:merge"),
                handler: () => openMerge(null)
            },
            {
                label: "配置声音",
                type: "outline",
                if: () => hasPermission("character:voice_match") || hasPermission("character:voice_design_auto"),
                handler: () => {
                    showAutoVoiceModal.value = true;
                }
            },
            {
                label: "智能录入",
                type: "outline",
                loading: smartImporting.value,
                if: () => hasPermission("character:add"),
                handler: () => {
                    smartImportText.value = "";
                    showSmartImportModal.value = true;
                }
            },
            {
                label: "添加角色",
                type: "primary",
                if: () => hasPermission("character:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getCharacterList,
        params: { project_id: props.projectId, ...filterData.value }
    };
});

async function onEdit(v: Record<string, any> | null) {
    await loadVoiceProfiles();
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑角色" : "添加角色",
            width: "650px"
        },
        value: tempValue || { project_id: props.projectId, status: "0", level: "main", gender: "unknown", aliases: [] },
        formConfig: [
            formHelper.input("角色名称", "name", {
                rules: [ruleHelper.require("请输入角色名称")]
            }),
            formHelper.select("性别", "gender", GENDERS),
            formHelper.select("层级", "level", LEVELS),
            formHelper.select(
                "声音配置",
                "voice_profile_id",
                voiceProfileOptions.value.map((v) => ({ label: v.name, value: v.id })),
                {
                    allowClear: true,
                    allowSearch: true,
                    placeholder: "选择声音配置（用于 TTS 合成）",
                    virtualListProps: { height: 200 }
                }
            ),
            formHelper.textarea("角色描述", "description", { autoSize: true }),
            formHelper.radio(
                "状态",
                "status",
                [
                    { label: "正常", value: "0" },
                    { label: "停用", value: "1" }
                ],
                { type: "radio", rules: [ruleHelper.require("请选择")] }
            ),
            formHelper.inputNumber("排序", "sort_order")
        ],
        ok: async (data: any) => {
            if (data.voice_profile_id === "" || data.voice_profile_id === undefined) {
                data.voice_profile_id = null;
            }
            if (tempValue) {
                await updateCharacter(data);
            } else {
                await addCharacter(data);
            }
            table.value.refresh();
        }
    });
}

const showMergeModal = ref(false);
const mergePresetSource = ref<CharacterDetailType | null>(null);
const showGenericPoolModal = ref(false);

function openMerge(row: CharacterDetailType | null) {
    mergePresetSource.value = row;
    showMergeModal.value = true;
}

function openGenericPool() {
    if (!hasPermission("character:merge_generic")) {
        Message.warning("没有归入路人池的权限");
        return;
    }
    showMergeModal.value = false;
    showGenericPoolModal.value = true;
}

function onMergeSuccess(res: CharacterMergeResult, target: CharacterDetailType) {
    Message.success(`已将 ${res.merged_count} 个角色合并到【${target.name}】，迁移 ${res.affected_segments} 个片段`);
    table.value.refresh();
    if (res.need_regenerate > 0) {
        Modal.confirm({
            title: "音色已变化",
            content: `有 ${res.need_regenerate} 个已生成的片段合并后音色发生变化，是否立即重新生成【${target.name}】的配音？已锁定的片段将被跳过。`,
            okText: "重新生成",
            cancelText: "稍后",
            onBeforeOk: async () => {
                const r = await regenerateByCharacter(target.id);
                Message.success(`已提交 ${r.total_count} 个片段`);
            }
        });
    }
}

function onGenericPoolSuccess(res: CharacterMergeGenericResult) {
    const parts = [`已将 ${res.merged_count} 个泛称角色归入路人池，迁移 ${res.affected_segments} 个片段`];
    if (res.need_regenerate > 0) {
        parts.push(`其中 ${res.need_regenerate} 个已生成片段音色变化，可在路人角色上重新生成配音`);
    }
    Message.success(parts.join("；"));
    table.value.refresh();
}

const showAutoVoiceModal = ref(false);
const showVoiceDesignModal = ref(false);
const voiceDesignCharacter = ref<CharacterDetailType | null>(null);

function openVoiceDesign(row: CharacterDetailType) {
    voiceDesignCharacter.value = row;
    showVoiceDesignModal.value = true;
}

function onVoiceDesigned(res: VoiceDesignApplyResult, character: CharacterDetailType) {
    Message.success(`已为【${character.name}】生成并绑定音色「${res.voice_profile_name}」`);
    loadVoiceProfiles();
    table.value.refresh();
    if (character.voice_profile_id) {
        Modal.confirm({
            title: "音色已更换",
            content: `是否立即用新音色重新生成【${character.name}】的配音？已锁定的片段将被跳过。`,
            okText: "重新生成",
            cancelText: "稍后",
            onBeforeOk: async () => {
                const r = await regenerateByCharacter(character.id);
                Message.success(`已提交 ${r.total_count} 个片段`);
            }
        });
    }
}

function onAutoVoiceChanged() {
    loadVoiceProfiles();
    table.value.refresh();
}

async function handleExtract() {
    Modal.confirm({
        title: "智能提取角色",
        content: "将使用 LLM 从全书文本中自动识别角色，已存在的角色会自动跳过。是否继续？",
        okText: "开始提取",
        cancelText: "取消",
        onBeforeOk: async () => {
            extracting.value = true;
            try {
                const res = await extractCharacters(props.projectId);
                Message.success(
                    `提取完成：识别 ${res.extracted_count} 个角色，新增 ${res.new_count} 个，跳过 ${res.skipped_count} 个`
                );
                table.value.refresh();
            } finally {
                extracting.value = false;
            }
        }
    });
}

async function handleSmartImport(done: (closed: boolean) => void) {
    if (!smartImportText.value.trim()) {
        Message.warning("请输入角色描述文字");
        done(false);
        return;
    }
    smartImporting.value = true;
    try {
        const res = await extractFromText(props.projectId, smartImportText.value);
        Message.success(
            `识别完成：发现 ${res.extracted_count} 个角色，新增 ${res.new_count} 个，跳过 ${res.skipped_count} 个`
        );
        table.value.refresh();
        done(true);
    } catch {
        done(false);
    } finally {
        smartImporting.value = false;
    }
}

const previewingId = ref<number | null>(null);
const previewLoadingId = ref<number | null>(null);
let previewAudioEl: HTMLAudioElement | null = null;
let previewBlobURL: string | null = null;

function stopPreview() {
    if (previewAudioEl) {
        previewAudioEl.pause();
        previewAudioEl = null;
    }
    if (previewBlobURL) {
        URL.revokeObjectURL(previewBlobURL);
        previewBlobURL = null;
    }
    previewingId.value = null;
    previewLoadingId.value = null;
}

async function togglePreview(row: CharacterDetailType) {
    const vpId = row.voice_profile_id;
    if (!vpId) return;

    if (previewingId.value === vpId) {
        stopPreview();
        return;
    }

    stopPreview();
    previewLoadingId.value = vpId;

    try {
        const blobURL = await fetchReferenceAudioBlob("voice-profile", vpId);
        if (previewLoadingId.value !== vpId) {
            URL.revokeObjectURL(blobURL);
            return;
        }

        previewBlobURL = blobURL;
        const audio = new Audio(blobURL);
        previewAudioEl = audio;
        previewingId.value = vpId;
        previewLoadingId.value = null;

        audio.addEventListener("ended", () => stopPreview());
        audio.addEventListener("error", () => {
            Message.warning("音频播放失败");
            stopPreview();
        });
        await audio.play();
    } catch {
        Message.warning("音频加载失败");
        stopPreview();
    }
}

onUnmounted(stopPreview);

function handleRegenerate(row: Record<string, any>) {
    Modal.confirm({
        title: "重新生成配音",
        content: `确认重新生成角色【${row.name}】的所有配音？已锁定的片段将被跳过。`,
        okText: "确认生成",
        cancelText: "取消",
        onBeforeOk: async () => {
            try {
                const res = await regenerateByCharacter(row.id);
                const parts = [`已提交 ${res.total_count} 个片段`];
                if (res.skipped_count > 0) parts.push(`跳过 ${res.skipped_count} 个`);
                Message.success(parts.join("，"));
            } catch (e: any) {
                Message.error(e?.message || "重新生成失败");
                throw e;
            }
        }
    });
}

async function handleToggle(row: CharacterDetailType) {
    try {
        if (row.status === "0") {
            await disableCharacter(row.id);
        } else {
            await enableCharacter(row.id);
        }
        table.value.refresh();
        return true;
    } catch {
        return false;
    }
}
</script>
<style lang="scss" scoped></style>
