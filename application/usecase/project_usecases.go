package usecase

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"openflp.com/model"
	"openflp.com/repository"
	"openflp.com/service"
)

type ProjectUseCase struct {
	projectRepo repository.ProjectRepository
	tagRepo     repository.TagRepository
	flpService  service.FLPService
}

func NewProjectUseCase(
	projectRepo repository.ProjectRepository,
	tagRepo repository.TagRepository,
	flpService service.FLPService,
) *ProjectUseCase {
	return &ProjectUseCase{
		projectRepo: projectRepo,
		tagRepo:     tagRepo,
		flpService:  flpService,
	}
}

func (uc *ProjectUseCase) ListAll(ctx context.Context) ([]model.Project, error) {
	return uc.projectRepo.List(ctx)
}

func (uc *ProjectUseCase) ListByAuthor(ctx context.Context, authorID string) ([]model.Project, error) {
	return uc.projectRepo.ListByAuthor(ctx, authorID)
}

func (uc *ProjectUseCase) GetDetails(ctx context.Context, id string) (*model.Project, error) {
	return uc.projectRepo.GetByID(ctx, id)
}

func (uc *ProjectUseCase) SearchTags(ctx context.Context, query string) ([]model.Tag, error) {
	return uc.tagRepo.Search(ctx, query)
}

func (uc *ProjectUseCase) CreateTag(ctx context.Context, name string) error {
	return uc.tagRepo.Create(ctx, &model.Tag{Name: name})
}

func (uc *ProjectUseCase) UpdateMetadata(ctx context.Context, id string, metadata map[string]interface{}) error {
	return uc.projectRepo.UpdateMetadata(ctx, id, metadata)
}

func (uc *ProjectUseCase) UploadBundle(ctx context.Context, user *model.User, fileName string, r io.Reader, size int64) (*model.Project, error) {
	// Create a storage directory
	storageDir := "./uploads/projects"
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		os.MkdirAll(storageDir, 0755)
	}

	id := uuid.New().String()
	ext := filepath.Ext(fileName)
	storedName := fmt.Sprintf("%s%s", id, ext)
	dst := filepath.Join(storageDir, storedName)

	out, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage file: %v", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		return nil, fmt.Errorf("failed to save bundle: %v", err)
	}

	// Resolve plugins if it's an FLP file directly or if we need to analyze
	// For now, let's assume we can resolve it if it's an .flp
	var plugins []model.Plugin
	if ext == ".flp" {
		res, err := uc.flpService.Resolve(dst)
		if err == nil {
			plugins = res.Plugins
		}
	}

	project := &model.Project{
		ID:           id,
		Name:         fileName,
		AuthorID:     user.ID,
		AuthorName:   user.Name,
		UploadDate:   time.Now(),
		Size:         size,
		SizeDisplay:  formatSize(size),
		PluginsCount: len(plugins),
		FilePath:     dst,
		Plugins:      plugins,
	}

	if err := uc.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (uc *ProjectUseCase) DownloadBundle(ctx context.Context, id string) (string, string, error) {
	project, err := uc.projectRepo.GetByID(ctx, id)
	if err != nil {
		return "", "", err
	}
	return project.FilePath, project.Name, nil
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
