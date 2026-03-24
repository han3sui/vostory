<template>
    <div class="login">
        <img class="info" src="@/assets/images/login/login.png" />
        <div class="content">
            <h3>管理系统 - 登录</h3>
            <arco-form ref="baseForm1" v-model="formData" :config="formConfig" size="medium"></arco-form>
            <a-button type="primary" long :loading="loading" :disabled="loading" @click="handleSubmit">登录</a-button>
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
            placeholder: "用户名",
            span: 24,
            rules: [ruleHelper.require("用户名必填", "blur")]
        }),
        formHelper.input("", "password", {
            onPressEnter: handleSubmit,
            hideLabel: true,
            span: 24,
            type: "password",
            placeholder: "密码"
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
.login {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    background: url("@/assets/images/login/login-bg.png") center no-repeat;
    background-size: cover;
    .info {
        width: 650px;
        position: fixed;
        top: 11%;
        right: 55%;
        animation: turn 100s linear infinite;
    }

    .content {
        position: fixed;
        top: 30%;
        right: 15%;
        width: 400px;
        padding: 60px 30px;
        border-radius: 8px;
        background-color: var(--color-bg-1);
        z-index: 9999;
        h3 {
            margin: 0 0 32px 0;
            font-size: 22px;
            font-weight: 600;
        }
        input {
            width: 100%;
        }
    }
}
@keyframes turn {
    0% {
        -webkit-transform: rotate(0deg);
    }
    25% {
        -webkit-transform: rotate(90deg);
    }
    50% {
        -webkit-transform: rotate(180deg);
    }
    75% {
        -webkit-transform: rotate(270deg);
    }
    100% {
        -webkit-transform: rotate(360deg);
    }
}
</style>
