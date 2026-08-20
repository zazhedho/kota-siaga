package hospitalcache

import (
	"fmt"
	"time"

	"kota-siaga/utils"
)

const defaultTTL = 24 * time.Hour

func TTL() time.Duration {
	return utils.DurationFromEnv([]string{"HOSPITAL_CACHE_TTL"}, defaultTTL)
}

func Key(kabupatenID string, page, perPage int) string {
	return fmt.Sprintf("hospital:%s:%d:%d", kabupatenID, page, perPage)
}
