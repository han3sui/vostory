package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsPromptTemplateService interface {
	Create(ctx context.Context, request *v1.VsPromptTemplateCreateRequest) error
	Update(ctx context.Context, request *v1.VsPromptTemplateUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsPromptTemplateDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsPromptTemplateListQuery) ([]*v1.VsPromptTemplateDetailResponse, int64, error)
	FindByType(ctx context.Context, templateType string) ([]*v1.VsPromptTemplateOptionResponse, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
	SeedDefaults(ctx context.Context) error
}

func NewVsPromptTemplateService(
	service *Service,
	repo repository.VsPromptTemplateRepository,
) VsPromptTemplateService {
	return &vsPromptTemplateService{
		Service: service,
		repo:    repo,
	}
}

type vsPromptTemplateService struct {
	*Service
	repo repository.VsPromptTemplateRepository
}

func (s *vsPromptTemplateService) Create(ctx context.Context, request *v1.VsPromptTemplateCreateRequest) error {
	id, err := s.sid.GenUint64()
	if err != nil {
		return fmt.Errorf("生成ID失败: %w", err)
	}

	template := &model.VsPromptTemplate{
		TemplateID:   id,
		Name:         request.Name,
		TemplateType: request.TemplateType,
		Content:      request.Content,
		Description:  request.Description,
		IsSystem:     "0",
		Version:      1,
		SortOrder:    request.SortOrder,
		Status:       request.Status,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
		},
	}

	return s.repo.Create(ctx, template)
}

func (s *vsPromptTemplateService) Update(ctx context.Context, request *v1.VsPromptTemplateUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("模板不存在")
	}

	existing.Name = request.Name
	existing.TemplateType = request.TemplateType
	existing.Content = request.Content
	existing.Description = request.Description
	existing.SortOrder = request.SortOrder
	existing.Status = request.Status
	existing.Version = existing.Version + 1
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsPromptTemplateService) Delete(ctx context.Context, id uint64) error {
	template, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("模板不存在")
	}
	if template.IsSystem == "1" {
		return fmt.Errorf("系统内置模板不允许删除")
	}
	return s.repo.Delete(ctx, id)
}

func (s *vsPromptTemplateService) FindByID(ctx context.Context, id uint64) (*v1.VsPromptTemplateDetailResponse, error) {
	template, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(template), nil
}

func (s *vsPromptTemplateService) FindWithPagination(ctx context.Context, query *v1.VsPromptTemplateListQuery) ([]*v1.VsPromptTemplateDetailResponse, int64, error) {
	templates, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsPromptTemplateDetailResponse
	for _, t := range templates {
		responses = append(responses, s.convertToDetailResponse(t))
	}
	return responses, total, nil
}

func (s *vsPromptTemplateService) FindByType(ctx context.Context, templateType string) ([]*v1.VsPromptTemplateOptionResponse, error) {
	templates, err := s.repo.FindByType(ctx, templateType)
	if err != nil {
		return nil, err
	}

	var responses []*v1.VsPromptTemplateOptionResponse
	for _, t := range templates {
		responses = append(responses, &v1.VsPromptTemplateOptionResponse{
			ID:           t.TemplateID,
			Name:         t.Name,
			TemplateType: t.TemplateType,
		})
	}
	return responses, nil
}

func (s *vsPromptTemplateService) Enable(ctx context.Context, id uint64) error {
	return s.repo.Enable(ctx, id)
}

func (s *vsPromptTemplateService) Disable(ctx context.Context, id uint64) error {
	return s.repo.Disable(ctx, id)
}

func (s *vsPromptTemplateService) SeedDefaults(ctx context.Context) error {
	defaults := []struct {
		Name         string
		TemplateType string
		Content      string
		Description  string
	}{
		{
			Name:         "默认角色抽取",
			TemplateType: "character_extract",
			Content:      "请从以下小说文本中抽取所有出现的角色。对于每个角色，请提供：\n1. 角色名称\n2. 可能的别名/称呼\n3. 性别（male/female/unknown）\n4. 角色层级（main=主角/supporting=配角/minor=路人）\n5. 简要描述\n\n请以JSON数组格式返回结果。\n\n---\n{{content}}",
			Description:  "从小说文本中自动抽取角色信息",
		},
		{
			Name:         "默认对白解析",
			TemplateType: "dialogue_parse",
			Content:      "请将以下小说文本拆分为脚本片段。对于每个片段，请标注：\n1. 片段类型：dialogue（对白）、narration（旁白）、monologue（内心独白）、description（场景描述）\n2. 说话人（对白和独白需要标注）\n3. 片段文本内容\n\n请以JSON数组格式返回结果，保持原文顺序。\n\n---\n{{content}}",
			Description:  "将小说文本拆分为对白、旁白等脚本片段",
		},
		{
			Name:         "默认情绪标注",
			TemplateType: "emotion_tag",
			Content:      "请为以下对白/独白片段标注情绪。对于每个片段，请提供：\n1. 情绪类型：happy/sad/angry/fear/surprise/neutral/disgust/contempt\n2. 情绪强度：light/medium/strong\n\n请以JSON数组格式返回结果。\n\n---\n{{segments}}",
			Description:  "为脚本片段自动标注情绪类型和强度",
		},
		{
			Name:         "默认场景切分",
			TemplateType: "scene_split",
			Content:      "请将以下章节文本按场景进行切分。场景切换的依据包括：\n1. 时间跳跃\n2. 地点变化\n3. 视角切换\n4. 明显的叙事断裂\n\n对于每个场景，请提供：\n1. 场景标题（简要概括）\n2. 场景描述\n3. 场景包含的文本范围（起始和结束位置）\n\n请以JSON数组格式返回结果。\n\n---\n{{content}}",
			Description:  "将章节文本按场景自动切分",
		},
		{
			Name:         "默认文本校正",
			TemplateType: "text_correct",
			Content:      "请对以下文本进行校正，确保：\n1. 不丢失任何原文内容\n2. 不添加原文没有的内容\n3. 修正明显的错别字\n4. 统一标点符号格式\n\n请返回校正后的完整文本。\n\n---\n{{content}}",
			Description:  "精准填充 - 确保LLM输出对齐回原文",
		},
	}

	for _, d := range defaults {
		count, err := s.repo.CountByType(ctx, d.TemplateType)
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		id, err := s.sid.GenUint64()
		if err != nil {
			return fmt.Errorf("生成ID失败: %w", err)
		}

		template := &model.VsPromptTemplate{
			TemplateID:   id,
			Name:         d.Name,
			TemplateType: d.TemplateType,
			Content:      d.Content,
			Description:  d.Description,
			IsSystem:     "1",
			Version:      1,
			SortOrder:    0,
			Status:       "0",
			BaseModel: model.BaseModel{
				CreatedBy: "system",
			},
		}

		if err := s.repo.Create(ctx, template); err != nil {
			return fmt.Errorf("创建默认模板[%s]失败: %w", d.Name, err)
		}
	}

	return nil
}

func (s *vsPromptTemplateService) convertToDetailResponse(t *model.VsPromptTemplate) *v1.VsPromptTemplateDetailResponse {
	return &v1.VsPromptTemplateDetailResponse{
		ID:           t.TemplateID,
		Name:         t.Name,
		TemplateType: t.TemplateType,
		Content:      t.Content,
		Description:  t.Description,
		IsSystem:     t.IsSystem,
		Version:      t.Version,
		SortOrder:    t.SortOrder,
		Status:       t.Status,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}
