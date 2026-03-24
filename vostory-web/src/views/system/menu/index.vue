<template>
    <frame-view>
        <arco-table
            ref="table"
            :table-config="getTableConfig"
            :filter-data="filterData"
            :filter-config="getFilterConfig"
            :req="getData"
            @export="onExport"
        >
            <template #s1>
                <a-table-column title="菜单类型">
                    <template #cell="{ record }">
                        <a-tag v-if="record.menu_type === 'M'" color="purple">目录</a-tag>
                        <a-tag v-if="record.menu_type === 'F'" color="green">按钮</a-tag>
                        <a-tag v-else-if="record.menu_type === 'C'" color="blue">菜单</a-tag>
                    </template>
                </a-table-column>
            </template>
            <template #s2>
                <a-table-column title="是否可见">
                    <template #cell="{ record }">
                        <a-tag v-if="record.visible === '1'" color="orangered">否</a-tag>
                        <a-tag v-else>是</a-tag>
                    </template>
                </a-table-column>
            </template>
            <template #tlBtns>
                <a-button @click="onExpend">展开/折叠</a-button>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import {
    ArcoTable,
    tableHelper,
    formHelper,
    ArcoModalFormShow,
    ruleHelper,
    ArcoModalTableShow
} from "@easyfe/admin-component";
import {
    getMenuTree,
    updateMenu,
    addMenu,
    MenuDetailType,
    deleteMenu,
    getApiList,
    muAddPers,
    MenuTreeType
} from "@/config/apis/system";
import { cloneDeep } from "lodash-es";
import { Modal } from "@arco-design/web-vue";
import { hasPermission, PageTableConfig } from "@/views/utils";
import { getApiTagList } from "@/config/apis/system";

const table = ref();

const apiTagList = ref<string[]>([]);

const getFilterConfig = computed(() => {
    return [
        formHelper.input("菜单名称", "menu_name", {
            span: 6,
            debounce: 500
        }),
        formHelper.select(
            "菜单类型",
            "menu_type",
            [
                { label: "目录", value: "M" },
                { label: "菜单", value: "C" },
                { label: "按钮", value: "F" }
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
        fn: getMenuTree,
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
                y: "65vh"
            }
        },
        trBtns: [
            {
                label: "新建",
                type: "primary",
                if: () => hasPermission("system:menu:add"),
                handler: () => {
                    editMenu(null);
                }
            }
        ],
        columns: [
            tableHelper.default("菜单名称", "menu_name"),
            tableHelper.custom("菜单/标识", (row: MenuDetailType) => {
                if (row.menu_type === "M") {
                    return "-";
                }
                if (row.menu_type === "C") {
                    return row.url ? row.url : "-";
                }
                if (row.menu_type === "F") {
                    return row.perms ? row.perms : "-";
                }
                return "-";
            }),
            tableHelper.slot("s1"),
            tableHelper.slot("s2"),
            tableHelper.default("排序", "order_num"),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: (row: MenuDetailType) => {
                        return hasPermission("system:menu:edit") && ["C", "M"].includes(row.menu_type);
                    },
                    handler: (row) => {
                        editMenu(row);
                    }
                },
                {
                    label: "删除",
                    color: "red",
                    if: () => hasPermission("system:menu:remove"),
                    handler(row: MenuDetailType) {
                        Modal.confirm({
                            title: "删除",
                            content: "确认删除该菜单？",
                            onBeforeOk: async () => {
                                await deleteMenu(row.id);
                                table.value.refresh();
                            }
                        });
                    }
                },
                {
                    label: "添加权限",
                    status: "warning",
                    if: (row) => {
                        return hasPermission("system:menu:muti:add") && ["C", "M"].includes(row.menu_type);
                    },
                    handler: (row) => {
                        addApi(row);
                    }
                }
            ])
        ]
    });
});

let menuList = ref<MenuTreeType[]>([]);

// 递归过滤菜单树，排除 menu_type 为 F 的节点
function filterMenuTree(menus: MenuTreeType[]): MenuTreeType[] {
    return menus
        .filter((menu) => menu.menu_type !== "F")
        .map((menu) => ({
            ...menu,
            children: menu.children ? filterMenuTree(menu.children) : []
        }));
}

