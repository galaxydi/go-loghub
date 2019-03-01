package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"time"
)

type ConsumerClient struct {
	option        LogHubConfig
	client        *sls.Client
	consumerGroup sls.ConsumerGroup
}

func initConsumerClient(option LogHubConfig) *ConsumerClient {
	// Setting configuration defaults
	if option.HeartbeatIntervalInSecond == 0 {
		option.HeartbeatIntervalInSecond = 20
	}
	if option.DataFetchInterval == 0 {
		option.DataFetchInterval = 2
	}
	if option.MaxFetchLogGroupCount == 0 {
		option.MaxFetchLogGroupCount = 1000
	}
	client := &sls.Client{
		Endpoint:        option.Endpoint,
		AccessKeyID:     option.AccessKeyID,
		AccessKeySecret: option.AccessKeySecret,
		// SecurityToken:   option.SecurityToken,
		UserAgent: option.ConsumerGroupName + "_" + option.ConsumerName,
	}
	consumerGroup := sls.ConsumerGroup{
		option.ConsumerGroupName,
		option.HeartbeatIntervalInSecond * 2,
		option.InOrder,
	}
	consumerClient := &ConsumerClient{
		option,
		client,
		consumerGroup,
	}

	return consumerClient
}

func (consumer *ConsumerClient) createConsumerGroup() {
	err := consumer.client.CreateConsumerGroup(consumer.option.Project, consumer.option.Logstore, consumer.consumerGroup)
	if err != nil {
		if x, ok := err.(sls.Error); ok {
			if x.Code == "ConsumerGroupAlreadyExist" {
				Info.Printf("New consumer %v join the consumer group %v ", consumer.option.ConsumerName, consumer.option.ConsumerGroupName)
			} else {
				Warning.Println(err)
			}
		}
	}
}

func (consumer *ConsumerClient) heartBeat(heart []int) []int {
	heldShard, err := consumer.client.HeartBeat(consumer.option.Project, consumer.option.Logstore, consumer.option.ConsumerGroupName, consumer.option.ConsumerName, heart)
	if err != nil {
		Warning.Println(err)
	}
	return heldShard
}

func (consumer *ConsumerClient) updateCheckPoint(shardId int, checkpoint string, forceSucess bool) {
	err := consumer.client.UpdateCheckpoint(consumer.option.Project, consumer.option.Logstore, consumer.option.ConsumerGroupName, consumer.option.ConsumerName, shardId, checkpoint, forceSucess)
	if err != nil {
		Warning.Println(err)
	}
}

// get a single shard checkpoint, if not，return ""
func (consumer *ConsumerClient) getChcekPoint(shardId int) string {
	checkPonitList := consumer.retryGetCheckPoint(shardId)
	for _, x := range checkPonitList {
		if x.ShardID == shardId {
			return x.CheckPoint
		}
	}
	return ""
}

// If a checkpoint error is reported, the shard will remain asynchronous and will not affect the consumption of other shards.
func (consumer *ConsumerClient) retryGetCheckPoint(shardId int) (checkPonitList []*sls.ConsumerGroupCheckPoint) {
	for {
		checkPonitList, err := consumer.client.GetCheckpoint(consumer.option.Project, consumer.option.Logstore, consumer.consumerGroup.ConsumerGroupName)
		if err != nil {
			if a, ok := err.(sls.Error); ok {
				Info.Printf("shard %v Get checkpoint gets %v errors, starts to try again", shardId, a.HTTPCode)
				time.Sleep(1 * time.Second)
			}
		} else {
			return checkPonitList
		}
	}
}

func (consumer *ConsumerClient) getCursor(shardId int, from string) (cursor string) {
	cursor, err := consumer.client.GetCursor(consumer.option.Project, consumer.option.Logstore, shardId, from)
	if err != nil {
		Warning.Println(err)
	}
	return cursor
}

func (consumer *ConsumerClient) pullLogs(shardId int, cursor string) (gl *sls.LogGroupList, nextCursor string) {
	for retry := 0; retry < 3; retry++ {
		gl, nextCursor, err := consumer.client.PullLogs(consumer.option.Project, consumer.option.Logstore, shardId, cursor, "", consumer.option.MaxFetchLogGroupCount)
		if err != nil {
			if a, ok := err.(sls.Error); ok {
				if a.HTTPCode == 500 {
					Info.Println(err)
					Info.Printf("Server gets 500 errors, starts to try again, try times %v", retry)
					time.Sleep(1 * time.Second)
				} else {
					Warning.Println(err)
				}
			}
		} else {
			return gl, nextCursor
		}
	}
	// If you can't retry the log three times, it will return to empty list and start pulling the log cursor,
	// so that next time you will come in and pull the function again, which is equivalent to a dead cycle.
	return gl, cursor
}
