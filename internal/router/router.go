package router

import (
	"net/http"
	"time"

	earthquakehandler "kota-siaga/internal/handlers/http/earthquake"
	hospitalhandler "kota-siaga/internal/handlers/http/hospital"
	locationhandler "kota-siaga/internal/handlers/http/location"
	warninghandler "kota-siaga/internal/handlers/http/warning"
	weatherhandler "kota-siaga/internal/handlers/http/weather"
	"kota-siaga/internal/integrations/apiindonesia"
	locationclient "kota-siaga/internal/integrations/locationservice"
	earthquakeservice "kota-siaga/internal/services/earthquake"
	"kota-siaga/middlewares"
	"kota-siaga/pkg/logger"
	"kota-siaga/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Routes struct {
	App *gin.Engine
}

func NewRoutes(redisClient *redis.Client, apiClient *apiindonesia.Client, locationClient *locationclient.Client, earthquakeClient earthquakeservice.UpstreamClient) *Routes {
	app := gin.New()
	app.ForwardedByClientIP = false

	app.Use(middlewares.CORS())
	app.Use(gin.CustomRecovery(middlewares.ErrorHandler))
	app.Use(middlewares.SetContextId())
	app.Use(middlewares.RequestLogger())
	app.Use(middlewares.IPRateLimitMiddleware(
		redisClient,
		"public",
		utils.GetEnv("PUBLIC_RATE_LIMIT", 30),
		utils.GetEnv("PUBLIC_RATE_WINDOW", time.Minute),
	))

	app.GET("/healthcheck", func(ctx *gin.Context) {
		logger.WriteLogWithContext(ctx, logger.LogLevelDebug, "Healthcheck requested")
		ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	locationhandler.Register(app, locationClient, redisClient)
	weatherhandler.Register(app, apiClient, redisClient)
	warninghandler.Register(app, apiClient, redisClient)
	earthquakehandler.Register(app, earthquakeClient, redisClient)
	hospitalhandler.Register(app, apiClient, redisClient)

	return &Routes{App: app}
}
