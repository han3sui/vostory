<template>
    <frame-view>
        <arco-table ref="table" :table-config="tableConfig" :req="getData">
            <template #tlBtns>
                <arco-form v-model="filterData" layout="row" :config="filterConfig"></arco-form>
            </template>
            <template #s1>
                <a-table-column title="数据权限">
                    <template #cell="{ record }">
                        <a-tag color="arcoblue">{{ getAuthDesc(record.data_scope) }}</a-tag>
                    </template>
                </a-table-column>
            </template>
            <template #s2>
                <a-table-column title="启用">
                    <template #cell="{ record }">
                        <a-switch
                            :disabled="!hasPermission('system:role:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleChangeIntercept(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
        <arco-modal v-model:visible="menuVisible" :config="{ title: '菜单授权' }" :ok="onMenuOk">
            <a-form-item label="菜单名称">
                <a-input v-model="searchKey3" placeholder="搜索菜单" allow-clear style="width: 200px">
                    <template #prefix>
                        <icon-search />
                    </template>
                </a-input>
                <a-button style="margin-left: 20px" @click="toggleExpanded">展开/折叠</a-button>
            </a-form-item>
            <a-tree
                v-model:checked-keys="menuCheckedKeys"
                v-model:half-checked-keys="halfCheckedKeys"
                v-model:expanded-keys="expandedKeys"
                style="margin-right: 20px"
                :block-node="true"
                :checkable="true"
                size="large"
                :data="getTreeData(menuList, searchKey3)"
                :field-names="{
                    key: 'id',
                    title: 'menu_name',
                    children: 'children'
                }"
            >
                <template #extra="v">
                    <span v-if="v.method" style="margin-right: 4px">
                        <a-tag :color="getMethodColor(v.method)" size="small">{{ v.method }}</a-tag>
                    </span>
                    <span style="color: #999">{{ v.url || v.perms }}</span>
                </template>
            </a-tree>
        </arco-modal>
    </frame-view>
</template>
<script lang="ts" setup>
import {
    ArcoTable,
    tableHelper,
    formHelper,
    ArcoModalFormShow,
    ruleHelper,
    ArcoModal,
    ArcoForm
} from "@easyfe/admin-component";
import {
    getRoleList,
    getDeptTree,
    getRoleMenuDetail,
    addRole,
    updateRole,
    updateRoleMenu,
    deleteRole,
    MenuDetailType,
    RoleDetailType,
    getMenuTree,
    MenuTreeType,
    enableRole,
    disableRole,
    getRole
} from "@/config/apis/system";
import { hasPermission, initGlobal, PageTableConfig, recuTree, splitSelection, treeToArray } from "@/views/utils";
import { Message, Modal } from "@arco-design/web-vue";
import i18n from "@/locales";

