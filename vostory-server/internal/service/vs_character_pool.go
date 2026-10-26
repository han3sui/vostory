package service

import (
	"strings"
	"unicode/utf8"

	"iot-alert-center/internal/model"
)

// poolRole 路人池角色：无名/泛称角色统一归入，按性别与年龄段共享音色
type poolRole struct {
	Name        string
	Gender      string
	Description string
}

const narratorName = "旁白"

// characterPool 路人池定义，SortOrder 从 poolSortOrderBase 开始依次排列
var characterPool = []poolRole{
	{Name: "路人-青年男", Gender: "male", Description: "路人池：无名或泛称的青年男性（少年、男子、伙计、侍卫等）"},
	{Name: "路人-青年女", Gender: "female", Description: "路人池：无名或泛称的青年女性（少女、姑娘、丫鬟、侍女等）"},
	{Name: "路人-中年男", Gender: "male", Description: "路人池：无名或泛称的中年男性（大叔、掌柜、汉子等）"},
	{Name: "路人-中年女", Gender: "female", Description: "路人池：无名或泛称的中年女性（妇人、大婶、老板娘等）"},
	{Name: "路人-老年男", Gender: "male", Description: "路人池：无名或泛称的老年男性（老者、老头、老翁等）"},
	{Name: "路人-老年女", Gender: "female", Description: "路人池：无名或泛称的老年女性（老妇、老妪、老太太等）"},
	{Name: "路人-孩童", Gender: "unknown", Description: "路人池：无名或泛称的孩童（小孩、童子等）"},
	{Name: "路人-群体", Gender: "unknown", Description: "路人池：多人齐声或群体发言（众人、众弟子、人群等）"},
}

const poolSortOrderBase = 9000

// 泛称 → 年龄段分类。分类为 "young"/"mid"/"old" 时再按性别选择男/女池
var genericTermCategory = map[string]string{
	// 老年
	"老者": "old_m", "老头": "old_m", "老头子": "old_m", "老翁": "old_m", "老汉": "old_m", "老叟": "old_m", "老僧": "old_m", "老和尚": "old_m",
	"老人": "old", "老人家": "old",
	"老妇": "old_f", "老妇人": "old_f", "老妪": "old_f", "老太": "old_f", "老太太": "old_f", "老奶奶": "old_f", "老婆子": "old_f",
	// 中年
	"中年人": "mid", "中年男子": "mid_m", "大叔": "mid_m", "汉子": "mid_m", "大汉": "mid_m", "壮汉": "mid_m", "掌柜": "mid_m", "掌柜的": "mid_m",
	"中年妇人": "mid_f", "妇人": "mid_f", "大婶": "mid_f", "大娘": "mid_f", "老板娘": "mid_f", "妇女": "mid_f",
	// 青年
	"少年": "young_m", "青年": "young_m", "男子": "young_m", "男人": "young_m", "书生": "young_m", "小厮": "young_m",
	"伙计": "young_m", "店小二": "young_m", "小二": "young_m", "侍卫": "young_m", "护卫": "young_m", "守卫": "young_m",
	"卫兵": "young_m", "士兵": "young_m", "官兵": "young_m", "衙役": "young_m", "士卒": "young_m", "军士": "young_m",
	"差役": "young_m", "书童": "young_m", "甲士": "young_m", "学子": "young_m", "酒客": "young_m",
	"管家": "mid_m", "官员": "mid_m", "大臣": "mid_m", "长老": "old_m", "观众": "young",
	"少女": "young_f", "女子": "young_f", "女人": "young_f", "姑娘": "young_f", "小姑娘": "young_f", "丫鬟": "young_f",
	"侍女": "young_f", "婢女": "young_f", "女弟子": "young_f",
	"年轻人": "young", "弟子": "young", "黑衣人": "young", "蒙面人": "young", "刺客": "young", "路人": "young",
	"旁人": "young", "某人": "young", "那人": "young", "来人": "young", "有人": "young", "陌生人": "young",
	"神秘人": "young", "声音": "young", "侍从": "young", "下人": "young", "仆人": "young", "村民": "young",
	"百姓": "young", "客人": "young", "食客": "young",
	// 孩童
	"孩子": "child", "小孩": "child", "孩童": "child", "童子": "child", "小童": "child", "小男孩": "child",
	"小女孩": "child", "男童": "child", "女童": "child",
	// 群体
	"众人": "group", "大家": "group", "众弟子": "group", "人群": "group", "群众": "group", "围观者": "group",
	"围观群众": "group", "众修士": "group", "众将士": "group", "两人": "group", "二人": "group", "几人": "group",
}

