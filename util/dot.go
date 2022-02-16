package util

import "regexp"

var (
	escapeTarget = regexp.MustCompile(`[/\$\.]`)
)

func AsDotID(v string) string {
	return string(escapeTarget.ReplaceAll([]byte(v), []byte("_")))
}
