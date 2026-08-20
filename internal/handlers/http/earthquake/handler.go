package earthquakehandler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"kota-siaga/internal/dto"
	handlercommon "kota-siaga/internal/handlers/http/common"
	"kota-siaga/internal/integrations/apiindonesia"
	earthquakeservice "kota-siaga/internal/services/earthquake"
	"kota-siaga/pkg/logger"
	"kota-siaga/pkg/messages"
	"kota-siaga/pkg/response"
	"kota-siaga/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	ListLatest(context.Context) ([]dto.Earthquake, error)
}

type Handler struct {
	Service Service
}

type EarthquakeHandler = Handler

func NewHandler(service Service) *Handler {
	return &Handler{Service: service}
}

func NewEarthquakeHandler(service Service) *Handler {
	return NewHandler(service)
}

func Register(router gin.IRouter, client *apiindonesia.Client, redisClient *redis.Client) {
	if router == nil {
		return
	}
	var upstream earthquakeservice.UpstreamClient
	if client != nil {
		upstream = client
	}
	handler := NewHandler(earthquakeservice.NewService(upstream, redisClient))
	router.GET("/api/earthquakes/latest", handler.GetLatest)
}

func (h *Handler) GetLatest(ctx *gin.Context) {
	if h == nil || h.Service == nil {
		writeDependencyUnavailable(ctx)
		return
	}

	requestContext := context.Background()
	if ctx.Request != nil {
		requestContext = ctx.Request.Context()
	}
	result, err := h.Service.ListLatest(requestContext)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Response(http.StatusOK, messages.MsgSuccess, utils.GenerateLogId(ctx), result))
}

func handleServiceError(ctx *gin.Context, err error) {
	if errors.Is(err, earthquakeservice.ErrEarthquakeClient) {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, "Earthquake upstream client unavailable")
		writeDependencyUnavailable(ctx)
		return
	}

	var upstreamErr *apiindonesia.UpstreamError
	if errors.As(err, &upstreamErr) && upstreamErr != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf(
			"Earthquake upstream request failed: status=%d code=%s",
			upstreamErr.StatusCode,
			handlercommon.SafeUpstreamCode(upstreamErr.Code),
		))
		if upstreamErr.StatusCode == http.StatusNotFound {
			writeNotFound(ctx)
			return
		}
		writeUpstreamError(ctx)
		return
	}

	logger.WriteLogWithContext(ctx, logger.LogLevelError, "Earthquake upstream request failed")
	writeUpstreamError(ctx)
}

func writeDependencyUnavailable(ctx *gin.Context) {
	ctx.JSON(http.StatusServiceUnavailable, response.ErrorResponse(
		http.StatusServiceUnavailable,
		"Earthquake service unavailable",
		utils.GenerateLogId(ctx),
		"Earthquake service unavailable",
	))
}

func writeNotFound(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, response.ErrorResponse(
		http.StatusNotFound,
		"Earthquake not found",
		utils.GenerateLogId(ctx),
		"Earthquake not found",
	))
}

func writeUpstreamError(ctx *gin.Context) {
	ctx.JSON(http.StatusBadGateway, response.ErrorResponse(
		http.StatusBadGateway,
		"Earthquake upstream unavailable",
		utils.GenerateLogId(ctx),
		"Earthquake upstream service unavailable",
	))
}
