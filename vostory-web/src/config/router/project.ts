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
            }
        ]
    }
];

export default routers;
