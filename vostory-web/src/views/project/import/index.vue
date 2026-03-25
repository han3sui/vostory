<template>
    <frame-view>
        <div class="p-6 max-w-2xl">
            <a-card title="文件导入">
                <a-form :model="formData" layout="vertical">
                    <a-form-item label="选择项目" required>
                        <a-select v-model="formData.project_id" placeholder="请选择项目" allow-clear>
                            <a-option v-for="p in projectOptions" :key="p.value" :value="p.value">
                                {{ p.label }}
                            </a-option>
                        </a-select>
                    </a-form-item>
                    <a-form-item label="上传源文件" required>
                        <a-upload
                            :custom-request="handleUpload"
                            :limit="1"
                            :disabled="!formData.project_id"
                            accept=".txt,.docx,.epub"
                        >
                            <template #upload-button>
                                <a-button type="primary" :disabled="!formData.project_id">
                                    选择文件（支持 txt / docx / epub）
                                </a-button>
                            </template>
                        </a-upload>
                    </a-form-item>
                    <a-form-item v-if="uploaded && canParse">
                        <a-button type="primary" :loading="parsing" @click="handleParse"> 解析章节 </a-button>
                    </a-form-item>
                </a-form>

                <a-divider v-if="uploadResult || parseResult" />

                <a-descriptions v-if="uploadResult" title="上传结果" :column="1" bordered size="small">
                    <a-descriptions-item label="文件名">{{ uploadResult.file_name }}</a-descriptions-item>
                    <a-descriptions-item label="文件大小"
                        >{{ (uploadResult.file_size / 1024).toFixed(1) }} KB</a-descriptions-item
                    >
                    <a-descriptions-item label="文件类型">{{ uploadResult.source_type }}</a-descriptions-item>
                </a-descriptions>

                <a-descriptions v-if="parseResult" title="解析结果" :column="1" bordered size="small" class="mt-4">
                    <a-descriptions-item label="章节数">{{ parseResult.total_chapters }}</a-descriptions-item>
                    <a-descriptions-item label="总字数">{{
                        parseResult.total_words.toLocaleString()
                    }}</a-descriptions-item>
                    <a-descriptions-item label="状态">
                        <a-tag color="green">{{ parseResult.message }}</a-tag>
                    </a-descriptions-item>
                </a-descriptions>
            </a-card>
        </div>
    </frame-view>
</template>
<script lang="ts" setup>
import { Message, RequestOption } from "@arco-design/web-vue";
import { uploadSourceFile, parseSourceFile, FileImportResponse, FileParseResponse } from "@/config/apis/file-import";
import { getProjectsByWorkspace, ProjectOptionType } from "@/config/apis/project";
import { getWorkspaceOptions, WorkspaceOptionType } from "@/config/apis/workspace";
import { useRoute } from "vue-router";

const route = useRoute();
const queryProjectId = route.query.project_id ? Number(route.query.project_id) : undefined;
const formData = ref<{ project_id: number | undefined }>({ project_id: queryProjectId });
const parsing = ref(false);
const uploaded = ref(false);
const uploadResult = ref<FileImportResponse | null>(null);
const parseResult = ref<FileParseResponse | null>(null);
const projectOptions = ref<{ label: string; value: number }[]>([]);
const supportedParseTypes = ["txt", "epub"];
const canParse = computed(
    () => uploadResult.value != null && supportedParseTypes.includes(uploadResult.value.source_type)
);

onMounted(async () => {
    const wsRes = await getWorkspaceOptions();
    for (const ws of wsRes as WorkspaceOptionType[]) {
        const projects = await getProjectsByWorkspace(ws.id);
        for (const p of projects as ProjectOptionType[]) {
            projectOptions.value.push({ label: `${ws.name} / ${p.name}`, value: p.id });
        }
    }
});

function handleUpload(option: RequestOption): any {
    uploadResult.value = null;
    parseResult.value = null;
    return uploadSourceFile(option, formData.value.project_id!)
        .then((res: FileImportResponse) => {
            uploadResult.value = res;
            uploaded.value = true;
            Message.success("文件上传成功");
        })
        .catch(() => {
            Message.error("上传失败");
        });
}

async function handleParse() {
    if (!formData.value.project_id) return;
    parsing.value = true;
    try {
        const result = await parseSourceFile(formData.value.project_id);
        parseResult.value = result;
        Message.success(`解析完成，共 ${result.total_chapters} 个章节`);
    } catch {
        Message.error("解析失败");
    } finally {
        parsing.value = false;
    }
}
</script>
<style lang="scss" scoped></style>
