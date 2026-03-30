<template>
    <frame-view>
        <a-spin :loading="pageLoading" style="width: 100%">
            <!-- 已激活：显示授权状态 -->
            <template v-if="statusInfo?.activated">
                <a-card title="授权信息">
                    <a-descriptions :column="2" bordered size="medium">
                        <a-descriptions-item label="激活状态">
                            <a-tag color="green">已激活</a-tag>
                        </a-descriptions-item>
                        <a-descriptions-item label="激活方式">
                            {{ statusInfo.mode === "online" ? "在线激活" : statusInfo.mode === "offline" ? "离线激活" : "-" }}
                        </a-descriptions-item>
                        <a-descriptions-item label="授权码">{{ statusInfo.license_code || "-" }}</a-descriptions-item>
                        <a-descriptions-item label="产品编码">{{ statusInfo.product_code || "-" }}</a-descriptions-item>
                        <a-descriptions-item label="授权类型">
                            {{ statusInfo.license_type === "permanent" ? "永久" : statusInfo.license_type === "subscription" ? "订阅" : "-" }}
                        </a-descriptions-item>
                        <a-descriptions-item label="到期时间">
                            <template v-if="statusInfo.license_type === 'permanent'">
                                <a-tag color="blue">永久有效</a-tag>
                            </template>
                            <template v-else-if="statusInfo.expires_at">
                                <span :class="{ 'expires-warn': isExpiringSoon }">{{ formatDate(statusInfo.expires_at) }}</span>
                                <a-tag v-if="isExpiringSoon" color="orange" style="margin-left: 8px">即将到期</a-tag>
                            </template>
                            <template v-else>-</template>
                        </a-descriptions-item>
                        <a-descriptions-item label="机器指纹" :span="2">
                            <div style="display: flex; align-items: center; gap: 8px">
                                <span class="fingerprint-text">{{ statusInfo.fingerprint || "-" }}</span>
                                <a-button v-if="statusInfo.fingerprint" type="text" size="mini" @click="copyText(statusInfo.fingerprint!, '机器指纹')">
                                    <icon-copy />
                                </a-button>
                            </div>
                        </a-descriptions-item>
                        <a-descriptions-item label="主机名称">{{ statusInfo.hostname || "-" }}</a-descriptions-item>
                        <a-descriptions-item label="功能特性">{{ statusInfo.features || "-" }}</a-descriptions-item>
                    </a-descriptions>
                    <div style="margin-top: 24px">
                        <a-popconfirm content="取消激活后系统将无法正常使用，需要重新激活才能恢复，确认继续？" @ok="handleDeactivate">
                            <a-button type="primary" status="danger" :loading="deactivateLoading">取消激活</a-button>
                        </a-popconfirm>
                    </div>
                </a-card>
            </template>

            <!-- 未激活：显示激活表单 -->
            <template v-else>
                <a-card title="系统授权激活" style="max-width: 600px">
                    <a-alert type="warning" style="margin-bottom: 20px">系统尚未激活，请完成授权激活以启用全部功能。</a-alert>

                    <div class="machine-info">
                        <a-descriptions :column="1" size="small">
                            <a-descriptions-item label="机器指纹">
                                <div style="display: flex; align-items: center; gap: 8px">
                                    <span class="fingerprint-text">{{ statusInfo?.fingerprint || "-" }}</span>
                                    <a-button
                                        v-if="statusInfo?.fingerprint"
                                        type="text"
                                        size="mini"
                                        @click="copyText(statusInfo!.fingerprint!, '机器指纹')"
                                    >
                                        <icon-copy />
                                    </a-button>
                                </div>
                            </a-descriptions-item>
                            <a-descriptions-item label="主机名称">{{ statusInfo?.hostname || "-" }}</a-descriptions-item>
                        </a-descriptions>
                    </div>

                    <a-tabs v-model:active-key="activeMode">
                        <a-tab-pane key="online" title="在线激活">
                            <a-form :model="{}" layout="vertical" style="margin-top: 8px">
                                <a-form-item label="授权码">
                                    <a-input v-model="licenseCode" placeholder="请输入授权码" allow-clear />
                                </a-form-item>
                                <a-form-item>
                                    <a-button type="primary" :loading="activateLoading" @click="handleOnlineActivate">在线激活</a-button>
                                </a-form-item>
                            </a-form>
                        </a-tab-pane>
                        <a-tab-pane key="offline" title="离线激活">
                            <div class="offline-steps">
                                <div class="step-item">
                                    <span class="step-num">1</span>
                                    <span>复制上方「机器指纹」发送给授权管理员</span>
                                </div>
                                <div class="step-item">
                                    <span class="step-num">2</span>
                                    <span>管理员在授权平台绑定指纹并生成 License 文件</span>
                                </div>
                                <div class="step-item">
                                    <span class="step-num">3</span>
                                    <span>选择或粘贴 .lic 文件内容到下方</span>
                                </div>
                            </div>
                            <a-form :model="{}" layout="vertical" style="margin-top: 12px">
                                <a-form-item label="License 文件内容">
                                    <a-textarea v-model="licenseFileContent" placeholder="请粘贴 .lic 文件中的内容，或点击下方按钮选择文件" :auto-size="{ minRows: 4, maxRows: 8 }" />
                                </a-form-item>
                                <a-form-item>
                                    <a-space>
                                        <a-button type="primary" :loading="activateLoading" @click="handleOfflineActivate">离线激活</a-button>
                                        <a-button @click="triggerFileSelect">选择 .lic 文件</a-button>
                                    </a-space>
                                    <input ref="fileInputRef" type="file" accept=".lic" style="display: none" @change="handleFileSelect" />
                                </a-form-item>
                            </a-form>
                        </a-tab-pane>
                    </a-tabs>
                </a-card>
            </template>
        </a-spin>
    </frame-view>
