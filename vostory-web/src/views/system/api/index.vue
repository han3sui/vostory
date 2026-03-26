<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="formData" :config="formConfig" layout="row"></arco-form>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { getApiList, getApiTagList } from "@/config/apis/system";
import { PageTableConfig } from "@/views/utils";
import { formHelper, ArcoTable, tableHelper, ArcoForm } from "@easyfe/admin-component";

const table = ref();
const formData = ref({});
const apiTagList = ref<string[]>([]);
const formConfig = computed(() => {
    return [
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
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: {
            rowKey: ""
        },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("名称", "name"),
            tableHelper.default("标签", "tag"),
            tableHelper.default("路径", "path"),
            tableHelper.default("权限标识", "perms"),
            tableHelper.default("方法", "method"),
            tableHelper.default("描述", "description")
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getApiList,
        params: formData.value
    };
});

async function getApiTagListFn() {
    apiTagList.value = await getApiTagList();
}

onMounted(() => {
    getApiTagListFn();
});
</script>
<style lang="scss" scoped></style>
