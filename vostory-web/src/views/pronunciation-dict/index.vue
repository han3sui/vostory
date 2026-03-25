<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #scopeSlot>
                <a-table-column title="范围">
                    <template #cell="{ record }">
                        <a-tag v-if="record.project_id" color="blue" size="small">
                            项目级 · {{ record.project_name }}
                        </a-tag>
                        <a-tag v-else color="orange" size="small">全局</a-tag>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { Message } from "@arco-design/web-vue";
import {
    formHelper,
    ArcoTable,
    tableHelper,
    ArcoForm,
    ArcoModalFormShow,
    ruleHelper
} from "@easyfe/admin-component";
import {
    getPronunciationDictList,
    addPronunciationDict,
    updatePronunciationDict,
    deletePronunciationDict,
    PronunciationDictDetailType
} from "@/config/apis/pronunciation-dict";
import { getWorkspaceOptions, WorkspaceOptionType } from "@/config/apis/workspace";
import { getProjectsByWorkspace, ProjectOptionType } from "@/config/apis/project";
import { hasPermission, PageTableConfig } from "@/views/utils";
import { cloneDeep } from "lodash-es";

const table = ref();
const filterData = ref<Record<string, any>>({});
const workspaceOptions = ref<{ label: string; value: number }[]>([]);
const projectOptions = ref<{ label: string; value: number }[]>([]);

onMounted(async () => {
    const wsRes = await getWorkspaceOptions();
    for (const ws of wsRes as WorkspaceOptionType[]) {
        workspaceOptions.value.push({ label: ws.name, value: ws.id });
        const projects = await getProjectsByWorkspace(ws.id);
        for (const p of projects as ProjectOptionType[]) {
            projectOptions.value.push({ label: `${ws.name} / ${p.name}`, value: p.id });
        }
    }
});

const getFilterConfig = computed(() => {
    return [
        formHelper.select("工作空间", "workspace_id", workspaceOptions.value, { span: 6, allowClear: true }),
        formHelper.input("原始词", "word", { span: 6, debounce: 500 })
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: { rowKey: "id" },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        trBtns: [
            {
                label: "新增词条",
                if: () => hasPermission("pronunciation-dict:add"),
                handler: () => {
                    handleAdd();
                }
            }
        ],
        columns: [
            tableHelper.default("原始词", "word"),
            tableHelper.default("发音标注", "phoneme"),
            tableHelper.default("工作空间", "workspace_name"),
            tableHelper.slot("scopeSlot"),
            tableHelper.default("备注", "remark"),
            tableHelper.date("创建时间", "created_at", { format: "YYYY-MM-DD HH:mm" }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("pronunciation-dict:edit"),
                    handler(row: Record<string, any>) {
                        handleEdit(row as PronunciationDictDetailType);
                    }
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("pronunciation-dict:remove"),
                    handler(row: Record<string, any>) {
                        handleDelete(row.id);
                    }
                }
            ])
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getPronunciationDictList,
        params: { ...filterData.value }
    };
});

function getFormConfig(isEdit: boolean) {
    return [
        formHelper.select("工作空间", "workspace_id", workspaceOptions.value, {
            rules: [ruleHelper.require("请选择工作空间")],
            disabled: isEdit
        }),
        formHelper.select("所属项目", "project_id", [
            { label: "全局（不绑定项目）", value: 0 },
            ...projectOptions.value
        ], {
            disabled: isEdit
        }),
        formHelper.input("原始词", "word", { rules: [ruleHelper.require("请输入原始词")] }),
        formHelper.input("发音标注", "phoneme", { rules: [ruleHelper.require("请输入发音标注")] }),
        formHelper.input("备注", "remark")
    ];
}

function handleAdd() {
    ArcoModalFormShow({
        modalConfig: { title: "新增词条" },
        value: {},
        formConfig: getFormConfig(false),
        ok: async (data: any) => {
            if (data.project_id === 0) {
                data.project_id = null;
            }
            await addPronunciationDict(data);
            Message.success("新增成功");
            table.value.refresh();
        }
    });
}

function handleEdit(row: PronunciationDictDetailType) {
    ArcoModalFormShow({
        modalConfig: { title: "编辑词条" },
        value: cloneDeep({ ...row, project_id: row.project_id || 0 }),
        formConfig: getFormConfig(true),
        ok: async (data: any) => {
            await updatePronunciationDict(data);
            Message.success("更新成功");
            table.value.refresh();
        }
    });
}

function handleDelete(id: number) {
    deletePronunciationDict(id).then(() => {
        Message.success("删除成功");
        table.value.refresh();
    });
}
</script>
<style lang="scss" scoped></style>
