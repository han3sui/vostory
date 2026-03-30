import { nanoid } from "nanoid";
import WebRequest from "@easyfe/web-request";
import loading from "./loading";
import envHelper from "@/utils/helper/env";
import storage from "@/utils/tools/storage";
import { errorLogout } from "@/views/utils";

const service = new WebRequest({
    base: {
        timeout: 60 * 1000,
        baseURL: envHelper.get("VITE_APP_API_URL"),
        prefix: "",
        enableCancel: false
    },
    loading,
    interceptors: {
        request: (config): Promise<any> => {
            if (config.method?.toLocaleUpperCase() === "GET") {
                config.params = {
                    ...config.params,
                    _t: nanoid()
                };
            }
            if (storage.getToken()) {
                config.headers = {
                    ...config.headers,
                    Authorization: `Bearer ${storage.getToken()}`
                };
            }
            if (config.url?.includes("/version.json")) {
                return Promise.resolve(config);
            }
            return Promise.resolve(config);
        },
        response: (response): Promise<any> => {
            return Promise.resolve(response.data);
        },
        responseError: (errorResponse): Promise<any> => {
            if (errorResponse.status === 401) {
                errorResponse.config.notify = false;
                errorLogout();
            }
            if (errorResponse.status === 403 && errorResponse.data?.code === 4031) {
                errorResponse.config.notify = false;
                window.location.href = "/activation";
                return Promise.reject(errorResponse);
            }
            return Promise.reject(errorResponse);
        }
    }
});

const request = service.request;

export function clearRequest(): void {
    service.clearQueue();
}

export default request;
