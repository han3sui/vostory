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
            sort: 40
        },
        component: Layout,
        children: [
            {
                path: "/project/list",
                name: "project-list",
                meta: {
                    title: "项目列表",
                    sort: 2,
                    permission: () => hasPermission("project:list")
                },
                component: () => import("@/views/project/index.vue")
            },
            {
                path: "/project/detail/:id",
                name: "project-detail",
                meta: {
                    title: "项目详情",
                    sort: 1,
                    hidden: true,
                    parentName: "project-list",
                    navTag: true,
                    permission: () => hasPermission("project:list")
                },
                component: () => import("@/views/project/detail/index.vue")
            }
        ]
    }
];

export default routers;
