import { RouteConfig } from "types";
import Layout from "@/layout/index.vue";
import { hasPermission } from "@/views/utils";

const routers: RouteConfig[] = [
    {
        path: "/ai",
        name: "ai",
        meta: {
            title: "AI 配置",
            icon: "menu-ai",
            sort: 20
        },
        component: Layout,
        children: [
            {
                path: "/ai/llm-provider",
                name: "ai-llm-provider",
                meta: {
                    title: "LLM 提供商",
                    sort: 4,
                    permission: () => hasPermission("ai:llm-provider:list")
                },
                component: () => import("@/views/ai/llm-provider/index.vue")
            },
            {
                path: "/ai/tts-provider",
                name: "ai-tts-provider",
                meta: {
                    title: "TTS 提供商",
                    sort: 3,
                    permission: () => hasPermission("ai:tts-provider:list")
                },
                component: () => import("@/views/ai/tts-provider/index.vue")
            },
            {
                path: "/ai/prompt-template",
                name: "ai-prompt-template",
                meta: {
                    title: "Prompt 模板",
                    sort: 2,
                    permission: () => hasPermission("ai:prompt-template:list")
                },
                component: () => import("@/views/ai/prompt-template/index.vue")
            },
            {
                path: "/ai/voice-asset",
                name: "ai-voice-asset",
                meta: {
                    title: "音色管理",
                    sort: 5,
                    permission: () => hasPermission("voice-asset:list")
                },
                component: () => import("@/views/ai/voice-asset/index.vue")
            },
            {
                path: "/ai/llm-log",
                name: "ai-llm-log",
                meta: {
                    title: "LLM 调用日志",
                    sort: 1,
                    permission: () => hasPermission("ai:llm-log:list")
                },
                component: () => import("@/views/ai/llm-log/index.vue")
            }
        ]
    }
];

export default routers;
