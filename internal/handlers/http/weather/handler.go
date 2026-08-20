package weatherhandler

import (
	"context"
	"errors"
	"net/http"

	"kota-siaga/internal/dto"
	"kota-siaga/internal/integrations/apiindonesia"
	weatherservice "kota-siaga/internal/services/weather"
	"kota-siaga/pkg/logger"
	"kota-siaga/pkg/messages"
	"kota-siaga/pkg/response"
	"kota-siaga/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	GetWeather(context.Context, string) ([]dto.WeatherForecast, error)
}

type Handler struct {
	Service Service
}

type WeatherHandler = Handler

func NewHandler(service Service) *Handler {
	return &Handler{Service: service}
}

func NewWeatherHandler(service Service) *Handler {
	return NewHandler(service)
}

func Register(router gin.IRouter, client *apiindonesia.Client, redisClient *redis.Client) {
	if router == nil {
		return
	}
	var upstream weatherservice.UpstreamClient
	if client != nil {
		upstream = client
	}
	handler := NewHandler(weatherservice.NewService(upstream, redisClient))
	router.GET("/api/weather", handler.GetWeather)
}

func (h *Handler) GetWeather(ctx *gin.Context) {
	adm4, _ := ctx.GetQuery("adm4")
	if err := weatherservice.ValidateADM4(adm4); err != nil {
		writeInvalidQuery(ctx)
		return
	}
	if h == nil || h.Service == nil {
		writeDependencyUnavailable(ctx)
		return
	}

	requestContext := context.Background()
	if ctx.Request != nil {
		requestContext = ctx.Request.Context()
	}
	result, err := h.Service.GetWeather(requestContext, adm4)
	if err != nil {
		if errors.Is(err, weatherservice.ErrWeatherClient) {
			logger.WriteLogWithContext(ctx, logger.LogLevelError, "Weather upstream client unavailable")
			writeDependencyUnavailable(ctx)
			return
		}
		logger.WriteLogWithContext(ctx, logger.LogLevelError, "Weather upstream request failed")
		writeUpstreamError(ctx)
		return
	}

	ctx.JSON(http.StatusOK, response.Response(http.StatusOK, messages.MsgSuccess, utils.GenerateLogId(ctx), result))
}

func writeInvalidQuery(ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, response.ErrorResponse(
		http.StatusBadRequest,
		"Invalid weather query",
		utils.GenerateLogId(ctx),
		"Invalid adm4",
	))
}

func writeDependencyUnavailable(ctx *gin.Context) {
	ctx.JSON(http.StatusServiceUnavailable, response.ErrorResponse(
		http.StatusServiceUnavailable,
		"Weather service unavailable",
		utils.GenerateLogId(ctx),
		"Weather service unavailable",
	))
}

func writeUpstreamError(ctx *gin.Context) {
	ctx.JSON(http.StatusBadGateway, response.ErrorResponse(
		http.StatusBadGateway,
		"Weather upstream unavailable",
		utils.GenerateLogId(ctx),
		"Weather upstream service unavailable",
	))
}
