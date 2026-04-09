<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #switchSlot>
                <a-table-column title="启用">
                    <template #cell="{ record }">
                        <a-switch
                            :disabled="!hasPermission('ai:tts-provider:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleToggle(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
            <template #statusSlot>
                <a-table-column title="系统负载" :width="340">
                    <template #cell="{ record }">
                        <div v-if="statusMap[record.id]?.loading" class="status-placeholder">
                            <a-spin :size="14" />
                        </div>
                        <div v-else-if="statusMap[record.id]?.error" class="status-placeholder">
                            <a-tooltip :content="statusMap[record.id].error ?? ''">
                                <a-link status="danger" :hoverable="false" @click="fetchStatus(record)">
                                    获取失败，点击重试
                                </a-link>
                            </a-tooltip>
                        </div>
                        <div v-else-if="statusMap[record.id]?.data" class="status-circles">
                            <a-tooltip position="top" :content-style="{ whiteSpace: 'pre-line' }">
                                <template #content>
                                    <div>{{ statusMap[record.id].data!.cpu.name }}</div>
                                    <div>
                                        {{ statusMap[record.id].data!.cpu.count_physical ?? "?" }} 核 /
                                        {{ statusMap[record.id].data!.cpu.count_logical }} 线程
                                    </div>
                                    <div>使用率 {{ statusMap[record.id].data!.cpu.percent }}%</div>
                                </template>
                                <div class="circle-item">
                                    <a-progress
                                        type="circle"
                                        :percent="statusMap[record.id].data!.cpu.percent / 100"
                                        :width="52"
                                        :stroke-width="4"
                                        :color="getProgressColor(statusMap[record.id].data!.cpu.percent)"
                                        :track-color="'var(--color-fill-2)'"
                                    >
                                        <template #text="{ percent }">
                                            <span class="circle-text">{{ Math.round(percent * 100) }}%</span>
                                        </template>
                                    </a-progress>
                                    <span class="circle-label">CPU</span>
                                </div>
                            </a-tooltip>
                            <a-tooltip position="top" :content-style="{ whiteSpace: 'pre-line' }">
                                <template #content>
                                    <div>
                                        内存 {{ formatSize(statusMap[record.id].data!.memory.used_mb) }} /
                                        {{ formatSize(statusMap[record.id].data!.memory.total_mb) }}
                                    </div>
                                    <div>使用率 {{ statusMap[record.id].data!.memory.percent }}%</div>
                                </template>
                                <div class="circle-item">
                                    <a-progress
                                        type="circle"
                                        :percent="statusMap[record.id].data!.memory.percent / 100"
                                        :width="52"
                                        :stroke-width="4"
                                        :color="getProgressColor(statusMap[record.id].data!.memory.percent)"
                                        :track-color="'var(--color-fill-2)'"
                                    >
                                        <template #text="{ percent }">
                                            <span class="circle-text">{{ Math.round(percent * 100) }}%</span>
                                        </template>
                                    </a-progress>
                                    <span class="circle-label">内存</span>
                                </div>
                            </a-tooltip>
                            <template v-if="statusMap[record.id].data!.gpu">
                                <a-tooltip position="top" :content-style="{ whiteSpace: 'pre-line' }">
                                    <template #content>
                                        <div>{{ statusMap[record.id].data!.gpu!.name }}</div>
                                        <div>
                                            GPU 使用率 {{ statusMap[record.id].data!.gpu!.gpu_utilization ?? "N/A" }}%
                                        </div>
                                    </template>
                                    <div class="circle-item">
                                        <a-progress
                                            type="circle"
                                            :percent="(statusMap[record.id].data!.gpu!.gpu_utilization ?? 0) / 100"
                                            :width="52"
                                            :stroke-width="4"
                                            :color="getProgressColor(statusMap[record.id].data!.gpu!.gpu_utilization ?? 0)"
                                            :track-color="'var(--color-fill-2)'"
                                        >
                                            <template #text="{ percent }">
                                                <span class="circle-text">{{ Math.round(percent * 100) }}%</span>
                                            </template>
                                        </a-progress>
                                        <span class="circle-label">GPU</span>
                                    </div>
                                </a-tooltip>
                                <a-tooltip position="top" :content-style="{ whiteSpace: 'pre-line' }">
                                    <template #content>
                                        <div>{{ statusMap[record.id].data!.gpu!.name }}</div>
                                        <div>
                                            显存 {{ formatSize(statusMap[record.id].data!.gpu!.memory_allocated_mb) }} /
                                            {{ formatSize(statusMap[record.id].data!.gpu!.memory_total_mb) }}
                                        </div>
                                    </template>
                                    <div class="circle-item">
                                        <a-progress
                                            type="circle"
                                            :percent="gpuMemPercent(statusMap[record.id].data!.gpu!)"
                                            :width="52"
                                            :stroke-width="4"
                                            :color="getProgressColor(gpuMemPercent(statusMap[record.id].data!.gpu!) * 100)"
                                            :track-color="'var(--color-fill-2)'"
                                        >
                                            <template #text="{ percent }">
                                                <span class="circle-text">{{ Math.round(percent * 100) }}%</span>
                                            </template>
                                        </a-progress>
                                        <span class="circle-label">显存</span>
                                    </div>
                                </a-tooltip>
                            </template>
                            <div class="circle-item circle-refresh">
                                <a-link :hoverable="false" @click="fetchStatus(record)">
                                    <icon-refresh />
                                </a-link>
                            </div>
                        </div>
                        <div v-else class="status-placeholder">
                            <a-link :hoverable="false" @click="fetchStatus(record)">查看状态</a-link>
                        </div>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { Modal, Message } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import {
    getTTSProviderList,
    addTTSProvider,
    updateTTSProvider,
    deleteTTSProvider,
    enableTTSProvider,
    disableTTSProvider,
    testTTSProvider,
    getTTSProviderStatus,
    TTSProviderDetailType,
    TTSProviderStatusResult,
    TTSProviderStatusGPU
} from "@/config/apis/ai";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const table = ref();
const filterData = ref({});

