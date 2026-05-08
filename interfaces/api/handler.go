package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"openflp.com/application/usecase"
	"openflp.com/infra"
)

type Handler struct {
	resolveFLPUseCase *usecase.ResolveFLPUseCase
	authClient        infra.CognitoClient
}

func NewHandler(resolveFLPUseCase *usecase.ResolveFLPUseCase, authClient infra.CognitoClient) *Handler {
	return &Handler{
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
		
		// Protected v1 routes
		protected := v1.Group("/")
		protected.Use(h.AuthMiddleware())
		{
			protected.POST("/upload", h.UploadFLP)
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

	// In a real app, you might want to redirect to the frontend with the token
	// or set a cookie. Here we'll return the token as JSON for now, 
	// but the user's webapp expects a redirect.
	
	// If you want to redirect back to frontend:
	// frontendURL := fmt.Sprintf("http://localhost:5173/callback#access_token=%s", authResp.Token)
	// c.Redirect(http.StatusFound, frontendURL)
	
	c.JSON(http.StatusOK, authResp)
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
