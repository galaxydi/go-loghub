package consumerLibrary

import (
	"sync"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type ShardConsumerWorker struct {
	client                    *ConsumerClient
	consumerCheckPointTracker *DefaultCheckPointTracker
	consumerShutDownFlag      bool
	lastFetchLogGroupList     *sls.LogGroupList
	nextFetchCursor           string
	lastFetchGroupCount       int
	lastFetchTime             time.Time
	consumerStatus            string
	process                   func(shard int, logGroup *sls.LogGroupList, checkpointTracker CheckPointTracer) string
	shardId                   int
	isCurrentDone             bool
	logger                    log.Logger
	// unix time
	lastCheckpointSaveTime time.Time
	rollBackCheckpoint     string

	statusLock   sync.RWMutex
	taskLock     sync.RWMutex
	shutDownLock sync.RWMutex
}

func (consumer *ShardConsumerWorker) setConsumerStatus(status string) {
	consumer.statusLock.Lock()
	defer consumer.statusLock.Unlock()
	consumer.consumerStatus = status
}

func (consumer *ShardConsumerWorker) getConsumerStatus() string {
	consumer.statusLock.RLock()
	defer consumer.statusLock.RUnlock()
	return consumer.consumerStatus
}

func initShardConsumerWorker(shardId int, consumerClient *ConsumerClient, consumerHeartBeat *ConsumerHeartBeat, do func(shard int, logGroup *sls.LogGroupList, checkpointTracker CheckPointTracer) string, logger log.Logger) *ShardConsumerWorker {
	shardConsumeWorker := &ShardConsumerWorker{
		consumerShutDownFlag:      false,
		process:                   do,
		consumerCheckPointTracker: initConsumerCheckpointTracker(shardId, consumerClient, consumerHeartBeat, logger),
		client:                    consumerClient,
		consumerStatus:            INITIALIZING,
		shardId:                   shardId,
		lastFetchTime:             time.Now(),
		isCurrentDone:             true,
		logger:                    logger,
		rollBackCheckpoint:        "",
	}
	return shardConsumeWorker
}

func (consumer *ShardConsumerWorker) consume() {
	if consumer.consumerShutDownFlag {
		go func() {
			defer consumer.setConsumerStatus(SHUTDOWN_COMPLETE)
			// if processing, we can wait it to switch status, otherwise we save checkpoint straightly
			if consumer.getConsumerStatus() == CONSUME_PROCESSING {
				level.Info(consumer.logger).Log("msg", "Consumption is in progress, sleep to wait for consumption to be completed")
				shutdownWaitTimes := 10
				// OPTIMIZE
				// now just sleep, won't wait until the end
				for i := 0; i < shutdownWaitTimes; i++ {
					time.Sleep(time.Millisecond * 10)
					if consumer.getConsumerStatus() != CONSUME_PROCESSING {
						break
					}
					if i == shutdownWaitTimes {
						level.Warn(consumer.logger).Log("msg", "wait many times, but last process may be not over yes", "retryTimes", shutdownWaitTimes)
					}
				}
			}

			var err error
			retryTimes := 3
			for i := 0; i < retryTimes; i++ {
				err = consumer.consumerCheckPointTracker.SaveCheckPoint(true)
				if err == nil {
					break
				}
			}
			if err == nil {
				level.Info(consumer.logger).Log("msg", "shardworker are shut down complete", "shardWorkerId", consumer.shardId)
			} else {
				level.Warn(consumer.logger).Log("msg", "failed after retry", "retryTimes", retryTimes, "err", err)
			}
		}()
	} else {
		consumer.setTaskDoneFlag(false)
		if consumer.getConsumerStatus() == INITIALIZING {
			go func() {
				defer consumer.setTaskDoneFlag(true)

				cursor, err := consumer.consumerInitializeTask()
				if err != nil {
					consumer.setConsumerStatus(INITIALIZING)
				} else {
					consumer.nextFetchCursor = cursor
					consumer.setConsumerStatus(INITIALIZING_DONE)
				}
			}()
		} else if consumer.getConsumerStatus() == INITIALIZING_DONE || consumer.getConsumerStatus() == CONSUME_PROCESSING_DONE {
			consumer.setConsumerStatus(PULL_PROCESSING)
			go func() {
				defer consumer.setTaskDoneFlag(true)
				if !consumer.shouldFetch() {
					level.Debug(consumer.logger).Log("msg", "Pull Log Current Limitation and Re-Pull Log")
					consumer.setConsumerStatus(INITIALIZING_DONE)
				}

				if err := consumer.nextFetchTask(); err != nil {
					consumer.setConsumerStatus(INITIALIZING_DONE)
				} else {
					consumer.setConsumerStatus(PULL_PROCESSING_DONE)
				}
			}()
		} else if consumer.getConsumerStatus() == PULL_PROCESSING_DONE {
			consumer.setConsumerStatus(CONSUME_PROCESSING)
			go func() {
				defer consumer.setTaskDoneFlag(true)
				defer consumer.setConsumerStatus(CONSUME_PROCESSING_DONE)
				rollBackCheckpoint := consumer.consumerProcessTask()
				if rollBackCheckpoint != "" {
					consumer.nextFetchCursor = rollBackCheckpoint
					level.Info(consumer.logger).Log(
						"msg", "Checkpoints set for users have been reset",
						"shardId", consumer.shardId,
						"rollBackCheckpoint", rollBackCheckpoint,
					)
				}
			}()
		}
	}
}

func (consumer *ShardConsumerWorker) shouldFetch() bool {
	if consumer.lastFetchGroupCount >= 1000 {
		return true
	}
	duration := time.Since(consumer.lastFetchTime)
	if consumer.lastFetchGroupCount < 100 {
		// The time used here is in milliseconds.
		return duration > 500*time.Millisecond
	} else if consumer.lastFetchGroupCount < 500 {
		return duration > 200*time.Millisecond
	} else { // 500 - 1000
		return duration > 50*time.Millisecond
	}
}

func (consumer *ShardConsumerWorker) saveCheckPointIfNeeded() {
	if !consumer.client.option.AutoCommitDisabled {
		return
	}
	if time.Since(consumer.lastCheckpointSaveTime) > time.Millisecond*time.Duration(consumer.client.option.AutoCommitIntervalInMS) {
		consumer.consumerCheckPointTracker.SaveCheckPoint(true)
	}
}

func (consumer *ShardConsumerWorker) consumerShutDown() {
	consumer.consumerShutDownFlag = true
	if !consumer.isShutDownComplete() {
		consumer.consume()
	}
}

func (consumer *ShardConsumerWorker) isShutDownComplete() bool {
	return consumer.getConsumerStatus() == SHUTDOWN_COMPLETE
}

func (consumer *ShardConsumerWorker) setTaskDoneFlag(done bool) {
	consumer.taskLock.Lock()
	defer consumer.taskLock.Unlock()
	consumer.isCurrentDone = done
}

func (consumer *ShardConsumerWorker) isTaskDone() bool {
	consumer.taskLock.RLock()
	defer consumer.taskLock.RUnlock()
	return consumer.isCurrentDone
}
