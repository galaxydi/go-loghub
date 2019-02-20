package consumerLibrary

import "github.com/aliyun/aliyun-log-go-sdk"

func (consumer *ShardConsumerWorker)ConsumerInitializeTask() string{

	checkpoint := consumer.MgetChcekPoint(consumer.ShardId)
	if checkpoint != ""{
		consumer.SetPersistentCheckPoint(checkpoint)
		return checkpoint
	}

	if consumer.CursorPosition == BEGIN_CURSOR{
		cursor := consumer.MgetBeginCursor(consumer.ShardId)
		return cursor
	}
	if consumer.CursorPosition == END_CURSOR{
		cursor := consumer.MgetEndCursor(consumer.ShardId)
		return cursor
	}
	if consumer.CursorPosition == SPECIAL_TIMER_CURSOR {
		cursor := consumer.MgetCursor(consumer.ShardId)
		return cursor
	}
	return ""
}


func (consumer *ShardConsumerWorker) ConsumerFetchTask()(*sls.LogGroupList, string){
	logGroup,next_cursor := consumer.MpullLogs(consumer.ShardId,consumer.NextFetchCursor)
	return logGroup,next_cursor
}

func (consumer *ShardConsumerWorker) ConsumerProcessTask(){
	// TODO 消费完以后 刷先检查点,如果距离上次持久化检查超过60s
	// TODO 需要一个回退的检查点
	consumer.Process(consumer.ShardId,consumer.LastFetchLogGroup)
	consumer.FlushCheck()
}