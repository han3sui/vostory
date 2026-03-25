<template>
    <div class="voice-emotion-wrap">
        <div class="emotion-header">
            <h4>情绪参考音频管理</h4>
            <a-button type="primary" size="small" @click="handleAdd">新增情绪音频</a-button>
        </div>

        <a-spin :loading="loading">
            <div v-if="emotions.length === 0" class="empty-area">
                <a-empty description="暂无情绪音频，点击上方按钮添加" />
            </div>
            <div v-else class="emotion-grid">
                <div v-for="item in emotions" :key="item.id" class="emotion-card">
                    <div class="emotion-card-header">
                        <a-tag :color="emotionColor(item.emotion_type)">{{ emotionLabel(item.emotion_type) }}</a-tag>
                        <a-tag size="small">{{ strengthLabel(item.emotion_strength) }}</a-tag>
                    </div>
                    <div class="emotion-card-body">
                        <div class="audio-url" :title="item.reference_audio_url">
                            {{ item.reference_audio_url ? extractFilenameFromPath(item.reference_audio_url) : '未设置' }}
                        </div>
                        <div v-if="item.reference_text" class="ref-text">{{ item.reference_text }}</div>
                    </div>
                    <div class="emotion-card-footer">
                        <a-button type="text" size="mini" @click="handleEdit(item)">编辑</a-button>
                        <a-popconfirm content="确定删除？" @ok="handleDelete(item.id)">
                            <a-button type="text" size="mini" status="danger">删除</a-button>
                        </a-popconfirm>
                    </div>
                </div>
            </div>
        </a-spin>

        <a-modal v-model:visible="formVisible" :title="formTitle" @ok="handleFormOk" @cancel="formVisible = false">
            <a-form :model="formData" layout="vertical">
                <a-form-item label="情绪类型" required>
                    <a-select v-model="formData.emotion_type" placeholder="选择情绪类型">
                        <a-option value="neutral">平静</a-option>
                        <a-option value="happy">开心</a-option>
                        <a-option value="sad">悲伤</a-option>
                        <a-option value="angry">愤怒</a-option>
                        <a-option value="fear">恐惧</a-option>
                        <a-option value="surprise">惊讶</a-option>
                        <a-option value="disgust">厌恶</a-option>
                    </a-select>
                </a-form-item>
                <a-form-item label="情绪强度" required>
                    <a-select v-model="formData.emotion_strength" placeholder="选择强度">
                        <a-option value="light">轻</a-option>
                        <a-option value="medium">中</a-option>
                        <a-option value="strong">强</a-option>
                    </a-select>
                </a-form-item>
                <a-form-item label="参考音频" required>
                    <div style="width: 100%">
                        <a-upload
                            :limit="1"
                            accept=".mp3,.wav,.flac,.ogg"
                            :custom-request="handleAudioUpload"
                            :show-file-list="!formData.reference_audio_url"
                        >
                            <template #upload-button>
                                <a-button type="outline" size="small">选择音频文件</a-button>
                            </template>
                        </a-upload>
                        <div v-if="formData.reference_audio_url" class="uploaded-file-info">
                            <icon-check-circle style="color: rgb(var(--green-6)); margin-right: 4px" />
                            <span class="file-name">{{ extractFilenameFromPath(formData.reference_audio_url as string) }}</span>
                        </div>
                    </div>
                </a-form-item>
                <a-form-item label="参考文本">
                    <a-textarea v-model="formData.reference_text" placeholder="参考音频对应的文本内容" :auto-size="{ minRows: 2 }" />
                </a-form-item>
            </a-form>
        </a-modal>
    </div>
</template>
<script lang="ts" setup>
import { Message, RequestOption } from "@arco-design/web-vue";
import { IconCheckCircle } from "@arco-design/web-vue/es/icon";
import {
    getVoiceEmotionsByProfile,
    addVoiceEmotion,
    updateVoiceEmotion,
    deleteVoiceEmotion,
    VoiceEmotionDetailType
} from "@/config/apis/voice-emotion";
import { uploadReferenceAudio, extractFilenameFromPath } from "@/config/apis/upload";

