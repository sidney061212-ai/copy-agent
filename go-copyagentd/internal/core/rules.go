package core

import (
	"regexp"
	"strings"
)

var copyCommand = regexp.MustCompile(`(?i)^(@\S+\s+)?(复制|拷贝|copy|cp)(?:\s*[:：]\s*|\s+)`)
var bareCopyCommand = regexp.MustCompile(`(?i)^(@\S+\s+)?(复制|拷贝|copy|cp)\s*$`)

func ExtractCopyText(text string) string {
	trimmed := strings.TrimSpace(text)
	if bareCopyCommand.MatchString(trimmed) {
		return ""
	}
	return copyCommand.ReplaceAllString(trimmed, "")
}

func ValidText(text string) bool {
	return strings.TrimSpace(text) != ""
}
