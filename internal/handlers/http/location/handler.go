package locationhandler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"kota-siaga/internal/dto"
	handlercommon "kota-siaga/internal/handlers/http/common"
	"kota-siaga/internal/integrations/apiindonesia"
	locationservice "kota-siaga/internal/services/location"
	"kota-siaga/pkg/logger"
	"kota-siaga/pkg/messages"
	"kota-siaga/pkg/response"
	"kota-siaga/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	ListProvinces(context.Context, int, int) (dto.Page[dto.Province], error)
	ListCities(context.Context, string, int, int) (dto.Page[dto.City], error)
	ListDistricts(context.Context, string, int, int) (dto.Page[dto.District], error)
	ListVillages(context.Context, string, int, int) (dto.Page[dto.Village], error)
	SearchLocations(context.Context, string, int) ([]dto.LocationSearchItem, error)
	ResolveLocation(context.Context, string) (dto.LocationPath, error)
}

type Handler struct {
	Service Service
}

type LocationHandler = Handler

func NewHandler(service Service) *Handler {
	return &Handler{Service: service}
}

func NewLocationHandler(service Service) *Handler {
	return NewHandler(service)
}

func Register(router gin.IRouter, client locationservice.UpstreamClient, redisClient *redis.Client) {
	if router == nil {
		return
	}
	handler := NewHandler(locationservice.NewService(client, redisClient))
	router.GET("/api/locations/province", handler.GetProvince)
	router.GET("/api/locations/city", handler.GetCity)
	router.GET("/api/locations/district", handler.GetDistrict)
	router.GET("/api/locations/village", handler.GetVillage)
	router.GET("/api/locations/search", handler.Search)
	router.GET("/api/locations/resolve", handler.Resolve)
}

func (h *Handler) GetProvince(ctx *gin.Context) {
	page, perPage, err := parsePagination(ctx)
	if err != nil {
		writeInvalidQuery(ctx, err)
		return
	}
	if h == nil || h.Service == nil {
		handleServiceError(ctx, locationservice.ErrLocationClient)
		return
	}

	result, err := h.Service.ListProvinces(ctx.Request.Context(), page, perPage)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	writePage(ctx, result.Total, result.Page, result.PerPage, result.Data)
}

func (h *Handler) GetCity(ctx *gin.Context) {
	page, perPage, err := parsePagination(ctx)
	if err != nil {
		writeInvalidQuery(ctx, err)
		return
	}
	provinceID, err := parentID(ctx, "province_id")
	if err != nil {
		writeInvalidQuery(ctx, err)
		return
	}
	if h == nil || h.Service == nil {
		handleServiceError(ctx, locationservice.ErrLocationClient)
		return
	}

	result, err := h.Service.ListCities(ctx.Request.Context(), provinceID, page, perPage)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	writePage(ctx, result.Total, result.Page, result.PerPage, result.Data)
}

func (h *Handler) GetDistrict(ctx *gin.Context) {
	page, perPage, err := parsePagination(ctx)
	if err != nil {
		writeInvalidQuery(ctx, err)
		return
	}
	regencyID, err := parentID(ctx, "kabupaten_id")
	if err != nil {
		writeInvalidQuery(ctx, err)
		return
	}
	if h == nil || h.Service == nil {
		handleServiceError(ctx, locationservice.ErrLocationClient)
		return
	}

	result, err := h.Service.ListDistricts(ctx.Request.Context(), regencyID, page, perPage)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	writePage(ctx, result.Total, result.Page, result.PerPage, result.Data)
}

func (h *Handler) GetVillage(ctx *gin.Context) {
	page, perPage, err := parsePagination(ctx)
	if err != nil {
		writeInvalidQuery(ctx, err)
		return
	}
	districtID, err := parentID(ctx, "kecamatan_id")
	if err != nil {
		writeInvalidQuery(ctx, err)
		return
	}
	if h == nil || h.Service == nil {
		handleServiceError(ctx, locationservice.ErrLocationClient)
		return
	}

	result, err := h.Service.ListVillages(ctx.Request.Context(), districtID, page, perPage)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	writePage(ctx, result.Total, result.Page, result.PerPage, result.Data)
}

