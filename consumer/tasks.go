package consumerLibrary

import "github.com/aliyun/aliyun-log-go-sdk"

func (consumer *ShardConsumerWorker) consumerInitializeTask() string {

	checkpoint := consumer.mGetChcekPoint(consumer.ShardId)
	if checkpoint != "" {
		consumer.setPersistentCheckPoint(checkpoint)
		return checkpoint
	}

	if consumer.CursorPosition == BEGIN_CURSOR {
		cursor := consumer.mGetBeginCursor(consumer.ShardId)
		return cursor
	}
	if consumer.CursorPosition == END_CURSOR {
		cursor := consumer.mGetEndCursor(consumer.ShardId)
		return cursor
	}
	if consumer.CursorPosition == SPECIAL_TIMER_CURSOR {
		cursor := consumer.mGetCursor(consumer.ShardId)
		return cursor
	}
	return ""
}

func (consumer *ShardConsumerWorker) consumerFetchTask() (*sls.LogGroupList, string) {
	logGroup, next_cursor := consumer.mPullLogs(consumer.ShardId, consumer.NextFetchCursor)
	return logGroup, next_cursor
}

func (consumer *ShardConsumerWorker) consumerProcessTask() {
	// If the user's consumption function reports a panic error, it will be captured and exited.
	defer func() {
		if r := recover(); r != nil {
			Error.Printf("get panic in your process function : %v", r)
		}
	}()
	consumer.Process(consumer.ShardId, consumer.LastFetchLogGroup)
	consumer.flushCheck()
}
