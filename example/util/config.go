package util

import (
	"github.com/aliyun/aliyun-log-go-sdk"
)

// When you use the file under example, please configure the required variables here.
// Project define Project for test
var (
	ProjectName     = "test-project"
	Endpoint        = "<endpoint>"
	LogStoreName    = "test-logstore"
	AccessKeyID     = "<accessKeyId>"
	AccessKeySecret = "<accessKeySecret>"
	Client          *sls.Client
)

// You can get the variable from the environment variable, or fill in the required configuration directly in the init function.
func init() {
	ProjectName = "your project name"
	AccessKeyID = "your ak id"
	AccessKeySecret = "your ak secret"
	Endpoint = "your endpoint" // just like cn-hangzhou.log.aliyuncs.com
	LogStoreName = "demo"

	Client = new(sls.Client)
	Client.Endpoint = Endpoint
	Client.AccessKeyID = AccessKeyID
	Client.AccessKeySecret = AccessKeySecret
}
