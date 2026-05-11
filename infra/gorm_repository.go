package infra

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"openflp.com/model"
)

type gormProjectRepository struct {
	db *gorm.DB
}

func NewGORMProjectRepository(db *gorm.DB) *gormProjectRepository {
	return &gormProjectRepository{db: db}
}

func (r *gormProjectRepository) Create(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *gormProjectRepository) GetByID(ctx context.Context, id string) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).Preload("Tags").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *gormProjectRepository) List(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	err := r.db.WithContext(ctx).Preload("Tags").Order("upload_date desc").Find(&projects).Error
	return projects, err
}

func (r *gormProjectRepository) ListByAuthor(ctx context.Context, authorID string) ([]model.Project, error) {
	var projects []model.Project
	err := r.db.WithContext(ctx).Preload("Tags").Where("author_id = ?", authorID).Order("upload_date desc").Find(&projects).Error
	return projects, err
}

func (r *gormProjectRepository) UpdateMetadata(ctx context.Context, id string, metadata map[string]interface{}) error {
	// Handle tags separately if they are in the metadata
	if tags, ok := metadata["tags"].([]string); ok {
		var tagModels []model.Tag
		for _, tagName := range tags {
			var tag model.Tag
			r.db.FirstOrCreate(&tag, model.Tag{Name: tagName})
			tagModels = append(tagModels, tag)
		}
		
		var project model.Project
		if err := r.db.First(&project, "id = ?", id).Error; err != nil {
			return err
		}
		if err := r.db.Model(&project).Association("Tags").Replace(tagModels); err != nil {
			return err
		}
		delete(metadata, "tags")
	}

	return r.db.WithContext(ctx).Model(&model.Project{}).Where("id = ?", id).Updates(metadata).Error
}

func (r *gormProjectRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, "id = ?", id).Error
}

type gormTagRepository struct {
	db *gorm.DB
}

func NewGORMTagRepository(db *gorm.DB) *gormTagRepository {
	return &gormTagRepository{db: db}
}

func (r *gormTagRepository) Create(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *gormTagRepository) FindByName(ctx context.Context, name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.WithContext(ctx).First(&tag, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *gormTagRepository) Search(ctx context.Context, query string) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.WithContext(ctx).Where("name LIKE ?", fmt.Sprintf("%%%s%%", query)).Limit(10).Find(&tags).Error
	return tags, err
}
