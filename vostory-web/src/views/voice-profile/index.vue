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

        <a-modal
            v-model:visible="importModalVisible"
            title="从音色库导入"
            :footer="false"
            width="780px"
        >
            <a-alert style="margin-bottom: 16px">选择一个全局音色，将自动创建对应的声音配置到当前项目。</a-alert>
            <a-spin :loading="loadingAssets" style="width: 100%">
                <a-empty v-if="!loadingAssets && voiceAssets.length === 0" description="暂无可用音色，请先在 AI 配置 > 音色管理 中添加" />
                <div v-else class="asset-grid">
                    <div
                        v-for="asset in voiceAssets"
                        :key="asset.id"
                        class="asset-card"
                        @click="handleImportAsset(asset)"
                    >
                        <div class="asset-card__header">
                            <span class="asset-card__name">{{ asset.name }}</span>
                            <a-tag size="small" :color="asset.gender === 'male' ? 'blue' : asset.gender === 'female' ? 'red' : 'gray'">
                                {{ asset.gender === "male" ? "男" : asset.gender === "female" ? "女" : "未知" }}
                            </a-tag>
                        </div>
                        <div class="asset-card__tags">
                            <a-tag v-for="t in (asset.tags || []).slice(0, 3)" :key="t" size="small" color="arcoblue">{{ t }}</a-tag>
                        </div>
                        <div class="asset-card__audio">
                            <icon-check-circle v-if="asset.reference_audio_url" style="color: var(--color-success-6)" />
                            <icon-close-circle v-else style="color: var(--color-text-4)" />
                            <span style="margin-left: 4px; font-size: 12px; color: var(--color-text-3)">
                                {{ asset.reference_audio_url ? "有参考音频" : "无参考音频" }}
                            </span>
                        </div>
                    </div>
                </div>
            </a-spin>
        </a-modal>
    </div>
</template>
<script lang="ts" setup>
import { Message } from "@arco-design/web-vue";
import {
    formHelper,
    ArcoTable,
    tableHelper,
    ArcoForm,
    ArcoModalFormShow,
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
import { getVoiceAssetOptions, VoiceAssetOptionType } from "@/config/apis/voice-asset";
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
const importModalVisible = ref(false);
const loadingAssets = ref(false);
const voiceAssets = ref<VoiceAssetOptionType[]>([]);

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
                        handleDelete(row.id);
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

function handleDelete(id: number) {
    deleteVoiceProfile(id).then(() => {
        Message.success("删除成功");
        table.value.refresh();
    });
}

function handleEmotionDrawer(row: VoiceProfileDetailType) {
    emotionDrawerProfileId.value = row.id;
    emotionDrawerProfileName.value = row.name;
    emotionDrawerVisible.value = true;
}

async function handleOpenImport() {
    importModalVisible.value = true;
    loadingAssets.value = true;
    try {
        voiceAssets.value = await getVoiceAssetOptions();
    } finally {
        loadingAssets.value = false;
    }
}

async function handleImportAsset(asset: VoiceAssetOptionType) {
    try {
        await addVoiceProfile({
            project_id: props.projectId,
            name: asset.name,
            voice_asset_id: asset.id,
            reference_audio_url: asset.reference_audio_url,
            reference_text: ""
        });
        Message.success(`已导入音色「${asset.name}」`);
        importModalVisible.value = false;
        table.value.refresh();
    } catch {
        Message.error("导入失败");
    }
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
<style lang="scss" scoped>
.asset-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
}

.asset-card {
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
    padding: 12px;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
        border-color: rgb(var(--primary-6));
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    }

    &__header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 8px;
    }

    &__name {
        font-weight: 500;
        font-size: 14px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        flex: 1;
        margin-right: 8px;
    }

    &__tags {
        display: flex;
        gap: 4px;
        margin-bottom: 8px;
        min-height: 22px;
    }

    &__audio {
        display: flex;
        align-items: center;
    }
}
</style>
