package warninghandler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"kota-siaga/internal/dto"
	handlercommon "kota-siaga/internal/handlers/http/common"
	"kota-siaga/internal/integrations/apiindonesia"
	warningservice "kota-siaga/internal/services/warning"
	"kota-siaga/pkg/logger"
	"kota-siaga/pkg/messages"
	"kota-siaga/pkg/response"
	"kota-siaga/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	ListWarnings(context.Context, string) ([]dto.Warning, error)
}

type Handler struct {
	Service Service
}

type WarningHandler = Handler

func NewHandler(service Service) *Handler {
	return &Handler{Service: service}
}

func NewWarningHandler(service Service) *Handler {
	return NewHandler(service)
}

func Register(router gin.IRouter, client *apiindonesia.Client, redisClient *redis.Client) {
	if router == nil {
		return
	}
	var upstream warningservice.UpstreamClient
	if client != nil {
		upstream = client
	}
	handler := NewHandler(warningservice.NewService(upstream, redisClient))
	router.GET("/api/warnings", handler.GetWarnings)
}

func (h *Handler) GetWarnings(ctx *gin.Context) {
	province, _ := ctx.GetQuery("provinsi")
	if err := warningservice.ValidateProvince(province); err != nil {
		writeInvalidProvince(ctx)
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
	result, err := h.Service.ListWarnings(requestContext, province)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Response(http.StatusOK, messages.MsgSuccess, utils.GenerateLogId(ctx), result))
}

func handleServiceError(ctx *gin.Context, err error) {
	if errors.Is(err, warningservice.ErrInvalidProvince) {
		writeInvalidProvince(ctx)
		return
	}
	if errors.Is(err, warningservice.ErrWarningClient) {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, "Warning upstream client unavailable")
		writeDependencyUnavailable(ctx)
		return
	}

	var upstreamErr *apiindonesia.UpstreamError
	if errors.As(err, &upstreamErr) && upstreamErr != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf(
			"Warning upstream request failed: status=%d code=%s",
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

	logger.WriteLogWithContext(ctx, logger.LogLevelError, "Warning upstream request failed")
	writeUpstreamError(ctx)
}

func writeInvalidProvince(ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, response.ErrorResponse(
		http.StatusBadRequest,
		"Invalid warning query",
		utils.GenerateLogId(ctx),
		"Invalid province",
	))
}

func writeDependencyUnavailable(ctx *gin.Context) {
	ctx.JSON(http.StatusServiceUnavailable, response.ErrorResponse(
		http.StatusServiceUnavailable,
		"Warning service unavailable",
		utils.GenerateLogId(ctx),
		"Warning service unavailable",
	))
}

func writeNotFound(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, response.ErrorResponse(
		http.StatusNotFound,
		"Warning not found",
		utils.GenerateLogId(ctx),
		"Warning not found",
	))
}

func writeUpstreamError(ctx *gin.Context) {
	ctx.JSON(http.StatusBadGateway, response.ErrorResponse(
		http.StatusBadGateway,
		"Warning upstream unavailable",
		utils.GenerateLogId(ctx),
		"Warning upstream service unavailable",
	))
}