const props = defineProps<{ voiceProfileId: number }>();

const loading = ref(false);
const emotions = ref<VoiceEmotionDetailType[]>([]);
const formVisible = ref(false);
const formTitle = ref("新增情绪音频");
const formData = ref<Partial<VoiceEmotionDetailType>>({});
const editingId = ref<number | null>(null);

async function loadEmotions() {
    if (!props.voiceProfileId) return;
    loading.value = true;
    try {
        emotions.value = await getVoiceEmotionsByProfile(props.voiceProfileId) || [];
    } finally {
        loading.value = false;
    }
}

onMounted(loadEmotions);
watch(() => props.voiceProfileId, loadEmotions);

function handleAdd() {
    editingId.value = null;
    formTitle.value = "新增情绪音频";
    formData.value = {
        voice_profile_id: props.voiceProfileId,
        emotion_type: "neutral",
        emotion_strength: "medium",
        reference_audio_url: "",
        reference_text: ""
    };
    formVisible.value = true;
}

function handleEdit(item: VoiceEmotionDetailType) {
    editingId.value = item.id;
    formTitle.value = "编辑情绪音频";
    formData.value = { ...item };
    formVisible.value = true;
}

async function handleFormOk() {
    if (!formData.value.emotion_type || !formData.value.reference_audio_url) {
        Message.warning("请填写必填项");
        return;
    }
    try {
        if (editingId.value) {
            await updateVoiceEmotion({ ...formData.value, id: editingId.value });
            Message.success("更新成功");
        } else {
            await addVoiceEmotion(formData.value);
            Message.success("新增成功");
        }
        formVisible.value = false;
        loadEmotions();
    } catch {
        Message.error("操作失败");
    }
}

function handleAudioUpload(option: RequestOption): any {
    return uploadReferenceAudio(option).then((res: any) => {
        formData.value.reference_audio_url = res.url;
        Message.success("音频上传成功");
    });
}

async function handleDelete(id: number) {
    await deleteVoiceEmotion(id);
    Message.success("删除成功");
    loadEmotions();
}

function emotionLabel(type: string) {
    const map: Record<string, string> = {
        neutral: "平静", happy: "开心", sad: "悲伤",
        angry: "愤怒", fear: "恐惧", surprise: "惊讶", disgust: "厌恶"
    };
    return map[type] || type;
}

function emotionColor(type: string) {
    const map: Record<string, string> = {
        neutral: "gray", happy: "orange", sad: "blue",
        angry: "red", fear: "purple", surprise: "magenta", disgust: "green"
    };
    return map[type] || "gray";
}

function strengthLabel(strength: string) {
    const map: Record<string, string> = { light: "轻", medium: "中", strong: "强" };
    return map[strength] || strength;
}
</script>
<style lang="scss" scoped>
.voice-emotion-wrap {
    padding: 16px 0;
}

.emotion-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;

    h4 {
        margin: 0;
        font-size: 15px;
        font-weight: 500;
    }
}

.empty-area {
    padding: 32px 0;
}

.emotion-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 12px;
}

.emotion-card {
    border: 1px solid var(--color-border);
    border-radius: 8px;
    padding: 12px;
    transition: box-shadow 0.2s;

    &:hover {
        box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
    }
}

.emotion-card-header {
    display: flex;
    gap: 6px;
    margin-bottom: 8px;
}

.emotion-card-body {
    .audio-url {
        font-size: 12px;
        color: var(--color-text-2);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .ref-text {
        margin-top: 4px;
        font-size: 12px;
        color: var(--color-text-3);
    }
}

.emotion-card-footer {
    display: flex;
    justify-content: flex-end;
    gap: 4px;
    margin-top: 8px;
    padding-top: 8px;
    border-top: 1px solid var(--color-border);
}

.uploaded-file-info {
    display: flex;
    align-items: center;
    margin-top: 6px;
    font-size: 13px;
    color: var(--color-text-2);

    .file-name {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        max-width: 300px;
    }
}
</style>