func (h *Handler) Search(ctx *gin.Context) {
	query := ctx.Query("q")
	limit := locationservice.DefaultSearchLimit
	if rawLimit, ok := ctx.GetQuery("limit"); ok {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			writeInvalidQuery(ctx, fmt.Errorf("%w: limit must be an integer", locationservice.ErrInvalidSearchLimit))
			return
		}
		limit = parsed
	}
	if err := locationservice.ValidateSearchQuery(query); err != nil {
		writeInvalidQuery(ctx, err)
		return
	}
	if err := locationservice.ValidateSearchLimit(limit); err != nil {
		writeInvalidQuery(ctx, err)
		return
	}
	if h == nil || h.Service == nil {
		handleServiceError(ctx, locationservice.ErrLocationClient)
		return
	}

	result, err := h.Service.SearchLocations(ctx.Request.Context(), query, limit)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	writeData(ctx, result)
}

func (h *Handler) Resolve(ctx *gin.Context) {
	code := ctx.Query("code")
	if err := locationservice.ValidateLocationCode(code); err != nil {
		writeInvalidQuery(ctx, err)
		return
	}
	if h == nil || h.Service == nil {
		handleServiceError(ctx, locationservice.ErrLocationClient)
		return
	}

	result, err := h.Service.ResolveLocation(ctx.Request.Context(), code)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	writeData(ctx, result)
}

func parsePagination(ctx *gin.Context) (int, int, error) {
	page, err := queryInt(ctx, "page")
	if err != nil {
		return 0, 0, err
	}
	perPage, err := queryInt(ctx, "per_page")
	if err != nil {
		return 0, 0, err
	}
	if err := locationservice.ValidatePagination(page, perPage); err != nil {
		return 0, 0, err
	}
	return page, perPage, nil
}

func queryInt(ctx *gin.Context, key string) (int, error) {
	value, ok := ctx.GetQuery(key)
	if !ok || value == "" {
		return 0, fmt.Errorf("%w: %s is required", locationservice.ErrInvalidPagination, key)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%w: %s must be an integer", locationservice.ErrInvalidPagination, key)
	}
	return parsed, nil
}

func parentID(ctx *gin.Context, key string) (string, error) {
	value, _ := ctx.GetQuery(key)
	if err := locationservice.ValidateParentID(value); err != nil {
		return "", fmt.Errorf("%w: %s must be a non-empty numeric string", locationservice.ErrInvalidParentID, key)
	}
	return value, nil
}

func writePage(ctx *gin.Context, total, page, perPage int, data any) {
	ctx.JSON(http.StatusOK, response.PaginationResponse(http.StatusOK, total, page, perPage, utils.GenerateLogId(ctx), data))
}

func writeData(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, response.Response(http.StatusOK, messages.MsgSuccess, utils.GenerateLogId(ctx), data))
}

func writeInvalidQuery(ctx *gin.Context, err error) {
	logID := utils.GenerateLogId(ctx)
	res := response.ErrorResponse(http.StatusBadRequest, "Invalid location query", logID, err.Error())
	ctx.JSON(http.StatusBadRequest, res)
}

func handleServiceError(ctx *gin.Context, err error) {
	if errors.Is(err, locationservice.ErrInvalidPagination) ||
		errors.Is(err, locationservice.ErrInvalidParentID) ||
		errors.Is(err, locationservice.ErrInvalidSearchQuery) ||
		errors.Is(err, locationservice.ErrInvalidSearchLimit) ||
		errors.Is(err, locationservice.ErrInvalidLocationCode) {
		writeInvalidQuery(ctx, err)
		return
	}

	var upstreamErr *apiindonesia.UpstreamError
	if errors.As(err, &upstreamErr) && upstreamErr != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("Location upstream request failed: status=%d code=%s", upstreamErr.StatusCode, handlercommon.SafeUpstreamCode(upstreamErr.Code)))
		if upstreamErr.StatusCode == http.StatusNotFound {
			writeLocationNotFound(ctx)
			return
		}
		writeUpstreamError(ctx)
		return
	}

	if errors.Is(err, locationservice.ErrLocationClient) {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, "Location upstream client unavailable")
		writeLocationClientUnavailable(ctx)
		return
	}

	logger.WriteLogWithContext(ctx, logger.LogLevelError, "Location upstream request failed")
	writeUpstreamError(ctx)
}

func writeLocationNotFound(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	res := response.ErrorResponse(http.StatusNotFound, "Location not found", logID, "Location not found")
	ctx.JSON(http.StatusNotFound, res)
}

func writeLocationClientUnavailable(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	res := response.ErrorResponse(http.StatusServiceUnavailable, "Location service unavailable", logID, "Location service unavailable")
	ctx.JSON(http.StatusServiceUnavailable, res)
}

func writeUpstreamError(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	res := response.ErrorResponse(http.StatusBadGateway, "Location upstream unavailable", logID, "Location upstream service unavailable")
	ctx.JSON(http.StatusBadGateway, res)
}
