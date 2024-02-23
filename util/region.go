package util

import (
	"fmt"
	"regexp"
	"strings"
)

const ENDPOINT_REGEX_PATTERN = `^(?:http[s]?:\/\/)?([a-z-0-9]+)\.(?:sls|log)\.aliyuncs\.com$`

func ParseRegion(endpoint string) (string, error) {
	var re = regexp.MustCompile(ENDPOINT_REGEX_PATTERN)
	groups := re.FindStringSubmatch(endpoint)
	if groups == nil {
		return "", fmt.Errorf("invalid endpoint format: %s", endpoint)
	}
	region := groups[1]
	region = strings.TrimSuffix(region, "-intranet")
	region = strings.TrimSuffix(region, "-share")
	return region, nil
}
