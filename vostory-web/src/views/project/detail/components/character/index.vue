<template>
    <div>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #switchSlot>
                <a-table-column title="启用">
                    <template #cell="{ record }">
                        <a-switch
                            :disabled="!hasPermission('character:enable')"
                            :default-checked="record.status === '0'"
                            :before-change="() => handleToggle(record)"
                        ></a-switch>
                    </template>
                </a-table-column>
            </template>
            <template #levelSlot>
                <a-table-column title="层级">
                    <template #cell="{ record }">
                        <a-tag :color="levelColor(record.level)" size="small">
                            {{ levelLabel(record.level) }}
                        </a-tag>
                    </template>
                </a-table-column>
            </template>
            <template #aliasSlot>
                <a-table-column title="别名">
                    <template #cell="{ record }">
                        <a-space wrap>
                            <a-tag v-for="a in record.aliases || []" :key="a" size="small">{{ a }}</a-tag>
                        </a-space>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
        <a-modal
            v-model:visible="showSmartImportModal"
            title="智能录入角色"
            :ok-loading="smartImporting"
            ok-text="开始识别"
            :ok-button-props="{ disabled: !smartImportText.trim() }"
            width="680px"
            @before-ok="handleSmartImport"
        >
            <a-typography-paragraph>
                粘贴角色介绍文字，LLM
                将自动识别角色名称、性别、层级、性格描述等信息并录入角色库。已存在的角色会自动跳过。
            </a-typography-paragraph>
            <a-textarea
                v-model="smartImportText"
                placeholder="请粘贴角色介绍文字，例如：&#10;张三：男，主角，性格沉稳冷静，外表俊朗&#10;李四：女，配角，活泼开朗，是张三的青梅竹马"
                :auto-size="{ minRows: 8, maxRows: 16 }"
                allow-clear
            />
        </a-modal>
    </div>
</template>
<script lang="ts" setup>
import { Message, Modal } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import {
    getCharacterList,
    addCharacter,
    updateCharacter,
    deleteCharacter,
    enableCharacter,
    disableCharacter,
    extractCharacters,
    extractFromText,
    CharacterDetailType
} from "@/config/apis/character";
import { getVoiceProfilesByProject, VoiceProfileOptionType } from "@/config/apis/voice-profile";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const props = defineProps<{ projectId: number }>();

const LEVELS = [
    { label: "主角", value: "main" },
    { label: "配角", value: "supporting" },
    { label: "龙套", value: "minor" }
];
const GENDERS = [
    { label: "男", value: "male" },
    { label: "女", value: "female" },
    { label: "未知", value: "unknown" }
];

function levelLabel(l: string) {
    return LEVELS.find((v) => v.value === l)?.label || l;
}
function levelColor(l: string) {
    const map: Record<string, string> = { main: "red", supporting: "blue", minor: "gray" };
    return map[l] || "gray";
}

const table = ref();
const filterData = ref<Record<string, any>>({});
const extracting = ref(false);
const showSmartImportModal = ref(false);
const smartImportText = ref("");
const smartImporting = ref(false);
const voiceProfileOptions = ref<VoiceProfileOptionType[]>([]);

async function loadVoiceProfiles() {
    if (!props.projectId) return;
    voiceProfileOptions.value = await getVoiceProfilesByProject(props.projectId);
}
onMounted(loadVoiceProfiles);
watch(() => props.projectId, loadVoiceProfiles);

