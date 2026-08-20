package earthquakecache

import (
	"time"

	"kota-siaga/utils"
)

const defaultTTL = time.Hour

func TTL() time.Duration {
	return utils.DurationFromEnv([]string{"EARTHQUAKE_CACHE_TTL"}, defaultTTL)
}

func Key() string {
	return "earthquake:latest"
}
