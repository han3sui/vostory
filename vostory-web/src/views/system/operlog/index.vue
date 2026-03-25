<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #statusSlot>
                <a-table-column title="操作状态">
                    <template #cell="{ record }">
                        <a-tag v-if="record.status === 0" color="green">成功</a-tag>
                        <a-tag v-else color="red">异常</a-tag>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { Modal } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper } from "@easyfe/admin-component";
import { getOperLogList, deleteOperLog, cleanOperLog, OperLogDetailType } from "@/config/apis/operlog";
import { hasPermission, PageTableConfig } from "@/views/utils";
import { ArcoModalShow } from "@easyfe/admin-component";

const table = ref();
const filterData = ref({});

const businessTypeOptions = [
    { label: "其他", value: 0 },
    { label: "新增", value: 1 },
    { label: "修改", value: 2 },
    { label: "删除", value: 3 }
];

const getFilterConfig = computed(() => {
    return [
        formHelper.input("模块标题", "title", { span: 6, debounce: 500 }),
        formHelper.input("操作人员", "oper_name", { span: 6, debounce: 500 }),
        formHelper.select("业务类型", "business_type", businessTypeOptions, { span: 6 }),
        formHelper.select(
            "操作状态",
            "status",
            [
                { label: "成功", value: 0 },
                { label: "异常", value: 1 }
            ],
            { span: 6 }
        )
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: {
            rowKey: "id"
        },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("模块标题", "title"),
            tableHelper.custom("业务类型", (row) => {
                const item = businessTypeOptions.find((o) => o.value === row.business_type);
                return item ? item.label : "其他";
            }),
            tableHelper.default("请求方式", "request_method"),
            tableHelper.default("操作人员", "oper_name"),
            tableHelper.default("操作IP", "oper_ip"),
            tableHelper.slot("statusSlot"),
            tableHelper.date("操作时间", "oper_time", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.default("耗时(ms)", "cost_time"),
            tableHelper.btns("操作", [
                {
                    label: "详情",
                    handler: (row: OperLogDetailType) => onDetail(row)
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("system:operlog:remove"),
                    handler: (row: Record<string, any>) => {
                        Modal.confirm({
                            title: "删除",
                            content: "确认删除该操作日志？",
                            onBeforeOk: async () => {
                                await deleteOperLog(row.id);
                                table.value.refresh();
                            }
                        });
                    }
                }
            ])
        ],
        trBtns: [
            {
                label: "清空",
                status: "danger",
                if: () => hasPermission("system:operlog:clean"),
                handler: () => {
                    Modal.confirm({
                        title: "清空",
                        content: "确认清空所有操作日志？此操作不可恢复！",
                        onBeforeOk: async () => {
                            await cleanOperLog();
                            table.value.refresh();
                        }
                    });
                }
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getOperLogList,
        params: { ...filterData.value }
    };
});

function onDetail(row: OperLogDetailType) {
    const items: Array<{ label: string; value: string; span?: number }> = [
        { label: "模块标题", value: row.title },
        { label: "请求方式", value: row.request_method },
        { label: "操作人员", value: row.oper_name },
        { label: "部门名称", value: row.dept_name },
        { label: "请求URL", value: row.oper_url, span: 2 },
        { label: "操作IP", value: row.oper_ip },
        { label: "操作状态", value: row.status === 0 ? "成功" : "异常" },
        { label: "耗时", value: `${row.cost_time}ms` },
        { label: "操作时间", value: row.oper_time }
    ];
    if (row.oper_param) items.push({ label: "请求参数", value: row.oper_param, span: 2 });
    if (row.json_result) items.push({ label: "返回结果", value: row.json_result, span: 2 });
    if (row.error_msg) items.push({ label: "错误信息", value: row.error_msg, span: 2 });

    ArcoModalShow({
        config: { title: "操作日志详情", hideCancel: true, width: "700px" },
        content: () =>
            h(
                "div",
                { style: "padding: 16px" },
                items.map((item) =>
                    h(
                        "div",
                        { style: "display: flex; padding: 8px 0; border-bottom: 1px solid var(--color-border-2)" },
                        [
                            h(
                                "span",
                                { style: "width: 100px; flex-shrink: 0; color: var(--color-text-3)" },
                                item.label
                            ),
                            h("span", { style: "flex: 1; word-break: break-all" }, item.value || "-")
                        ]
                    )
                )
            )
    });
}
</script>
<style lang="scss" scoped></style>
