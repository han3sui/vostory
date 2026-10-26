package repository

import (
	"context"
	"errors"
	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

type SysOSSConfigRepository interface {
	GetByKey(ctx context.Context, key string) (string, error)
	SetByKey(ctx context.Context, key, name, value string) error
	GetAll(ctx context.Context) (map[string]string, error)
}

func NewSysOSSConfigRepository(repository *Repository) SysOSSConfigRepository {
	return &sysOSSConfigRepository{Repository: repository}
}

type sysOSSConfigRepository struct {
	*Repository
}

func (r *sysOSSConfigRepository) GetByKey(ctx context.Context, key string) (string, error) {
	var config model.SysConfig
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}
	return config.ConfigValue, nil
}

func (r *sysOSSConfigRepository) SetByKey(ctx context.Context, key, name, value string) error {
	var config model.SysConfig
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&config).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		config = model.SysConfig{
			ConfigKey:   key,
			ConfigName:  name,
			ConfigValue: value,
			ConfigType:  "Y",
		}
		return r.db.WithContext(ctx).Create(&config).Error
	}
	return r.db.WithContext(ctx).Model(&config).Update("config_value", value).Error
}

func (r *sysOSSConfigRepository) GetAll(ctx context.Context) (map[string]string, error) {
	var configs []model.SysConfig
	err := r.db.WithContext(ctx).Where("config_key LIKE ?", "oss.%").Find(&configs).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, c := range configs {
		result[c.ConfigKey] = c.ConfigValue
	}
	return result, nil
}
