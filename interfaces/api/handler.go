package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"openflp.com/application/usecase"
	"openflp.com/infra"
	"openflp.com/model"
)

type Handler struct {
	projectUseCase    *usecase.ProjectUseCase
	resolveFLPUseCase *usecase.ResolveFLPUseCase
	authClient        infra.CognitoClient
}

func NewHandler(projectUseCase *usecase.ProjectUseCase, resolveFLPUseCase *usecase.ResolveFLPUseCase, authClient infra.CognitoClient) *Handler {
	return &Handler{
		projectUseCase:    projectUseCase,
		resolveFLPUseCase: resolveFLPUseCase,
		authClient:        authClient,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// Auth Group
	auth := r.Group("/auth")
	{
		auth.GET("/login", h.Login)
		auth.GET("/callback", h.Callback)
	}

	// API V1 Group
	v1 := r.Group("/v1")
	{
		v1.GET("/health", h.HealthCheck)
		v1.GET("/list", h.ListProjects)
		v1.GET("/details/:id", h.GetProjectDetails)
		v1.GET("/tags/search", h.SearchTags)
		
		// Protected v1 routes
		protected := v1.Group("/")
		protected.Use(h.AuthMiddleware())
		{
			protected.GET("/my-list", h.ListMyProjects)
			protected.POST("/upload", h.UploadFLP)
			protected.POST("/upload-bundle", h.UploadBundle)
			protected.POST("/tags", h.CreateTag)
			protected.POST("/project/:id/metadata", h.UpdateMetadata)
			protected.GET("/project/:id/download", h.DownloadBundle)
		}
	}
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		user, err := h.authClient.VerifyToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func (h *Handler) Login(c *gin.Context) {
	state := "state" // In production, use a secure random string
	url := h.authClient.GetAuthURL(state)
	c.Redirect(http.StatusFound, url)
}

func (h *Handler) Callback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
		return
	}

	authResp, err := h.authClient.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResp)
}

func (h *Handler) ListProjects(c *gin.Context) {
	projects, err := h.projectUseCase.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, projects)
}

func (h *Handler) ListMyProjects(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(*model.User)
	
	projects, err := h.projectUseCase.ListByAuthor(c.Request.Context(), u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, projects)
}

func (h *Handler) GetProjectDetails(c *gin.Context) {
	id := c.Param("id")
	project, err := h.projectUseCase.GetDetails(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Format to ResolutionResult if needed, or just return project
	// Frontend expects ResolutionResult in getFLPDetails
	tagNames := make([]string, 0, len(project.Tags))
	for _, t := range project.Tags {
		tagNames = append(tagNames, t.Name)
	}

	result := model.ResolutionResult{
		ID:           project.ID,
		Name:         project.Name,
		Description:  project.Description,
		Genre:        project.Genre,
		Tags:         tagNames,
		Plugins:      project.Plugins,
		PluginsCount: project.PluginsCount,
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) SearchTags(c *gin.Context) {
	q := c.Query("q")
	tags, err := h.projectUseCase.SearchTags(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}
	c.JSON(http.StatusOK, names)
}

func (h *Handler) CreateTag(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.projectUseCase.CreateTag(c.Request.Context(), body.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *Handler) UpdateMetadata(c *gin.Context) {
	id := c.Param("id")
	var metadata map[string]interface{}
	if err := c.ShouldBindJSON(&metadata); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.projectUseCase.UpdateMetadata(c.Request.Context(), id, metadata); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) UploadFLP(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(*model.User)

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

	project, err := h.projectUseCase.UploadBundle(c.Request.Context(), u, file.Filename, f, file.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload: %v", err)})
		return
	}

	result := model.ResolutionResult{
		ID:           project.ID,
		Name:         project.Name,
		Description:  project.Description,
		Genre:        project.Genre,
		Plugins:      project.Plugins,
		PluginsCount: project.PluginsCount,
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) UploadBundle(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(*model.User)

	file, err := c.FormFile("bundle")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No bundle uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to open uploaded bundle: %v", err)})
		return
	}
	defer f.Close()

	project, err := h.projectUseCase.UploadBundle(c.Request.Context(), u, file.Filename, f, file.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload bundle: %v", err)})
		return
	}

	result := model.ResolutionResult{
		ID:           project.ID,
		Name:         project.Name,
		Description:  project.Description,
		Genre:        project.Genre,
		Plugins:      project.Plugins,
		PluginsCount: project.PluginsCount,
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DownloadBundle(c *gin.Context) {
	id := c.Param("id")
	filePath, fileName, err := h.projectUseCase.DownloadBundle(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.FileAttachment(filePath, fileName)
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
