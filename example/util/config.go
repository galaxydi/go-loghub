package util

import sls "github.com/aliyun/aliyun-log-go-sdk"

// Project define Project for test
var Project = &sls.LogProject{
	Name:            "test-project",
	Endpoint:        "cn-hangzhou.log.aliyuncs.com",
	AccessKeyID:     "",
	AccessKeySecret: "",
}
