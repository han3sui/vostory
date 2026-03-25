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
import {
    getProjectList,
    addProject,
    updateProject,
    deleteProject,
    ProjectDetailType
} from "@/config/apis/project";
import { getWorkspaceOptions, WorkspaceOptionType } from "@/config/apis/workspace";
import { getLLMProviderList, LLMProviderDetailType } from "@/config/apis/ai";
import { getTTSProviderList, TTSProviderDetailType } from "@/config/apis/ai";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const STATUS_MAP: Record<string, { label: string; color: string }> = {
    draft: { label: "草稿", color: "gray" },
    parsing: { label: "解析中", color: "orangered" },
    parsed: { label: "已解析", color: "blue" },
    generating: { label: "生成中", color: "orange" },
    completed: { label: "已完成", color: "green" }
};

function statusLabel(status: string) {
    return STATUS_MAP[status]?.label || status;
}

function statusColor(status: string) {
    return STATUS_MAP[status]?.color || "gray";
}

const table = ref();
const filterData = ref({});
const workspaceOptions = ref<{ label: string; value: number }[]>([]);
const llmOptions = ref<{ label: string; value: number }[]>([]);
const ttsOptions = ref<{ label: string; value: number }[]>([]);

onMounted(async () => {
    const [wsRes, llmRes, ttsRes] = await Promise.all([
        getWorkspaceOptions(),
        getLLMProviderList({ status: "0", page: 1, size: 100 }),
        getTTSProviderList({ status: "0", page: 1, size: 100 })
    ]);
    workspaceOptions.value = (wsRes as WorkspaceOptionType[]).map((w) => ({
        label: w.name,
        value: w.id
    }));
    llmOptions.value = ((llmRes as any).data as LLMProviderDetailType[]).map((p) => ({
        label: p.name,
        value: p.id
    }));
    ttsOptions.value = ((ttsRes as any).data as TTSProviderDetailType[]).map((p) => ({
        label: p.name,
        value: p.id
    }));
});

const getFilterConfig = computed(() => {
    return [
        formHelper.select("工作空间", "workspace_id", workspaceOptions.value, { span: 6 }),
        formHelper.input("项目名称", "name", { span: 6, debounce: 500 }),
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
            tableHelper.default("项目名称", "name"),
            tableHelper.default("工作空间", "workspace_name"),
            tableHelper.slot("statusSlot"),
            tableHelper.default("LLM 提供商", "llm_provider_name"),
            tableHelper.default("TTS 提供商", "tts_provider_name"),
            tableHelper.default("章节数", "total_chapters"),
            tableHelper.default("角色数", "total_characters"),
            tableHelper.date("创建时间", "created_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("project:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("project:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除项目【${row.name}】？删除后相关章节、脚本等数据将一并删除。`,
                            onBeforeOk: async () => {
                                await deleteProject(row.id);
                                table.value.refresh();
                            }
                        });
                    }
                }
            ])
        ],
        trBtns: [
            {
                label: "新建项目",
                type: "primary",
                if: () => hasPermission("project:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getProjectList,
        params: { ...filterData.value }
    };
});

function onEdit(v: Record<string, any> | null) {
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑项目" : "新建项目",
            width: "700px"
        },
        value: tempValue || {},
        formConfig: [
            ...(tempValue
                ? []
                : [
                      formHelper.select("工作空间", "workspace_id", workspaceOptions.value, {
                          rules: [ruleHelper.require("请选择工作空间")]
                      })
                  ]),
            formHelper.input("项目名称", "name", {
                rules: [ruleHelper.require("请输入项目名称")]
            }),
            formHelper.textarea("项目描述", "description"),
            formHelper.select("LLM 提供商", "llm_provider_id", llmOptions.value, {
                inputTips: "用于文本解析、角色抽取等 AI 功能"
            }),
            formHelper.select("TTS 提供商", "tts_provider_id", ttsOptions.value, {
                inputTips: "用于语音合成"
            }),
            formHelper.textarea("备注", "remark")
        ],
        ok: async (data: any) => {
            if (tempValue) {
                await updateProject(data);
            } else {
                await addProject(data);
            }
            table.value.refresh();
        }
    });
}
</script>
<style lang="scss" scoped></style>