async function editMenu(row: null | MenuDetailType) {
    const tempValue = cloneDeep(row);
    const menuTree = [
        {
            id: 0,
            menu_name: "根菜单",
            children: filterMenuTree(menuList.value)
        }
    ];
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑菜单" : "新建菜单"
        },
        value: tempValue || { visible: "0", menu_type: "M" },
        formConfig: [
            formHelper.treeSelect("上级菜单", "parent_id", menuTree, {
                allowSearch: true,
                filterTreeNode(searchKey: string, nodeData: any) {
                    return nodeData?.menu_name?.toLowerCase().indexOf(searchKey.toLowerCase()) > -1;
                },
                fieldNames: {
                    key: "id",
                    title: "menu_name",
                    children: "children"
                },
                rules: [ruleHelper.require("请选择上级菜单")]
            }),
            formHelper.radio(
                "菜单类型",
                "menu_type",
                [
                    { label: "目录", value: "M" },
                    { label: "菜单", value: "C" }
                    // { label: "按钮", value: "F" }
                ],
                {
                    type: "radio",
                    rules: [ruleHelper.require("请选择")]
                }
            ),
            formHelper.input("菜单名称", "menu_name", { rules: [ruleHelper.require("请输入")] }),
            formHelper.input("路由", "url", {
                rules: [ruleHelper.require("请输入")],
                if: (value: any) => {
                    return ["C"].includes(value?.menu_type);
                }
            }),
            formHelper.input("权限标识", "perms", {
                inputTips: "权限标识是唯一的，如：sys:menu:add",
                rules: [ruleHelper.require("请输入")],
                if: (value: any) => {
                    return ["F"].includes(value?.menu_type);
                }
            }),
            formHelper.radio(
                "显示状态",
                "visible",
                [
                    { label: "显示", value: "0" },
                    { label: "隐藏", value: "1" }
                ],
                {
                    type: "radio",
                    rules: [ruleHelper.require("请选择")],
                    if: (value: any) => {
                        return ["C"].includes(value?.menu_type);
                    }
                }
            ),
            formHelper.inputNumber("排序", "order_num")
        ],
        ok: async (data: any) => {
            if (tempValue) {
                await updateMenu({ ...data, id: tempValue.id });
            } else {
                await addMenu(data);
            }
            table.value.refresh();
        }
    });
}

async function addApi(row: MenuDetailType) {
    const filterData = ref<any>({});
    ArcoModalTableShow({
        modalConfig: {
            title: "选择接口",
            width: "1200px"
        },
        defaultSelected: [],
        tableConfig: {
            tableConfig: {
                arcoProps: {
                    rowKey: "id",
                    rowSelection: {
                        type: "checkbox",
                        showCheckedAll: true
                    }
                },
                ...PageTableConfig,
                showRefresh: false,
                maxHeight: "40vh",
                columns: [
                    tableHelper.default("名称", "name"),
                    tableHelper.default("标签", "tag"),
                    tableHelper.default("路径", "path"),
                    tableHelper.default("权限标识", "perms"),
                    tableHelper.default("方法", "method"),
                    tableHelper.default("描述", "description")
                ]
            },
            filterConfig: [
                formHelper.input("名称", "name", { span: 5, debounce: 500 }),
                formHelper.input("路径", "path", { span: 5, debounce: 500 }),
                formHelper.input("权限标识", "perms", { span: 5, debounce: 500 }),
                formHelper.select(
                    "标签",
                    "tag",
                    apiTagList.value.map((item) => ({ label: item, value: item })),
                    {
                        span: 6,
                        allowSearch: true
                    }
                )
            ],
            filterData: filterData.value,
            req: {
                fn: getApiList,
                params: { ...filterData.value }
            }
        },
        ok: async (v: any[]) => {
            const data = v.map((item) => ({
                parent_id: row.id,
                menu_name: item.name,
                perms: item.perms
            }));
            await muAddPers(data);
            table.value.refresh();
        }
    });
}

let expendStatus = false;

function onExpend() {
    expendStatus = !expendStatus;
    table.value.baseTable.expandAll(expendStatus);
}

function onExport(v: MenuTreeType[]) {
    menuList.value = v;
}

async function getApiTagListFn() {
    apiTagList.value = await getApiTagList();
}

onMounted(() => {
    getApiTagListFn();
});
</script>
<style lang="scss" scoped></style>
