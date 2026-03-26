<template>
    <div>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #statusSlot>
                <a-table-column title="状态">
                    <template #cell="{ record }">
                        <a-switch
                            :model-value="record.status === '0'"
                            @change="(val: string | number | boolean) => handleToggleStatus(record, !!val)"
                        />
                    </template>
                </a-table-column>
            </template>
            <template #previewSlot>
                <a-table-column title="参考音频">
                    <template #cell="{ record }">
                        <a-button
                            v-if="record.reference_audio_url"
                            type="text"
                            size="mini"
                            @click="togglePreview(record)"
                        >
                            <template #icon>
                                <icon-play-arrow v-if="previewingId !== record.id" />
                                <icon-pause v-else />
                            </template>
                            {{ previewingId === record.id ? "停止" : "试听" }}
                        </a-button>
                        <span v-else style="color: var(--color-text-3)">未上传</span>
                    </template>
                </a-table-column>
            </template>
        </arco-table>

        <a-drawer
            v-model:visible="emotionDrawerVisible"
            :title="`情绪音频 - ${emotionDrawerProfileName}`"
            :width="600"
            :footer="false"
        >
            <VoiceEmotionManager v-if="emotionDrawerProfileId" :voice-profile-id="emotionDrawerProfileId" />
        </a-drawer>
    </div>
</template>
<script lang="ts" setup>
import { Message, Modal } from "@arco-design/web-vue";
import {
    formHelper,
    ArcoTable,
    tableHelper,
    ArcoForm,
    ArcoModalFormShow,
    ArcoModalTableShow,
    ruleHelper
} from "@easyfe/admin-component";
import {
    getVoiceProfileList,
    addVoiceProfile,
    updateVoiceProfile,
    deleteVoiceProfile,
    enableVoiceProfile,
    disableVoiceProfile,
    VoiceProfileDetailType
} from "@/config/apis/voice-profile";
import { getTTSProviderList, TTSProviderDetailType } from "@/config/apis/ai";
import { getVoiceAssetList, VoiceAssetDetailType } from "@/config/apis/voice-asset";
import { uploadReferenceAudio, extractUploadUrl, pathToFileList, fetchReferenceAudioBlob } from "@/config/apis/upload";
import { hasPermission, PageTableConfig } from "@/views/utils";
import { IconPlayArrow, IconPause } from "@arco-design/web-vue/es/icon";
import { cloneDeep } from "lodash-es";
import VoiceEmotionManager from "@/views/voice-emotion/index.vue";

const props = defineProps<{ projectId: number }>();

const table = ref();
const filterData = ref<Record<string, any>>({});
const emotionDrawerVisible = ref(false);
const emotionDrawerProfileId = ref<number>(0);
const emotionDrawerProfileName = ref("");
const ttsProviderOptions = ref<{ label: string; value: number }[]>([]);

async function loadTTSProviders() {
    const res = await getTTSProviderList({ page: 1, size: 100, status: "0" });
    ttsProviderOptions.value = (res.data || []).map((p: TTSProviderDetailType) => ({
        label: `${p.name} (${p.provider_type})`,
        value: p.id
    }));
}
onMounted(loadTTSProviders);

const getFilterConfig = computed(() => {
    return [formHelper.input("名称", "name", { span: 6, debounce: 500 })];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: { rowKey: "id" },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        trBtns: [
            {
                label: "从音色库导入",
                type: "outline",
                if: () => hasPermission("voice-profile:add"),
                handler: () => {
                    handleOpenImport();
                }
            },
            {
                label: "新增声音配置",
                if: () => hasPermission("voice-profile:add"),
                handler: () => {
                    handleAdd();
                }
            }
        ],
        columns: [
            tableHelper.default("名称", "name"),
            tableHelper.default("参考文本", "reference_text"),
            tableHelper.slot("previewSlot"),
            tableHelper.default("TTS 提供商", "tts_provider_name"),
            tableHelper.slot("statusSlot"),
            tableHelper.date("创建时间", "created_at", { format: "YYYY-MM-DD HH:mm" }),
            tableHelper.btns("操作", [
                {
                    label: "情绪音频",
                    handler(row: Record<string, any>) {
                        handleEmotionDrawer(row as VoiceProfileDetailType);
                    }
                },
                {
                    label: "编辑",
                    if: () => hasPermission("voice-profile:edit"),
                    handler(row: Record<string, any>) {
                        handleEdit(row as VoiceProfileDetailType);
                    }
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("voice-profile:remove"),
                    handler(row: Record<string, any>) {
                        handleDelete(row as VoiceProfileDetailType);
                    }
                }
            ])
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getVoiceProfileList,
        params: { project_id: props.projectId, ...filterData.value }
    };
});

