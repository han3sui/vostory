import request from "@/packages/request";

export type OperLogDetailType = {
    id: number;
    title: string;
    business_type: number;
    method: string;
    request_method: string;
    operator_type: number;
    oper_name: string;
    dept_name: string;
    oper_url: string;
    oper_ip: string;
    oper_location: string;
    oper_param: string;
    json_result: string;
    status: number;
    error_msg: string;
    oper_time: string;
    cost_time: number;
};

export function getOperLogList(params?: Record<string, any>): Promise<{
    data: OperLogDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({
        url: "/api/v1/system/operlog/list",
        params
    });
}

export function getOperLog(id: number): Promise<OperLogDetailType> {
    return request({
        url: `/api/v1/system/operlog/${id}`
    });
}

export function deleteOperLog(id: number) {
    return request({
        url: `/api/v1/system/operlog/${id}`,
        method: "delete"
    });
}

export function cleanOperLog() {
    return request({
        url: "/api/v1/system/operlog/clean",
        method: "delete"
    });
}
