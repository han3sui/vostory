<template>
    <arco-modal v-model:visible="localVisible" :config="modalConfig" :ok="handleApply">
        <a-typography-paragraph type="secondary">
            用文字描述生成该角色的基础音色（Breeze-TTS-2 Voice
            Design），保存后作为角色的参考音频，所有台词都基于这个音色按场景演绎。
        </a-typography-paragraph>
        <a-alert v-if="character?.voice_profile_id" type="warning" style="margin-bottom: 12px">
            当前已绑定「{{
                character.voice_profile_name || "声音配置"
            }}」，保存后将改为绑定新生成的音色，已生成的配音需要重新生成。
        </a-alert>
        <a-form :model="form" layout="vertical">
            <a-form-item label="音色描述" required>
                <a-textarea
                    v-model="form.instruction"
                    :auto-size="{ minRows: 3, maxRows: 6 }"
                    placeholder="如：二十岁左右的青年男性，声音清亮干净，中等音高，语速偏快，带一点少年意气"
                />
            </a-form-item>
            <a-form-item label="参考台词（试听文本，同时作为参考音频的转写）" required>
                <a-input v-model="form.sampleText" placeholder="一句 20~40 字、语气平稳的台词" />
            </a-form-item>
        </a-form>
        <a-space>
            <a-button :loading="suggesting" @click="handleSuggest">
                <template #icon><icon-robot /></template>
                AI 撰写描述
            </a-button>
            <a-button type="primary" :loading="previewing" :disabled="!canPreview" @click="handlePreview">
                <template #icon><icon-sound /></template>
                {{ preview ? "换一个音色" : "生成试听" }}
            </a-button>
        </a-space>
        <div v-if="preview" class="preview-box">
            <audio :src="previewUrl" controls style="width: 100%" />
            <a-typography-text type="secondary" style="font-size: 12px">
                种子 {{ preview.seed }}，满意后点击「保存并绑定」；不满意可再生成一次，或调整描述后重试。
            </a-typography-text>
        </div>
    </arco-modal>
</template>
<script lang="ts" setup>
import { Message } from "@arco-design/web-vue";
import { IconRobot, IconSound } from "@arco-design/web-vue/es/icon";
import { ArcoModal } from "@easyfe/admin-component";
import {
    suggestVoiceDesign,
    previewVoiceDesign,
    applyVoiceDesign,
    CharacterDetailType,
    VoiceDesignPreviewResult,
    VoiceDesignApplyResult
} from "@/config/apis/character";

const props = defineProps<{
    visible: boolean;
    character: CharacterDetailType | null;
}>();

const emit = defineEmits<{
    (e: "update:visible", v: boolean): void;
    (e: "success", res: VoiceDesignApplyResult, character: CharacterDetailType): void;
}>();

const localVisible = computed({
    get: () => props.visible,
    set: (v: boolean) => emit("update:visible", v)
});

const form = reactive({ instruction: "", sampleText: "" });
const suggesting = ref(false);
const previewing = ref(false);
const preview = ref<VoiceDesignPreviewResult | null>(null);
const previewUrl = ref("");

const canPreview = computed(() => !!form.instruction.trim() && !!form.sampleText.trim());

const modalConfig = computed(() => ({
    title: `设计音色 · ${props.character?.name || ""}`,
    width: "680px",
    top: "8vh",
    okText: "保存并绑定",
    okButtonProps: { disabled: !preview.value || previewing.value },
    onOpen,
    onClose: releaseAudio
}));

// 描述或台词改动后，旧试听与新内容不再对应，需要重新生成
watch(
    () => [form.instruction, form.sampleText],
    () => {
        if (preview.value) {
            preview.value = null;
            releaseAudio();
        }
    }
);

function releaseAudio() {
    if (previewUrl.value) {
        URL.revokeObjectURL(previewUrl.value);
        previewUrl.value = "";
    }
}

function base64ToBlobUrl(b64: string) {
    const bin = atob(b64);
    const bytes = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    return URL.createObjectURL(new Blob([bytes], { type: "audio/wav" }));
}

function onOpen() {
    form.instruction = "";
    form.sampleText = "";
    preview.value = null;
    releaseAudio();
    if (props.character) handleSuggest();
}

async function handleSuggest() {
    if (!props.character) return;
    suggesting.value = true;
    try {
        const res = await suggestVoiceDesign(props.character.id);
        form.instruction = res.instruction;
        form.sampleText = res.sample_text;
    } finally {
        suggesting.value = false;
    }
}

async function handlePreview() {
    if (!props.character || !canPreview.value) return;
    previewing.value = true;
    try {
        const res = await previewVoiceDesign(props.character.id, {
            instruction: form.instruction.trim(),
            sample_text: form.sampleText.trim()
        });
        releaseAudio();
        previewUrl.value = base64ToBlobUrl(res.audio_base64);
        preview.value = res;
    } finally {
        previewing.value = false;
    }
}

async function handleApply() {
    const character = props.character;
    if (!character || !preview.value) {
        Message.warning("请先生成试听");
        throw new Error("no preview");
    }
    const res = await applyVoiceDesign(character.id, {
        preview_path: preview.value.preview_path,
        instruction: form.instruction.trim(),
        sample_text: form.sampleText.trim(),
        seed: preview.value.seed
    });
    releaseAudio();
    emit("success", res, character);
}

onUnmounted(releaseAudio);
</script>
<style lang="scss" scoped>
.preview-box {
    margin-top: 16px;
    padding: 12px;
    border-radius: 4px;
    background-color: var(--color-fill-1);
}
</style>
