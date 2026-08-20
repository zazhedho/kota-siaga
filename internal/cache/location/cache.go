package locationcache

import (
	"fmt"
	"time"

	"kota-siaga/utils"
)

const defaultTTL = 24 * time.Hour

func TTL() time.Duration {
	return utils.DurationFromEnv([]string{"LOCATION_CACHE_TTL"}, defaultTTL)
}

func ProvinceKey(page, perPage int) string {
	return fmt.Sprintf("locations:province:%d:%d", page, perPage)
}

func CityKey(provinceID string, page, perPage int) string {
	return fmt.Sprintf("locations:city:%s:%d:%d", provinceID, page, perPage)
}

func DistrictKey(regencyID string, page, perPage int) string {
	return fmt.Sprintf("locations:district:%s:%d:%d", regencyID, page, perPage)
}

func VillageKey(districtID string, page, perPage int) string {
	return fmt.Sprintf("locations:village:%s:%d:%d", districtID, page, perPage)
}
