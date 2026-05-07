package main

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"openflp.com/application/usecase"
	"openflp.com/infra"
	"openflp.com/interfaces/api"
)

func main() {
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Max upload size 50MB
	r.MaxMultipartMemory = 50 << 20

	// Initialize infrastructure
	flpResolver := infra.NewFLPResolver("infra/flp_resolver.py")

	// Initialize application usecases
	resolveFLPUseCase := usecase.NewResolveFLPUseCase(flpResolver)

	// Initialize handlers
	handler := api.NewHandler(resolveFLPUseCase)

	// Register routes
	handler.RegisterRoutes(r)

	fmt.Println("Server starting on :8080")
	r.Run(":8080")
}
