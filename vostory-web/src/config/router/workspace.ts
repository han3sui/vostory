import { RouteConfig } from "types";
import Layout from "@/layout/index.vue";
import { hasPermission } from "@/views/utils";

const routers: RouteConfig[] = [
    {
        path: "/workspace",
        name: "workspace",
        meta: {
            title: "工作空间",
            icon: "menu-workspace",
            sort: 0
        },
        component: Layout,
        children: [
            {
                path: "/workspace/list",
                name: "workspace-list",
                meta: {
                    title: "空间管理",
                    sort: 1,
                    permission: () => hasPermission("workspace:list")
                },
                component: () => import("@/views/workspace/index.vue")
            }
        ]
    }
];

export default routers;
