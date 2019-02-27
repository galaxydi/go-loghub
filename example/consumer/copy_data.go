package main

import (
	"fmt"
	"github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/consumer"
)

var option consumerLibrary.LogHubConfig
var client sls.Client
var logStore *sls.LogStore

func main() {
	option = consumerLibrary.LogHubConfig{
		Endpoint:           "",
		AccessKeyID:        "",
		AccessKeySecret:    "",
		Project:            "",
		Logstore:           "",
		ConsumerGroupName: "",
		ConsumerName:       "",
		// This options is used for initialization, will be ignored once consumer group is created and each shard has been started to be consumed.
		// Could be "begin", "end", "specific time format in time stamp", it's log receiving time.
		CursorPosition: consumerLibrary.BEGIN_CURSOR,
	}
	client = sls.Client{
		Endpoint:        option.Endpoint,
		AccessKeyID:     option.AccessKeyID,
		AccessKeySecret: option.AccessKeySecret,
	}
	logStore = &sls.LogStore{
		Name:       "copy-logstore",
		TTL:        1,
		ShardCount: 2,
	}
	err := client.CreateLogStoreV2(option.Project, logStore)
	if err != nil {
		fmt.Println(err)
	}
	consumer := consumerLibrary.InitConsumerWorker(option, process)
	consumer.Start()
}

func process(shardId int, logGroupList *sls.LogGroupList) {
	for _, logGroup := range logGroupList.LogGroups {
		err2 := client.PutLogs(option.Project, "copy-logstore", logGroup)
		if err2 != nil {
			fmt.Println(err2)
		}
	}
	fmt.Println("shardId %v processing works sucess", shardId)
}
