<template>
    <frame-view>
        <arco-table ref="table" :table-config="getTableConfig" :req="getData">
            <template #tlBtns>
                <arco-form v-model="filterData" layout="row" :config="getFilterConfig" />
            </template>
            <template #s1>
                <a-table-column title="岗位" :width="250">
                    <template #cell="{ record }">
                        <a-space wrap>
                            <a-tag v-for="item in record.posts" :key="item.postId">{{ item.post_name }}</a-tag>
                        </a-space>
                    </template>
                </a-table-column>
            </template>
            <template #s2>
                <a-table-column title="角色">
                    <template #cell="{ record }">
                        <a-space wrap>
                            <a-tag v-for="item in record.roles" :key="item.roleId">{{ item.role_name }}</a-tag>
                        </a-space>
                    </template>
                </a-table-column>
            </template>
            <template #s4>
                <a-table-column title="直属上级" :width="120">
                    <template #cell="{ record }">
                        <span v-if="record.superior">{{ record.superior.user_name }}</span>
                        <span v-else-if="record.superior_name">{{ record.superior_name }}</span>
                        <span v-else class="text-gray-400">-</span>
                    </template>
                </a-table-column>
            </template>
            <template #s3>
                <a-table-column title="启用">
                    <template #cell="{ record }">
                        <a-switch
                            :disabled="!hasPermission('system:user:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleChangeIntercept(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
        <a-upload style="display: none" accept=".xls,.xlsx" :custom-request="uploadFile">
            <template #upload-button>
                <a-button ref="myUpload"></a-button>
            </template>
        </a-upload>
        <arco-modal-table
            v-model:visible="errorModal"
            :modal-config="{ title: '校验失败数据', hideCancel: true, width: 800 }"
            :table-config="errorTableConfig"
        >
            <template #s1>
                <a-table-column title="错误描述">
                    <template #cell="{ record }">
                        <a-space>
                            <a-tag v-for="item in record.errors" :key="item" color="red">{{ item }}</a-tag>
                        </a-space>
                    </template>
                </a-table-column>
            </template>
        </arco-modal-table>
    </frame-view>
</template>
<script lang="ts" setup>
import {
    ArcoTable,
    tableHelper,
    formHelper,
    ArcoModalFormShow,
    ruleHelper,
    ArcoModalTable,
    ArcoForm
} from "@easyfe/admin-component";
import {
    getUserList,
    getUserOptions,
    getDeptTree,
    addUser,
    updateUser,
    getUser,
    deleteUser,
    UserDetailType,
    getRoleList,
    getPostList,
    enableUser,
    disableUser,
    resetUserPassword,
    importUsers,
    downloadUserImportTemplate
} from "@/config/apis/system";
import { Message, Modal } from "@arco-design/web-vue";
import { cloneDeep, isArray } from "lodash-es";
import { encryptPassword, hasPermission, PageTableConfig } from "@/views/utils";

const myUpload = ref();
const table = ref();
const errorModal = ref(false);
const errorData = ref<any[]>([]);

const errorTableConfig = computed(() => {
    return {
        list: errorData.value,
        tableConfig: {
            rowKey: "",
            showRefresh: false,
            arcoProps: {
                pagination: false,
                rowKey: "lineNum"
            },
            columns: [tableHelper.default("行号", "lineNum", { width: 80 }), tableHelper.slot("s1")]
        }
    };
});

const getFilterConfig = computed(() => {
    return [
        formHelper.input("用户名", "login_name", {
            span: 5,
            debounce: 500
        }),
        formHelper.input("姓名", "user_name", {
            span: 5,
            debounce: 500
        }),
        formHelper.select(
            "状态",
            "status",
            [
                { label: "正常", value: "0" },
                { label: "禁用", value: "1" }
            ],
            {
                span: 5
            }
        ),
        formHelper.treeSelect("组织机构", "dept_id", originTreeData.value, {
            allowSearch: true,
            filterTreeNode(searchKey: string, nodeData: any) {
                return nodeData?.dept_name?.toLowerCase().indexOf(searchKey.toLowerCase()) > -1;
            },
            placeholder: "请选择",
            span: 5,
            fieldNames: {
                key: "id",
                title: "dept_name",
                children: "children"
            }
        })
    ];
});

