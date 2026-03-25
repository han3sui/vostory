import { RouteConfig } from "types";
import Layout from "@/layout/index.vue";
import { hasPermission } from "@/views/utils";

const routers: RouteConfig[] = [
    {
        path: "/project",
        name: "project",
        meta: {
            title: "项目管理",
            icon: "menu-project",
            sort: 0
        },
        component: Layout,
        children: [
            {
                path: "/project/list",
                name: "project-list",
                meta: {
                    title: "项目列表",
                    sort: 1,
                    permission: () => hasPermission("project:list")
                },
                component: () => import("@/views/project/index.vue")
            },
            {
                path: "/project/chapter",
                name: "project-chapter",
                meta: {
                    title: "章节管理",
                    sort: 2,
                    permission: () => hasPermission("chapter:list")
                },
                component: () => import("@/views/chapter/index.vue")
            },
            {
                path: "/project/script-segment",
                name: "project-script-segment",
                meta: {
                    title: "脚本片段",
                    sort: 3,
                    permission: () => hasPermission("script-segment:list")
                },
                component: () => import("@/views/script-segment/index.vue")
            },
            {
                path: "/project/character",
                name: "project-character",
                meta: {
                    title: "角色管理",
                    sort: 4,
                    permission: () => hasPermission("character:list")
                },
                component: () => import("@/views/character/index.vue")
            },
            {
                path: "/project/import",
                name: "project-import",
                meta: {
                    title: "文件导入",
                    sort: 5,
                    permission: () => hasPermission("project:import:upload")
                },
                component: () => import("@/views/project/import/index.vue")
            },
            {
                path: "/project/script-editor",
                name: "project-script-editor",
                meta: {
                    title: "脚本编辑",
                    sort: 6,
                    permission: () => hasPermission("script-segment:list")
                },
                component: () => import("@/views/script-editor/index.vue")
            },
            {
                path: "/project/voice-profile",
                name: "project-voice-profile",
                meta: {
                    title: "声音配置",
                    sort: 7,
                    permission: () => hasPermission("voice-profile:list")
                },
                component: () => import("@/views/voice-profile/index.vue")
            },
            {
                path: "/project/pronunciation-dict",
                name: "project-pronunciation-dict",
                meta: {
                    title: "发音词典",
                    sort: 8,
                    permission: () => hasPermission("pronunciation-dict:list")
                },
                component: () => import("@/views/pronunciation-dict/index.vue")
            }
        ]
    }
];

export default routers;
