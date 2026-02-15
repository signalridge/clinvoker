package policy

import (
	"strconv"
	"strings"
)

func parseInt64(v string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(v), 10, 64)
}
