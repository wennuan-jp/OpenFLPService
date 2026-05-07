package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"openflp.com/application/usecase"
	"openflp.com/infra"
	"openflp.com/interfaces/api"
)

func main() {
	r := gin.Default()

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
