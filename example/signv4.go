package main

import (
	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/util"
)

func CreateSignV4Client() {
	accessKeyId, accessKeySecret := "", ""           // replace with your access key and secret
	endpoint := "cn-hangzhou-share.log.aliyuncs.com" // replace with your endpoint

	client := sls.CreateNormalInterfaceV2(endpoint,
		sls.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, ""))
	region, err := util.ParseRegion(endpoint) // parse region from endpoint
	if err != nil {
		panic(err)
	}
	client.SetRegion(region)          // region must be set if using signature v4
	client.SetAuthVersion(sls.AuthV4) // set signature v4

	client.GetProject("example-project") // call client API
}
