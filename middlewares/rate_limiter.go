package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"kota-siaga/pkg/logger"
	"kota-siaga/pkg/messages"
	"kota-siaga/pkg/response"
	"kota-siaga/utils"
)

const rateLimitScript = `
local current = redis.call("INCR", KEYS[1])
if current == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current`

// IPRateLimitMiddleware applies a Redis-backed rate limit per client IP.
func IPRateLimitMiddleware(redisClient *redis.Client, prefix string, limit int, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if redisClient == nil || limit <= 0 || window <= 0 {
			ctx.Next()
			return
		}

		logID := utils.GenerateLogId(ctx)
		logPrefix := fmt.Sprintf("[RateLimiter][%s]", prefix)
		key := fmt.Sprintf("rate_limit:%s:%s", prefix, ctx.ClientIP())

		reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
		defer cancel()

		current, err := redisClient.Eval(reqCtx, rateLimitScript, []string{key}, window.Milliseconds()).Int64()
		if err != nil {
			logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; redis.Eval error: %v", logPrefix, err))
			ctx.Next()
			return
		}

		if current > int64(limit) {
			ttl, _ := redisClient.TTL(reqCtx, key).Result()
			if ttl > 0 {
				ctx.Header("Retry-After", strconv.Itoa(int(ttl.Seconds())))
			}

			res := response.Response(http.StatusTooManyRequests, messages.MsgSomethingWrong, logID, nil)
			res.Error = response.Errors{
				Code:    http.StatusTooManyRequests,
				Message: "Too many requests from this IP, please try again later",
			}
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, res)
			return
		}

		ctx.Next()
	}
}
