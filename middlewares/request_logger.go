package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"kota-siaga/pkg/logger"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		ctx.Next()

		status := ctx.Writer.Status()
		latency := time.Since(start)
		path := ctx.FullPath()
		if path == "" {
			path = ctx.Request.URL.Path
		}

		msg := fmt.Sprintf(
			"[Request]; %s %s; status=%d; latency_ms=%d; ip=%s",
			ctx.Request.Method,
			path,
			status,
			latency.Milliseconds(),
			ctx.ClientIP(),
		)

		switch {
		case status >= http.StatusInternalServerError:
			logger.WriteLogWithContext(ctx, logger.LogLevelError, msg)
		case status >= http.StatusBadRequest:
			logger.WriteLogWithContext(ctx, logger.LogLevelWarn, msg)
		default:
			logger.WriteLogWithContext(ctx, logger.LogLevelInfo, msg)
		}
	}
}
