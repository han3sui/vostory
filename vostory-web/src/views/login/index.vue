<template>
    <div class="login-page">
        <div class="login-left">
            <div class="brand-area">
                <img class="brand-logo" src="@/assets/images/vostory-logo.svg" alt="Vostory" />
                <p class="brand-slogan">用声音，讲述你的故事</p>
            </div>
            <div class="decorative-elements">
                <div class="float-card card-1">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
                    <span>剧本编辑</span>
                </div>
                <div class="float-card card-2">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4M12 15a3 3 0 003-3V5a3 3 0 00-6 0v7a3 3 0 003 3z"/></svg>
                    <span>语音合成</span>
                </div>
                <div class="float-card card-3">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M15.536 8.464a5 5 0 010 7.072M18.364 5.636a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707A1 1 0 0112 5.586v12.828a1 1 0 01-1.707.707L5.586 15z"/></svg>
                    <span>角色配音</span>
                </div>
            </div>
        </div>

        <div class="login-right">
            <div class="login-card">
                <div class="login-header">
                    <img class="login-icon" src="@/assets/images/vostory-mark.svg" alt="Vostory" />
                    <h2>欢迎回来</h2>
                    <p class="login-subtitle">登录你的 Vostory 账户</p>
                </div>
                <div class="login-form">
                    <arco-form ref="baseForm1" v-model="formData" :config="formConfig" size="large"></arco-form>
                    <a-button type="primary" long size="large" :loading="loading" :disabled="loading" @click="handleSubmit">
                        登录
                    </a-button>
                </div>
                <div class="login-footer">
                    <span class="footer-text">Vostory · 小说有声化创作平台</span>
                </div>
            </div>
        </div>
    </div>
</template>
<script setup lang="ts">
import { ArcoForm, formHelper, ruleHelper } from "@easyfe/admin-component";
import { Message } from "@arco-design/web-vue";
import { getDefaultRoute } from "@/packages/vue-router";
import { initGlobal, clearLoingInfo, encryptPassword } from "@/views/utils/index";
import { login } from "@/config/apis";
import global from "@/config/pinia/global";
import storage from "@/utils/tools/storage";
const router = useRouter();

const loading = ref(false);

const formData = ref({
    username: "",
    password: ""
});
const formConfig = computed(() => {
    return [
        formHelper.input("", "username", {
            onPressEnter: handleSubmit,
            hideLabel: true,
            placeholder: "请输入用户名",
            span: 24,
            rules: [ruleHelper.require("用户名必填", "blur")]
        }),
        formHelper.input("", "password", {
            onPressEnter: handleSubmit,
            hideLabel: true,
            span: 24,
            type: "password",
            placeholder: "请输入密码"
        })
    ];
});
const baseForm1 = ref();
const handleSubmit = async (): Promise<any> => {
    const v = await baseForm1.value.validate();
    if (v) return;
    global().initSuccess = false;
    loading.value = true;
    try {
        const loginRes = await login({
            login_name: formData.value.username,
            password: encryptPassword(formData.value.password)
        });
        storage.setToken(loginRes.token);
        await initGlobal();
        Message.success("登录成功");
        router.replace(getDefaultRoute() || { path: "/" });
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    clearLoingInfo();
});
</script>

<style scoped lang="scss">
.login-page {
    position: fixed;
    inset: 0;
    display: flex;
    background: #faf8f6;
}

.login-left {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    position: relative;
    background: linear-gradient(145deg, #fff5ed 0%, #fff0f0 40%, #f5edff 100%);
    overflow: hidden;

    &::before {
        content: "";
        position: absolute;
        width: 600px;
        height: 600px;
        border-radius: 50%;
        background: radial-gradient(circle, rgba(255, 138, 61, 0.08) 0%, transparent 70%);
        top: -100px;
        left: -100px;
    }

    &::after {
        content: "";
        position: absolute;
        width: 500px;
        height: 500px;
        border-radius: 50%;
        background: radial-gradient(circle, rgba(199, 58, 245, 0.06) 0%, transparent 70%);
        bottom: -80px;
        right: -80px;
    }
}

.brand-area {
    position: relative;
    z-index: 2;
    text-align: center;

    .brand-logo {
        width: 320px;
        filter: drop-shadow(0 8px 32px rgba(255, 91, 77, 0.12));
    }

    .brand-slogan {
        margin-top: 24px;
        font-size: 18px;
        color: #6b6476;
        letter-spacing: 4px;
        font-weight: 300;
    }
}

.decorative-elements {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 1;
}

.float-card {
    position: absolute;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 14px 22px;
    background: rgba(255, 255, 255, 0.85);
    backdrop-filter: blur(12px);
    border-radius: 16px;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.04), 0 1px 4px rgba(0, 0, 0, 0.02);
    font-size: 14px;
    color: #3d3650;
    font-weight: 500;
    animation: float 6s ease-in-out infinite;

    svg {
        width: 22px;
        height: 22px;
        flex-shrink: 0;
    }

    &.card-1 {
        top: 18%;
        left: 8%;
        color: #ff8a3d;
        animation-delay: 0s;
    }

    &.card-2 {
        bottom: 28%;
        left: 12%;
        color: #ff5b4d;
        animation-delay: -2s;
    }

    &.card-3 {
        top: 30%;
        right: 8%;
        color: #c73af5;
        animation-delay: -4s;
    }
}

@keyframes float {
    0%,
    100% {
        transform: translateY(0);
    }
    50% {
        transform: translateY(-12px);
    }
}

.login-right {
    width: 520px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px;
    background: #fff;
    box-shadow: -8px 0 40px rgba(0, 0, 0, 0.03);
}

.login-card {
    width: 100%;
    max-width: 380px;
}

.login-header {
    text-align: center;
    margin-bottom: 40px;

    .login-icon {
        width: 56px;
        height: 56px;
        margin-bottom: 20px;
        filter: drop-shadow(0 4px 12px rgba(255, 91, 77, 0.2));
    }

    h2 {
        margin: 0 0 8px;
        font-size: 26px;
        font-weight: 700;
        color: #18181b;
    }

    .login-subtitle {
        margin: 0;
        font-size: 14px;
        color: #9ca3af;
    }
}

.login-form {
    :deep(.arco-form) {
        .arco-form-item {
            margin-bottom: 20px;
        }

        .arco-input-wrapper {
            border-radius: 10px;
            background: #f9fafb;
            border: 1px solid #f0f0f0;
            transition: all 0.2s;

            &:hover {
                border-color: #ff8a3d;
            }

            &.arco-input-focus {
                border-color: #ff5b4d;
                background: #fff;
                box-shadow: 0 0 0 3px rgba(255, 91, 77, 0.08);
            }
        }
    }

    .arco-btn-primary {
        margin-top: 8px;
        height: 44px;
        border-radius: 10px;
        font-size: 15px;
        font-weight: 600;
        background: linear-gradient(135deg, #ff8a3d 0%, #ff5b4d 100%);
        border: none;
        transition: all 0.3s;

        &:hover {
            transform: translateY(-1px);
            box-shadow: 0 6px 20px rgba(255, 91, 77, 0.3);
        }

        &:active {
            transform: translateY(0);
        }
    }
}

.login-footer {
    margin-top: 48px;
    text-align: center;

    .footer-text {
        font-size: 12px;
        color: #c4c4c4;
        letter-spacing: 1px;
    }
}

@media (max-width: 960px) {
    .login-left {
        display: none;
    }

    .login-right {
        width: 100%;
    }
}
</style>