const originTreeData = ref<any[]>([]);

const filterData = ref<any>({});

const importLoading = ref(false);

const getData = computed(() => {
    return {
        fn: getUserList,
        params: {
            ...filterData.value
        }
    };
});

const getTableConfig = computed(() => {
    return tableHelper.create({
        ...PageTableConfig,
        maxHeight: "auto",
        arcoProps: {
            rowKey: "id"
        },
        trBtns: [
            {
                label: "新建",
                type: "primary",
                if: () => hasPermission("system:user:add"),
                handler: () => {
                    editUser(null);
                }
            },
            {
                label: "导入用户",
                type: "outline",
                loading: importLoading.value,
                if: () => hasPermission("system:user:import"),
                handler: () => {
                    handleImportUser();
                }
            },
            {
                label: "下载模板",
                type: "text",
                if: () => hasPermission("system:user:import:template"),
                handler: () => {
                    handleDownloadTemplate();
                }
            }
        ],
        columns: [
            tableHelper.default("用户名", "login_name"),
            tableHelper.default("姓名", "user_name"),
            tableHelper.slot("s1"),
            tableHelper.slot("s2"),
            tableHelper.slot("s4"),
            tableHelper.slot("s3"),
            tableHelper.default("部门", "dept_name"),
            // tableHelper.default("联系电话", "phone"),
            tableHelper.default("邮箱", "email"),
            tableHelper.date("更新时间", "update_at", {
                format: "YYYY-MM-DD HH:mm:ss",
                width: 150
            }),
            tableHelper.btns("操作", [
                {
                    label: "密码",
                    if: () => hasPermission("system:user:updatepwd"),
                    handler: (row) => {
                        resetPassword(row);
                    }
                },
                {
                    label: "编辑",
                    if: (row) => hasPermission("system:user:edit") && row.user_type !== "99",
                    handler: (row) => {
                        editUser(row);
                    }
                },
                {
                    label: "删除",
                    color: "red",
                    if: (row) => hasPermission("system:user:remove") && row.user_type !== "99",
                    handler: (row: UserDetailType) => {
                        onDelUser(row.user_id);
                    }
                }
            ])
        ]
    });
});

function resetPassword(row: UserDetailType) {
    let changeData: any = {};
    ArcoModalFormShow({
        modalConfig: {
            title: "重置密码"
        },
        formConfig: [
            formHelper.input("新密码", "password", {
                formType: "password",
                rules: [{ required: true, message: "请输入密码" }]
            }),
            formHelper.input("确认密码", "REPASSWORD", {
                formType: "password",
                rules: [
                    {
                        required: true,
                        validator: (value: string, callback) => {
                            if (!value) {
                                callback("请输入确认密码");
                                return;
                            }
                            if (value !== changeData?.password) {
                                callback("两次密码不一致");
                                return;
                            }
                            callback();
                        }
                    }
                ]
            })
        ],
        ok: async (data) => {
            await resetUserPassword(row.user_id, encryptPassword(data.password));
            table.value.refresh();
        },
        change: (data) => {
            changeData = data;
        }
    });
}

function onDelUser(ids: number) {
    Modal.confirm({
        title: "提示",
        content: "确认删除该用户？",
        onBeforeOk: async () => {
            await deleteUser(ids);
            Message.success("删除成功");
            table.value.refresh();
        }
    });
}

function initDeptList() {
    getDeptTree({}).then((res) => {
        originTreeData.value = res;
    });
    getRoleList({}).then((res) => {
        roleList.value = res.data;
    });
    getPostList({}).then((res) => {
        postList.value = res.data;
    });
    // 获取用户列表用于选择直属上级
    getUserOptions().then((res) => {
        userList.value = res.map((item) => {
            return {
                label: `${item.user_name}-${item.login_name}`,
                value: item.user_id
            };
        });
    });
}

let roleList = ref<any[]>([]);
let postList = ref<any[]>([]);
let userList = ref<{ label: string; value: number }[]>([]);

