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
                            :disabled="!hasPermission('ai:provider:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleToggle(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup name="AiProvider">
import { Modal, Message } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import {
    getProviderList,
    addProvider,
    updateProvider,
    deleteProvider,
    enableProvider,
    disableProvider,
    testProvider,
    AIProviderType
} from "@/config/apis/ai-provider";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const table = ref();
const filterData = ref({});

const getFilterConfig = computed(() => {
    return [
        formHelper.input("厂家名称", "name", { span: 6, debounce: 500 }),
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
            tableHelper.default("厂家名称", "name"),
            tableHelper.default("API 地址", "api_base_url"),
            tableHelper.default("API Key", "api_key"),
            tableHelper.default("最大并发", "max_concurrency"),
            tableHelper.slot("switchSlot"),
            tableHelper.default("备注", "remark"),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("ai:provider:edit"),
                    handler: onEdit
                },
                {
                    label: "测试",
                    if: () => hasPermission("ai:provider:test"),
                    handler: onTest
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("ai:provider:remove"),
                    handler(row: AIProviderType) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除厂家【${row.name}】？`,
                            onBeforeOk: async () => {
                                await deleteProvider(row.id);
                                Message.success("删除成功");
                                table.value.refresh();
                            }
                        });
                    }
                }
            ])
        ],
        trBtns: [
            {
                label: "新增厂家",
                if: () => hasPermission("ai:provider:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getProviderList,
        params: { ...filterData.value }
    };
});

function onEdit(v: AIProviderType | null) {
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: { title: tempValue ? "编辑厂家" : "新增厂家", width: "600px" },
        value: tempValue || { status: "0", sort_order: 0, max_concurrency: 1 },
        formConfig: [
            formHelper.input("厂家名称", "name", { rules: [ruleHelper.require("请输入")] }),
            formHelper.input("API 地址", "api_base_url", {
                rules: [ruleHelper.require("请输入")],
                inputTips: "例如：https://api.openai.com 或 https://api.deepseek.com/v1"
            }),
            formHelper.input("API Key", "api_key", {
                formType: "password",
                inputTips: "编辑时留空表示不修改"
            }),
            formHelper.inputNumber("最大并发数", "max_concurrency", {
                inputTips: "同时处理的最大任务数，默认为 1",
                min: 1,
                max: 100
            }),
            formHelper.inputNumber("排序", "sort_order"),
            formHelper.radio(
                "状态",
                "status",
                [
                    { label: "正常", value: "0" },
                    { label: "停用", value: "1" }
                ],
                { type: "radio", rules: [ruleHelper.require("请选择")] }
            ),
            formHelper.textarea("备注", "remark")
        ],
        ok: async (data: any) => {
            if (tempValue) {
                await updateProvider(data);
            } else {
                await addProvider(data);
            }
            Message.success("保存成功");
            table.value.refresh();
        }
    });
}

// 测试连接：再填一个模型标识，用 Chat Completions 探活
function onTest(row: AIProviderType) {
    ArcoModalFormShow({
        modalConfig: { title: `测试连接：${row.name}` },
        value: { model: "" },
        formConfig: [
            formHelper.input("测试模型", "model", {
                rules: [ruleHelper.require("请输入用于测试的模型标识，如 gpt-4o/deepseek-chat")],
                inputTips: "将用该模型发起一次极简对话以验证连通性"
            })
        ],
        ok: async (data: any) => {
            const result = await testProvider({
                id: row.id,
                api_base_url: row.api_base_url,
                api_key: row.api_key,
                model: data.model
            });
            if (result.success) {
                Message.success(`${result.message}（耗时 ${result.duration}ms）`);
            } else {
                Message.error(`连接失败：${result.message}`);
                throw new Error(result.message);
            }
        }
    });
}

async function handleToggle(row: AIProviderType) {
    try {
        if (row.status === "0") {
            await disableProvider(row.id);
        } else {
            await enableProvider(row.id);
        }
        table.value.refresh();
        return true;
    } catch {
        return false;
    }
}
</script>
<style lang="scss" scoped></style>
