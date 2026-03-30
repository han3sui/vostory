export const baseRouter = [
    {
        path: "/:pathMatch(.*)*",
        name: "404",
        meta: {
            title: "找不到页面",
            hidden: true
        },
        // eslint-disable-next-line @syyfe/validate-router
        component: (): any => import("@/views/404/index.vue")
    },
    {
        path: "/login",
        name: "login",
        meta: {
            title: "登录",
            hidden: true
        },
        component: (): any => import("@/views/login/index.vue")
    },
    {
        path: "/activation",
        name: "activation",
        meta: {
            title: "授权激活",
            hidden: true
        },
        component: (): any => import("@/views/activation/index.vue")
    }
];
