import request from "@/packages/request";

export type OSSConfigType = {
    type: string;
    endpoint: string;
    ak: string;
    sk: string;
    bucket: string;
    prefix: string;
    use_ssl: string;
    cdn_base?: string;
};

export function getOSSConfig(): Promise<OSSConfigType> {
    return request({
        url: "/api/v1/system/oss-config"
    });
}

export function updateOSSConfig(data: OSSConfigType) {
    return request({
        url: "/api/v1/system/oss-config",
        method: "put",
        data
    });
}

export function testOSSConnection(data: OSSConfigType) {
    return request({
        url: "/api/v1/system/oss-config/test",
        method: "post",
        data
    });
}