// 泛称前缀修饰词，按长度从长到短匹配
var genericPrefixes = []string{"那一个", "一个", "一名", "一位", "那个", "这个", "那位", "这位", "那名", "这名", "某个", "某位", "某", "那", "这"}

func isPoolCharacterName(name string) bool {
	for _, p := range characterPool {
		if p.Name == name {
			return true
		}
	}
	return false
}

func findPoolRole(name string) (poolRole, bool) {
	for _, p := range characterPool {
		if p.Name == name {
			return p, true
		}
	}
	return poolRole{}, false
}

// resolveGenericPoolName 判断角色名是否为泛称，是则返回应归入的路人池名称。
// 仅做整词匹配（允许去掉"那个/一名/某"等前缀与"们"后缀），避免误伤真实人名。
func resolveGenericPoolName(name, gender string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}
	if isPoolCharacterName(name) {
		return name, true
	}

	candidates := []string{name}
	for _, prefix := range genericPrefixes {
		if rest, ok := strings.CutPrefix(name, prefix); ok {
			candidates = append(candidates, rest)
		}
	}

	for _, cand := range candidates {
		isPlural := false
		if strings.HasSuffix(cand, "们") && utf8.RuneCountInString(cand) > 1 {
			cand = strings.TrimSuffix(cand, "们")
			isPlural = true
		}
		category, ok := genericTermCategory[cand]
		if !ok {
			continue
		}
		if isPlural {
			category = "group"
		}
		return poolNameByCategory(category, normalizeGender(gender)), true
	}
	return "", false
}

func poolNameByCategory(category, gender string) string {
	switch category {
	case "old_m":
		return "路人-老年男"
	case "old_f":
		return "路人-老年女"
	case "mid_m":
		return "路人-中年男"
	case "mid_f":
		return "路人-中年女"
	case "young_m":
		return "路人-青年男"
	case "young_f":
		return "路人-青年女"
	case "child":
		return "路人-孩童"
	case "group":
		return "路人-群体"
	}

	female := gender == "female"
	switch category {
	case "old":
		if female {
			return "路人-老年女"
		}
		return "路人-老年男"
	case "mid":
		if female {
			return "路人-中年女"
		}
		return "路人-中年男"
	default:
		if female {
			return "路人-青年女"
		}
		return "路人-青年男"
	}
}

func poolSortOrder(name string) int {
	for i, p := range characterPool {
		if p.Name == name {
			return poolSortOrderBase + i
		}
	}
	return poolSortOrderBase
}

func newPoolCharacter(projectID uint64, name, loginName string, deptID uint) *model.VsCharacter {
	role, _ := findPoolRole(name)
	return &model.VsCharacter{
		ProjectID:   projectID,
		Name:        role.Name,
		Gender:      role.Gender,
		Description: role.Description,
		Level:       "minor",
		Status:      "0",
		SortOrder:   poolSortOrder(name),
		BaseModel: model.BaseModel{
			CreatedBy: loginName,
			DeptID:    deptID,
		},
	}
}

const (
	maxKnownCharacters       = 300
	maxKnownCharacterDescLen = 30
)

