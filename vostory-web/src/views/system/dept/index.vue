<template>
    <frame-view>
        <arco-table ref="table" :table-config="getTableConfig" :req="getData" @export="onExport">
            <template #tlBtns>
                <div style="display: flex; align-items: center; width: 100%">
                    <a-button style="margin-right: 20px" @click="onExpend">展开/折叠</a-button>
                    <arco-form v-model="filterData" :config="getFilterConfig" layout="row" style="flex: 1"></arco-form>
                </div>
            </template>
            <template #leader>
                <a-table-column title="负责人">
                    <template #cell="{ record }">
                        <span v-if="record.leader_user">{{ record.leader_user.user_name }}</span>
                        <span v-else-if="record.leader">{{ record.leader }}</span>
                        <span v-else>-</span>
                    </template>
                </a-table-column>
            </template>
            <template #s2>
                <a-table-column title="启用">
                    <template #cell="{ record }">
                        <a-switch
                            :disabled="!hasPermission('system:dept:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleChangeIntercept(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { ArcoTable, tableHelper, formHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import {
    getDeptTree,
    updateDept,
    addDept,
    DeptDetailType,
    deleteDept,
    enableDept,
    disableDept
} from "@/config/apis/system";
import { cloneDeep } from "lodash-es";
import { Modal } from "@arco-design/web-vue";
import { hasPermission } from "@/views/utils";
import { useUserListOptions } from "@/hooks/useCommon";

const table = ref();
const { userListOptions } = useUserListOptions();
const getFilterConfig = computed(() => {
    return [
        formHelper.input("部门名称", "dept_name", {
            debounce: 500,
            span: 6
        }),
        formHelper.select(
            "状态",
            "status",
            [
                { label: "正常", value: "0" },
                { label: "停用", value: "1" }
            ],
            {
                span: 6
            }
        )
    ];
});

const filterData = ref<any>({});

const getData = computed(() => {
    return {
        fn: getDeptTree,
        params: { ...filterData.value }
    };
});

const getTableConfig = computed(() => {
    return tableHelper.create({
        rowKey: "",
        arcoProps: {
            rowKey: "id",
            pagination: false,
            scroll: {
                y: "70vh"
            }
        },
        trBtns: [
            {
                label: "新建",
                type: "primary",
                if: () => hasPermission("system:dept:add"),
                handler: () => {
                    editMenu(null);
                }
            }
        ],
        columns: [
            tableHelper.default("部门名称", "dept_name"),
            tableHelper.slot("leader"),
            tableHelper.default("排序", "order_num"),
            tableHelper.slot("s2"),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("system:dept:edit"),
                    handler: (row) => {
                        editMenu(row);
                    }
                },
                {
                    label: "删除",
                    color: "red",
                    if: () => hasPermission("system:dept:remove"),
                    handler(row: DeptDetailType) {
                        Modal.confirm({
                            title: "删除",
                            content: "确认删除该部门？",
                            onBeforeOk: async () => {
                                try {
                                    await deleteDept(row.id);
                                    table.value.refresh();
                                } catch (error) {
                                    return false;
                                }
                            }
                        });
                    }
                }
            ])
        ]
    });
});

let menuList = ref<any[]>([]);

async function editMenu(row: null | DeptDetailType) {
    const tempValue = cloneDeep(row);
    const deptTree = [
        {
            id: 0,
            dept_name: "根部门",
            children: menuList.value
        }
    ];
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑部门" : "新建部门"
        },
        value: tempValue || { status: "0", order_num: 1 },
        formConfig: [
            formHelper.input("部门名称", "dept_name", { rules: [ruleHelper.require("请输入")] }),
            formHelper.treeSelect("上级部门", "parent_id", deptTree, {
                allowSearch: true,
                filterTreeNode(searchKey: string, nodeData: any) {
                    return nodeData?.dept_name?.toLowerCase().indexOf(searchKey.toLowerCase()) > -1;
                },
                fieldNames: {
                    key: "id",
                    title: "dept_name",
                    children: "children"
                },
                rules: [ruleHelper.require("请选择上级部门")]
            }),
            formHelper.select("负责人", "leader_id", userListOptions.value, {
                allowSearch: true,
                allowClear: true,
                placeholder: "请选择负责人"
            }),
            formHelper.inputNumber("排序", "order_num", { rules: [ruleHelper.require("请输入")] }),
            formHelper.radio(
                "状态",
                "status",
                [
                    { label: "正常", value: "0" },
                    { label: "停用", value: "1" }
                ],
                {
                    type: "radio",
                    rules: [ruleHelper.require("请选择")]
                }
            ),
            formHelper.textarea("备注", "remark")
        ],
        ok: async (data: any) => {
            // 根据选择的用户ID获取用户名称作为冗余字段
            const selectedUser = userListOptions.value.find((u) => u.value === data.leader_id);
            const submitData = {
                ...data,
                leader: selectedUser?.label || "",
                leader_id: selectedUser?.value || null
            };
            if (tempValue) {
                await updateDept({ ...submitData, id: tempValue.id });
            } else {
                await addDept(submitData);
            }
            table.value.refresh();
        }
    });
}
let expendStatus = false;

async function handleChangeIntercept(v2: DeptDetailType) {
    try {
        if (v2.status === "1") {
            await enableDept(v2.id);
        } else {
            await disableDept(v2.id);
        }
        table.value.refresh();
        return true;
    } catch (error) {
        return false;
    }
}

function onExpend() {
    expendStatus = !expendStatus;
    table.value.baseTable.expandAll(expendStatus);
}

function onExport(v: any[]) {
    menuList.value = v;
}
</script>
<style lang="scss" scoped></style>