function getFormConfig(isEdit: boolean) {
    return [
        formHelper.input("配置名称", "name", { rules: [ruleHelper.require("请输入名称")] }),
        formHelper.select("TTS 提供商", "tts_provider_id", ttsProviderOptions.value, {
            allowClear: true,
            placeholder: "选择 TTS 提供商"
        }),
        formHelper.upload("参考音频", "reference_audio_url", {
            accept: ".mp3,.wav,.flac,.ogg",
            limit: 1,
            customRequest: uploadReferenceAudio
        }),
        formHelper.input("参考文本", "reference_text", {
            placeholder: "参考音频对应的文本内容"
        })
    ];
}

function handleAdd() {
    ArcoModalFormShow({
        modalConfig: { title: "新增声音配置" },
        value: { project_id: props.projectId },
        formConfig: getFormConfig(false),
        ok: async (data: any) => {
            data.reference_audio_url = extractUploadUrl(data.reference_audio_url);
            await addVoiceProfile(data);
            Message.success("新增成功");
            table.value.refresh();
        }
    });
}

function handleEdit(row: VoiceProfileDetailType) {
    const editValue = cloneDeep(row) as any;
    if (editValue.reference_audio_url) {
        editValue.reference_audio_url = pathToFileList(editValue.reference_audio_url);
    }
    ArcoModalFormShow({
        modalConfig: { title: "编辑声音配置" },
        value: editValue,
        formConfig: getFormConfig(true),
        ok: async (data: any) => {
            data.reference_audio_url = extractUploadUrl(data.reference_audio_url);
            await updateVoiceProfile(data);
            Message.success("更新成功");
            table.value.refresh();
        }
    });
}

function handleDelete(row: VoiceProfileDetailType) {
    Modal.confirm({
        title: "删除",
        content: `确认删除声音配置【${row.name}】？`,
        onBeforeOk: async () => {
            await deleteVoiceProfile(row.id);
            Message.success("删除成功");
            table.value.refresh();
        }
    });
}

function handleEmotionDrawer(row: VoiceProfileDetailType) {
    emotionDrawerProfileId.value = row.id;
    emotionDrawerProfileName.value = row.name;
    emotionDrawerVisible.value = true;
}

const GENDERS = [
    { label: "男", value: "male" },
    { label: "女", value: "female" },
    { label: "未知", value: "unknown" }
];

function handleOpenImport() {
    const importFilterData = ref<any>({});
    ArcoModalTableShow({
        modalConfig: {
            title: "从音色库导入",
            width: "900px"
        },
        defaultSelected: [],
        tableConfig: {
            tableConfig: {
                arcoProps: {
                    rowKey: "id",
                    rowSelection: {
                        type: "checkbox",
                        showCheckedAll: true
                    }
                },
                ...PageTableConfig,
                showRefresh: false,
                maxHeight: "40vh",
                columns: [
                    tableHelper.default("音色名称", "name"),
                    tableHelper.status("性别", "gender", (item: any) => {
                        const found = GENDERS.find((g) => g.value === item.gender);
                        return { text: found?.label || item.gender, status: "normal" };
                    }),
                    tableHelper.default("描述", "description"),
                    tableHelper.default("参考文本", "reference_text"),
                    tableHelper.default("TTS 提供商", "tts_provider_name")
                ]
            },
            filterConfig: [
                formHelper.input("名称", "name", { span: 6, debounce: 500 }),
                formHelper.select("性别", "gender", GENDERS, { span: 5 })
            ],
            filterData: importFilterData.value,
            req: {
                fn: getVoiceAssetList,
                params: { status: "0", ...importFilterData.value }
            }
        },
        ok: async (selected: VoiceAssetDetailType[]) => {
            let successCount = 0;
            for (const asset of selected) {
                try {
                    await addVoiceProfile({
                        project_id: props.projectId,
                        name: asset.name,
                        voice_asset_id: asset.id,
                        reference_audio_url: asset.reference_audio_url,
                        reference_text: asset.reference_text,
                        tts_provider_id: asset.tts_provider_id
                    });
                    successCount++;
                } catch {
                    // skip failed
                }
            }
            Message.success(`成功导入 ${successCount} 个音色配置`);
            table.value.refresh();
        }
    });
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

async function togglePreview(row: VoiceProfileDetailType) {
    if (previewingId.value === row.id) {
        stopPreview();
        return;
    }

    stopPreview();
    previewLoadingId.value = row.id;

    try {
        const blobURL = await fetchReferenceAudioBlob("voice-profile", row.id);
        if (previewLoadingId.value !== row.id) {
            URL.revokeObjectURL(blobURL);
            return;
        }

        previewBlobURL = blobURL;
        const audio = new Audio(blobURL);
        previewAudioEl = audio;
        previewingId.value = row.id;
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

async function handleToggleStatus(record: VoiceProfileDetailType, enabled: boolean) {
    if (enabled) {
        await enableVoiceProfile(record.id);
    } else {
        await disableVoiceProfile(record.id);
    }
    table.value.refresh();
}
</script>
<style lang="scss" scoped>
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

.no-audio-text {
    font-size: 12px;
    color: var(--color-text-3);
}
</style>
