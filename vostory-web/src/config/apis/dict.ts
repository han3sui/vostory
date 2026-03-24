import request from "@/packages/request";

// ============= 字典类型 =============

export type DictTypeDetailType = {
    id: number;
    dict_name: string;
    dict_type: string;
    status: string;
    remark: string;
    created_at: string;
    updated_at: string;
};

export type DictTypeCreateParams = {
    dict_name: string;
    dict_type: string;
    status: string;
    remark?: string;
};

export type DictTypeUpdateParams = DictTypeCreateParams & {
    id: number;
};

export function getDictTypeList(params?: Record<string, any>): Promise<{
    data: DictTypeDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({
        url: "/api/v1/system/dict/type/list",
        params
    });
}

export function getDictType(id: number): Promise<DictTypeDetailType> {
    return request({
        url: `/api/v1/system/dict/type/${id}`
    });
}

export function addDictType(data: DictTypeCreateParams) {
    return request({
        url: "/api/v1/system/dict/type",
        method: "post",
        data
    });
}

export function updateDictType(data: DictTypeUpdateParams) {
    return request({
        url: `/api/v1/system/dict/type/${data.id}`,
        method: "put",
        data
    });
}

export function deleteDictType(id: number) {
    return request({
        url: `/api/v1/system/dict/type/${id}`,
        method: "delete"
    });
}

export function enableDictType(id: number) {
    return request({
        url: `/api/v1/system/dict/type/${id}/enable`,
        method: "put"
    });
}

export function disableDictType(id: number) {
    return request({
        url: `/api/v1/system/dict/type/${id}/disable`,
        method: "put"
    });
}

// ============= 字典数据 =============

export type DictDataDetailType = {
    id: number;
    dict_sort: number;
    dict_label: string;
    dict_value: string;
    dict_type: string;
    css_class: string;
    list_class: string;
    is_default: string;
    status: string;
    remark: string;
    created_at: string;
    updated_at: string;
};

export type DictDataCreateParams = {
    dict_sort: number;
    dict_label: string;
    dict_value: string;
    dict_type: string;
    css_class?: string;
    list_class?: string;
    is_default?: string;
    status: string;
    remark?: string;
};

export type DictDataUpdateParams = DictDataCreateParams & {
    id: number;
};

export function getDictDataList(params?: Record<string, any>): Promise<{
    data: DictDataDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({
        url: "/api/v1/system/dict/data/list",
        params
    });
}

export function getDictData(id: number): Promise<DictDataDetailType> {
    return request({
        url: `/api/v1/system/dict/data/${id}`
    });
}

export function getDictDataByType(dictType: string): Promise<DictDataDetailType[]> {
    return request({
        url: `/api/v1/common/dict/data/type/${dictType}`
    });
}

export function addDictData(data: DictDataCreateParams) {
    return request({
        url: "/api/v1/system/dict/data",
        method: "post",
        data
    });
}

export function updateDictData(data: DictDataUpdateParams) {
    return request({
        url: `/api/v1/system/dict/data/${data.id}`,
        method: "put",
        data
    });
}

export function deleteDictData(id: number) {
    return request({
        url: `/api/v1/system/dict/data/${id}`,
        method: "delete"
    });
}

export function enableDictData(id: number) {
    return request({
        url: `/api/v1/system/dict/data/${id}/enable`,
        method: "put"
    });
}

export function disableDictData(id: number) {
    return request({
        url: `/api/v1/system/dict/data/${id}/disable`,
        method: "put"
    });
}
