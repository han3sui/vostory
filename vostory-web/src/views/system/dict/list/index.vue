<template>
    <frame-view>
        <arco-table ref="table" :table-config="tableConfig" :req="getData"></arco-table>
    </frame-view>
</template>
<script lang="ts" setup name="AliveDictList-1">
import { getDictList } from "@/config/apis";
import { ArcoModalFormShow, ArcoTable, formHelper, ruleHelper, tableHelper } from "@easyfe/admin-component";
import { cloneDeep } from "lodash-es";
import { updateDict, addDict, deleteDict } from "@/config/apis/index";
import { Modal } from "@arco-design/web-vue";
import router from "@/packages/vue-router";

const table = ref();

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: {
            rowKey: "id",
            pagination: false
            // scroll: { y: "calc(100vh - 250px)" }
        },
        maxHeight: "auto",
        trBtns: [
            {
                label: "新增",
                type: "primary",
                handler: () => {
                    onEdit(null);
                }
            }
        ],
        rowKey: "",
        columns: [
            tableHelper.default("字典名称", "description"),
            tableHelper.link("字典类型", "dictType", (row) => {
                router.push({
                    name: "admin-dict-detail",
                    query: {
                        dictId: row.id,
                        dictType: row.dictType
                    }
                });
            }),
            tableHelper.status("配置类型", "systemFlag", (item) => {
                if (item.systemFlag === "1") {
                    return {
                        text: "系统类",
                        status: "normal"
                    };
                } else {
                    return {
                        text: "业务类",
                        status: "success"
                    };
                }
            }),
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
                                await deleteDict([row.id]);
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
        fn: getDictList
    };
});

function onEdit(v: null | Record<string, any>) {
    const tempValue = cloneDeep(v) || {};
    ArcoModalFormShow({
        modalConfig: {
            title: v ? "编辑" : "新增"
        },
        value: tempValue,
        formConfig: [
            formHelper.radio(
                "配置类型",
                "systemFlag",
                [
                    { label: "系统类", value: "1" },
                    { label: "业务类", value: "0" }
                ],
                {
                    disabled: !!v,
                    type: "radio",
                    rules: ruleHelper.require("请选择")
                }
            ),
            formHelper.input("字典类型", "dictType", {
                disabled: !!v,
                rules: ruleHelper.require("请输入")
            }),
            formHelper.input("描述", "description", {
                rules: ruleHelper.require("请输入")
            }),
            formHelper.textarea("备注", "remarks")
        ],
        ok: async (data) => {
            if (v) {
                await updateDict(data);
            } else {
                await addDict(data);
            }
            table.value.refresh();
        }
    });
}
</script>
<style lang="scss" scoped></style>
