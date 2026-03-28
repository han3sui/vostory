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
                        <a-button v-if="record.reference_audio_url" type="text" size="mini" @click="playAudio(record)">
                            <template #icon>
                                <icon-play-arrow v-if="playingId !== record.id" />
                                <icon-pause v-else />
                            </template>
                            {{ playingId === record.id ? "停止" : "试听" }}
                        </a-button>
                        <span v-else style="color: var(--color-text-3)">未上传</span>
                    </template>
                </a-table-column>
            </template>
        </arco-table>

        <a-modal
            v-model:visible="batchVisible"
            title="批量添加音色"
            width="800px"
            :mask-closable="false"
            :footer="false"
            :align-center="false"
            top="10vh"
            title-align="start"
            @close="onBatchClose"
        >
            <div style="margin-bottom: 12px">
                <a-space>
                    <a-typography-text type="secondary">
                        共 {{ batchFiles.length }} 个文件，成功 {{ batchSuccessCount }} 个，失败 {{ batchFailCount }} 个
                    </a-typography-text>
                    <a-tag v-if="batchProcessing" color="blue">处理中...</a-tag>
                    <a-tag v-else-if="batchFiles.length > 0 && batchPendingCount === 0" color="green">已完成</a-tag>
                </a-space>
            </div>
            <a-space>
                <a-upload
                    ref="batchUploadRef"
                    :auto-upload="false"
                    multiple
                    accept=".mp3,.wav,.flac,.ogg"
                    :file-list="batchFiles"
                    :disabled="batchProcessing"
                    list-type="text"
                    :show-file-list="false"
                    @change="onBatchFilesChange"
                >
                    <template #upload-button>
                        <a-button type="outline" :disabled="batchProcessing">
                            <template #icon><icon-upload /></template>
                            选择音频文件
                        </a-button>
                    </template>
                </a-upload>
                <a-button
                    v-if="batchFiles.length > 0"
                    type="text"
                    status="danger"
                    :disabled="batchProcessing"
                    @click="batchFiles = []"
                >
                    <template #icon><icon-delete /></template>
                    清空
                </a-button>
            </a-space>
            <div style="max-height: 400px; overflow-y: auto; overflow-x: hidden; margin-top: 12px">
                <a-upload
                    :auto-upload="false"
                    :file-list="batchFiles"
                    :disabled="batchProcessing"
                    list-type="text"
                    :show-upload-button="false"
                    @change="onBatchFilesChange"
                />
            </div>
            <div style="margin-top: 16px; text-align: right">
                <a-space>
                    <a-button @click="batchVisible = false" :disabled="batchProcessing">取消</a-button>
                    <a-button
                        type="primary"
                        :loading="batchProcessing"
                        :disabled="batchPendingCount === 0"
                        @click="startBatchUpload"
                    >
                        开始上传 ({{ batchPendingCount }})
                    </a-button>
                </a-space>
            </div>
        </a-modal>
    </frame-view>
</template>
<script lang="ts" setup>
import { Modal, Message, FileItem } from "@arco-design/web-vue";
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
import { uploadReferenceAudio, extractUploadUrl, pathToFileList, fetchReferenceAudioBlob } from "@/config/apis/upload";
import request from "@/packages/request";
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
let currentAudio: HTMLAudioElement | null = null;

let currentBlobURL = "";

async function playAudio(record: VoiceAssetDetailType) {
    if (playingId.value === record.id) {
        currentAudio?.pause();
        currentAudio = null;
        playingId.value = 0;
        if (currentBlobURL) {
            URL.revokeObjectURL(currentBlobURL);
            currentBlobURL = "";
        }
        return;
    }
    currentAudio?.pause();
    if (currentBlobURL) {
        URL.revokeObjectURL(currentBlobURL);
        currentBlobURL = "";
    }
    try {
        playingId.value = record.id;
        currentBlobURL = await fetchReferenceAudioBlob("voice-asset", record.id);
        currentAudio = new Audio(currentBlobURL);
        currentAudio.play();
        currentAudio.onended = () => {
            playingId.value = 0;
            currentAudio = null;
            if (currentBlobURL) {
                URL.revokeObjectURL(currentBlobURL);
                currentBlobURL = "";
            }
        };
    } catch {
        playingId.value = 0;
        Message.error("播放失败");
    }
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
            },
            {
                label: "批量添加",
                type: "outline",
                if: () => hasPermission("voice-asset:add"),
                handler: () => {
                    batchFiles.value = [];
                    batchVisible.value = true;
                }
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
    if (tempValue?.reference_audio_url) {
        tempValue.reference_audio_url = pathToFileList(tempValue.reference_audio_url);
    }
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
            data.reference_audio_url = extractUploadUrl(data.reference_audio_url);
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

// ========== 批量上传 ==========

const batchVisible = ref(false);
const batchFiles = ref<FileItem[]>([]);
const batchProcessing = ref(false);

const batchSuccessCount = computed(() => batchFiles.value.filter((f) => f.status === "done").length);
const batchFailCount = computed(() => batchFiles.value.filter((f) => f.status === "error").length);
const batchPendingCount = computed(() => batchFiles.value.filter((f) => !f.status || f.status === "init").length);

function onBatchFilesChange(fileList: FileItem[]) {
    batchFiles.value = fileList;
}

function detectGender(filename: string): string {
    const lower = filename.toLowerCase();
    if (/(男|male|boy|少年|书生|君子|大叔|老翁|公子|兄弟|父亲|爷爷|哥哥|叔叔|伯伯)/.test(lower)) return "male";
    if (/(女|female|girl|少女|姐姐|姑娘|夫人|萝莉|御姐|母亲|奶奶|妹妹|阿姨|婆婆)/.test(lower)) return "female";
    return "unknown";
}

function extractVoiceName(filename: string): string {
    return filename.replace(/\.[^.]+$/, "");
}

const CONCURRENCY = 4;

async function startBatchUpload() {
    const pending = batchFiles.value.filter((f) => !f.status || f.status === "init");
    if (pending.length === 0) return;

    batchProcessing.value = true;

    let idx = 0;
    async function next(): Promise<void> {
        if (idx >= pending.length) return;
        const file = pending[idx++];
        file.status = "uploading";
        file.percent = 0;
        try {
            const formData = new FormData();
            formData.append("file", file.file as File);
            const uploadRes: any = await request({
                url: "/api/v1/common/upload/reference-audio",
                method: "POST",
                headers: { "Content-Type": "multipart/form-data" },
                enableCancel: false,
                timeout: 0,
                data: formData,
                onUploadProgress: (e: ProgressEvent) => {
                    file.percent = Math.round((e.loaded / e.total) * 100) / 100;
                }
            });

            const voiceName = extractVoiceName(file.name || "");
            const gender = detectGender(file.name || "");
            await addVoiceAsset({
                name: voiceName,
                gender,
                reference_audio_url: uploadRes.path
            });

            file.status = "done";
            file.percent = 1;
        } catch {
            file.status = "error";
        }
        await next();
    }

    const workers = Array.from({ length: Math.min(CONCURRENCY, pending.length) }, () => next());
    await Promise.all(workers);

    batchProcessing.value = false;
    const successCount = pending.filter((f) => f.status === "done").length;
    if (successCount > 0) {
        Message.success(`成功添加 ${successCount} 个音色`);
        table.value.refresh();
    }
}

function onBatchClose() {
    if (!batchProcessing.value) {
        batchFiles.value = [];
    }
}
</script>
<style lang="scss" scoped></style>
