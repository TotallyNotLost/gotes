package markdown

import (
	"regexp"
)

func RemoveMetadata(md string, key string) string {
	r, _ := regexp.Compile("\\[_metadata_:" + key + "\\]:# \"[^\"]*\"")

	return r.ReplaceAllString(md, "")
}
