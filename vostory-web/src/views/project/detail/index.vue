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
import { getProject, ProjectDetailType } from "@/config/apis/project";
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
</script>
<style lang="scss" scoped></style>
