package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"openflp.com/application/usecase"
)

type Handler struct {
	resolveFLPUseCase *usecase.ResolveFLPUseCase
}

func NewHandler(resolveFLPUseCase *usecase.ResolveFLPUseCase) *Handler {
	return &Handler{
		resolveFLPUseCase: resolveFLPUseCase,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/upload", h.UploadFLP)
	r.GET("/health", h.HealthCheck)
}

func (h *Handler) UploadFLP(c *gin.Context) {
	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to open uploaded file: %v", err)})
		return
	}
	defer f.Close()

	// Resolve FLP using usecase
	result, err := h.resolveFLPUseCase.Execute(file.Filename, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to resolve FLP: %v", err)})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