type StatusEntry = {
    loading: boolean;
    error: string | null;
    data: TTSProviderStatusResult | null;
};
const statusMap = ref<Record<number, StatusEntry>>({});

async function fetchStatus(row: TTSProviderDetailType) {
    statusMap.value[row.id] = { loading: true, error: null, data: null };
    try {
        const data = await getTTSProviderStatus(row.id);
        statusMap.value[row.id] = { loading: false, error: null, data };
    } catch (e: any) {
        statusMap.value[row.id] = { loading: false, error: e?.message || "获取失败", data: null };
    }
}

function getProgressColor(percent: number): string {
    if (percent >= 90) return "var(--color-danger-6)";
    if (percent >= 70) return "var(--color-warning-6)";
    return "rgb(var(--arcoblue-6))";
}

function formatSize(mb: number): string {
    if (mb >= 1024) return (mb / 1024).toFixed(1) + " GB";
    return mb + " MB";
}

function gpuMemPercent(gpu: TTSProviderStatusGPU): number {
    if (!gpu.memory_total_mb) return 0;
    return gpu.memory_allocated_mb / gpu.memory_total_mb;
}

const getFilterConfig = computed(() => {
    return [
        formHelper.input("名称", "name", { span: 6, debounce: 500 }),
        formHelper.select(
            "状态",
            "status",
            [
                { label: "正常", value: "0" },
                { label: "停用", value: "1" }
            ],
            { span: 6 }
        )
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: { rowKey: "id" },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("名称", "name"),
            tableHelper.default("API 地址", "api_base_url"),
            tableHelper.slot("switchSlot"),
            tableHelper.slot("statusSlot"),
            tableHelper.date("创建时间", "created_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns("操作", [
                {
                    label: "测试",
                    type: "outline",
                    if: () => hasPermission("ai:tts-provider:test"),
                    handler: handleTest
                },
                {
                    label: "编辑",
                    if: () => hasPermission("ai:tts-provider:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("ai:tts-provider:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除【${row.name}】？`,
                            onBeforeOk: async () => {
                                await deleteTTSProvider(row.id);
                                table.value.refresh();
                            }
                        });
                    }
                }
            ])
        ],
        trBtns: [
            {
                label: "添加",
                type: "primary",
                if: () => hasPermission("ai:tts-provider:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getTTSProviderList,
        params: { ...filterData.value }
    };
});

function onEdit(v: Record<string, any> | null) {
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑 TTS 提供商" : "添加 TTS 提供商",
            width: "650px"
        },
        value: tempValue || {
            status: "0",
            provider_type: "local",
            supported_features: ["clone", "multi_speaker"],
            custom_params: {},
            max_concurrency: 1
        },
        formConfig: [
            formHelper.input("名称", "name", { rules: [ruleHelper.require("请输入名称")] }),
            formHelper.input("API 地址", "api_base_url", {
                rules: [ruleHelper.require("请输入API地址")],
                inputTips: "例如：http://localhost:8080"
            }),
            formHelper.input("API 密钥", "api_key", { formType: "password" }),
            formHelper.radio(
                "状态",
                "status",
                [
                    { label: "正常", value: "0" },
                    { label: "停用", value: "1" }
                ],
                { type: "radio", rules: [ruleHelper.require("请选择")] }
            ),
            formHelper.inputNumber("排序", "sort_order")
        ],
        ok: async (data: any) => {
            if (tempValue) {
                await updateTTSProvider(data);
            } else {
                await addTTSProvider(data);
            }
            table.value.refresh();
        }
    });
}

async function handleToggle(row: TTSProviderDetailType) {
    try {
        if (row.status === "0") {
            await disableTTSProvider(row.id);
        } else {
            await enableTTSProvider(row.id);
        }
        table.value.refresh();
        return true;
    } catch {
        return false;
    }
}

async function handleTest(row: TTSProviderDetailType) {
    const loading = Message.loading({ content: "正在测试连通性...", duration: 0 });
    try {
        const result = await testTTSProvider({
            provider_type: row.provider_type,
            api_base_url: row.api_base_url,
            api_key: row.api_key
        });
        loading.close();
        if (result.success) {
            Message.success(`连接成功（耗时 ${result.duration}ms）`);
        } else {
            Message.error(`连接失败：${result.message}`);
        }
    } catch {
        loading.close();
        Message.error("测试请求失败");
    }
}
</script>
<style lang="scss" scoped>
.status-placeholder {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--color-text-3);
}

.status-circles {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 4px 0;
}

.circle-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    cursor: pointer;
}

.circle-text {
    font-size: 12px;
    font-weight: 500;
    font-variant-numeric: tabular-nums;
    color: var(--color-text-1);
}

.circle-label {
    font-size: 11px;
    color: var(--color-text-3);
    line-height: 1;
}

.circle-refresh {
    align-self: center;
    cursor: pointer;
    font-size: 16px;
    color: var(--color-text-3);

    &:hover {
        color: rgb(var(--arcoblue-6));
    }
}
</style>
