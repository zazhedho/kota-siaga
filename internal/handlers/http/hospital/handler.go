package hospitalhandler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"kota-siaga/internal/dto"
	handlercommon "kota-siaga/internal/handlers/http/common"
	"kota-siaga/internal/integrations/satusehat"
	hospitalservice "kota-siaga/internal/services/hospital"
	"kota-siaga/pkg/logger"
	"kota-siaga/pkg/response"
	"kota-siaga/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	ListHospitals(context.Context, string, string, int, int) (dto.Page[dto.Hospital], error)
}

type Handler struct {
	Service Service
}

type HospitalHandler = Handler

func NewHandler(service Service) *Handler {
	return &Handler{Service: service}
}

func NewHospitalHandler(service Service) *Handler {
	return NewHandler(service)
}

func Register(router gin.IRouter, client *satusehat.Client, redisClient *redis.Client) {
	if router == nil {
		return
	}
	var upstream hospitalservice.UpstreamClient
	if client != nil {
		upstream = client
	}
	handler := NewHandler(hospitalservice.NewService(upstream, redisClient))
	router.GET("/api/hospitals", handler.GetHospitals)
}

func (h *Handler) GetHospitals(ctx *gin.Context) {
	kabupatenID, search, page, perPage, err := parseQuery(ctx)
	if err != nil {
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
	result, err := h.Service.ListHospitals(requestContext, kabupatenID, search, page, perPage)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, response.PaginationResponseWithTotalPages(
		http.StatusOK,
		result.Total,
		result.Page,
		result.PerPage,
		result.TotalPages,
		utils.GenerateLogId(ctx),
		result.Data,
	))
}

func parseQuery(ctx *gin.Context) (string, string, int, int, error) {
	kabupatenID, ok := ctx.GetQuery("kabupaten_id")
	if !ok || kabupatenID == "" {
		return "", "", 0, 0, errors.New("invalid hospital query")
	}
	if err := hospitalservice.ValidateKabupatenID(kabupatenID); err != nil {
		return "", "", 0, 0, err
	}

	search, err := hospitalservice.NormalizeSearch(ctx.Query("search"))
	if err != nil {
		return "", "", 0, 0, err
	}

	page, err := queryInt(ctx, "page")
	if err != nil {
		return "", "", 0, 0, err
	}
	perPage, err := queryInt(ctx, "per_page")
	if err != nil {
		return "", "", 0, 0, err
	}
	if err := hospitalservice.ValidatePagination(page, perPage); err != nil {
		return "", "", 0, 0, err
	}
	return kabupatenID, search, page, perPage, nil
}

func queryInt(ctx *gin.Context, key string) (int, error) {
	value, ok := ctx.GetQuery(key)
	if !ok || value == "" {
		return 0, fmt.Errorf("invalid hospital query: %s is required", key)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid hospital query: %s must be an integer", key)
	}
	return parsed, nil
}

func handleServiceError(ctx *gin.Context, err error) {
	if errors.Is(err, hospitalservice.ErrInvalidKabupatenID) || errors.Is(err, hospitalservice.ErrInvalidPagination) || errors.Is(err, hospitalservice.ErrInvalidSearch) {
		writeInvalidQuery(ctx)
		return
	}
	if errors.Is(err, hospitalservice.ErrHospitalClient) {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, "Hospital upstream client unavailable")
		writeDependencyUnavailable(ctx)
		return
	}

	var upstreamErr *satusehat.UpstreamError
	if errors.As(err, &upstreamErr) && upstreamErr != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("Hospital upstream request failed: status=%d code=%s", upstreamErr.StatusCode, handlercommon.SafeUpstreamCode(upstreamErr.Code)))
		if upstreamErr.StatusCode == http.StatusNotFound && upstreamErr.IsResourceError {
			writeNotFound(ctx)
			return
		}
		writeUpstreamError(ctx)
		return
	}

	logger.WriteLogWithContext(ctx, logger.LogLevelError, "Hospital upstream request failed")
	writeUpstreamError(ctx)
}

func writeInvalidQuery(ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, response.ErrorResponse(
		http.StatusBadRequest,
		"Invalid hospital query",
		utils.GenerateLogId(ctx),
		"Invalid hospital query",
	))
}

func writeDependencyUnavailable(ctx *gin.Context) {
	ctx.JSON(http.StatusServiceUnavailable, response.ErrorResponse(
		http.StatusServiceUnavailable,
		"Hospital service unavailable",
		utils.GenerateLogId(ctx),
		"Hospital service unavailable",
	))
}

func writeNotFound(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, response.ErrorResponse(
		http.StatusNotFound,
		"Hospital not found",
		utils.GenerateLogId(ctx),
		"Hospital not found",
	))
}

func writeUpstreamError(ctx *gin.Context) {
	ctx.JSON(http.StatusBadGateway, response.ErrorResponse(
		http.StatusBadGateway,
		"Hospital upstream unavailable",
		utils.GenerateLogId(ctx),
		"Hospital upstream service unavailable",
	))
}
