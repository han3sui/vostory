<template>
    <frame-view>
        <a-card title="对象存储配置">
            <a-alert type="info" style="margin-bottom: 24px">
                配置 S3 兼容的对象存储（华为云 OBS、阿里云 OSS、腾讯云 COS、MinIO 等）。
                将后端 <code>storage.type</code> 设为 <code>oss</code> 后，合成音频、参考音色、导出文件会写入这里。
            </a-alert>
            <a-form
                ref="formRef"
                :model="formData"
                :rules="rules"
                auto-label-width
                layout="vertical"
                style="max-width: 600px"
            >
                <a-form-item label="存储类型" field="type">
                    <a-select v-model="formData.type" placeholder="请选择存储类型" @change="onTypeChange">
                        <a-option value="huawei">华为云 OBS</a-option>
                        <a-option value="aliyun">阿里云 OSS</a-option>
                        <a-option value="tencent">腾讯云 COS</a-option>
                        <a-option value="minio">MinIO</a-option>
                        <a-option value="other">其他S3兼容</a-option>
                    </a-select>
                </a-form-item>

                <a-form-item label="服务端点 (Endpoint)" field="endpoint">
                    <a-input v-model.trim="formData.endpoint" placeholder="如 obs.cn-south-1.myhuaweicloud.com" />
                    <template #extra>
                        <span class="text-gray-400">不含协议前缀（http:// 或 https://），协议由下方 SSL 开关决定</span>
                    </template>
                </a-form-item>

                <a-form-item label="AccessKey (AK)" field="ak">
                    <a-input v-model.trim="formData.ak" placeholder="请输入 AccessKey" />
                </a-form-item>

                <a-form-item label="SecretKey (SK)" field="sk">
                    <a-input-password v-model="formData.sk" placeholder="请输入 SecretKey，不修改请留空" />
                </a-form-item>

                <a-form-item label="桶名称 (Bucket)" field="bucket">
                    <a-input v-model.trim="formData.bucket" placeholder="请输入桶名称" />
                </a-form-item>

                <a-form-item label="路径前缀 (Prefix)" field="prefix">
                    <a-input v-model="formData.prefix" placeholder="建议留空" />
                    <template #extra>
                        <span class="text-gray-400">
                            全局前缀，会加在所有对象 Key 之前，一般留空；如需收拢到某目录须以 / 结尾。
                        </span>
                    </template>
                </a-form-item>

                <a-form-item label="CDN 访问域名" field="cdn_base">
                    <a-input
                        v-model.trim="formData.cdn_base"
                        placeholder="如 https://cdn.example.com，音频播放仍走后端流接口"
                    />
                </a-form-item>

                <a-form-item label="启用 SSL" field="use_ssl">
                    <a-switch
                        :model-value="formData.use_ssl !== 'false'"
                        @change="(val: boolean) => (formData.use_ssl = val ? 'true' : 'false')"
                    />
                </a-form-item>

                <a-form-item>
                    <a-space>
                        <a-button type="primary" :loading="saving" @click="handleSave">保存配置</a-button>
                        <a-button :loading="testing" @click="handleTest">测试连接</a-button>
                    </a-space>
                </a-form-item>
            </a-form>
        </a-card>
    </frame-view>
</template>

<script lang="ts" setup>
import { Message } from "@arco-design/web-vue";
import { getOSSConfig, updateOSSConfig, testOSSConnection, OSSConfigType } from "@/config/apis/oss-config";

const formRef = ref();
const saving = ref(false);
const testing = ref(false);

const formData = ref<OSSConfigType>({
    type: "",
    endpoint: "",
    ak: "",
    sk: "",
    bucket: "",
    prefix: "",
    use_ssl: "true",
    cdn_base: ""
});

const rules = {
    type: [{ required: true, message: "请选择存储类型" }],
    endpoint: [{ required: true, message: "请输入服务端点" }],
    ak: [{ required: true, message: "请输入 AccessKey" }],
    bucket: [{ required: true, message: "请输入桶名称" }]
};

const endpointMap: Record<string, string> = {
    huawei: "obs.cn-south-1.myhuaweicloud.com",
    aliyun: "oss-cn-hangzhou.aliyuncs.com",
    tencent: "cos.ap-guangzhou.myqcloud.com",
    minio: "localhost:9000"
};

function onTypeChange(val: string) {
    if (endpointMap[val] && !formData.value.endpoint) {
        formData.value.endpoint = endpointMap[val];
    }
}

async function loadConfig() {
    try {
        const res = await getOSSConfig();
        if (res) {
            formData.value = { ...formData.value, ...res };
        }
    } catch {
        // 初次加载可能没有配置
    }
}

async function handleSave() {
    const valid = await formRef.value?.validate();
    if (valid) return;
    saving.value = true;
    try {
        await updateOSSConfig(formData.value);
        Message.success("保存成功");
        await loadConfig();
    } finally {
        saving.value = false;
    }
}

async function handleTest() {
    const valid = await formRef.value?.validate();
    if (valid) return;
    testing.value = true;
    try {
        await testOSSConnection(formData.value);
        Message.success("连接成功");
    } finally {
        testing.value = false;
    }
}

onMounted(() => {
    loadConfig();
});
</script>
<style lang="scss" scoped></style>
