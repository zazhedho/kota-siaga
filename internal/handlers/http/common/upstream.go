package handlercommon

import "strings"

const maxLoggedUpstreamCodeLength = 64

func SafeUpstreamCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" || len(code) > maxLoggedUpstreamCodeLength {
		return "UNKNOWN"
	}
	for i := 0; i < len(code); i++ {
		char := code[i]
		if (char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') && char != '_' {
			return "UPSTREAM_ERROR"
		}
	}
	return code
}
