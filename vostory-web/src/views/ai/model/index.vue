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
                            :disabled="!hasPermission('ai:model:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleToggle(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup name="AiModel">
import { Modal, Message } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import {
    getModelList,
    addModel,
    updateModel,
    deleteModel,
    enableModel,
    disableModel,
    AIModelType
} from "@/config/apis/ai-model";
import { getProviderList } from "@/config/apis/ai-provider";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const table = ref();
const filterData = ref({});
const providerOptions = ref<{ label: string; value: number }[]>([]);

async function loadProviders() {
    const res = await getProviderList({ page: 1, size: 200 });
    providerOptions.value = (res.data || []).map((p) => ({ label: p.name, value: p.id }));
}

const getFilterConfig = computed(() => {
    return [
        formHelper.select("所属厂家", "provider_id", providerOptions.value, { span: 6 }),
        formHelper.input("模型标识", "model_name", { span: 6, debounce: 500 }),
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
            tableHelper.default("模型标识", "model_name"),
            tableHelper.default("所属厂家", "provider_name"),
            tableHelper.slot("switchSlot"),
            tableHelper.default("备注", "remark"),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("ai:model:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("ai:model:remove"),
                    handler(row: AIModelType) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除模型【${row.model_name}】？`,
                            onBeforeOk: async () => {
                                await deleteModel(row.id);
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
                label: "新增模型",
                if: () => hasPermission("ai:model:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getModelList,
        params: { ...filterData.value }
    };
});

function onEdit(v: AIModelType | null) {
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: { title: tempValue ? "编辑模型" : "新增模型", width: "600px" },
        value: tempValue || { status: "0", sort_order: 0 },
        formConfig: [
            formHelper.select("所属厂家", "provider_id", providerOptions.value, {
                rules: [ruleHelper.require("请选择")],
                allowSearch: true
            }),
            formHelper.input("模型标识", "model_name", {
                rules: [ruleHelper.require("请输入")],
                inputTips: "如 gpt-4o / deepseek-chat / glm-4.6"
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
                await updateModel(data);
            } else {
                await addModel(data);
            }
            Message.success("保存成功");
            table.value.refresh();
        }
    });
}

async function handleToggle(row: AIModelType) {
    try {
        if (row.status === "0") {
            await disableModel(row.id);
        } else {
            await enableModel(row.id);
        }
        table.value.refresh();
        return true;
    } catch {
        return false;
    }
}

onMounted(() => {
    loadProviders();
});
</script>
<style lang="scss" scoped></style>
