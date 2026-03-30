<template>
    <div class="activation-page">
        <div class="activation-card">
            <div class="activation-header">
                <img class="activation-icon" src="@/assets/images/vostory-mark.svg" alt="Vostory" />
                <h2>系统授权激活</h2>
                <p class="activation-subtitle">请选择激活方式以启用系统功能</p>
            </div>

            <!-- 机器信息 -->
            <div class="machine-info" v-if="statusInfo">
                <div class="info-row">
                    <span class="info-label">机器指纹</span>
                    <div class="info-value-group">
                        <span class="info-value fingerprint" :title="statusInfo.fingerprint">{{ statusInfo.fingerprint || '-' }}</span>
                        <a-button
                            v-if="statusInfo.fingerprint"
                            type="text"
                            size="mini"
                            class="copy-btn"
                            @click="copyToClipboard(statusInfo.fingerprint!, '机器指纹')"
                        >
                            <icon-copy />
                        </a-button>
                    </div>
                </div>
                <div class="info-row">
                    <span class="info-label">主机名称</span>
                    <span class="info-value">{{ statusInfo.hostname || '-' }}</span>
                </div>
            </div>

            <!-- 激活方式切换 -->
            <div class="mode-tabs">
                <div class="mode-tab" :class="{ active: activeMode === 'online' }" @click="activeMode = 'online'">
                    <icon-cloud />
                    在线激活
                </div>
                <div class="mode-tab" :class="{ active: activeMode === 'offline' }" @click="activeMode = 'offline'">
                    <icon-file />
                    离线激活
                </div>
            </div>

            <!-- 在线激活表单 -->
            <div class="activation-form" v-if="activeMode === 'online'">
                <a-input v-model="licenseCode" placeholder="请输入授权码" size="large" allow-clear @press-enter="handleOnlineActivate">
                    <template #prefix>
                        <icon-lock />
                    </template>
                </a-input>
                <a-button type="primary" long size="large" :loading="loading" @click="handleOnlineActivate">
                    在线激活
                </a-button>
            </div>

            <!-- 离线激活表单 -->
            <div class="activation-form" v-if="activeMode === 'offline'">
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
                        <span>将获取到的 License 内容和公钥粘贴到下方</span>
                    </div>
                </div>
                <a-textarea v-model="licenseFileContent" placeholder="请粘贴 License 文件内容（JSON 格式）" :auto-size="{ minRows: 4, maxRows: 8 }" />
                <a-textarea v-model="publicKey" placeholder="请粘贴产品公钥（PEM 格式）" :auto-size="{ minRows: 3, maxRows: 6 }" class="mt-12" />
                <div class="upload-hint">
                    <icon-info-circle />
                    <span>从授权管理平台下载 License 文件，并复制对应产品的公钥</span>
                </div>
                <a-button type="primary" long size="large" :loading="loading" @click="handleOfflineActivate">
                    离线激活
                </a-button>
            </div>

            <div class="activation-footer">
                <span class="footer-text">Vostory · 小说有声化创作平台</span>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { Message } from "@arco-design/web-vue";
import { getLicenseStatus, activateOnline, activateOffline, type LicenseStatusResponse } from "@/config/apis/license";
import { setLicenseActivated } from "@/packages/vue-router";

const router = useRouter();

const activeMode = ref<"online" | "offline">("online");
const licenseCode = ref("");
const licenseFileContent = ref("");
const publicKey = ref("");
const loading = ref(false);
const statusInfo = ref<LicenseStatusResponse | null>(null);

const fetchStatus = async () => {
    try {
        statusInfo.value = await getLicenseStatus();
        if (statusInfo.value?.activated) {
            router.replace("/login");
        }
    } catch {
        // 忽略错误
    }
};

const handleOnlineActivate = async () => {
    if (!licenseCode.value.trim()) {
        Message.warning("请输入授权码");
        return;
    }
    loading.value = true;
    try {
        await activateOnline(licenseCode.value.trim());
        setLicenseActivated(true);
        Message.success("激活成功");
        router.replace("/login");
    } catch (err: any) {
        Message.error(err?.data?.message || err?.message || "激活失败");
    } finally {
        loading.value = false;
    }
};

const handleOfflineActivate = async () => {
    if (!licenseFileContent.value.trim()) {
        Message.warning("请粘贴 License 文件内容");
        return;
    }
    if (!publicKey.value.trim()) {
        Message.warning("请粘贴产品公钥");
        return;
    }
    loading.value = true;
    try {
        await activateOffline(licenseFileContent.value.trim(), publicKey.value.trim());
        setLicenseActivated(true);
        Message.success("激活成功");
        router.replace("/login");
    } catch (err: any) {
        Message.error(err?.data?.message || err?.message || "激活失败");
    } finally {
        loading.value = false;
    }
};

