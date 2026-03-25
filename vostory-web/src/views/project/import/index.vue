<template>
    <div class="p-6 max-w-2xl">
        <a-card title="文件导入">
            <a-form :model="{}" layout="vertical">
                <a-form-item label="上传源文件" required>
                    <a-upload
                        :custom-request="handleUpload"
                        :limit="1"
                        accept=".txt,.docx,.epub"
                    >
                        <template #upload-button>
                            <a-button type="primary">
                                选择文件（支持 txt / docx / epub）
                            </a-button>
                        </template>
                    </a-upload>
                </a-form-item>
            </a-form>

            <a-divider v-if="uploadResult" />

            <a-descriptions v-if="uploadResult" title="上传结果" :column="1" bordered size="small">
                <a-descriptions-item label="文件名">{{ uploadResult.file_name }}</a-descriptions-item>
                <a-descriptions-item label="文件大小"
                    >{{ (uploadResult.file_size / 1024).toFixed(1) }} KB</a-descriptions-item
                >
                <a-descriptions-item label="文件类型">{{ uploadResult.source_type }}</a-descriptions-item>
                <a-descriptions-item label="解析状态">
                    <a-tag color="blue">后台自动解析中，请在「章节管理」中查看结果</a-tag>
                </a-descriptions-item>
            </a-descriptions>
        </a-card>
    </div>
</template>
<script lang="ts" setup>
import { Message, RequestOption } from "@arco-design/web-vue";
import { uploadSourceFile, FileImportResponse } from "@/config/apis/file-import";

const props = defineProps<{ projectId: number }>();

const uploadResult = ref<FileImportResponse | null>(null);

function handleUpload(option: RequestOption): any {
    uploadResult.value = null;
    return uploadSourceFile(option, props.projectId)
        .then((res: FileImportResponse) => {
            uploadResult.value = res;
            Message.success("文件上传成功，后台正在自动解析");
        })
        .catch(() => {
            Message.error("上传失败");
        });
}
</script>
<style lang="scss" scoped></style>
