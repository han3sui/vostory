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
            }
        ]
    }
];

export default routers;
