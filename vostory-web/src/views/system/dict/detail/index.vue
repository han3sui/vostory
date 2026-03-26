<template>
    <frame-view>
        <a-page-header :title="`字典数据 - ${dictType}`" @back="onBack"></a-page-header>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #switchSlot>
                <a-table-column title="启用">
                    <template #cell="{ record }">
                        <a-switch
                            :disabled="!hasPermission('system:dict:data:enable')"
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
import { Modal } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper } from "@easyfe/admin-component";
import {
    getDictDataList,
    addDictData,
    updateDictData,
    deleteDictData,
    enableDictData,
    disableDictData,
    DictDataDetailType
} from "@/config/apis/dict";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";
import router from "@/packages/vue-router";
import QS from "@/utils/tools/qs";

const table = ref();
const dictId = computed(() => QS.getUrlkey("dictId"));
const dictType = computed(() => QS.getUrlkey("dictType") || "");

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: {
            rowKey: "id"
        },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("字典标签", "dict_label"),
            tableHelper.default("字典键值", "dict_value"),
            tableHelper.default("排序", "dict_sort"),
            tableHelper.slot("switchSlot"),
            tableHelper.default("备注", "remark"),
            tableHelper.date("创建时间", "created_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("system:dict:data:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("system:dict:data:remove"),
                    handler: (row: Record<string, any>) => {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除【${row.dict_label}】？`,
                            onBeforeOk: async () => {
                                await deleteDictData(row.id);
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
                if: () => hasPermission("system:dict:data:add"),
                handler: () => {
                    onEdit(null);
                }
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getDictDataList,
        params: {
            dict_type: dictType.value
        }
    };
});

function onEdit(v: Record<string, any> | null) {
    const tempValue = cloneDeep(v) || {};
    if (!v) {
        tempValue.dict_type = dictType.value;
        tempValue.status = "0";
        tempValue.is_default = "N";
        tempValue.dict_sort = 0;
    }
    ArcoModalFormShow({
        modalConfig: {
            title: v ? "编辑" : "添加"
        },
        value: tempValue,
        formConfig: [
            formHelper.input("字典类型", "dict_type", { disabled: true }),
            formHelper.input("字典标签", "dict_label", { rules: [ruleHelper.require("请输入")] }),
            formHelper.input("字典键值", "dict_value", { rules: [ruleHelper.require("请输入")] }),
            formHelper.inputNumber("排序", "dict_sort"),
            formHelper.input("样式属性", "css_class"),
            formHelper.select("回显样式", "list_class", [
                { label: "默认", value: "default" },
                { label: "主要", value: "primary" },
                { label: "成功", value: "success" },
                { label: "信息", value: "info" },
                { label: "警告", value: "warning" },
                { label: "危险", value: "danger" }
            ]),
            formHelper.radio(
                "是否默认",
                "is_default",
                [
                    { label: "是", value: "Y" },
                    { label: "否", value: "N" }
                ],
                { type: "radio" }
            ),
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
            if (v) {
                await updateDictData(data);
            } else {
                await addDictData(data);
            }
            table.value.refresh();
        }
    });
}

async function handleChangeIntercept(v2: DictDataDetailType) {
    try {
        if (v2.status === "1") {
            await enableDictData(v2.id);
        } else {
            await disableDictData(v2.id);
        }
        table.value.refresh();
        return true;
    } catch (error) {
        return false;
    }
}

function onBack() {
    router.back();
}
</script>
<style lang="scss" scoped></style>
