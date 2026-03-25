<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #statusSlot>
                <a-table-column title="状态">
                    <template #cell="{ record }">
                        <a-tag :color="statusColor(record.status)" size="small">
                            {{ statusLabel(record.status) }}
                        </a-tag>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { Modal } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import { getChapterList, addChapter, updateChapter, deleteChapter } from "@/config/apis/chapter";
import { getProjectsByWorkspace, ProjectOptionType } from "@/config/apis/project";
import { getWorkspaceOptions, WorkspaceOptionType } from "@/config/apis/workspace";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const STATUS_MAP: Record<string, { label: string; color: string }> = {
    raw: { label: "原始", color: "gray" },
    parsed: { label: "已解析", color: "blue" },
    edited: { label: "已编辑", color: "cyan" },
    generated: { label: "已生成", color: "green" },
    exported: { label: "已导出", color: "purple" }
};

function statusLabel(s: string) {
    return STATUS_MAP[s]?.label || s;
}
function statusColor(s: string) {
    return STATUS_MAP[s]?.color || "gray";
}

const table = ref();
const filterData = ref<Record<string, any>>({});
const projectOptions = ref<{ label: string; value: number }[]>([]);

onMounted(async () => {
    const wsRes = await getWorkspaceOptions();
    for (const ws of wsRes as WorkspaceOptionType[]) {
        const projects = await getProjectsByWorkspace(ws.id);
        for (const p of projects as ProjectOptionType[]) {
            projectOptions.value.push({ label: `${ws.name} / ${p.name}`, value: p.id });
        }
    }
});

const getFilterConfig = computed(() => {
    return [
        formHelper.select("项目", "project_id", projectOptions.value, { span: 8 }),
        formHelper.input("章节标题", "title", { span: 6, debounce: 500 }),
        formHelper.select(
            "状态",
            "status",
            Object.entries(STATUS_MAP).map(([k, v]) => ({ label: v.label, value: k })),
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
            tableHelper.default("序号", "chapter_num"),
            tableHelper.default("章节标题", "title"),
            tableHelper.default("字数", "word_count"),
            tableHelper.slot("statusSlot"),
            tableHelper.default("备注", "remark"),
            tableHelper.date("更新时间", "updated_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("chapter:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("chapter:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除【${row.title || "第" + row.chapter_num + "章"}】？`,
                            onBeforeOk: async () => {
                                await deleteChapter(row.id);
                                table.value.refresh();
                            }
                        });
                    }
                }
            ])
        ],
        trBtns: [
            {
                label: "添加章节",
                type: "primary",
                if: () => hasPermission("chapter:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getChapterList,
        params: { ...filterData.value }
    };
});

function onEdit(v: Record<string, any> | null) {
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑章节" : "添加章节",
            width: "700px"
        },
        value: tempValue || {},
        formConfig: [
            ...(tempValue
                ? []
                : [
                      formHelper.select("所属项目", "project_id", projectOptions.value, {
                          rules: [ruleHelper.require("请选择项目")]
                      })
                  ]),
            formHelper.input("章节标题", "title"),
            formHelper.inputNumber("章节序号", "chapter_num", {
                rules: [ruleHelper.require("请输入序号")]
            }),
            formHelper.textarea("章节原文", "content", {
                inputTips: "粘贴或输入章节文本内容"
            }),
            formHelper.textarea("备注", "remark")
        ],
        ok: async (data: any) => {
            if (tempValue) {
                await updateChapter(data);
            } else {
                await addChapter(data);
            }
            table.value.refresh();
        }
    });
}
</script>
<style lang="scss" scoped></style>