const table = ref();
const expandedKeys = ref<number[]>([]);
const searchKey3 = ref("");
const menuCheckedKeys = ref<number[]>([]);
const halfCheckedKeys = ref<number[]>([]);
const menuVisible = ref(false);
const filterData = ref({
    role_name: "",
    role_key: "",
    status: ""
});
const filterConfig = computed(() => {
    return [
        formHelper.input("角色名称", "role_name", { span: 6, debounce: 500 }),
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
        arcoProps: {
            rowKey: "role_id"
        },
        ...PageTableConfig,
        columns: [
            tableHelper.default("角色名称", "role_name"),
            tableHelper.default("角色权限字符串", "role_key"),
            tableHelper.default("显示顺序", "role_sort"),
            tableHelper.slot("s1"),
            tableHelper.slot("s2"),
            tableHelper.default("备注", "remark"),
            tableHelper.date("创建时间", "created_at", {
                format: "YYYY-MM-DD HH:mm:ss"
            }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("system:role:edit"),
                    handler: (row) => {
                        editRole(row);
                    }
                },
                {
                    label: "授权",
                    if: () => hasPermission("system:role:edit:menus"),
                    handler: (row) => {
                        onMenu(row);
                    }
                },
                {
                    label: "删除",
                    status: "danger",
                    if: (item) => hasPermission("system:role:remove") && item.role_key !== "ROLE_ADMIN",
                    handler: (row: RoleDetailType) => {
                        Modal.confirm({
                            title: "提示",
                            content: "确认删除该角色吗？",
                            onBeforeOk: async () => {
                                await deleteRole(row.role_id);
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
                label: "添加",
                if: () => hasPermission("system:role:add"),
                handler: () => {
                    editRole(null);
                }
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getRoleList,
        params: { ...filterData.value }
    };
});

const getTreeData = computed(() => (v: any[], k: string) => {
    if (!k) return v;
    return recuTree((item) => item.name.toLowerCase().indexOf(k.toLowerCase()) > -1, v, "children");
});

const authMap = [
    { label: "全部数据权限", value: "1" },
    { label: "自定数据权限", value: "2" },
    { label: "本部门数据权限", value: "3" },
    { label: "本部门及以下数据权限", value: "4" },
    { label: "仅本人数据权限", value: "5" }
];

function getAuthDesc(v: string) {
    return authMap.find((item) => item.value === v)?.label || "未知";
}

function getMethodColor(method: string) {
    const colorMap: Record<string, string> = {
        GET: "green",
        POST: "arcoblue",
        PUT: "orange",
        DELETE: "red",
        PATCH: "purple"
    };
    return colorMap[method?.toUpperCase()] || "gray";
}

async function editRole(v: null | RoleDetailType) {
    let detail: null | RoleDetailType = null;
    if (v) {
        const loading = Message.loading("加载中...");
        detail = await getRole(v.role_id);
        loading.close();
    }
    ArcoModalFormShow({
        modalConfig: {
            title: v ? "编辑角色" : "添加角色"
        },
        value: detail || { status: "0", data_scope: "1" },
        formConfig: [
            formHelper.input("角色名称", "role_name", { rules: ruleHelper.require("请输入") }),
            formHelper.input("角色标识", "role_key", {
                inputTips: "请输入大写英文、下划线",
                rules: [
                    ruleHelper.require("请输入"),
                    {
                        validator: (value: any, callback: any) => {
                            if (/^[A-Z_][A-Z_]{0,31}$/.test(value)) {
                                callback();
                            } else {
                                callback("请输入大写英文、下划线");
                            }
                        }
                    }
                ],
                disabled: !!detail
            }),
            formHelper.inputNumber("显示顺序", "role_sort", { rules: ruleHelper.require("请输入") }),
            formHelper.select("数据范围", "data_scope", authMap, { rules: ruleHelper.require("请选择") }),
            formHelper.treeSelect("权限范围", "data_scope_ids", deptTree.value, {
                multiple: true,
                if: (v: any) => v.data_scope === "2",
                allowSearch: true,
                filterTreeNode(searchKey: string, nodeData: any) {
                    return nodeData?.dept_name?.toLowerCase().indexOf(searchKey.toLowerCase()) > -1;
                },
                fieldNames: {
                    key: "id",
                    title: "dept_name",
                    children: "children"
                },
                rules: ruleHelper.require("请选择")
            }),
            formHelper.radio(
                "状态",
                "status",
                [
                    { label: "正常", value: "0" },
                    { label: "停用", value: "1" }
                ],
                { rules: ruleHelper.require("请选择"), type: "radio" }
            ),
            formHelper.textarea("备注", "remark")
        ],
        ok: async (data: any) => {
            if (v) {
                await updateRole({ ...data, id: v.role_id });
            } else {
                await addRole(data);
            }
            Message.success(v ? "编辑成功" : "添加成功");
            table.value.refresh();
        }
    });
}

const allExpandedKeys = computed(() => {
    return flatMenuList.value.map((item) => item.id);
});

async function onMenu(row: RoleDetailType) {
    editRoleDetail.value = row;
    const loading = Message.loading("加载中...");
    const res = await getRoleMenuDetail(row.role_id);
    const { fullySelected, partiallySelected } = splitSelection(menuList.value, res);
    halfCheckedKeys.value = partiallySelected;
    menuCheckedKeys.value = fullySelected;
    loading.close();
    expandedKeys.value = [];
    menuVisible.value = true;
}

function toggleExpanded() {
    expandedKeys.value = expandedKeys.value.length ? [] : allExpandedKeys.value;
}

let deptTree = ref<any[]>([]);
let menuList = ref<MenuTreeType[]>([]);
let editRoleDetail = ref<RoleDetailType>();

const flatMenuList = computed<MenuDetailType[]>(() => {
    return treeToArray(menuList.value, "children") as MenuDetailType[];
});
async function init() {
    menuList.value = await getMenuTree({});
    deptTree.value = await getDeptTree({});
}

async function handleChangeIntercept(v2: RoleDetailType) {
    try {
        if (v2.status === "1") {
            await enableRole(v2.role_id);
        } else {
            await disableRole(v2.role_id);
        }
        table.value.refresh();
        return true;
    } catch (error) {
        return false;
    }
}

async function onMenuOk() {
    const roleId = editRoleDetail.value?.role_id;
    if (!roleId) {
        Message.error("请选择角色");
        return;
    }
    await updateRoleMenu({
        menu_ids: [...halfCheckedKeys.value, ...menuCheckedKeys.value],
        role_id: roleId
    });
    Message.success(i18n.global.t("授权成功"));
    initGlobal();
}

onMounted(() => {
    init();
});
</script>
<style lang="scss" scoped></style>
