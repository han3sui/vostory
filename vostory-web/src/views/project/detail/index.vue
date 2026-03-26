<template>
    <frame-view>
        <a-page-header :title="project?.name || '加载中...'" :show-back="true" @back="router.push('/project/list')">
            <template #subtitle>
                <a-space v-if="project">
                    <a-tag :color="statusColor(project.status)" size="small">
                        {{ statusLabel(project.status) }}
                    </a-tag>
                    <a-tag size="small" color="arcoblue">{{ project.total_chapters }} 章</a-tag>
                    <a-tag size="small" color="arcoblue">{{ project.total_characters }} 角色</a-tag>
                </a-space>
            </template>
        </a-page-header>

        <!-- Tab 区域 -->
        <a-tabs v-model:active-key="activeTab">
            <a-tab-pane key="import" title="文件导入">
                <project-import :project-id="projectId" />
            </a-tab-pane>
            <a-tab-pane key="chapter" title="章节管理">
                <project-chapter :project-id="projectId" />
            </a-tab-pane>
            <a-tab-pane key="script-editor" title="脚本编辑">
                <project-script-editor :project-id="projectId" />
            </a-tab-pane>
            <a-tab-pane key="script-segment" title="脚本片段">
                <project-script-segment :project-id="projectId" />
            </a-tab-pane>
            <a-tab-pane key="character" title="角色管理">
                <project-character :project-id="projectId" />
            </a-tab-pane>
            <a-tab-pane key="voice-profile" title="声音配置">
                <project-voice-profile :project-id="projectId" />
            </a-tab-pane>
            <a-tab-pane key="pronunciation-dict" title="发音词典">
                <project-pronunciation-dict :project-id="projectId" />
            </a-tab-pane>
        </a-tabs>
    </frame-view>
</template>
<script lang="ts" setup>
import { useRoute, useRouter } from "vue-router";
import { fetchEventSource } from "@fortaine/fetch-event-source";
import { getProject, ProjectDetailType } from "@/config/apis/project";
import { getProjectEventsURL, TTSSegmentEvent } from "@/config/apis/tts";
import storage from "@/utils/tools/storage";
import ProjectImport from "@/views/project/import/index.vue";
import ProjectChapter from "@/views/chapter/index.vue";
import ProjectScriptEditor from "@/views/script-editor/index.vue";
import ProjectScriptSegment from "@/views/script-segment/index.vue";
import ProjectCharacter from "@/views/character/index.vue";
import ProjectVoiceProfile from "@/views/voice-profile/index.vue";
import ProjectPronunciationDict from "@/views/pronunciation-dict/index.vue";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.id));
const project = ref<ProjectDetailType | null>(null);
const activeTab = ref("chapter");

const STATUS_MAP: Record<string, { label: string; color: string }> = {
    draft: { label: "草稿", color: "gray" },
    parsing: { label: "解析中", color: "orangered" },
    parsed: { label: "已解析", color: "blue" },
    parse_failed: { label: "解析失败", color: "red" },
    generating: { label: "生成中", color: "orange" },
    completed: { label: "已完成", color: "green" }
};

function statusLabel(status: string) {
    return STATUS_MAP[status]?.label || status;
}
function statusColor(status: string) {
    return STATUS_MAP[status]?.color || "gray";
}

async function loadProject() {
    if (!projectId.value) return;
    project.value = await getProject(projectId.value);
}

onMounted(loadProject);

watch(projectId, loadProject);

defineExpose({ refreshProject: loadProject });

// ── 项目级 SSE ──

type TTSEventHandler = (evt: TTSSegmentEvent) => void;

const ttsEventHandlers = ref<Set<TTSEventHandler>>(new Set());
let sseController: AbortController | null = null;

function onTTSEvent(handler: TTSEventHandler) {
    ttsEventHandlers.value.add(handler);
    return () => ttsEventHandlers.value.delete(handler);
}

function connectProjectSSE(pid: number) {
    disconnectProjectSSE();

    const controller = new AbortController();
    sseController = controller;
    const token = storage.getToken();

    fetchEventSource(getProjectEventsURL(pid), {
        headers: { Authorization: `Bearer ${token}` },
        openWhenHidden: true,
        signal: controller.signal,
        onmessage(event: any) {
            if (event.event === "segment") {
                const data: TTSSegmentEvent = JSON.parse(event.data);
                ttsEventHandlers.value.forEach((fn) => fn(data));
            }
        },
        onerror(error) {
            console.error("Project SSE error", error);
            throw error;
        }
    });
}

function disconnectProjectSSE() {
    if (sseController) {
        sseController.abort();
        sseController = null;
    }
}

watch(
    projectId,
    (pid) => {
        if (pid) connectProjectSSE(pid);
        else disconnectProjectSSE();
    },
    { immediate: true }
);

onUnmounted(disconnectProjectSSE);

provide("onTTSEvent", onTTSEvent);
</script>
<style lang="scss" scoped></style>
