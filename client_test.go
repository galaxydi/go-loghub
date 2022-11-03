package sls

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRegion(t *testing.T) {
	assert.Equal(t, parseRegionFromEndpoint(
		"http://cn-hangzhou-devcommon-intranet.sls.aliyuncs.com"),
		"cn-hangzhou-devcommon")
	assert.Equal(t, parseRegionFromEndpoint(
		"cn-hangzhou-share.log.aliyuncs.com"),
		"cn-hangzhou")
	assert.Equal(t, parseRegionFromEndpoint(
		"https://cn-shanghai.log.aliyuncs.com"),
		"cn-shanghai")
	assert.Equal(t, parseRegionFromEndpoint("cn-hangzhou-stg.log.aliyuncs.com"),
		"cn-hangzhou-stg")
	assert.Equal(t, parseRegionFromEndpoint(
		"http://cn-zhangjiakou-stg-intranet.log.aliyuncs.com"),
		"cn-zhangjiakou-stg")
	assert.Equal(t, parseRegionFromEndpoint(""), "")
	assert.Equal(t, parseRegionFromEndpoint("192.168.1.1"), "")
	assert.Equal(t, parseRegionFromEndpoint("http://192.168.1.1"), "")
}
