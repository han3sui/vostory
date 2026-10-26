// Package llmconfig 把项目绑定的模型解析成可直接调用的厂家凭证 + 模型标识。
package llmconfig

import (
	"context"
	"fmt"

	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

// Resolved 一次解析结果：厂家凭证 + 模型名
type Resolved struct {
	ProviderID uint64
	BaseURL    string
	APIKey     string
	ModelName  string
}

// Resolver 按项目绑定的模型解析厂家
type Resolver interface {
	ResolveByProject(ctx context.Context, project *model.VsProject) (*Resolved, error)
}

func NewResolver(
	providerRepo repository.VsLLMProviderRepository,
	modelRepo repository.VsLLMModelRepository,
) Resolver {
	return &resolver{
		providerRepo: providerRepo,
		modelRepo:    modelRepo,
	}
}

type resolver struct {
	providerRepo repository.VsLLMProviderRepository
	modelRepo    repository.VsLLMModelRepository
}

func (r *resolver) ResolveByProject(ctx context.Context, project *model.VsProject) (*Resolved, error) {
	if project == nil || project.LLMModelID == nil || *project.LLMModelID == 0 {
		return nil, fmt.Errorf("项目未配置 LLM 模型，请先在项目设置中绑定")
	}

	m, err := r.modelRepo.FindByID(ctx, *project.LLMModelID)
	if err != nil {
		return nil, fmt.Errorf("LLM 模型不存在")
	}
	if m.Status != "0" {
		return nil, fmt.Errorf("LLM 模型已停用")
	}

	p, err := r.providerRepo.FindByID(ctx, m.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("模型所属厂家不存在")
	}
	if p.Status != "0" {
		return nil, fmt.Errorf("LLM 厂家已停用")
	}

	return &Resolved{
		ProviderID: p.ProviderID,
		BaseURL:    p.APIBaseURL,
		APIKey:     p.APIKey,
		ModelName:  m.ModelName,
	}, nil
}
