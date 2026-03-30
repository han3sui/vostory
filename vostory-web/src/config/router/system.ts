import { RouteConfig } from "types";
import Layout from "@/layout/index.vue";
import { hasPermission } from "@/views/utils";

const routers: RouteConfig[] = [
    {
        path: "/system",
        name: "system",
        meta: {
            title: "系统管理",
            icon: "menu-system",
            sort: 2
        },
        component: Layout,
        children: [
            {
                path: "/system/post",
                name: "system-post",
                meta: {
                    title: "岗位管理",
                    sort: 5,
                    permission: () => hasPermission("system:post:list")
                },
                component: () => import("@/views/system/post/index.vue")
            },
            {
                path: "/system/dept",
                name: "system-dept",
                meta: {
                    title: "部门管理",
                    sort: 4,
                    permission: () => hasPermission("system:dept:list")
                },
                component: () => import("@/views/system/dept/index.vue")
            },
            {
                path: "/system/menu",
                name: "system-menu",
                meta: {
                    title: "菜单管理",
                    sort: 2,
                    permission: () => hasPermission("system:menu:list")
                },
                component: () => import("@/views/system/menu/index.vue")
            },
            {
                path: "/system/role",
                name: "system-role",
                meta: {
                    title: "角色管理",
                    sort: 3,
                    permission: () => hasPermission("system:role:list")
                },
                component: () => import("@/views/system/role/index.vue")
            },

            {
                path: "/system/dict/list",
                name: "system-dict-list",
                meta: {
                    title: "字典管理",
                    sort: 6,
                    keepAliveName: "AliveDictList",
                    permission: () => hasPermission("system:dict:list")
                },
                component: () => import("@/views/system/dict/list/index.vue")
            },
            {
                path: "/system/dict/detail",
                name: "system-dict-detail",
                meta: {
                    title: "字典详情",
                    sort: 6,
                    parentName: "system-dict-list",
                    hidden: true,
                    navTag: true,
                    permission: () => hasPermission("system:dict:list")
                },
                component: () => import("@/views/system/dict/detail/index.vue")
            },
            {
                path: "/system/operlog",
                name: "system-operlog",
                meta: {
                    title: "操作日志",
                    sort: 8,
                    permission: () => hasPermission("system:operlog:list")
                },
                component: () => import("@/views/system/operlog/index.vue")
            },
            {
                path: "/system/user",
                name: "system-user",
                meta: {
                    title: "用户管理",
                    sort: 1,
                    permission: () => hasPermission("system:user:list")
                },
                component: () => import("@/views/system/user/index.vue")
            },
            {
                path: "/system/logininfor",
                name: "system-logininfor",
                meta: {
                    title: "登录日志",
                    sort: 1,
                    permission: () => hasPermission("system:logininfor:list")
                },
                component: () => import("@/views/system/logininfor/index.vue")
            },
            {
                path: "/system/api",
                name: "system-api",
                meta: {
                    title: "接口管理",
                    sort: 1,
                    permission: () => hasPermission("system:api:list")
                },
                component: () => import("@/views/system/api/index.vue")
            },
            {
                path: "/system/license",
                name: "system-license",
                meta: {
                    title: "授权管理",
                    sort: 10
                },
                component: () => import("@/views/system/license/index.vue")
            }
        ]
    }
];

export default routers;
