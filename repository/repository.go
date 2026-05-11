package repository

import (
	"context"
	"openflp.com/model"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *model.Project) error
	GetByID(ctx context.Context, id string) (*model.Project, error)
	List(ctx context.Context) ([]model.Project, error)
	ListByAuthor(ctx context.Context, authorID string) ([]model.Project, error)
	UpdateMetadata(ctx context.Context, id string, metadata map[string]interface{}) error
	Delete(ctx context.Context, id string) error
}

type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	FindByName(ctx context.Context, name string) (*model.Tag, error)
	Search(ctx context.Context, query string) ([]model.Tag, error)
}
