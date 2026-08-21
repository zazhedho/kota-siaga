package hospitalcache

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"kota-siaga/utils"
)

const defaultTTL = 24 * time.Hour

func TTL() time.Duration {
	return utils.DurationFromEnv([]string{"HOSPITAL_CACHE_TTL"}, defaultTTL)
}

func Key(kabupatenID, search string, page, perPage int) string {
	search = url.QueryEscape(strings.ToLower(strings.TrimSpace(search)))
	if search != "" {
		return fmt.Sprintf("hospital:search:%s:%s:%d:%d", kabupatenID, search, page, perPage)
	}
	return fmt.Sprintf("hospital:%s:%s:%d:%d", kabupatenID, search, page, perPage)
}
