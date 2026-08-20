package warningcache

import (
	"strings"
	"time"

	"kota-siaga/utils"
)

const defaultTTL = time.Hour

func TTL() time.Duration {
	return utils.DurationFromEnv([]string{"WARNING_CACHE_TTL"}, defaultTTL)
}

func Key(province string) string {
	return "warning:" + strings.ToLower(strings.TrimSpace(province))
}
