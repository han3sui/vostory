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
                            :disabled="!hasPermission('ai:llm-provider:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleToggle(record)"
                        ></a-switch>
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
    getLLMProviderList,
    addLLMProvider,
    updateLLMProvider,
    deleteLLMProvider,
    enableLLMProvider,
    disableLLMProvider,
    testLLMProvider,
    LLMProviderDetailType
} from "@/config/apis/ai";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const LLM_TYPES = [
    { label: "OpenAI", value: "openai" },
    { label: "DeepSeek", value: "deepseek" },
    { label: "Anthropic", value: "anthropic" },
    { label: "Gemini", value: "gemini" },
    { label: "Ollama", value: "ollama" },
    { label: "Azure", value: "azure" },
    { label: "阿里云", value: "aliyun" },
    { label: "自定义", value: "custom" }
];

const table = ref();
const filterData = ref({});

const getFilterConfig = computed(() => {
    return [
        formHelper.input("名称", "name", { span: 6, debounce: 500 }),
        formHelper.select("提供商类型", "provider_type", LLM_TYPES, { span: 6 }),
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
            tableHelper.status("类型", "provider_type", (item: any) => {
                const found = LLM_TYPES.find((t) => t.value === item.provider_type);
                return { text: found?.label || item.provider_type, status: "normal" };
            }),
            tableHelper.default("API 地址", "api_base_url"),
            tableHelper.default("默认模型", "default_model"),
            tableHelper.default("排序", "sort_order"),
            tableHelper.slot("switchSlot"),
            tableHelper.date("创建时间", "created_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns("操作", [
                {
                    label: "测试",
                    type: "outline",
                    if: () => hasPermission("ai:llm-provider:test"),
                    handler: handleTest
                },
                {
                    label: "编辑",
                    if: () => hasPermission("ai:llm-provider:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("ai:llm-provider:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除【${row.name}】？`,
                            onBeforeOk: async () => {
                                await deleteLLMProvider(row.id);
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
                if: () => hasPermission("ai:llm-provider:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getLLMProviderList,
        params: { ...filterData.value }
    };
});

function onEdit(v: Record<string, any> | null) {
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑 LLM 提供商" : "添加 LLM 提供商",
            width: "650px"
        },
        value: tempValue || { status: "0", model_list: [], custom_params: {} },
        formConfig: [
            formHelper.input("名称", "name", { rules: [ruleHelper.require("请输入名称")] }),
            formHelper.select("提供商类型", "provider_type", LLM_TYPES, {
                rules: [ruleHelper.require("请选择类型")]
            }),
            formHelper.input("API 地址", "api_base_url", {
                rules: [ruleHelper.require("请输入API地址")],
                inputTips: "例如：https://api.openai.com"
            }),
            formHelper.input("API 密钥", "api_key", { formType: "password" }),
            formHelper.input("默认模型", "default_model", {
                inputTips: "例如：gpt-4o、deepseek-chat"
            }),
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
                await updateLLMProvider(data);
            } else {
                await addLLMProvider(data);
            }
            table.value.refresh();
        }
    });
}

async function handleToggle(row: LLMProviderDetailType) {
    try {
        if (row.status === "0") {
            await disableLLMProvider(row.id);
        } else {
            await enableLLMProvider(row.id);
        }
        table.value.refresh();
        return true;
    } catch {
        return false;
    }
}

async function handleTest(row: LLMProviderDetailType) {
    const loading = Message.loading("正在测试连通性...");
    try {
        const result = await testLLMProvider({
            provider_type: row.provider_type,
            api_base_url: row.api_base_url,
            api_key: row.api_key,
            model: row.default_model
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
<style lang="scss" scoped></style>
