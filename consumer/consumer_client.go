package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"time"
)

type ConsumerClient struct {
	LogHubConfig
	*sls.Client
	sls.ConsumerGroup
}

func InitConsumerClient(option LogHubConfig) *ConsumerClient {
	// Setting configuration defaults
	if option.HeartbeatInterval == 0 {
		option.HeartbeatInterval = 20
	}
	if option.DataFetchInterval == 0 {
		option.DataFetchInterval = 2
	}
	if option.MaxFetchLogGroupSize == 0 {
		option.MaxFetchLogGroupSize = 1000
	}
	client := &sls.Client{
		Endpoint:        option.Endpoint,
		AccessKeyID:     option.AccessKeyID,
		AccessKeySecret: option.AccessKeySecret,
		SecurityToken:   option.SecurityToken,
		// TODO  UserAgent Whether to add ？
	}
	consumerGroup := sls.ConsumerGroup{
		option.MConsumerGroupName,
		option.HeartbeatInterval * 2,
		option.InOrder,
	}
	consumerClient := &ConsumerClient{
		option,
		client,
		consumerGroup,
	}

	return consumerClient
}

func (consumer *ConsumerClient) mCreateConsumerGroup() {
	err := consumer.CreateConsumerGroup(consumer.Project, consumer.Logstore, consumer.ConsumerGroup)
	if err != nil {
		if x, ok := err.(sls.Error); ok {
			if x.Code == "ConsumerGroupAlreadyExist" {
				Info.Printf("New consumer %v join the consumer group %v ", consumer.ConsumerName, consumer.ConsumerGroupName)
			} else {
				Warning.Println(err)
			}
		}
	}
}

func (consumer *ConsumerClient) mHeartBeat(heart []int) []int {
	held_shard, err := consumer.HeartBeat(consumer.Project, consumer.Logstore, consumer.ConsumerGroup.ConsumerGroupName, consumer.ConsumerName, heart)
	if err != nil {
		Warning.Println(err)
	}
	return held_shard
}

func (consumer *ConsumerClient) mUpdateCheckPoint(shardId int, checkpoint string, forceSucess bool) {
	err := consumer.UpdateCheckpoint(consumer.Project, consumer.Logstore, consumer.ConsumerGroup.ConsumerGroupName, consumer.ConsumerName, shardId, checkpoint, forceSucess)
	if err != nil {
		Warning.Println(err)
	}
}

// get a single shard checkpoint, if not，return ""
func (consumer *ConsumerClient) mGetChcekPoint(shardId int) string {
	checkPonitList, err := consumer.GetCheckpoint(consumer.Project, consumer.Logstore, consumer.ConsumerGroup.ConsumerGroupName)
	if err != nil {
		Warning.Println(err)
	}
	for _, x := range checkPonitList {
		if x.ShardID == shardId {
			return x.CheckPoint
		}
	}
	return ""
}

func (consumer *ConsumerClient) mGetCursor(shardId int) (cursor string) {
	tm, _ := time.Parse("2006-01-02 03:04:05", consumer.CursorStartTime)
	timeUnix := tm.Unix()
	cursor, err := consumer.GetCursor(consumer.Project, consumer.Logstore, shardId, string(timeUnix))
	if err != nil {
		Warning.Println(err)
	}
	return cursor
}

func (consumer *ConsumerClient) mGetBeginCursor(shardId int) string {
	cursor, err := consumer.GetCursor(consumer.Project, consumer.Logstore, shardId, "begin")
	if err != nil {
		Warning.Println(err)
	}
	return cursor
}

func (consumer *ConsumerClient) mGetEndCursor(shardId int) string {
	cursor, err := consumer.GetCursor(consumer.Project, consumer.Logstore, shardId, "end")
	if err != nil {
		Warning.Println(err)
	}
	return cursor
}

// TODO if error code = InvalidCursor , get cursor just like python ?
func (consumer *ConsumerClient) mPullLogs(shardId int, cursor string) (gl *sls.LogGroupList, nextCursor string) {
	for retry := 0; retry < 3; retry++ {
		gl, nextCursor, err := consumer.PullLogs(consumer.Project, consumer.Logstore, shardId, cursor, "", consumer.MaxFetchLogGroupSize)
		if err != nil {
			if a, ok := err.(sls.Error); ok {
				if a.HTTPCode == 500 {
					Info.Printf("Server gets 500 errors, starts to try again, try times %v", retry)
				} else {
					Error.Println(err)
				}
			}
		} else {
			return gl, nextCursor
		}
	}
	return
}
