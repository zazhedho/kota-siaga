package middlewares

import (
	"fmt"
	"kota-siaga/pkg/logger"
	"kota-siaga/pkg/response"
	"kota-siaga/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler(c *gin.Context, err any) {
	logId := utils.GenerateLogId(c)
	logger.WriteLogWithContext(c, logger.LogLevelPanic, fmt.Sprintf("RECOVERY; Error: %+v;", err))

	res := response.InternalServerError(logId)
	c.AbortWithStatusJSON(http.StatusInternalServerError, res)
}
