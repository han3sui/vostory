<template>
    <div class="logo" :style="{ maxWidth: !collapsed ? `${LAYOUT_SIZE.SIDER_WIDTH}px` : '' }" @click="toDefaultPage">
        <img class="logo-icon" src="@/assets/images/vostory-mark.svg" alt="Vostory" />
        <span v-show="!collapsed" class="title">Vostory</span>
    </div>
</template>
<script lang="ts" setup name="AppLogo">
import router, { getDefaultRoute } from "@/packages/vue-router";
import global from "@/config/pinia/global";
import { LAYOUT_SIZE } from "@/layout/constants";

const collapsed = computed(() => global().collapsed);

function toDefaultPage() {
    const defaultPage = getDefaultRoute();
    if (!defaultPage) {
        return;
    }
    router.push(defaultPage);
}
</script>
<style lang="scss" scoped>
.logo {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 60px;
    cursor: pointer;
    user-select: none;
    background-color: var(--color-menu-light-bg);
    gap: 10px;
    transition: opacity 0.2s;

    &:hover {
        opacity: 0.75;
    }

    .logo-icon {
        width: 32px;
        height: 32px;
        flex-shrink: 0;
    }

    .title {
        font-size: 17px;
        font-weight: 700;
        letter-spacing: 0.5px;
        background: linear-gradient(135deg, #ff8a3d 0%, #ff5b4d 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
        white-space: nowrap;
    }
}
</style>
