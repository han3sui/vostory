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
                            :disabled="!hasPermission('character:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleToggle(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
            <template #levelSlot>
                <a-table-column title="层级">
                    <template #cell="{ record }">
                        <a-tag :color="levelColor(record.level)" size="small">
                            {{ levelLabel(record.level) }}
                        </a-tag>
                    </template>
                </a-table-column>
            </template>
            <template #aliasSlot>
                <a-table-column title="别名">
                    <template #cell="{ record }">
                        <a-space wrap>
                            <a-tag v-for="a in record.aliases || []" :key="a" size="small">{{ a }}</a-tag>
                        </a-space>
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
    getCharacterList,
    addCharacter,
    updateCharacter,
    deleteCharacter,
    enableCharacter,
    disableCharacter,
    CharacterDetailType
} from "@/config/apis/character";
import { getProjectsByWorkspace, ProjectOptionType } from "@/config/apis/project";
import { getWorkspaceOptions, WorkspaceOptionType } from "@/config/apis/workspace";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const LEVELS = [
    { label: "主角", value: "main" },
    { label: "配角", value: "supporting" },
    { label: "龙套", value: "minor" }
];
const GENDERS = [
    { label: "男", value: "male" },
    { label: "女", value: "female" },
    { label: "未知", value: "unknown" }
];

function levelLabel(l: string) {
    return LEVELS.find((v) => v.value === l)?.label || l;
}
function levelColor(l: string) {
    const map: Record<string, string> = { main: "red", supporting: "blue", minor: "gray" };
    return map[l] || "gray";
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
        formHelper.select("项目", "project_id", projectOptions.value, { span: 6 }),
        formHelper.input("角色名称", "name", { span: 5, debounce: 500 }),
        formHelper.select("层级", "level", LEVELS, { span: 4 }),
        formHelper.select("性别", "gender", GENDERS, { span: 4 })
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: { rowKey: "id" },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("角色名称", "name"),
            tableHelper.slot("aliasSlot"),
            tableHelper.status("性别", "gender", (item: any) => {
                const found = GENDERS.find((g) => g.value === item.gender);
                return { text: found?.label || item.gender, status: "normal" };
            }),
            tableHelper.slot("levelSlot"),
            tableHelper.default("描述", "description"),
            tableHelper.default("排序", "sort_order"),
            tableHelper.slot("switchSlot"),
            tableHelper.date("更新时间", "updated_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("character:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("character:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除角色【${row.name}】？`,
                            onBeforeOk: async () => {
                                await deleteCharacter(row.id);
                                table.value.refresh();
                            }
                        });
                    }
                }
            ])
        ],
        trBtns: [
            {
                label: "添加角色",
                type: "primary",
                if: () => hasPermission("character:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getCharacterList,
        params: { ...filterData.value }
    };
});

function onEdit(v: Record<string, any> | null) {
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑角色" : "添加角色",
            width: "650px"
        },
        value: tempValue || { status: "0", level: "main", gender: "unknown", aliases: [] },
        formConfig: [
            ...(tempValue
                ? []
                : [
                      formHelper.select("所属项目", "project_id", projectOptions.value, {
                          rules: [ruleHelper.require("请选择项目")]
                      })
                  ]),
            formHelper.input("角色名称", "name", {
                rules: [ruleHelper.require("请输入角色名称")]
            }),
            formHelper.select("性别", "gender", GENDERS),
            formHelper.select("层级", "level", LEVELS),
            formHelper.textarea("角色描述", "description"),
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
                await updateCharacter(data);
            } else {
                await addCharacter(data);
            }
            table.value.refresh();
        }
    });
}

async function handleToggle(row: CharacterDetailType) {
    try {
        if (row.status === "0") {
            await disableCharacter(row.id);
        } else {
            await enableCharacter(row.id);
        }
        table.value.refresh();
        return true;
    } catch {
        return false;
    }
}
</script>
<style lang="scss" scoped></style>
