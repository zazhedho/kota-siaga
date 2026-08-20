package weathercache

import (
	"fmt"
	"time"

	"kota-siaga/utils"
)

const defaultTTL = 6 * time.Hour

func TTL() time.Duration {
	return utils.DurationFromEnv([]string{"WEATHER_CACHE_TTL"}, defaultTTL)
}

func Key(adm4 string) string {
	return fmt.Sprintf("weather:%s", adm4)
}
