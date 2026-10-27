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
                        <audio-preview-button
                            :id="record.id"
                            :available="!!record.reference_audio_url"
                            :preview="audioPreview"
                            :load="() => fetchReferenceAudioBlob('voice-profile', record.id)"
                        />
                    </template>
                </a-table-column>
            </template>
        </arco-table>
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
    importVoiceProfiles,
    VoiceProfileDetailType
} from "@/config/apis/voice-profile";
import { getVoiceAssetList, VoiceAssetDetailType } from "@/config/apis/voice-asset";
import { uploadReferenceAudio, extractUploadUrl, pathToFileList, fetchReferenceAudioBlob } from "@/config/apis/upload";
import { regenerateByVoiceProfile } from "@/config/apis/tts";
import { hasPermission, PageTableConfig } from "@/views/utils";
import { useAudioPreview } from "@/hooks/useAudioPreview";
import AudioPreviewButton from "@/views/components/audio-preview-button/index.vue";
import { cloneDeep } from "lodash-es";

const props = defineProps<{ projectId: number }>();

const table = ref();
const filterData = ref<Record<string, any>>({});

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
            tableHelper.status("性别", "gender", (item: any) => {
                const found = GENDERS.find((g) => g.value === item.gender);
                return { text: found?.label || item.gender || "-", status: "normal" };
            }),
            tableHelper.default("描述", "description"),
            // tableHelper.default("参考文本", "reference_text"),
            tableHelper.slot("previewSlot"),
            tableHelper.slot("statusSlot"),
            tableHelper.date("创建时间", "created_at", { format: "YYYY-MM-DD HH:mm" }),
            tableHelper.btns(
                "操作",
                [
                    {
                        label: "重新生成",
                        if: () => hasPermission("tts:synthesize"),
                        handler(row: Record<string, any>) {
                            handleRegenerate(row as VoiceProfileDetailType);
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
                ],
                {
                    width: 300
                }
            )
        ]
    });
});

const GENDERS = [
    { label: "男", value: "male" },
    { label: "女", value: "female" },
    { label: "未知", value: "unknown" }
];

const getData = computed(() => {
    return {
        fn: getVoiceProfileList,
        params: { project_id: props.projectId, ...filterData.value }
    };
});

function getFormConfig(isEdit: boolean) {
    return [
        formHelper.input("配置名称", "name", { rules: [ruleHelper.require("请输入名称")] }),
        formHelper.select("性别", "gender", GENDERS, { allowClear: true }),
        formHelper.textarea("描述", "description", {
            placeholder: "描述该声音的特征，如音色温暖、低沉、清脆等"
        }),
        formHelper.upload("参考音频", "reference_audio_url", {
            accept: ".mp3,.wav,.flac,.ogg",
            limit: 1,
            customRequest: uploadReferenceAudio
        }),
        formHelper.input("参考文本", "reference_text", {
            placeholder: "参考音频的逐字转写（Breeze-TTS-2 克隆音色必填）",
            rules: [ruleHelper.require("请输入参考音频的逐字转写")]
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

function handleRegenerate(row: VoiceProfileDetailType) {
    Modal.confirm({
        title: "重新生成配音",
        content: `将重新生成项目中所有使用声音配置【${row.name}】的角色对应的片段配音（已锁定的片段不受影响）。是否继续？`,
        okText: "确认生成",
        cancelText: "取消",
        onOk: async () => {
            try {
                const res = await regenerateByVoiceProfile(row.id);
                Message.success(`已提交 ${res.total_count} 个片段到生成队列`);
            } catch (e: any) {
                const msg = e?.response?.data?.message || e?.message || "重新生成失败";
                Message.error(msg);
            }
        }
    });
}

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
                    tableHelper.default("参考文本", "reference_text")
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
            if (selected.length === 0) {
                Message.warning("请选择要导入的音色");
                return;
            }
            const res = await importVoiceProfiles({
                project_id: props.projectId,
                voice_asset_ids: selected.map((a) => a.id)
            });
            Message.success(`成功导入 ${res.imported} 个音色配置`);
            table.value.refresh();
        }
    });
}

const audioPreview = useAudioPreview();

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