async function editUser(row: UserDetailType | null) {
    let detail: null | UserDetailType = null;
    if (row) {
        const loading = Message.loading("加载中...");
        detail = await getUser(row.user_id);
        loading.close();
    }
    ArcoModalFormShow({
        modalConfig: {
            title: detail ? "编辑用户" : "新建用户"
        },
        value: detail || { status: "0" },
        formConfig: [
            formHelper.input("用户名", "login_name", {
                labelTips: "用户名只能由字母、数字、下划线组成",
                disabled: !!detail,
                rules: [
                    {
                        required: true,
                        validator: (value: string, callback) => {
                            if (!value) {
                                callback("请输入用户名");
                                return;
                            }
                            if (!/^[a-zA-Z0-9_]+$/.test(value)) {
                                callback("只能由字母、数字、下划线组成");
                                return;
                            }
                            callback();
                        }
                    }
                ]
            }),
            formHelper.input("昵称", "user_name", {
                rules: [{ required: true, message: "请输入姓名" }]
            }),

            formHelper.input("密码", "password", {
                formType: "password",
                if: !detail,
                rules: [{ required: true, message: "请输入密码" }]
            }),
            formHelper.treeSelect("部门", "dept_id", originTreeData.value, {
                allowSearch: true,
                placeholder: "请选择",
                fieldNames: {
                    key: "id",
                    title: "dept_name",
                    children: "children"
                },
                rules: [ruleHelper.require("请选择")]
            }),
            formHelper.select("角色", "role_ids", roleList.value, {
                multiple: true,
                fieldNames: {
                    label: "role_name",
                    value: "role_id"
                },
                rules: [ruleHelper.require("请选择")]
            }),
            formHelper.radio(
                "启用状态",
                "status",
                [
                    { label: "启用", value: "0" },
                    { label: "禁用", value: "1" }
                ],
                {
                    type: "radio",
                    rules: [ruleHelper.require("请选择")]
                }
            ),
            formHelper.select("直属上级", "superior_id", userList.value, {
                allowClear: true,
                allowSearch: true,
                placeholder: "请选择直属上级（非必填）"
            }),
            formHelper.select("岗位", "post_ids", postList.value, {
                multiple: true,
                fieldNames: {
                    label: "post_name",
                    value: "id"
                }
            }),
            formHelper.input("联系电话", "phonenumber"),
            formHelper.input("邮箱", "email"),
            formHelper.textarea("备注", "remark")
        ],
        ok: async (data: any) => {
            if (detail) {
                const temValue = cloneDeep(data);
                delete temValue.password;
                if (!detail.superior_id) {
                    delete temValue.superior_id;
                }
                await updateUser(temValue);
            } else {
                await addUser({
                    ...data,
                    superior_id: data.superior_id ? Number(data.superior_id) : null,
                    password: encryptPassword(data.password)
                });
            }
            table.value.refresh();
        }
    });
}

async function handleChangeIntercept(v2: UserDetailType) {
    try {
        if (v2.user_type === "99") {
            Message.warning("超级管理员不能禁用");
            return false;
        }
        if (v2.status === "0") {
            await enableUser(v2.user_id);
        } else {
            await disableUser(v2.user_id);
        }
        table.value.refresh();
        return true;
    } catch (error) {
        return false;
    }
}

function handleImportUser() {
    myUpload.value?.$el.click();
}

function uploadFile(options: any) {
    importLoading.value = true;
    const file = options.fileItem.file;

    importUsers(file)
        .then((res) => {
            Message.success(`导入成功，成功${res.success_count}条`);
            table.value.refresh();
        })
        .catch((err) => {
            if (isArray(err.data)) {
                errorData.value = err.data;
                errorModal.value = true;
            } else {
                Message.error(err.msg || "导入失败");
            }
        })
        .finally(() => {
            importLoading.value = false;
        });
    return {};
}

async function handleDownloadTemplate() {
    try {
        const res = await downloadUserImportTemplate();
        const blob = new Blob([res as any], {
            type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
        });
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = "用户导入模板.xlsx";
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);
    } catch (error) {
        Message.error("下载模板失败");
    }
}

onMounted(() => {
    initDeptList();
});
</script>
<style lang="scss" scoped></style>
