package consumerLibrary

import (
	"fmt"
	"github.com/aliyun/aliyun-log-go-sdk"
)

type ConsumerClient struct{
	LogHubConfig
	*sls.Client
	sls.ConsumerGroup
}

func InitConsumerClient(option LogHubConfig)*ConsumerClient{
	client := &sls.Client{
		Endpoint:option.Endpoint,
		AccessKeyID:option.AccessKeyID,
		AccessKeySecret:option.AccessKeySecret,
		SecurityToken:option.SecurityToken,
		// TODO  UserAgent 是否用添加？
	}
	consumerGroup := sls.ConsumerGroup{
		option.MConsumerGroupName,
		option.HeartbeatInterval*2,
		option.InOrder,

	}
	consumerClient := &ConsumerClient{
		option,
		client,
		consumerGroup,
	}
	fmt.Println("jianlaile a ",consumerClient.Project)
	return consumerClient
}



func(consumer *ConsumerClient) McreateConsumerGroup(){
	err := consumer.CreateConsumerGroup(consumer.Project,consumer.Logstore,consumer.ConsumerGroup)
	if err != nil{
		Info.Println(err)
	}
}

func (consumer *ConsumerClient) MheartBeat(heart []int) []int {

	fmt.Println(consumer.Project,consumer.Logstore,consumer.ConsumerGroup.ConsumerGroupName,consumer.ConsumerName,heart)
	held_shard,err:=consumer.HeartBeat(consumer.Project,consumer.Logstore,consumer.ConsumerGroup.ConsumerGroupName,consumer.ConsumerName,heart)
	if err != nil {
		Info.Println(err)
	}
	return held_shard
}

func (consumer *ConsumerClient) MupdateCheckPoint(shardId int,checkpoint string,forceSucess bool){
	err := consumer.UpdateCheckpoint(consumer.Project,consumer.Logstore,consumer.ConsumerGroup.ConsumerGroupName,consumer.ConsumerName,shardId,checkpoint,forceSucess)
	if err != nil{
		Info.Println(err)
	}
}
// TODO 这个获得的是当前logstore 下面的所有分区的检查点,我写成只获取一个的,获取不到返回空字符串
func (consumer *ConsumerClient) MgetChcekPoint(shardId int) string {
	checkPonitList,err:=consumer.GetCheckpoint(consumer.Project,consumer.Logstore,consumer.ConsumerGroup.ConsumerGroupName)
	if err != nil{
		Info.Println(err)
	}
	for _,x:= range checkPonitList{
		if x.ShardID == shardId {
			return x.CheckPoint // TODO 问题，如果有这个分区一样没有检查点，是不是也为空字符串？
		}
	}
	return ""
}

func (consumer *ConsumerClient) MgetCursor(shardId int) (cursor string) {
	cursor,err:=consumer.GetCursor(consumer.Project,consumer.Logstore,shardId,consumer.CursorStarttime)
	if err != nil{
		Info.Println(err)
	}
	return cursor
}

func (consumer *ConsumerClient) MgetBeginCursor(shardId int)string{
	cursor,err:=consumer.GetCursor(consumer.Project,consumer.Logstore,shardId,"begin")
	if err != nil{
		Info.Println(err)
	}
	return cursor
}

func (consumer *ConsumerClient) MgetEndCursor(shardId int) string{
	cursor,err:=consumer.GetCursor(consumer.Project,consumer.Logstore,shardId,"end")
	if err != nil{
		Info.Println(err)
	}
	return cursor
}

func (consumer *ConsumerClient) MpullLogs(shardId int,cursor string)(*sls.LogGroupList,string){
	gl,next_cursor,err:=consumer.PullLogs(consumer.Project,consumer.Logstore,shardId,cursor,"",consumer.MaxFetchLogGroupSize)
	if err != nil{
		Info.Println(err)
	}
	return gl ,next_cursor
}























