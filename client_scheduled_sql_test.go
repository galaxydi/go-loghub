package sls

import (
	"fmt"
	"os"
	"testing"
)

func TestList(t *testing.T) {
	AccessKeyID := os.Getenv("ALICLOUD_ACCESS_KEY")
	AccessKeySecret := os.Getenv("ALICLOUD_SECRET_KEY")
	Endpoint := "cn-hangzhou.log.aliyuncs.com"
	client := CreateNormalInterface(Endpoint, AccessKeyID, AccessKeySecret, "")
	scheduledSQL, err := client.GetScheduledSQL("", "")
	if err != nil {
		fmt.Printf("%v", err)
	}
	fmt.Printf("%v", scheduledSQL)
}