</template>

<script lang="ts" setup>
import { Message } from "@arco-design/web-vue";
import { getLicenseStatus, activateOnline, activateOffline, deactivateLicense, type LicenseStatusResponse } from "@/config/apis/license";
import dayjs from "dayjs";

const router = useRouter();

const pageLoading = ref(false);
const statusInfo = ref<LicenseStatusResponse | null>(null);
const activeMode = ref("online");
const licenseCode = ref("");
const licenseFileContent = ref("");
const fileInputRef = ref<HTMLInputElement>();
const activateLoading = ref(false);
const deactivateLoading = ref(false);

const isExpiringSoon = computed(() => {
    if (!statusInfo.value?.expires_at || statusInfo.value.license_type === "permanent") return false;
    return dayjs(statusInfo.value.expires_at).diff(dayjs(), "day") <= 30;
});

async function fetchStatus() {
    pageLoading.value = true;
    try {
        statusInfo.value = await getLicenseStatus();
    } catch {
        // ignore
    } finally {
        pageLoading.value = false;
    }
}

async function handleOnlineActivate() {
    if (!licenseCode.value.trim()) {
        Message.warning("请输入授权码");
        return;
    }
    activateLoading.value = true;
    try {
        await activateOnline(licenseCode.value.trim());
        Message.success("激活成功");
        await fetchStatus();
    } finally {
        activateLoading.value = false;
    }
}

function triggerFileSelect() {
    fileInputRef.value?.click();
}

function handleFileSelect(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
        licenseFileContent.value = (reader.result as string).trim();
        Message.success("文件内容已读取");
    };
    reader.readAsText(file);
    (e.target as HTMLInputElement).value = "";
}

async function handleOfflineActivate() {
    if (!licenseFileContent.value.trim()) {
        Message.warning("请粘贴 License 文件内容");
        return;
    }
    activateLoading.value = true;
    try {
        await activateOffline(licenseFileContent.value.trim());
        Message.success("激活成功");
        await fetchStatus();
    }finally {
        activateLoading.value = false;
    }
}

async function handleDeactivate() {
    deactivateLoading.value = true;
    try {
        await deactivateLicense();
        Message.success("已取消激活");
        await fetchStatus();
    } finally {
        deactivateLoading.value = false;
    }
}

function formatDate(v: string) {
    const d = dayjs(v);
    return d.isValid() ? d.format("YYYY-MM-DD HH:mm:ss") : v;
}

async function copyText(text: string, label: string) {
    try {
        await navigator.clipboard.writeText(text);
        Message.success(`${label}已复制到剪贴板`);
    } catch {
        const el = document.createElement("textarea");
        el.value = text;
        document.body.appendChild(el);
        el.select();
        document.execCommand("copy");
        document.body.removeChild(el);
        Message.success(`${label}已复制到剪贴板`);
    }
}

onMounted(() => {
    fetchStatus();
});
</script>

<style lang="scss" scoped>
.fingerprint-text {
    font-family: "SF Mono", "Monaco", "Menlo", monospace;
    font-size: 12px;
    color: #333;
    word-break: break-all;
}

.expires-warn {
    color: #ff7d00;
    font-weight: 500;
}

.machine-info {
    background: #f8fafc;
    border-radius: 8px;
    padding: 12px 16px;
    margin-bottom: 16px;
    border: 1px solid #f0f0f0;
}

.offline-steps {
    background: #f0f7ff;
    border-radius: 8px;
    padding: 12px 14px;
    margin-top: 8px;
    border: 1px solid #e0edff;

    .step-item {
        display: flex;
        align-items: flex-start;
        gap: 10px;
        font-size: 13px;
        color: #4a5568;
        line-height: 1.5;

        & + .step-item {
            margin-top: 6px;
        }
    }

    .step-num {
        flex-shrink: 0;
        width: 18px;
        height: 18px;
        border-radius: 50%;
        background: #0f6fff;
        color: #fff;
        font-size: 10px;
        font-weight: 600;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-top: 2px;
    }
}
</style>
