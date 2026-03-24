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
                            :disabled="!hasPermission('system:dict:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleChangeIntercept(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup name="AliveDictList">
import { Modal } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import {
    getDictTypeList,
    addDictType,
    updateDictType,
    deleteDictType,
    enableDictType,
    disableDictType,
    DictTypeDetailType
} from "@/config/apis/dict";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";
import routerHelper from "@/utils/helper/router";

const table = ref();
const filterData = ref({});

const getFilterConfig = computed(() => {
    return [
        formHelper.input("字典名称", "dict_name", { span: 6, debounce: 500 }),
        formHelper.input("字典类型", "dict_type", { span: 6, debounce: 500 }),
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
            rowKey: "id"
        },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("字典名称", "dict_name"),
            tableHelper.link("字典类型", "dict_type", (row) => {
                routerHelper.push({
                    name: "system-dict-detail",
                    query: {
                        dictId: row.id,
                        dictType: row.dict_type
                    }
                });
            }),
            tableHelper.slot("switchSlot"),
            tableHelper.default("备注", "remark"),
            tableHelper.date("创建时间", "created_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("system:dict:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("system:dict:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除字典【${row.dict_name}】？`,
                            onBeforeOk: async () => {
                                await deleteDictType(row.id);
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
                if: () => hasPermission("system:dict:add"),
                handler: () => {
                    onEdit(null);
                }
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getDictTypeList,
        params: { ...filterData.value }
    };
});

function onEdit(v: Record<string, any> | null) {
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑" : "添加"
        },
        value: tempValue || {},
        formConfig: [
            formHelper.input("字典名称", "dict_name", { rules: [ruleHelper.require("请输入")] }),
            formHelper.input("字典类型", "dict_type", {
                disabled: !!tempValue,
                rules: [ruleHelper.require("请输入")]
            }),
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
            if (tempValue) {
                await updateDictType(data);
            } else {
                await addDictType(data);
            }
            table.value.refresh();
        }
    });
}

async function handleChangeIntercept(v2: DictTypeDetailType) {
    try {
        if (v2.status === "1") {
            await enableDictType(v2.id);
        } else {
            await disableDictType(v2.id);
        }
        table.value.refresh();
        return true;
    } catch (error) {
        return false;
    }
}
</script>
<style lang="scss" scoped></style>
