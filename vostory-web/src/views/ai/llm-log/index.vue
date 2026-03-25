<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #statusSlot>
                <a-table-column title="状态">
                    <template #cell="{ record }">
                        <a-tag v-if="record.status === 0" color="green" size="small">成功</a-tag>
                        <a-tooltip v-else :content="record.error_message">
                            <a-tag color="red" size="small">失败</a-tag>
                        </a-tooltip>
                    </template>
                </a-table-column>
            </template>
            <template #tokensSlot>
                <a-table-column title="Token">
                    <template #cell="{ record }">
                        <span class="text-xs">
                            入 {{ record.input_tokens }} / 出 {{ record.output_tokens }}
                        </span>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { Modal } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoForm } from "@easyfe/admin-component";
import { getLLMLogList, deleteLLMLog } from "@/config/apis/llm-log";
import { hasPermission, PageTableConfig } from "@/views/utils";

const table = ref();
const filterData = ref<Record<string, any>>({});

const getFilterConfig = computed(() => {
    return [
        formHelper.input("模型名称", "model_name", { span: 6, debounce: 500 }),
        formHelper.select(
            "状态",
            "status",
            [
                { label: "全部", value: -1 },
                { label: "成功", value: 0 },
                { label: "失败", value: 1 }
            ],
            { span: 5 }
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
            tableHelper.default("项目", "project_name"),
            tableHelper.default("提供商", "provider_name"),
            tableHelper.default("模型", "model_name"),
            tableHelper.default("模板", "template_name"),
            tableHelper.slot("tokensSlot"),
            tableHelper.default("耗时(ms)", "cost_time"),
            tableHelper.slot("statusSlot"),
            tableHelper.date("时间", "created_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns("操作", [
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("ai:llm-log:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: "确认删除该日志？",
                            onBeforeOk: async () => {
                                await deleteLLMLog(row.id);
                                table.value.refresh();
                            }
                        });
                    }
                }
            ])
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getLLMLogList,
        params: { ...filterData.value }
    };
});
</script>
<style lang="scss" scoped></style>