// buildKnownCharactersBlock 构造注入切分 prompt 的「已知角色表 + 路人池」说明。
// 主角/配角全量注入；路人级角色只注入名字或别名在本章出现过的，控制 prompt 长度。
func buildKnownCharactersBlock(characters []*model.VsCharacter, chapterContent string) string {
	var primary, hits []*model.VsCharacter
	for _, c := range characters {
		if c.Name == narratorName || isPoolCharacterName(c.Name) {
			continue
		}
		if c.Level == "main" || c.Level == "supporting" {
			primary = append(primary, c)
			continue
		}
		if characterMentioned(c, chapterContent) {
			hits = append(hits, c)
		}
	}
	selected := append(primary, hits...)
	if len(selected) > maxKnownCharacters {
		selected = selected[:maxKnownCharacters]
	}

	var sb strings.Builder
	sb.WriteString("【已知角色表】\n")
	sb.WriteString("以下是本书已识别的角色。character 字段必须优先使用表中的「标准名」；文中出现的昵称、尊称、代称若指向表中角色，也输出其标准名，不要另起新名字：\n")
	if len(selected) == 0 {
		sb.WriteString("（暂无）\n")
	}
	for _, c := range selected {
		sb.WriteString("- ")
		sb.WriteString(c.Name)
		if len(c.Aliases) > 0 {
			sb.WriteString("｜别名：")
			sb.WriteString(strings.Join(c.Aliases, "、"))
		}
		if c.Gender != "" && c.Gender != "unknown" {
			sb.WriteString("｜性别：")
			sb.WriteString(c.Gender)
		}
		if desc := strings.TrimSpace(c.Description); desc != "" {
			sb.WriteString("｜")
			sb.WriteString(truncate(desc, maxKnownCharacterDescLen))
		}
		sb.WriteString("\n")
	}

	poolNames := make([]string, len(characterPool))
	for i, p := range characterPool {
		poolNames[i] = p.Name
	}
	sb.WriteString("\n【路人池】\n")
	sb.WriteString("没有名字、只有泛称的角色（如老者、店小二、黑衣人、某弟子、众人）不要作为新角色，character 字段从下列路人池中选择最接近的一个：\n")
	sb.WriteString(strings.Join(poolNames, "、"))
	sb.WriteString("\n（「路人-群体」用于多人齐声或群体发言）\n")
	return sb.String()
}

func characterMentioned(c *model.VsCharacter, content string) bool {
	if c.Name != "" && strings.Contains(content, c.Name) {
		return true
	}
	for _, alias := range c.Aliases {
		if alias != "" && strings.Contains(content, alias) {
			return true
		}
	}
	return false
}

const maxInstructionRunes = 40

// instructionRequirement 旧版切分模板没有 instruction 字段时追加的输出要求
const instructionRequirement = `每个片段额外输出 instruction 字段：结合场景上下文，用一句不超过20字的中文描述这句话该怎么演绎（语气、语速、音量、停顿等，如"压低声音，急促地"、"冷笑着，慢条斯理"），不要描述音色和性别；平铺直叙的旁白可为空字符串。`

// injectKnownCharacters 将已知角色表写入 prompt：模板含 {{known_characters}} 占位符时就地替换，
// 否则（旧版自定义模板）追加到末尾作为补充要求；旧模板未要求 instruction 字段时一并追加。
// 必须在替换 {{content}} 之前调用。
func injectKnownCharacters(promptTemplate, block string) string {
	var extra []string
	if strings.Contains(promptTemplate, "{{known_characters}}") {
		promptTemplate = strings.ReplaceAll(promptTemplate, "{{known_characters}}", block)
	} else {
		extra = append(extra, "【角色识别】\n"+block)
	}
	if !strings.Contains(promptTemplate, "instruction") {
		extra = append(extra, "【演绎指令】\n"+instructionRequirement)
	}
	if len(extra) == 0 {
		return promptTemplate
	}
	return promptTemplate + "\n\n---\n补充要求：\n" + strings.Join(extra, "\n\n")
}
