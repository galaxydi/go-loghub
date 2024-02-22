package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRegion(t *testing.T) {
	region, err := ParseRegion("cn-qingdao-acdr-ut-1-intranet.log.aliyuncs.com")
	assert.NoError(t, err)
	assert.Equal(t, "cn-qingdao-acdr-ut-1", region)

	region, err = ParseRegion("cn-chengdu-acdr-ut-1-intranet.log.aliyuncs.com")
	assert.NoError(t, err)
	assert.Equal(t, "cn-chengdu-acdr-ut-1", region)

	region, err = ParseRegion("http://cn-hangzhou-share.log.aliyuncs.com")
	assert.NoError(t, err)
	assert.Equal(t, "cn-hangzhou", region)

	region, err = ParseRegion("https://cn-hangzhou.log.aliyuncs.com")
	assert.NoError(t, err)
	assert.Equal(t, "cn-hangzhou", region)

	region, err = ParseRegion("ap-southease-1-intranet.log.aliyuncs.com")
	assert.NoError(t, err)
	assert.Equal(t, "ap-southease-1", region)

	region, err = ParseRegion("cn-shanghai-corp-share.sls.aliyuncs.com")
	assert.NoError(t, err)
	assert.Equal(t, "cn-shanghai-corp", region)

	_, err = ParseRegion("sls.aliyuncs.com")
	assert.Error(t, err)
}
