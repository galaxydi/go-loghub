package main

import (
	"fmt"
	"time"

	"github.com/aliyun/aliyun-log-go-sdk/example/util"
)

func main() {
	logstore, err := util.Client.GetLogStore(util.ProjectName, util.LogStoreName)
	if err != nil {
		panic(err)
	}
	fmt.Println("get logstore successfully:", logstore.Name)

	resp, err := logstore.GetLogs("", time.Now().Unix()-10, time.Now().Unix(), "*", 100, 1, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp)
}
