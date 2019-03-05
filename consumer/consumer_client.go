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
		if slsError, ok := err.(sls.Error); ok {
			if slsError.Code == "ConsumerGroupAlreadyExist" {
				Info.Printf("New consumer %v join the consumer group %v ", consumer.option.ConsumerName, consumer.option.ConsumerGroupName)
			} else {
				Warning.Println(err)
			}
		}
	}
}

func (consumer *ConsumerClient) heartBeat(heart []int) ([]int, error) {
	heldShard, err := consumer.client.HeartBeat(consumer.option.Project, consumer.option.Logstore, consumer.option.ConsumerGroupName, consumer.option.ConsumerName, heart)
	return heldShard, err
}

func (consumer *ConsumerClient) updateCheckPoint(shardId int, checkpoint string, forceSucess bool) error {
	err := consumer.client.UpdateCheckpoint(consumer.option.Project, consumer.option.Logstore, consumer.option.ConsumerGroupName, consumer.option.ConsumerName, shardId, checkpoint, forceSucess)
	if err != nil {
		return err
	}
	return nil
}

// get a single shard checkpoint, if not，return ""
func (consumer *ConsumerClient) getCheckPoint(shardId int) string {
	checkPonitList := consumer.retryGetCheckPoint(shardId)
	for _, checkPoint := range checkPonitList {
		if checkPoint.ShardID == shardId {
			return checkPoint.CheckPoint
		}
	}
	return ""
}

// If a checkpoint error is reported, the shard will remain asynchronous and will not affect the consumption of other shards.
func (consumer *ConsumerClient) retryGetCheckPoint(shardId int) (checkPonitList []*sls.ConsumerGroupCheckPoint) {
	for {
		checkPonitList, err := consumer.client.GetCheckpoint(consumer.option.Project, consumer.option.Logstore, consumer.consumerGroup.ConsumerGroupName)
		if err != nil {
			Info.Printf("shard %v Get checkpoint gets errors, starts to try again, error : %v", shardId, err)
			time.Sleep(1 * time.Second)
		} else {
			return checkPonitList
		}
	}
}

func (consumer *ConsumerClient) getCursor(shardId int, from string) (string, error) {
	cursor, err := consumer.client.GetCursor(consumer.option.Project, consumer.option.Logstore, shardId, from)
	return cursor, err
}

func (consumer *ConsumerClient) pullLogs(shardId int, cursor string) (gl *sls.LogGroupList, nextCursor string) {
	for retry := 0; retry < 3; retry++ {
		gl, nextCursor, err := consumer.client.PullLogs(consumer.option.Project, consumer.option.Logstore, shardId, cursor, "", consumer.option.MaxFetchLogGroupCount)
		if err != nil {
			slsError, ok := err.(sls.Error)
			if ok {
				if slsError.HTTPCode == 403 {
					Info.Printf("shard %v Get checkpoint gets errors, starts to try again, error : %v", shardId, slsError)
					time.Sleep(5 * time.Second)
				} else {
					Info.Printf("shard %v Get checkpoint gets errors, starts to try again, error : %v", shardId, slsError)
					time.Sleep(200 * time.Millisecond)
				}
			} else {
				Info.Println("xxx.logger ...") //TODO log
			}
		} else {
			return gl, nextCursor
		}
	}
	// If you can't retry the log three times, it will return to empty list and start pulling the log cursor,
	// so that next time you will come in and pull the function again, which is equivalent to a dead cycle.
	return gl, "PullLogFailed"
}
