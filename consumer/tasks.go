package consumerLibrary

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/go-kit/kit/log/level"
)

func (consumer *ShardConsumerWorker) consumerInitializeTask() (string, error) {
	// read checkpoint firstly
	checkpoint, err := consumer.client.getCheckPoint(consumer.shardId)
	if err != nil {
		return "", err
	}
	if checkpoint != "" && err == nil {
		consumer.consumerCheckPointTracker.initCheckPoint(checkpoint)
		return checkpoint, nil
	}

	if consumer.client.option.CursorPosition == BEGIN_CURSOR {
		cursor, err := consumer.client.getCursor(consumer.shardId, "begin")
		if err != nil {
			level.Warn(consumer.logger).Log("msg", "get beginCursor error", "shard", consumer.shardId, "error", err)
		}
		return cursor, err
	}
	if consumer.client.option.CursorPosition == END_CURSOR {
		cursor, err := consumer.client.getCursor(consumer.shardId, "end")
		if err != nil {
			level.Warn(consumer.logger).Log("msg", "get endCursor error", "shard", consumer.shardId, "error", err)
		}
		return cursor, err
	}
	if consumer.client.option.CursorPosition == SPECIAL_TIMER_CURSOR {
		cursor, err := consumer.client.getCursor(consumer.shardId, fmt.Sprintf("%v", consumer.client.option.CursorStartTime))
		if err != nil {
			level.Warn(consumer.logger).Log("msg", "get specialCursor error", "shard", consumer.shardId, "error", err)
		}
		return cursor, err
	}
	level.Warn(consumer.logger).Log("msg", "CursorPosition setting error, please reset with BEGIN_CURSOR or END_CURSOR or SPECIAL_TIMER_CURSOR")
	return "", errors.New("CursorPositionError")
}

func (consumer *ShardConsumerWorker) nextFetchTask() error {
	// update last fetch time, for control fetch frequency
	consumer.lastFetchTime = time.Now()
	
	logGroup, nextCursor, err := consumer.client.pullLogs(consumer.shardId, consumer.nextFetchCursor)
	if err != nil {
		return err
	}
	// set cursors user to decide whether to save according to the execution of `process`
	consumer.consumerCheckPointTracker.setCurrentCheckPoint(consumer.nextFetchCursor)
	consumer.lastFetchLogGroupList = logGroup
	consumer.nextFetchCursor = nextCursor
	consumer.lastFetchGroupCount = GetLogGroupCount(consumer.lastFetchLogGroupList)
	consumer.consumerCheckPointTracker.setNextCheckPoint(consumer.nextFetchCursor)
	level.Debug(consumer.logger).Log(
		"shardId", consumer.shardId,
		"fetch log count", GetLogCount(consumer.lastFetchLogGroupList),
	)
	if consumer.lastFetchGroupCount == 0 {
		consumer.lastFetchLogGroupList = nil
		// may no new data can be pulled, no process func can trigger checkpoint saving
		consumer.saveCheckPointIfNeeded()
	}

	return nil
}

func (consumer *ShardConsumerWorker) consumerProcessTask() string {
	// If the user's consumption function reports a panic error, it will be captured and retry until sucessed.
	defer func() {
		if r := recover(); r != nil {
			stackBuf := make([]byte, 1<<16)
			runtime.Stack(stackBuf, false)
			level.Error(consumer.logger).Log("msg", "get panic in your process function", "error", r, "stack", string(stackBuf))
			for {
				if consumer.consumerRetryProcessTask() {
					break
				}
				time.Sleep(time.Second * 2)
			}
		}
	}()
	if consumer.lastFetchLogGroupList != nil {
		consumer.rollBackCheckpoint = consumer.processor.Process(consumer.shardId, consumer.lastFetchLogGroupList, consumer.consumerCheckPointTracker)
		consumer.saveCheckPointIfNeeded()
		consumer.lastFetchLogGroupList = nil
	}
	return consumer.rollBackCheckpoint
}

func (consumer *ShardConsumerWorker) consumerRetryProcessTask() bool {
	level.Info(consumer.logger).Log("msg", "Start retrying the process function")
	defer func() {
		if r := recover(); r != nil {
			stackBuf := make([]byte, 1<<16)
			runtime.Stack(stackBuf, false)
			level.Error(consumer.logger).Log("msg", "get panic in your process function", "error", r, "stack", string(stackBuf))
		}
	}()
	consumer.rollBackCheckpoint = consumer.processor.Process(consumer.shardId, consumer.lastFetchLogGroupList, consumer.consumerCheckPointTracker)
	consumer.saveCheckPointIfNeeded()
	return true
}
