<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #statusSlot>
                <a-table-column title="状态">
                    <template #cell="{ record }">
                        <a-switch
                            :model-value="record.status === '0'"
                            @change="(val: boolean) => handleToggleStatus(record, val)"
                        />
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { Message } from "@arco-design/web-vue";
import {
    formHelper,
    ArcoTable,
    tableHelper,
    ArcoForm,
    ArcoModalFormShow,
    ruleHelper
} from "@easyfe/admin-component";
import {
    getVoiceProfileList,
    addVoiceProfile,
    updateVoiceProfile,
    deleteVoiceProfile,
    enableVoiceProfile,
    disableVoiceProfile,
    VoiceProfileDetailType
} from "@/config/apis/voice-profile";
import { getWorkspaceOptions, WorkspaceOptionType } from "@/config/apis/workspace";
import { getProjectsByWorkspace, ProjectOptionType } from "@/config/apis/project";
import { hasPermission, PageTableConfig } from "@/views/utils";
import { cloneDeep } from "lodash-es";

const table = ref();
const filterData = ref<Record<string, any>>({});
const projectOptions = ref<{ label: string; value: number }[]>([]);

onMounted(async () => {
    const wsRes = await getWorkspaceOptions();
    for (const ws of wsRes as WorkspaceOptionType[]) {
        const projects = await getProjectsByWorkspace(ws.id);
        for (const p of projects as ProjectOptionType[]) {
            projectOptions.value.push({ label: `${ws.name} / ${p.name}`, value: p.id });
        }
    }
});

const getFilterConfig = computed(() => {
    return [
        formHelper.select("项目", "project_id", projectOptions.value, { span: 8, allowClear: true }),
        formHelper.input("名称", "name", { span: 6, debounce: 500 })
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: { rowKey: "id" },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        trBtns: [
            {
                label: "新增声音配置",
                if: () => hasPermission("voice-profile:add"),
                handler: () => {
                    handleAdd();
                }
            }
        ],
        columns: [
            tableHelper.default("名称", "name"),
            tableHelper.default("参考文本", "reference_text"),
            tableHelper.default("TTS 提供商", "tts_provider_name"),
            tableHelper.slot("statusSlot"),
            tableHelper.date("创建时间", "created_at", { format: "YYYY-MM-DD HH:mm" }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("voice-profile:edit"),
                    handler(row: Record<string, any>) {
                        handleEdit(row as VoiceProfileDetailType);
                    }
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("voice-profile:remove"),
                    handler(row: Record<string, any>) {
                        handleDelete(row.id);
                    }
                }
            ])
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getVoiceProfileList,
        params: { ...filterData.value }
    };
});

function getFormConfig(isEdit: boolean) {
    return [
        formHelper.select("所属项目", "project_id", projectOptions.value, {
            rules: [ruleHelper.require("请选择项目")],
            disabled: isEdit
        }),
        formHelper.input("配置名称", "name", { rules: [ruleHelper.require("请输入名称")] }),
        formHelper.input("参考音频 URL", "reference_audio_url"),
        formHelper.input("参考文本", "reference_text")
    ];
}

function handleAdd() {
    ArcoModalFormShow({
        modalConfig: { title: "新增声音配置" },
        value: {},
        formConfig: getFormConfig(false),
        ok: async (data: any) => {
            await addVoiceProfile(data);
            Message.success("新增成功");
            table.value.refresh();
        }
    });
}

function handleEdit(row: VoiceProfileDetailType) {
    ArcoModalFormShow({
        modalConfig: { title: "编辑声音配置" },
        value: cloneDeep(row),
        formConfig: getFormConfig(true),
        ok: async (data: any) => {
            await updateVoiceProfile(data);
            Message.success("更新成功");
            table.value.refresh();
        }
    });
}

function handleDelete(id: number) {
    deleteVoiceProfile(id).then(() => {
        Message.success("删除成功");
        table.value.refresh();
    });
}

async function handleToggleStatus(record: VoiceProfileDetailType, enabled: boolean) {
    if (enabled) {
        await enableVoiceProfile(record.id);
    } else {
        await disableVoiceProfile(record.id);
    }
    table.value.refresh();
}
</script>
<style lang="scss" scoped></style>
