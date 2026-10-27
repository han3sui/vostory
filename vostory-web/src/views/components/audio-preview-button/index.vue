<template>
    <a-button v-if="available" type="text" size="mini" :loading="loading" @click="handleClick">
        <template #icon>
            <icon-play-arrow v-if="!playing" />
            <icon-pause v-else />
        </template>
        {{ preview.buttonText(id) }}
    </a-button>
    <span v-else class="empty-text">{{ emptyText }}</span>
</template>
<script lang="ts" setup name="AudioPreviewButton">
import type { AudioPreviewController } from "@/hooks/useAudioPreview";

const props = withDefaults(
    defineProps<{
        id: number | null;
        available: boolean;
        preview: AudioPreviewController;
        load: () => Promise<string>;
        emptyText?: string;
    }>(),
    { emptyText: "未上传" }
);

const loading = computed(() => !!props.id && props.preview.isLoading(props.id));
const playing = computed(() => !!props.id && props.preview.isPlaying(props.id));

function handleClick() {
    if (!props.id) return;
    props.preview.toggle(props.id, props.load);
}
</script>
<style lang="scss" scoped>
.empty-text {
    color: var(--color-text-3);
}
</style>
