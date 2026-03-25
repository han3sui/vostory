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
import { hasPermission, PageTableConfig } from "@/views/utils";
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
    return [
        formHelper.input("名称", "name", { span: 6, debounce: 500 })
    ];
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
        formHelper.input("参考音频 URL", "reference_audio_url", {
            placeholder: "参考音频文件路径或 URL"
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
            await addVoiceProfile(data);
            Message.success("新增成功");
            table.value.refresh();
        }
    });
}

function handleEdit(row: VoiceProfileDetailType) {
    ArcoModalFormShow({
        modalConfig: { title: "编辑声音配置" },
        value: cloneDeep(row),
        formConfig: getFormConfig(true),
        ok: async (data: any) => {
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

async function handleToggleStatus(record: VoiceProfileDetailType, enabled: boolean) {
    if (enabled) {
        await enableVoiceProfile(record.id);
    } else {
        await disableVoiceProfile(record.id);
    }
    table.value.refresh();
}
</script>
<style lang="scss" scoped></style>
