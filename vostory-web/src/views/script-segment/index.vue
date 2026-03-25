<template>
    <frame-view>
        <arco-table ref="table" :req="getData" :table-config="tableConfig">
            <template #tlBtns>
                <arco-form v-model="filterData" :config="getFilterConfig" layout="row"></arco-form>
            </template>
            <template #typeSlot>
                <a-table-column title="类型">
                    <template #cell="{ record }">
                        <a-tag :color="typeColor(record.segment_type)" size="small">
                            {{ typeLabel(record.segment_type) }}
                        </a-tag>
                    </template>
                </a-table-column>
            </template>
            <template #emotionSlot>
                <a-table-column title="情绪">
                    <template #cell="{ record }">
                        <span v-if="record.emotion_type">
                            {{ record.emotion_type }}
                            <a-tag size="small" color="arcoblue">{{ record.emotion_strength }}</a-tag>
                        </span>
                        <span v-else class="text-gray-400">—</span>
                    </template>
                </a-table-column>
            </template>
        </arco-table>
    </frame-view>
</template>
<script lang="ts" setup>
import { Modal } from "@arco-design/web-vue";
import { formHelper, ArcoTable, tableHelper, ArcoModalFormShow, ruleHelper, ArcoForm } from "@easyfe/admin-component";
import { getScriptSegmentList, updateScriptSegment, deleteScriptSegment } from "@/config/apis/script-segment";
import { cloneDeep } from "lodash-es";
import { hasPermission, PageTableConfig } from "@/views/utils";

const TYPE_MAP: Record<string, { label: string; color: string }> = {
    dialogue: { label: "对白", color: "blue" },
    narration: { label: "旁白", color: "gray" },
    monologue: { label: "独白", color: "cyan" },
    description: { label: "描述", color: "orangered" }
};

const SEGMENT_TYPES = Object.entries(TYPE_MAP).map(([k, v]) => ({ label: v.label, value: k }));
const EMOTION_TYPES = [
    { label: "开心", value: "happy" },
    { label: "悲伤", value: "sad" },
    { label: "愤怒", value: "angry" },
    { label: "恐惧", value: "fear" },
    { label: "惊讶", value: "surprise" },
    { label: "中性", value: "neutral" }
];
const EMOTION_STRENGTHS = [
    { label: "轻", value: "light" },
    { label: "中", value: "medium" },
    { label: "强", value: "strong" }
];

function typeLabel(t: string) {
    return TYPE_MAP[t]?.label || t;
}
function typeColor(t: string) {
    return TYPE_MAP[t]?.color || "gray";
}

const table = ref();
const filterData = ref<Record<string, any>>({});

const getFilterConfig = computed(() => {
    return [
        formHelper.input("章节ID", "chapter_id", { span: 4, debounce: 500 }),
        formHelper.select("片段类型", "segment_type", SEGMENT_TYPES, { span: 5 }),
        formHelper.select(
            "状态",
            "status",
            [
                { label: "原始", value: "raw" },
                { label: "已编辑", value: "edited" },
                { label: "已生成", value: "generated" }
            ],
            { span: 5 }
        )
    ];
});

const tableConfig = computed(() => {
    return tableHelper.create({
        arcoProps: { rowKey: "id" },
        ...PageTableConfig,
        showRefresh: true,
        maxHeight: "auto",
        columns: [
            tableHelper.default("序号", "segment_num"),
            tableHelper.slot("typeSlot"),
            tableHelper.default("内容", "content"),
            tableHelper.default("角色", "character_name"),
            tableHelper.slot("emotionSlot"),
            tableHelper.default("版本", "version"),
            tableHelper.btns("操作", [
                {
                    label: "编辑",
                    if: () => hasPermission("script-segment:edit"),
                    handler: onEdit
                },
                {
                    label: "删除",
                    status: "danger",
                    if: () => hasPermission("script-segment:remove"),
                    handler(row: Record<string, any>) {
                        Modal.confirm({
                            title: "删除",
                            content: "确认删除该片段？",
                            onBeforeOk: async () => {
                                await deleteScriptSegment(row.id);
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
        fn: getScriptSegmentList,
        params: { ...filterData.value }
    };
});

function onEdit(v: Record<string, any> | null) {
    const tempValue = cloneDeep(v);
    if (!tempValue) return;
    ArcoModalFormShow({
        modalConfig: { title: "编辑脚本片段", width: "700px" },
        value: tempValue,
        formConfig: [
            formHelper.select("片段类型", "segment_type", SEGMENT_TYPES, {
                rules: [ruleHelper.require("请选择类型")]
            }),
            formHelper.textarea("片段内容", "content", {
                rules: [ruleHelper.require("请输入内容")]
            }),
            formHelper.select("情绪类型", "emotion_type", EMOTION_TYPES),
            formHelper.select("情绪强度", "emotion_strength", EMOTION_STRENGTHS),
            formHelper.inputNumber("片段序号", "segment_num")
        ],
        ok: async (data: any) => {
            await updateScriptSegment(data);
            table.value.refresh();
        }
    });
}
</script>
<style lang="scss" scoped></style>
