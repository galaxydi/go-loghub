package consumerLibrary

import (
	"errors"
	"github.com/aliyun/aliyun-log-go-sdk"
)

func (consumer *ShardConsumerWorker) consumerInitializeTask() (string, error) {
	checkpoint := consumer.client.getChcekPoint(consumer.shardId)

	if checkpoint != "" {
		consumer.consumerCheckPointTracker.setPersistentCheckPoint(checkpoint)
		return checkpoint, nil
	}

	if consumer.client.option.CursorPosition == BEGIN_CURSOR {
		cursor, err := consumer.client.getCursor(consumer.shardId, "begin")

		return cursor, err
	}
	if consumer.client.option.CursorPosition == END_CURSOR {
		cursor, err := consumer.client.getCursor(consumer.shardId, "end")
		if err != nil {
			Warning.Println(err)
		}
		return cursor, err
	}
	if consumer.client.option.CursorPosition == SPECIAL_TIMER_CURSOR {
		cursor, err := consumer.client.getCursor(consumer.shardId, string(consumer.client.option.CursorStartTime))
		if err != nil {
			Warning.Println(err)
		}
		return cursor, err
	}
	Info.Println("CursorPosition setting error, please reset with BEGIN_CURSOR or END_CURSOR or SPECIAL_TIMER_CURSOR")
	return "", errors.New("CursorPositionError")
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
	if consumer.lastFetchLogGroupList != nil {
		consumer.process(consumer.shardId, consumer.lastFetchLogGroupList)
		consumer.consumerCheckPointTracker.flushCheck()
	}
}
