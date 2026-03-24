<template>
    <frame-view>
        <a-page-header title="字典详情" @back="onBack"> </a-page-header>
        <arco-table ref="table" :table-config="tableConfig" :req="getData"></arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { getDictItemList, addDictItem, updateDictItem, deleteDictItem } from "@/config/apis";
import { ArcoModalFormShow, ArcoTable, formHelper, ruleHelper, tableHelper } from "@easyfe/admin-component";
import { cloneDeep } from "lodash-es";
import { Modal } from "@arco-design/web-vue";
import router from "@/packages/vue-router";
import QS from "@/utils/tools/qs";

const table = ref();
const dictId = computed(() => {
    return QS.getUrlkey("dictId");
});
const dictType = computed(() => {
    return QS.getUrlkey("dictType");
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: {
            rowKey: "id",
            scroll: { y: "calc(100vh - 450px)" }
        },
        showRefresh: true,
        pageKey: "current",
        trBtns: [
            {
                label: "新增",
                type: "primary",
                handler: () => {
                    onEdit(null);
                }
            }
        ],
        rowKey: "records",
        columns: [
            tableHelper.default("类型", "dictType"),
            tableHelper.default("数据值", "value"),
            tableHelper.default("标签名", "label"),
            tableHelper.default("描述", "description"),
            tableHelper.default("排序", "sortOrder"),
            tableHelper.default("备注", "remarks"),
            tableHelper.default("操作人", "createBy"),
            tableHelper.default("更新时间", "updateTime"),
            tableHelper.btns("操作", [
                { label: "编辑", handler: onEdit },
                {
                    label: "删除",
                    status: "danger",
                    if: (row) => row.systemFlag !== "1",
                    handler: (row) => {
                        Modal.confirm({
                            title: "删除",
                            content: "确认删除吗？",
                            onBeforeOk: async () => {
                                await deleteDictItem(row.id);
                                table.value.refresh();
                            }
                        });
                    }
                }
            ])
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getDictItemList,
        params: {
            dictId: dictId.value
        }
    };
});

function onEdit(v: null | Record<string, any>) {
    const tempValue = cloneDeep(v) || {};
    tempValue.dictType = dictType.value;
    tempValue.dictId = dictId.value;
    ArcoModalFormShow({
        modalConfig: {
            title: v ? "编辑" : "新增"
        },
        value: tempValue,
        formConfig: [
            formHelper.input("字典类型", "dictType", {
                disabled: true
            }),
            formHelper.input("标签名", "label", {
                rules: ruleHelper.require("请输入")
            }),
            formHelper.input("数据值", "value", {
                rules: ruleHelper.require("请输入")
            }),
            formHelper.input("描述", "description", {
                rules: ruleHelper.require("请输入")
            }),
            formHelper.inputNumber("排序", "sortOrder", {
                rules: ruleHelper.require("请输入")
            }),
            formHelper.textarea("备注", "remarks")
        ],
        ok: async (data) => {
            if (v) {
                await updateDictItem(data);
            } else {
                await addDictItem(data);
            }
            table.value.refresh();
        }
    });
}

function onBack() {
    router.back();
}
</script>
<style lang="scss" scoped></style>
