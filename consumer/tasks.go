package consumerLibrary

import "github.com/aliyun/aliyun-log-go-sdk"

func (consumer *ShardConsumerWorker) ConsumerInitializeTask() string {

	checkpoint := consumer.MgetChcekPoint(consumer.ShardId)
	if checkpoint != "" {
		consumer.SetPersistentCheckPoint(checkpoint)
		return checkpoint
	}

	if consumer.CursorPosition == BEGIN_CURSOR {
		cursor := consumer.MgetBeginCursor(consumer.ShardId)
		return cursor
	}
	if consumer.CursorPosition == END_CURSOR {
		cursor := consumer.MgetEndCursor(consumer.ShardId)
		return cursor
	}
	if consumer.CursorPosition == SPECIAL_TIMER_CURSOR {
		cursor := consumer.MgetCursor(consumer.ShardId)
		return cursor
	}
	return ""
}

func (consumer *ShardConsumerWorker) ConsumerFetchTask() (*sls.LogGroupList, string) {
	logGroup, next_cursor := consumer.MpullLogs(consumer.ShardId, consumer.NextFetchCursor)
	return logGroup, next_cursor
}

func (consumer *ShardConsumerWorker) ConsumerProcessTask() {
	// If the user's consumption function reports a panic error, it will be captured and exited.
	defer func() {
		if r := recover(); r != nil {
			Error.Printf("get panic in your process function : %v", r)
		}
	}()
	consumer.Process(consumer.ShardId, consumer.LastFetchLogGroup)
	consumer.FlushCheck()
}
