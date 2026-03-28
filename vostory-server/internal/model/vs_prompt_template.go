package model

// VsPromptTemplate Prompt 模板
type VsPromptTemplate struct {
	BaseModel
	TemplateID   uint64 `json:"template_id" gorm:"primaryKey;autoIncrement;comment:模板ID"`
	Name         string `json:"name" gorm:"size:100;not null;uniqueIndex:uk_template_type_name;comment:模板名称"`
	TemplateType string `json:"template_type" gorm:"size:30;not null;uniqueIndex:uk_template_type_name;comment:模板类型（character_extract/dialogue_parse/emotion_tag/scene_split/text_correct）"`
	Content      string `json:"content" gorm:"type:text;not null;comment:Prompt内容"`
	Description  string `json:"description" gorm:"size:500;comment:模板描述"`
	IsSystem     string `json:"is_system" gorm:"size:1;default:'0';comment:是否系统内置（1是 0否）"`
	Version      int    `json:"version" gorm:"default:1;comment:版本号"`
	SortOrder    int    `json:"sort_order" gorm:"default:0;comment:排序"`
	Status       string `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`
}

func (VsPromptTemplate) TableName() string {
	return "vs_prompt_template"
}

// PromptTemplateSeed 种子数据结构
type PromptTemplateSeed struct {
	Name         string
	TemplateType string
	Content      string
	Description  string
}

// DefaultPromptTemplateSeeds 系统内置 Prompt 模板种子数据（唯一数据源）
var DefaultPromptTemplateSeeds = []PromptTemplateSeed{
	{
		Name:         "默认角色抽取",
		TemplateType: "character_extract",
		Content: `你是一个专业的小说文本分析助手。请从以下小说文本中抽取所有出现的角色。

要求：
1. 识别所有有名字的角色（包括只出现一次的）
2. 不要把地名、物品名当作角色
3. 同一个角色的不同称呼要合并为一个角色

请严格以JSON格式返回，不要包含任何其他文字，不要使用markdown代码块包裹，结构如下：
{"characters":[{"name":"角色主要名称","aliases":["别名1","称呼2"],"gender":"male|female|unknown","level":"main|supporting|minor","description":"一句话角色描述"}]}

---
{{content}}`,
		Description: "从小说文本中自动抽取角色信息",
	},
	{
		Name:         "默认对白解析",
		TemplateType: "dialogue_parse",
		Content: `你是一个专业的小说文本分析助手。请将以下章节文本进行结构化切分。

要求：
1. 识别场景切换（基于时间跳跃、地点变化、视角切换）
2. 在每个场景内，将文本切分为独立片段
3. 每个片段标注类型：dialogue(对白)、narration(旁白)、monologue(独白)、description(描述)
4. 对白和独白片段需识别说话人名称，并根据上下文推断该角色的性别和固有特征描述
5. character_description只描述角色固有特征（身份、外貌、性格、年龄等），不要包含当前情绪状态，情绪由emotion字段单独表达
6. 标注每个片段的情绪：neutral/happy/sad/angry/fear/surprise/disgust
7. 标注情绪强度：light/medium/strong
8. content字段中的特殊字符（引号、反斜杠、换行等）必须按照JSON标准进行转义

请严格以JSON格式返回，不要包含任何其他文字，不要使用markdown代码块包裹，结构如下：
{"scenes":[{"title":"场景标题","description":"场景简述","segments":[{"type":"dialogue|narration|monologue|description","content":"片段文本内容","character":"说话人名称（非对白/独白时为空字符串）","character_gender":"male|female|unknown（根据上下文推断，非对白/独白时为空字符串）","character_description":"角色固有特征描述，如身份、外貌、性格、年龄等，不要包含当前情绪状态（非对白/独白时为空字符串）","emotion":"neutral|happy|sad|angry|fear|surprise|disgust","emotion_strength":"light|medium|strong"}]}]}

---
{{content}}`,
		Description: "将章节文本按场景和片段进行结构化切分，识别类型、说话人和情绪",
	},
	{
		Name:         "默认情绪标注",
		TemplateType: "emotion_tag",
		Content: `请为以下对白/独白片段标注情绪。对于每个片段，请提供：
1. 情绪类型：happy/sad/angry/fear/surprise/neutral/disgust/contempt
2. 情绪强度：light/medium/strong

请严格以JSON数组格式返回结果，不要包含任何其他文字，不要使用markdown代码块包裹。

---
{{segments}}`,
		Description: "为脚本片段自动标注情绪类型和强度",
	},
	{
		Name:         "默认场景切分",
		TemplateType: "scene_split",
		Content: `请将以下章节文本按场景进行切分。场景切换的依据包括：
1. 时间跳跃
2. 地点变化
3. 视角切换
4. 明显的叙事断裂

对于每个场景，请提供：
1. 场景标题（简要概括）
2. 场景描述
3. 场景包含的文本范围（起始和结束位置）

请严格以JSON数组格式返回结果，不要包含任何其他文字，不要使用markdown代码块包裹。

---
{{content}}`,
		Description: "将章节文本按场景自动切分",
	},
	{
		Name:         "默认文本校正",
		TemplateType: "text_correct",
		Content: `请对以下文本进行校正，确保：
1. 不丢失任何原文内容
2. 不添加原文没有的内容
3. 修正明显的错别字
4. 统一标点符号格式

请返回校正后的完整文本。

---
{{content}}`,
		Description: "精准填充 - 确保LLM输出对齐回原文",
	},
	{
		Name:         "默认声音匹配",
		TemplateType: "voice_match",
		Content: `你是一个专业的有声书制作助手。请根据角色信息和声音信息，为每个角色匹配最合适的声音配置。

匹配规则（按优先级排序）：
1. 性别必须一致：male角色只能匹配male声音，female角色只能匹配female声音，unknown性别的角色可匹配任意声音
2. 综合参考声音的名称和描述进行语义匹配：
   - 声音名称通常蕴含音色特征（如"温柔少女""沧桑大叔""少年音"等），是重要的匹配依据
   - 声音描述提供更详细的音色特征，若存在则优先参考
   - 当描述为空时，完全依据声音名称进行匹配
3. 角色匹配时同样综合角色名称和描述：
   - 角色描述包含年龄、性格、身份等信息，是匹配的核心依据
   - 当角色描述为空时，根据角色名称推断其可能的特征（如"老王"暗示中老年男性）
4. 一个声音可以被多个角色共用
5. 如果没有合适的声音，voice_profile_id返回null

角色列表：
{{characters}}

声音配置列表：
{{voices}}

请严格以JSON格式返回，不要包含任何其他文字，不要使用markdown代码块包裹，结构如下：
{"matches":[{"character_id":1,"voice_profile_id":2,"reason":"简要匹配理由"}]}`,
		Description: "根据角色信息与声音信息（名称+描述）自动匹配最合适的声音配置",
	},
}
