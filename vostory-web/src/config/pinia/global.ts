import { defineStore } from "pinia";
import { UserDetailType } from "../apis/system";

export default defineStore({
    id: "global",
    state: () => ({
        //项目配置
        app: {
            layout: <"left" | "top" | "mix">"left"
        },
        // 滚动条位置
        scrollTop: <Record<string, number>>{},
        userInfo: <UserDetailType | null>null,
        userMenu: <string[]>[],
        //是否收起菜单
        collapsed: false,
        initSuccess: false
    }),
    actions: {
        SET_SCROLL(res: { name: string; value: number }): void {
            this.scrollTop[res.name] = res.value;
        }
    }
});
