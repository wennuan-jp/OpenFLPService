package usecase

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"openflp.com/model"
	"openflp.com/service"
)

type ResolveFLPUseCase struct {
	flpService service.FLPService
}

func NewResolveFLPUseCase(flpService service.FLPService) *ResolveFLPUseCase {
	return &ResolveFLPUseCase{
		flpService: flpService,
	}
}

func (uc *ResolveFLPUseCase) Execute(fileName string, r io.Reader) (*model.ResolutionResult, error) {
	// Create a temp directory for uploads if not exists
	tempDir := "./uploads"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		os.Mkdir(tempDir, 0755)
	}

	// Save file
	dst := filepath.Join(tempDir, fileName)
	out, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	
	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		return nil, fmt.Errorf("failed to save file: %v", err)
	}
	out.Close()
	defer os.Remove(dst) // Clean up after processing

	// Resolve FLP
	return uc.flpService.Resolve(dst)
}
