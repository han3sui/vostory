<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #s2>
                <a-table-column title="启用">
                    <template #cell="{ record }">
                        <a-switch
                            :disabled="!hasPermission('system:post:enable')"
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
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import {
    getPostList,
    addPost,
    updatePost,
    deletePost,
    enablePost,
    disablePost,
    PostListDetailType
} from "@/config/apis/system";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const table = ref();
const filterData = ref({});

const getFilterConfig = computed(() => {
    return [
        formHelper.input("岗位编码", "post_code", { span: 6, debounce: 500 }),
        formHelper.input("岗位名称", "post_name", { span: 6, debounce: 500 }),
        formHelper.select(
            "岗位状态",
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
const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: {
            rowKey: "id"
        },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("岗位编号", "post_code"),
            tableHelper.default("岗位名称", "post_name"),
            tableHelper.default("岗位排序", "post_sort"),
            tableHelper.slot("s2"),
            tableHelper.default("岗位描述", "remark"),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("system:post:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("system:post:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: "确认删除？",
                            onBeforeOk: async () => {
                                await deletePost(row.id);
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
                if: () => hasPermission("system:post:add"),
                handler: () => {
                    onEdit(null);
                }
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getPostList,
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
            formHelper.input("岗位编码", "post_code", { rules: [ruleHelper.require("请输入")] }),
            formHelper.input("岗位名称", "post_name", { rules: [ruleHelper.require("请输入")] }),
            formHelper.radio(
                "岗位状态",
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
            formHelper.inputNumber("岗位排序", "post_sort"),
            formHelper.textarea("岗位描述", "remark")
        ],
        ok: async (data: any) => {
            if (tempValue) {
                await updatePost(data);
            } else {
                await addPost(data);
            }
            table.value.refresh();
        }
    });
}

async function handleChangeIntercept(v2: PostListDetailType) {
    try {
        if (v2.status === "1") {
            await enablePost(v2.id);
        } else {
            await disablePost(v2.id);
        }
        table.value.refresh();
        return true;
    } catch (error) {
        return false;
    }
}
</script>
<style lang="scss" scoped></style>