const copyToClipboard = async (text: string, label: string) => {
    try {
        await navigator.clipboard.writeText(text);
        Message.success(`${label}已复制到剪贴板`);
    } catch {
        const input = document.createElement("textarea");
        input.value = text;
        document.body.appendChild(input);
        input.select();
        document.execCommand("copy");
        document.body.removeChild(input);
        Message.success(`${label}已复制到剪贴板`);
    }
};

onMounted(() => {
    fetchStatus();
});
</script>

<style scoped lang="scss">
.activation-page {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f5f7fa;
    overflow-y: auto;
    padding: 40px 20px;
}

.activation-card {
    width: 100%;
    max-width: 460px;
    background: #fff;
    border-radius: 16px;
    padding: 40px 36px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.activation-header {
    text-align: center;
    margin-bottom: 28px;

    .activation-icon {
        width: 48px;
        height: 48px;
        margin-bottom: 16px;
    }

    h2 {
        margin: 0 0 8px;
        font-size: 22px;
        font-weight: 700;
        color: #1d2129;
    }

    .activation-subtitle {
        margin: 0;
        font-size: 14px;
        color: #9ca3af;
    }
}

.machine-info {
    background: #f8fafc;
    border-radius: 10px;
    padding: 14px 16px;
    margin-bottom: 20px;
    border: 1px solid #f0f0f0;

    .info-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 4px 0;

        & + .info-row {
            margin-top: 8px;
        }
    }

    .info-label {
        font-size: 13px;
        color: #8c8c8c;
        flex-shrink: 0;
    }

    .info-value-group {
        display: flex;
        align-items: center;
        gap: 4px;
        min-width: 0;
        flex: 1;
        justify-content: flex-end;
    }

    .info-value {
        font-size: 13px;
        color: #333;
        text-align: right;
        word-break: break-all;

        &.fingerprint {
            font-family: "SF Mono", "Monaco", "Menlo", monospace;
            font-size: 11px;
            line-height: 1.6;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            margin-left: 20px;
        }
    }

    .copy-btn {
        flex-shrink: 0;
        color: #0f6fff !important;
        padding: 2px 4px;

        :deep(svg) {
            width: 14px;
            height: 14px;
        }
    }
}

.mode-tabs {
    display: flex;
    gap: 12px;
    margin-bottom: 20px;

    .mode-tab {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        padding: 10px;
        border-radius: 8px;
        border: 1px solid #e5e7eb;
        background: #fafafa;
        cursor: pointer;
        font-size: 14px;
        color: #666;
        transition: all 0.2s;

        &:hover {
            border-color: #0f6fff;
            color: #0f6fff;
        }

        &.active {
            border-color: #0f6fff;
            background: rgba(15, 111, 255, 0.04);
            color: #0f6fff;
            font-weight: 500;
        }

        :deep(svg) {
            width: 16px;
            height: 16px;
        }
    }
}

.offline-steps {
    background: #f0f7ff;
    border-radius: 8px;
    padding: 12px 14px;
    margin-bottom: 16px;
    border: 1px solid #e0edff;

    .step-item {
        display: flex;
        align-items: flex-start;
        gap: 10px;
        font-size: 12px;
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
        margin-top: 1px;
    }
}

.activation-form {
    .arco-input-wrapper,
    :deep(.arco-textarea-wrapper) {
        border-radius: 8px;
        background: #f9fafb;
        border: 1px solid #e5e7eb;
        transition: all 0.2s;

        &:hover {
            border-color: #0f6fff;
        }

        &.arco-input-focus,
        &.arco-textarea-focus {
            border-color: #0f6fff;
            background: #fff;
            box-shadow: 0 0 0 2px rgba(15, 111, 255, 0.08);
        }
    }

    .mt-12 {
        margin-top: 12px;
    }

    .upload-hint {
        display: flex;
        align-items: flex-start;
        gap: 6px;
        margin: 10px 0 4px;
        font-size: 12px;
        color: #999;
        line-height: 1.5;

        :deep(svg) {
            width: 14px;
            height: 14px;
            flex-shrink: 0;
            margin-top: 2px;
        }
    }

    .arco-btn-primary {
        margin-top: 16px;
        height: 42px;
        border-radius: 8px;
        font-size: 14px;
        font-weight: 600;
        background: #0f6fff;
        border: none;
        transition: all 0.2s;

        &:hover {
            background: #0a56d6;
        }
    }
}

.activation-footer {
    margin-top: 32px;
    text-align: center;

    .footer-text {
        font-size: 12px;
        color: #c4c4c4;
        letter-spacing: 1px;
    }
}
</style>