const getFilterConfig = computed(() => {
    return [
        formHelper.input("角色名称", "name", { span: 5, debounce: 500 }),
        formHelper.select("层级", "level", LEVELS, { span: 4 }),
        formHelper.select("性别", "gender", GENDERS, { span: 4 })
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: { rowKey: "id" },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("角色名称", "name"),
            tableHelper.slot("aliasSlot"),
            tableHelper.status("性别", "gender", (item: any) => {
                const found = GENDERS.find((g) => g.value === item.gender);
                return { text: found?.label || item.gender, status: "normal" };
            }),
            tableHelper.slot("levelSlot"),
            tableHelper.default("描述", "description"),
            tableHelper.default("声音配置", "voice_profile_name"),
            tableHelper.default("排序", "sort_order"),
            tableHelper.slot("switchSlot"),
            tableHelper.date("更新时间", "updated_at", { format: "YYYY-MM-DD HH:mm:ss" }),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("character:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("character:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: `确认删除角色【${row.name}】？`,
                            onBeforeOk: async () => {
                                await deleteCharacter(row.id);
                                table.value.refresh();
                            }
                        });
                    }
                }
            ])
        ],
        trBtns: [
            {
                label: "智能录入",
                type: "outline",
                loading: smartImporting.value,
                if: () => hasPermission("character:add"),
                handler: () => {
                    smartImportText.value = "";
                    showSmartImportModal.value = true;
                }
            },
            {
                label: "添加角色",
                type: "primary",
                if: () => hasPermission("character:add"),
                handler: () => onEdit(null)
            }
        ]
    });
});

const getData = computed(() => {
    return {
        fn: getCharacterList,
        params: { project_id: props.projectId, ...filterData.value }
    };
});

async function onEdit(v: Record<string, any> | null) {
    await loadVoiceProfiles();
    const tempValue = cloneDeep(v);
    ArcoModalFormShow({
        modalConfig: {
            title: tempValue ? "编辑角色" : "添加角色",
            width: "650px"
        },
        value: tempValue || { project_id: props.projectId, status: "0", level: "main", gender: "unknown", aliases: [] },
        formConfig: [
            formHelper.input("角色名称", "name", {
                rules: [ruleHelper.require("请输入角色名称")]
            }),
            formHelper.select("性别", "gender", GENDERS),
            formHelper.select("层级", "level", LEVELS),
            formHelper.select(
                "声音配置",
                "voice_profile_id",
                voiceProfileOptions.value.map((v) => ({ label: v.name, value: v.id })),
                { allowClear: true, placeholder: "选择声音配置（用于 TTS 合成）" }
            ),
            formHelper.textarea("角色描述", "description"),
            formHelper.radio(
                "状态",
                "status",
                [
                    { label: "正常", value: "0" },
                    { label: "停用", value: "1" }
                ],
                { type: "radio", rules: [ruleHelper.require("请选择")] }
            ),
            formHelper.inputNumber("排序", "sort_order")
        ],
        ok: async (data: any) => {
            if (tempValue) {
                await updateCharacter(data);
            } else {
                await addCharacter(data);
            }
            table.value.refresh();
        }
    });
}

async function handleExtract() {
    Modal.confirm({
        title: "智能提取角色",
        content: "将使用 LLM 从全书文本中自动识别角色，已存在的角色会自动跳过。是否继续？",
        okText: "开始提取",
        cancelText: "取消",
        onBeforeOk: async () => {
            extracting.value = true;
            try {
                const res = await extractCharacters(props.projectId);
                Message.success(
                    `提取完成：识别 ${res.extracted_count} 个角色，新增 ${res.new_count} 个，跳过 ${res.skipped_count} 个`
                );
                table.value.refresh();
            } finally {
                extracting.value = false;
            }
        }
    });
}

async function handleSmartImport(done: (closed: boolean) => void) {
    if (!smartImportText.value.trim()) {
        Message.warning("请输入角色描述文字");
        done(false);
        return;
    }
    smartImporting.value = true;
    try {
        const res = await extractFromText(props.projectId, smartImportText.value);
        Message.success(
            `识别完成：发现 ${res.extracted_count} 个角色，新增 ${res.new_count} 个，跳过 ${res.skipped_count} 个`
        );
        table.value.refresh();
        done(true);
    } catch {
        done(false);
    } finally {
        smartImporting.value = false;
    }
}

async function handleToggle(row: CharacterDetailType) {
    try {
        if (row.status === "0") {
            await disableCharacter(row.id);
        } else {
            await enableCharacter(row.id);
        }
        table.value.refresh();
        return true;
    } catch {
        return false;
    }
}
</script>
<style lang="scss" scoped></style>
