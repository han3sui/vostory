<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #switchSlot>
                <a-table-column title="启用">
                    <template #cell="{ record }">
                        <a-switch
                            :disabled="!hasPermission('voice-asset:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleToggle(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
            <template #tagsSlot>
                <a-table-column title="标签">
                    <template #cell="{ record }">
                        <a-space wrap>
                            <a-tag v-for="t in record.tags || []" :key="t" size="small" color="arcoblue">
                                {{ t }}
                            </a-tag>
                            <span v-if="!record.tags?.length" style="color: var(--color-text-3)">-</span>
                        </a-space>
                    </template>
                </a-table-column>
            </template>
            <template #audioSlot>
                <a-table-column title="参考音频">
                    <template #cell="{ record }">
                        <template v-if="record.reference_audio_url">
                            <a-button type="text" size="mini" @click="playAudio(record)">
                                <template #icon>
                                    <icon-play-arrow v-if="playingId !== record.id" />
                                    <icon-pause v-else />
                                </template>
                                {{ playingId === record.id ? "停止" : "试听" }}
                            </a-button>
                        </template>
                        <span v-else style="color: var(--color-text-3)">未上传</span>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { Modal, Message } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import {
    getVoiceAssetList,
    addVoiceAsset,
    updateVoiceAsset,
    deleteVoiceAsset,
    enableVoiceAsset,
    disableVoiceAsset,
    VoiceAssetDetailType
} from "@/config/apis/voice-asset";
import { getTTSProviderList, TTSProviderDetailType } from "@/config/apis/ai";
import { uploadReferenceAudio } from "@/config/apis/upload";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const GENDERS = [
    { label: "男", value: "male" },
    { label: "女", value: "female" },
    { label: "未知", value: "unknown" }
];

const table = ref();
const filterData = ref({});
const playingId = ref<number>(0);
const ttsProviderOptions = ref<{ label: string; value: number }[]>([]);
let currentAudio: HTMLAudioElement | null = null;

async function loadTTSProviders() {
    const res = await getTTSProviderList({ page: 1, size: 100, status: "0" });
    ttsProviderOptions.value = (res.data || []).map((p: TTSProviderDetailType) => ({
        label: `${p.name} (${p.provider_type})`,
        value: p.id
    }));
}
onMounted(loadTTSProviders);

function playAudio(record: VoiceAssetDetailType) {
    if (playingId.value === record.id) {
        currentAudio?.pause();
        currentAudio = null;
        playingId.value = 0;
        return;
    }
    currentAudio?.pause();
    currentAudio = new Audio(record.reference_audio_url);
    playingId.value = record.id;
    currentAudio.play();
    currentAudio.onended = () => {
        playingId.value = 0;
        currentAudio = null;
    };
}

const getFilterConfig = computed(() => {
    return [
        formHelper.input("名称", "name", { span: 5, debounce: 500 }),
        formHelper.select("性别", "gender", GENDERS, { span: 4 }),
        formHelper.select(
            "状态",
            "status",
            [
                { label: "正常", value: "0" },
                { label: "停用", value: "1" }
            ],
            { span: 4 }
        )
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: { rowKey: "id" },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("音色名称", "name"),
            tableHelper.status("性别", "gender", (item: any) => {
                const found = GENDERS.find((g) => g.value === item.gender);
                return { text: found?.label || item.gender, status: "normal" };
            }),
            tableHelper.default("描述", "description"),
            tableHelper.slot("audioSlot"),
            tableHelper.default("参考文本", "reference_text"),
            tableHelper.default("TTS 提供商", "tts_provider_name"),
            tableHelper.slot("tagsSlot"),
            tableHelper.slot("switchSlot"),
            tableHelper.date("创建时间", "created_at", { format: "YYYY-MM-DD HH:mm" }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("voice-asset:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("voice-asset:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除音色【${row.name}】？`,
                            onBeforeOk: async () => {
                                await deleteVoiceAsset(row.id);
                                table.value.refresh();
                            }
                        });
                    }
                }
            ])
        ],
        trBtns: [
            {
                label: "添加音色",
                type: "primary",
                if: () => hasPermission("voice-asset:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getVoiceAssetList,
        params: { ...filterData.value }
    };
});

function onEdit(v: Record<string, any> | null) {
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑音色" : "添加音色",
            width: "650px"
        },
        value: tempValue || { status: "0", gender: "unknown", tags: [] },
        formConfig: [
            formHelper.input("音色名称", "name", {
                rules: [ruleHelper.require("请输入名称")]
            }),
            formHelper.select("性别", "gender", GENDERS),
            formHelper.textarea("描述", "description", {
                placeholder: "描述该音色的特征，如音色温暖、低沉、清脆等"
            }),
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
        ],
        ok: async (data: any) => {
            if (tempValue) {
                await updateVoiceAsset(data);
            } else {
                await addVoiceAsset(data);
            }
            table.value.refresh();
        }
    });
}

async function handleToggle(row: VoiceAssetDetailType) {
    try {
        if (row.status === "0") {
            await disableVoiceAsset(row.id);
        } else {
            await enableVoiceAsset(row.id);
        }
        table.value.refresh();
        return true;
    } catch {
        return false;
    }
}
</script>
<style lang="scss" scoped></style>
