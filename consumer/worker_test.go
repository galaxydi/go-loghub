package consumerLibrary

import (
	"fmt"
	"testing"

	sls "github.com/aliyun/aliyun-log-go-sdk"
)

func TestStartAndStop(t *testing.T) {
	option := LogHubConfig{
		Endpoint:          "",
		AccessKeyID:       "",
		AccessKeySecret:   "",
		Project:           "",
		Logstore:          "",
		ConsumerGroupName: "",
		ConsumerName:      "",
		// This options is used for initialization, will be ignored once consumer group is created and each shard has been started to be consumed.
		// Could be "begin", "end", "specific time format in time stamp", it's log receiving time.
		CursorPosition: BEGIN_CURSOR,
	}

	worker := InitConsumerWorker(option, process)

	worker.Start()
	worker.StopAndWait()
}

func process(shardId int, logGroupList *sls.LogGroupList) string {
	fmt.Printf("shardId %d processing works sucess", shardId)
	return ""
}
