package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsVoiceEmotionRepository interface {
	Create(ctx context.Context, emotion *model.VsVoiceEmotion) error
	BatchCreate(ctx context.Context, emotions []*model.VsVoiceEmotion) error
	Update(ctx context.Context, emotion *model.VsVoiceEmotion) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsVoiceEmotion, error)
	FindWithPagination(ctx context.Context, query *v1.VsVoiceEmotionListQuery) ([]*model.VsVoiceEmotion, int64, error)
	FindByVoiceProfileID(ctx context.Context, voiceProfileID uint64) ([]*model.VsVoiceEmotion, error)
	FindByVoiceAssetID(ctx context.Context, voiceAssetID uint64) ([]*model.VsVoiceEmotion, error)
	FindByVoiceAssetIDs(ctx context.Context, ids []uint64) ([]*model.VsVoiceEmotion, error)
	DeleteByVoiceProfileID(ctx context.Context, voiceProfileID uint64) error
	DeleteByVoiceAssetID(ctx context.Context, voiceAssetID uint64) error
	FindByMatch(ctx context.Context, voiceProfileID uint64, emotionType, emotionStrength string) (*model.VsVoiceEmotion, error)
}

func NewVsVoiceEmotionRepository(repository *Repository) VsVoiceEmotionRepository {
	return &vsVoiceEmotionRepository{Repository: repository}
}

type vsVoiceEmotionRepository struct {
	*Repository
}

func (r *vsVoiceEmotionRepository) Create(ctx context.Context, emotion *model.VsVoiceEmotion) error {
	return r.db.WithContext(ctx).Create(emotion).Error
}

func (r *vsVoiceEmotionRepository) BatchCreate(ctx context.Context, emotions []*model.VsVoiceEmotion) error {
	if len(emotions) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&emotions).Error
}

func (r *vsVoiceEmotionRepository) Update(ctx context.Context, emotion *model.VsVoiceEmotion) error {
	return r.db.WithContext(ctx).Model(emotion).
		Where("voice_emotion_id = ?", emotion.VoiceEmotionID).
		Omit("created_by", "created_at", "voice_emotion_id", "voice_profile_id", "voice_asset_id").
		Updates(emotion).Error
}

func (r *vsVoiceEmotionRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("voice_emotion_id = ?", id).Delete(&model.VsVoiceEmotion{}).Error
}

func (r *vsVoiceEmotionRepository) FindByID(ctx context.Context, id uint64) (*model.VsVoiceEmotion, error) {
	var emotion model.VsVoiceEmotion
	if err := r.db.WithContext(ctx).
		Where("voice_emotion_id = ?", id).First(&emotion).Error; err != nil {
		return nil, err
	}
	return &emotion, nil
}

func (r *vsVoiceEmotionRepository) FindWithPagination(ctx context.Context, query *v1.VsVoiceEmotionListQuery) ([]*model.VsVoiceEmotion, int64, error) {
	var emotions []*model.VsVoiceEmotion
	db := r.db.WithContext(ctx).Model(&model.VsVoiceEmotion{})

	if query.VoiceProfileID > 0 {
		db = db.Where("voice_profile_id = ?", query.VoiceProfileID)
	}
	if query.VoiceAssetID > 0 {
		db = db.Where("voice_asset_id = ?", query.VoiceAssetID)
	}
	if query.EmotionType != "" {
		db = db.Where("emotion_type = ?", query.EmotionType)
	}
	if query.EmotionStrength != "" {
		db = db.Where("emotion_strength = ?", query.EmotionStrength)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	if err := db.Order("voice_emotion_id DESC").Find(&emotions).Error; err != nil {
		return nil, 0, err
	}

	return emotions, total, nil
}

func (r *vsVoiceEmotionRepository) FindByVoiceProfileID(ctx context.Context, voiceProfileID uint64) ([]*model.VsVoiceEmotion, error) {
	var emotions []*model.VsVoiceEmotion
	if err := r.db.WithContext(ctx).
		Where("voice_profile_id = ?", voiceProfileID).
		Order("emotion_type ASC, emotion_strength ASC").Find(&emotions).Error; err != nil {
		return nil, err
	}
	return emotions, nil
}

func (r *vsVoiceEmotionRepository) FindByVoiceAssetID(ctx context.Context, voiceAssetID uint64) ([]*model.VsVoiceEmotion, error) {
	var emotions []*model.VsVoiceEmotion
	if err := r.db.WithContext(ctx).
		Where("voice_asset_id = ?", voiceAssetID).
		Order("emotion_type ASC, emotion_strength ASC").Find(&emotions).Error; err != nil {
		return nil, err
	}
	return emotions, nil
}

func (r *vsVoiceEmotionRepository) FindByVoiceAssetIDs(ctx context.Context, ids []uint64) ([]*model.VsVoiceEmotion, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var emotions []*model.VsVoiceEmotion
	if err := r.db.WithContext(ctx).
		Where("voice_asset_id IN ?", ids).
		Find(&emotions).Error; err != nil {
		return nil, err
	}
	return emotions, nil
}

func (r *vsVoiceEmotionRepository) DeleteByVoiceProfileID(ctx context.Context, voiceProfileID uint64) error {
	return r.db.WithContext(ctx).Where("voice_profile_id = ?", voiceProfileID).Delete(&model.VsVoiceEmotion{}).Error
}

func (r *vsVoiceEmotionRepository) DeleteByVoiceAssetID(ctx context.Context, voiceAssetID uint64) error {
	return r.db.WithContext(ctx).Where("voice_asset_id = ?", voiceAssetID).Delete(&model.VsVoiceEmotion{}).Error
}

// FindByMatch finds a specific emotion reference audio by profile + type + strength.
func (r *vsVoiceEmotionRepository) FindByMatch(ctx context.Context, voiceProfileID uint64, emotionType, emotionStrength string) (*model.VsVoiceEmotion, error) {
	var emotion model.VsVoiceEmotion
	if err := r.db.WithContext(ctx).
		Where("voice_profile_id = ? AND emotion_type = ? AND emotion_strength = ?",
			voiceProfileID, emotionType, emotionStrength).
		First(&emotion).Error; err != nil {
		return nil, err
	}
	return &emotion, nil
}
