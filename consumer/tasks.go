package consumerLibrary

import "github.com/aliyun/aliyun-log-go-sdk"

func (consumer *ShardConsumerWorker) consumerInitializeTask() string {
	checkpoint := consumer.client.getChcekPoint(consumer.shardId)
	if checkpoint != "" {
		consumer.consumerCheckPointTracker.setPersistentCheckPoint(checkpoint)
		return checkpoint
	}

	if consumer.client.option.CursorPosition == BEGIN_CURSOR {
		cursor := consumer.client.getCursor(consumer.shardId, "begin")
		return cursor
	}
	if consumer.client.option.CursorPosition == END_CURSOR {
		cursor := consumer.client.getCursor(consumer.shardId, "end")
		return cursor
	}
	if consumer.client.option.CursorPosition == SPECIAL_TIMER_CURSOR {
		cursor := consumer.client.getCursor(consumer.shardId, string(consumer.client.option.CursorStartTime))
		return cursor
	}
	return ""
}

func (consumer *ShardConsumerWorker) consumerFetchTask() (*sls.LogGroupList, string) {
	logGroup, next_cursor := consumer.client.pullLogs(consumer.shardId, consumer.nextFetchCursor)
	return logGroup, next_cursor
}

func (consumer *ShardConsumerWorker) consumerProcessTask() {
	// If the user's consumption function reports a panic error, it will be captured and exited.
	defer func() {
		if r := recover(); r != nil {
			Error.Printf("get panic in your process function : %v", r)
		}
	}()
	consumer.process(consumer.shardId, consumer.lastFetchLogGroupList)
	consumer.consumerCheckPointTracker.flushCheck()
}
