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
                            :disabled="!hasPermission('ai:prompt-template:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleToggle(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
            <template #systemSlot>
                <a-table-column title="内置">
                    <template #cell="{ record }">
                        <a-tag v-if="record.is_system === '1'" color="orangered" size="small">系统</a-tag>
                        <a-tag v-else color="cyan" size="small">自定义</a-tag>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { Modal } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import {
    getPromptTemplateList,
    addPromptTemplate,
    updatePromptTemplate,
    deletePromptTemplate,
    enablePromptTemplate,
    disablePromptTemplate,
    PromptTemplateDetailType
} from "@/config/apis/ai";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const TEMPLATE_TYPES = [
    { label: "角色抽取", value: "character_extract" },
    { label: "对白解析", value: "dialogue_parse" },
    { label: "情绪标注", value: "emotion_tag" },
    { label: "场景切分", value: "scene_split" },
    { label: "文本校正", value: "text_correct" }
];

const table = ref();
const filterData = ref({});

const getFilterConfig = computed(() => {
    return [
        formHelper.input("模板名称", "name", { span: 6, debounce: 500 }),
        formHelper.select("模板类型", "template_type", TEMPLATE_TYPES, { span: 6 }),
        formHelper.select(
            "来源",
            "is_system",
            [
                { label: "系统内置", value: "1" },
                { label: "自定义", value: "0" }
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
            tableHelper.default("模板名称", "name"),
            tableHelper.status("模板类型", "template_type", (item: any) => {
                const found = TEMPLATE_TYPES.find((t) => t.value === item.template_type);
                return { text: found?.label || item.template_type, status: "normal" };
            }),
            tableHelper.slot("systemSlot"),
            tableHelper.default("版本", "version"),
            tableHelper.default("描述", "description"),
            tableHelper.default("排序", "sort_order"),
            tableHelper.slot("switchSlot"),
            tableHelper.date("更新时间", "updated_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("ai:prompt-template:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: (row: any) => hasPermission("ai:prompt-template:remove") && row?.is_system !== "1",
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除【${row.name}】？`,
                            onBeforeOk: async () => {
                                await deletePromptTemplate(row.id);
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
                if: () => hasPermission("ai:prompt-template:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getPromptTemplateList,
        params: { ...filterData.value }
    };
});

function onEdit(v: Record<string, any> | null) {
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑 Prompt 模板" : "添加 Prompt 模板",
            width: "750px"
        },
        value: tempValue || { status: "0" },
        formConfig: [
            formHelper.input("模板名称", "name", { rules: [ruleHelper.require("请输入名称")] }),
            formHelper.select("模板类型", "template_type", TEMPLATE_TYPES, {
                rules: [ruleHelper.require("请选择类型")]
            }),
            formHelper.textarea("Prompt 内容", "content", {
                rules: [ruleHelper.require("请输入Prompt内容")],
                inputTips: "支持变量占位符：{{content}}、{{segments}} 等"
            }),
            formHelper.textarea("描述", "description"),
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
                await updatePromptTemplate(data);
            } else {
                await addPromptTemplate(data);
            }
            table.value.refresh();
        }
    });
}

async function handleToggle(row: PromptTemplateDetailType) {
    try {
        if (row.status === "0") {
            await disablePromptTemplate(row.id);
        } else {
            await enablePromptTemplate(row.id);
        }
        table.value.refresh();
        return true;
    } catch {
        return false;
    }
}
</script>
<style lang="scss" scoped></style>
