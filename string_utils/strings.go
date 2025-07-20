package string_utils

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/tidwall/gjson"
)

// ReplaceTags receive a string and a map[string]interface{} and call GetTagsFromString on the string
// for each string that GetTagsFromString returns it will try to get the value from the map and add it to the string
func ReplaceTags(ctx context.Context, str string, input []byte) (string, error) {
	newStr := str
	for _, tag := range GetTagsFromString(str) {
		jsonStr := gjson.GetBytes(input, tag)
		if jsonStr.Exists() {
			newStr = strings.ReplaceAll(newStr, "{"+tag+"}", jsonStr.String())
		} else {
			log.Printf("Tag not found: %s for str %s", tag, str)
			return "", errors.New("tag not found")
		}
	}
	return newStr, nil
}

// GetTagsFromString obtain all the values wrapped by '{' and '}' from the string
func GetTagsFromString(str string) []string {
	var tags []string
	for _, tag := range strings.Split(str, "{") {
		if strings.Contains(tag, "}") {
			tags = append(tags, tag[:strings.Index(tag, "}")])
		}
	}
	return tags
}
