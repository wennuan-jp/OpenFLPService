package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"openflp.com/application/usecase"
	"openflp.com/config"
	"openflp.com/infra"
	"openflp.com/interfaces/api"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Max upload size 50MB
	r.MaxMultipartMemory = 50 << 20

	// Initialize infrastructure
	flpResolver := infra.NewFLPResolver("infra/flp_resolver.py")
	authClient, err := infra.NewRealCognitoClient(
		cfg.Cognito.ClientID,
		cfg.Cognito.ClientSecret,
		cfg.Cognito.RedirectURL,
		cfg.Cognito.IssuerURL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Cognito client: %v", err)
	}

	// Initialize application usecases
	resolveFLPUseCase := usecase.NewResolveFLPUseCase(flpResolver)

	// Initialize handlers
	handler := api.NewHandler(resolveFLPUseCase, authClient)

	// Register routes
	handler.RegisterRoutes(r)

	fmt.Printf("Server starting on :%s\n", cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)
}

