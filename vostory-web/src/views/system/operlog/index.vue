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
import { formHelper, ArcoTable, tableHelper, ArcoForm } from "@easyfe/admin-component";
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
    ArcoModalShow({
        config: { title: "操作日志详情", hideCancel: true, width: "700px" },
        content: () =>
            h("div", { class: "p-4" }, [
                h("a-descriptions", { bordered: true, column: 2, size: "small" }, [
                    h("a-descriptions-item", { label: "模块标题" }, row.title),
                    h("a-descriptions-item", { label: "请求方式" }, row.request_method),
                    h("a-descriptions-item", { label: "操作人员" }, row.oper_name),
                    h("a-descriptions-item", { label: "部门名称" }, row.dept_name),
                    h("a-descriptions-item", { label: "请求URL", span: 2 }, row.oper_url),
                    h("a-descriptions-item", { label: "操作IP" }, row.oper_ip),
                    h("a-descriptions-item", { label: "操作状态" }, row.status === 0 ? "成功" : "异常"),
                    h("a-descriptions-item", { label: "耗时" }, `${row.cost_time}ms`),
                    h("a-descriptions-item", { label: "操作时间" }, row.oper_time),
                    row.oper_param
                        ? h("a-descriptions-item", { label: "请求参数", span: 2 }, row.oper_param)
                        : null,
                    row.json_result
                        ? h("a-descriptions-item", { label: "返回结果", span: 2 }, row.json_result)
                        : null,
                    row.error_msg
                        ? h("a-descriptions-item", { label: "错误信息", span: 2 }, row.error_msg)
                        : null
                ])
            ])
    });
}
</script>
<style lang="scss" scoped></style>
