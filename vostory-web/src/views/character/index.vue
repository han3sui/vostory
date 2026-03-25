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

        <!-- 智能录入弹窗 -->
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
                粘贴角色介绍文字，LLM 将自动识别角色名称、性别、层级、性格描述等信息并录入角色库。已存在的角色会自动跳过。
            </a-typography-paragraph>
            <a-textarea
                v-model="smartImportText"
                placeholder="请粘贴角色介绍文字，例如：&#10;张三：男，主角，性格沉稳冷静，外表俊朗&#10;李四：女，配角，活泼开朗，是张三的青梅竹马"
                :auto-size="{ minRows: 8, maxRows: 16 }"
                allow-clear
            />
        </a-modal>

        <!-- 角色编辑弹窗 -->
        <a-modal
            v-model:visible="editModalVisible"
            :title="editForm.id ? '编辑角色' : '添加角色'"
            width="680px"
            :ok-loading="editSaving"
            @before-ok="handleEditSave"
        >
            <a-form :model="editForm" layout="vertical">
                <a-form-item label="角色名称" required>
                    <a-input v-model="editForm.name" placeholder="请输入角色名称" />
                </a-form-item>
                <a-row :gutter="16">
                    <a-col :span="12">
                        <a-form-item label="性别">
                            <a-select v-model="editForm.gender">
                                <a-option v-for="g in GENDERS" :key="g.value" :value="g.value">{{ g.label }}</a-option>
                            </a-select>
                        </a-form-item>
                    </a-col>
                    <a-col :span="12">
                        <a-form-item label="层级">
                            <a-select v-model="editForm.level">
                                <a-option v-for="l in LEVELS" :key="l.value" :value="l.value">{{ l.label }}</a-option>
                            </a-select>
                        </a-form-item>
                    </a-col>
                </a-row>
                <a-form-item label="声音配置">
                    <div class="voice-picker">
                        <div
                            v-if="selectedProfileName"
                            class="voice-picker__selected"
                            @click="showVoicePicker = true"
                        >
                            <icon-sound style="color: rgb(var(--primary-6))" />
                            <span class="voice-picker__name">{{ selectedProfileName }}</span>
                            <a-button type="text" size="mini" status="danger" @click.stop="clearVoiceProfile">
                                <template #icon><icon-close /></template>
                            </a-button>
                        </div>
                        <a-button v-else type="dashed" long @click="showVoicePicker = true">
                            <template #icon><icon-plus /></template>
                            选择声音配置
                        </a-button>
                    </div>
                </a-form-item>
                <a-form-item label="角色描述">
                    <a-textarea v-model="editForm.description" :auto-size="{ minRows: 2, maxRows: 6 }" />
                </a-form-item>
                <a-row :gutter="16">
                    <a-col :span="12">
                        <a-form-item label="状态">
                            <a-radio-group v-model="editForm.status">
                                <a-radio value="0">正常</a-radio>
                                <a-radio value="1">停用</a-radio>
                            </a-radio-group>
                        </a-form-item>
                    </a-col>
                    <a-col :span="12">
                        <a-form-item label="排序">
                            <a-input-number v-model="editForm.sort_order" :min="0" />
                        </a-form-item>
                    </a-col>
                </a-row>
            </a-form>
        </a-modal>

        <!-- 声音配置卡片选择弹窗 -->
        <a-modal
            v-model:visible="showVoicePicker"
            title="选择声音配置"
            :footer="false"
            width="700px"
        >
            <a-spin :loading="loadingProfiles" style="width: 100%">
                <a-empty v-if="!loadingProfiles && voiceProfileOptions.length === 0" description="当前项目暂无声音配置，请先在「声音配置」Tab 中添加" />
                <div v-else class="profile-grid">
                    <div
                        v-for="p in voiceProfileOptions"
                        :key="p.id"
                        class="profile-card"
                        :class="{ 'profile-card--active': editForm.voice_profile_id === p.id }"
                        @click="selectVoiceProfile(p)"
                    >
                        <div class="profile-card__name">
                            <icon-sound style="margin-right: 6px;" />
                            {{ p.name }}
                        </div>
                        <div v-if="editForm.voice_profile_id === p.id" class="profile-card__check">
                            <icon-check-circle-fill style="color: rgb(var(--primary-6)); font-size: 18px;" />
                        </div>
                    </div>
                </div>
            </a-spin>
        </a-modal>
    </div>
</template>
<script lang="ts" setup>
import { Message, Modal } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoForm } from "@easyfe/admin-component";
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

// 编辑弹窗状态
const editModalVisible = ref(false);
const editSaving = ref(false);
const editForm = ref<Record<string, any>>({});
const showVoicePicker = ref(false);
const loadingProfiles = ref(false);

const selectedProfileName = computed(() => {
    if (!editForm.value.voice_profile_id) return "";
    return voiceProfileOptions.value.find((p) => p.id === editForm.value.voice_profile_id)?.name || "已选择";
});

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

function onEdit(v: Record<string, any> | null) {
    if (v) {
        editForm.value = { ...v };
    } else {
        editForm.value = {
            project_id: props.projectId,
            status: "0",
            level: "main",
            gender: "unknown",
            aliases: []
        };
    }
    loadVoiceProfiles();
    editModalVisible.value = true;
}

function selectVoiceProfile(p: VoiceProfileOptionType) {
    editForm.value.voice_profile_id = p.id;
    showVoicePicker.value = false;
}

function clearVoiceProfile() {
    editForm.value.voice_profile_id = null;
}

async function handleEditSave(done: (closed: boolean) => void) {
    if (!editForm.value.name?.trim()) {
        Message.warning("请输入角色名称");
        done(false);
        return;
    }
    editSaving.value = true;
    try {
        if (editForm.value.id) {
            await updateCharacter(editForm.value);
        } else {
            await addCharacter(editForm.value);
        }
        Message.success(editForm.value.id ? "更新成功" : "添加成功");
        table.value.refresh();
        done(true);
    } catch {
        done(false);
    } finally {
        editSaving.value = false;
    }
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
        Message.success(`识别完成：发现 ${res.extracted_count} 个角色，新增 ${res.new_count} 个，跳过 ${res.skipped_count} 个`);
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
<style lang="scss" scoped>
.voice-picker {
    &__selected {
        display: flex;
        align-items: center;
        padding: 8px 12px;
        border: 1px solid rgb(var(--primary-6));
        border-radius: 6px;
        background: var(--color-primary-light-1);
        cursor: pointer;
        transition: all 0.2s;

        &:hover {
            background: var(--color-primary-light-2);
        }
    }

    &__name {
        flex: 1;
        margin-left: 8px;
        font-weight: 500;
    }
}

.profile-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
}

.profile-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px;
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
        border-color: rgb(var(--primary-6));
        background: var(--color-primary-light-1);
    }

    &--active {
        border-color: rgb(var(--primary-6));
        background: var(--color-primary-light-1);
    }

    &__name {
        display: flex;
        align-items: center;
        font-weight: 500;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    &__check {
        flex-shrink: 0;
        margin-left: 8px;
    }
}
</style>
