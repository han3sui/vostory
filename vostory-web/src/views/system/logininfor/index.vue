<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" layout="row" :config="filterConfig" />
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { formHelper, ArcoTable, tableHelper, ArcoForm } from "@easyfe/admin-component";
import { getLoginInfoList } from "@/config/apis/system";
import { PageTableConfig } from "@/views/utils";
import dayjs from "dayjs";

const table = ref();
const filterData = ref({
    start_time: "",
    end_time: ""
});
const filterConfig = computed(() => {
    return [
        formHelper.input("用户名", "login_name", {
            span: 6,
            debounce: 500
        }),
        formHelper.select(
            "登录状态",
            "status",
            [
                {
                    label: "成功",
                    value: "0"
                },
                {
                    label: "失败",
                    value: "1"
                }
            ],
            {
                span: 6
            }
        ),
        formHelper.rangePicker("日期", "key10", {
            type: "date",
            span: 6,
            onChange(value?: any[]) {
                if (!value) {
                    filterData.value.start_time = "";
                    filterData.value.end_time = "";
                } else {
                    filterData.value.start_time = dayjs(value[0]).startOf("day").format("YYYY-MM-DD HH:mm:ss");
                    filterData.value.end_time = dayjs(value[1]).endOf("day").format("YYYY-MM-DD HH:mm:ss");
                }
            }
        })
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: {
            rowKey: "info_id"
        },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("用户名", "login_name"),
            tableHelper.status("登录状态", "status", (item) => {
                return {
                    status: item.status === "0" ? "success" : "danger",
                    text: item.status === "0" ? "成功" : "失败"
                };
            }),
            tableHelper.default("登录地址", "ipaddr"),
            tableHelper.default("登录地点", "login_location"),
            tableHelper.default("浏览器", "browser"),
            tableHelper.default("操作系统", "os"),
            tableHelper.default("备注", "msg"),
            tableHelper.date("登录时间", "login_time", {
                format: "YYYY-MM-DD HH:mm:ss"
            })
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getLoginInfoList,
        params: filterData.value
    };
});
</script>
<style lang="scss" scoped></style>
