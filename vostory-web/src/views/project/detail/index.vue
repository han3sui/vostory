<template>
    <frame-view>
        <a-page-header :title="project?.name || '加载中...'" :show-back="true" @back="router.push('/project/list')">
            <template #subtitle>
                <a-space v-if="project" wrap>
                    <a-tag :color="statusColor(project.status)" size="small">
                        {{ statusLabel(project.status) }}
                    </a-tag>
                    <a-tag size="small" color="arcoblue">{{ project.total_chapters }} 章</a-tag>
                    <a-tag size="small" color="arcoblue">{{ project.total_characters }} 角色</a-tag>
                    <template v-if="activeTaskList.length > 0">
                        <a-tag v-for="t in activeTaskList" :key="t.task_id" size="small" color="orange">
                            {{ t.chapter_title || '未知章节' }} {{ t.completed_count + t.failed_count }}/{{ t.total_count }}
                        </a-tag>
                        <a-popconfirm content="确认取消所有排队中的生成任务？" @ok="handleCancelAll">
                            <a-button type="text" size="mini" status="danger">取消全部队列</a-button>
                        </a-popconfirm>
                    </template>
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
import { Message } from "@arco-design/web-vue";
import { useRoute, useRouter } from "vue-router";
import { fetchEventSource } from "@fortaine/fetch-event-source";
import { getProject, ProjectDetailType } from "@/config/apis/project";
import {
    getProjectEventsURL,
    getActiveTasksByProject,
    cancelProjectQueue,
    TTSSegmentEvent,
    ProjectTaskProgress
} from "@/config/apis/tts";
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

// ── 项目级 SSE + 全局任务进度 ──

type TTSEventHandler = (evt: TTSSegmentEvent) => void;

const ttsEventHandlers = ref<Set<TTSEventHandler>>(new Set());
const activeTasks = ref<Map<number, ProjectTaskProgress>>(new Map());
let sseController: AbortController | null = null;

const activeTaskList = computed(() => Array.from(activeTasks.value.values()));

function onTTSEvent(handler: TTSEventHandler) {
    ttsEventHandlers.value.add(handler);
    return () => ttsEventHandlers.value.delete(handler);
}

async function loadActiveTasks(pid: number) {
    try {
        const tasks = await getActiveTasksByProject(pid);
        const map = new Map<number, ProjectTaskProgress>();
        if (tasks) {
            for (const t of tasks) map.set(t.task_id, t);
        }
        activeTasks.value = map;
    } catch {
        // ignore
    }
}

function handleSSEEvent(evt: TTSSegmentEvent) {
    ttsEventHandlers.value.forEach((fn) => fn(evt));

    const existing = activeTasks.value.get(evt.task_id);
    if (existing) {
        existing.completed_count = evt.completed;
        existing.failed_count = evt.failed;
        existing.total_count = evt.total;
        existing.progress = evt.progress;
        existing.status = evt.task_status;
    } else {
        activeTasks.value.set(evt.task_id, {
            task_id: evt.task_id,
            chapter_id: evt.chapter_id,
            chapter_title: "",
            status: evt.task_status,
            progress: evt.progress,
            total_count: evt.total,
            completed_count: evt.completed,
            failed_count: evt.failed
        });
    }

    if (evt.task_done) {
        activeTasks.value.delete(evt.task_id);
    }
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
                handleSSEEvent(JSON.parse(event.data));
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

async function handleCancelAll() {
    if (!projectId.value) return;
    try {
        const res = await cancelProjectQueue(projectId.value);
        Message.success(`已取消 ${res.cancelled_count} 个排队片段`);
        await loadActiveTasks(projectId.value);
    } catch {
        Message.error("取消失败");
    }
}

watch(
    projectId,
    (pid) => {
        if (pid) {
            loadActiveTasks(pid);
            connectProjectSSE(pid);
        } else {
            disconnectProjectSSE();
        }
    },
    { immediate: true }
);

onUnmounted(disconnectProjectSSE);

provide("onTTSEvent", onTTSEvent);
</script>
<style lang="scss" scoped></style>
