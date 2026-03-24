import request from "@/packages/request";
import { UserDetailType } from "./system";

/**
 * 获取用户信息
 * @returns
 */
export function getUserInfo(): Promise<{ user: UserDetailType; permissions: string[] }> {
    return request({
        url: "/api/v1/user/info",
        method: "get"
    });
}

/**
 * 登录
 * @param data
 * @returns
 */
export function login(data: { login_name: string; password: string }): Promise<{ token: string }> {
    return request({
        url: "/api/v1/user/login",
        method: "post",
        data
    });
}

/**
 * 退出登录
 * @returns
 */
export function logout() {
    return request({
        url: "/api/v1/user/logout",
        method: "post"
    });
}
