package main

import (
	"aoa-inventory/api"
	"aoa-inventory/config"
	squareClient "aoa-inventory/squareUtils/client"
	"aoa-inventory/utils"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var log = utils.NewLogger("MAIN")

func main() {
	// load config
	config.Load()

	// setup gin
	gin.SetMode(gin.ReleaseMode)

	// setup square client
	squareClient.Init(config.SquareAccessToken, config.SquareEnv)

	// setup healthcheck route before setting cors
	ginEngine := gin.Default()
	ginEngine.GET("/healthcheck", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	ginEngine.Use(cors.New(cors.Config{
		AllowOrigins:     config.AllowedOrigins,
		AllowMethods:     []string{"GET", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           10 * time.Minute,
	}))

	// setup endpoints for the api
	apiGroup := ginEngine.Group("/api")
	api.SetupEndpoints(apiGroup)

	// start server
	log.Printf("Server started on port %s...\n", config.Port)
	if err := ginEngine.Run(":" + config.Port); err != nil {
		log.Println("ERROR: Could not start server:", err)
		return
	}
}
