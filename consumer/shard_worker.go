package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"time"
)

type ShardConsumerWorker struct {
	client                    *ConsumerClient
	consumerCheckPointTracker *ConsumerCheckPointTracker
	consumerShutDownFlag      bool
	lastFetchLogGroupList     *sls.LogGroupList
	nextFetchCursor           string
	lastFetchGroupCount       int
	lastFetchtime             int64
	consumerStatus            string
	process                   func(shard int, logGroup *sls.LogGroupList) string
	shardId                   int
	tempCheckPoint            string
	isCurrentDone             bool
	isShutDowning             bool
	logger                    log.Logger
	lastFetchTimeForForceFlushCpt int64

	startShutDownTime         int64
}

func (consumer *ShardConsumerWorker) setConsumerStatus(status string) {
	m.Lock()
	defer m.Unlock()
	consumer.consumerStatus = status
}

func (consumer *ShardConsumerWorker) getConsumerStatus() string {
	m.RLock()
	defer m.RUnlock()
	return consumer.consumerStatus
}

func initShardConsumerWorker(shardId int, consumerClient *ConsumerClient, do func(shard int, logGroup *sls.LogGroupList) string, logger log.Logger) *ShardConsumerWorker {
	shardConsumeWorker := &ShardConsumerWorker{
		consumerShutDownFlag:      false,
		process:                   do,
		consumerCheckPointTracker: initConsumerCheckpointTracker(shardId, consumerClient, logger),
		client:                    consumerClient,
		consumerStatus:            INITIALIZING,
		shardId:                   shardId,
		lastFetchtime:             0,
		isCurrentDone:             true,
		isShutDowning:             false,
		logger:                    logger,
		lastFetchTimeForForceFlushCpt: 0,
		startShutDownTime:         0,
	}
	return shardConsumeWorker
}

func (consumer *ShardConsumerWorker) consume() {
	if consumer.consumerShutDownFlag {
		consumer.isShutDowning = true
		go func() {
			// If the data is not consumed, save the tempCheckPoint to the server
			if consumer.getConsumerStatus() == PULL_PROCESSING_DONE {
				consumer.consumerCheckPointTracker.tempCheckPoint = consumer.tempCheckPoint
			}
			if consumer.getConsumerStatus() == CONSUME_PROCESSING {
				for {
					if consumer.getConsumerStatus() == CONSUME_PROCESSING_DONE {
						break
					} else {
						time.Sleep(500 * time.Millisecond)
					}
				}
			}

			for {
				err := consumer.consumerCheckPointTracker.flushCheckPoint()
				if err != nil {
					time.Sleep(time.Second)
				} else {
					break
				}
			}
			consumer.setConsumerStatus(SHUTDOWN_COMPLETE)
			consumer.isShutDowning = false
			level.Info(consumer.logger).Log("msg", "shardworker are shut down complete", "shardWorkerId", consumer.shardId)
		}()
	} else if consumer.getConsumerStatus() == INITIALIZING {
		consumer.isCurrentDone = false
		go func() {
			cursor, err := consumer.consumerInitializeTask()
			if err != nil {
				consumer.setConsumerStatus(INITIALIZING)
			} else {
				consumer.nextFetchCursor = cursor
				consumer.setConsumerStatus(INITIALIZING_DONE)
			}
			consumer.isCurrentDone = true
		}()
	} else if consumer.getConsumerStatus() == INITIALIZING_DONE || consumer.getConsumerStatus() == CONSUME_PROCESSING_DONE {
		consumer.isCurrentDone = false
		consumer.setConsumerStatus(PULL_PROCESSING)
		go func() {
			var isGenerateFetchTask = true
			// throttling control, similar as Java's SDK
			if consumer.lastFetchGroupCount < 100 {
				// The time used here is in milliseconds.
				isGenerateFetchTask = (time.Now().UnixNano()/1e6 - consumer.lastFetchtime) > 500
			} else if consumer.lastFetchGroupCount < 500 {
				isGenerateFetchTask = (time.Now().UnixNano()/1e6 - consumer.lastFetchtime) > 200
			} else if consumer.lastFetchGroupCount < 1000 {
				isGenerateFetchTask = (time.Now().UnixNano()/1e6 - consumer.lastFetchtime) > 50
			}
			if isGenerateFetchTask {
				consumer.lastFetchtime = time.Now().UnixNano() / 1e6
				// Set the logback cursor. If the logs are not consumed, save the logback cursor to the server.
				consumer.tempCheckPoint = consumer.nextFetchCursor

				logGroupList, nextCursor, err := consumer.consumerFetchTask()
				if err != nil {
					consumer.setConsumerStatus(INITIALIZING_DONE)
				} else {
					consumer.lastFetchLogGroupList = logGroupList
					consumer.nextFetchCursor = nextCursor
					consumer.consumerCheckPointTracker.setMemoryCheckPoint(consumer.nextFetchCursor)
					consumer.lastFetchGroupCount = GetLogCount(consumer.lastFetchLogGroupList)
					if consumer.lastFetchGroupCount == 0 {
						consumer.lastFetchLogGroupList = nil
					} else {
						consumer.lastFetchTimeForForceFlushCpt = time.Now().Unix()
					}
					if consumer.lastFetchTimeForForceFlushCpt != 0 && time.Now().Unix()-consumer.lastFetchTimeForForceFlushCpt > 30 {
						err := consumer.consumerCheckPointTracker.flushCheckPoint()
						if err != nil {
							level.Warn(consumer.logger).Log("msg", "Failed to save the final checkpoint", "error:", err)
						}else{
							consumer.lastFetchTimeForForceFlushCpt = 0
						}

					}
					consumer.setConsumerStatus(PULL_PROCESSING_DONE)
				}
			}
			consumer.isCurrentDone = true
		}()
	} else if consumer.getConsumerStatus() == PULL_PROCESSING_DONE {
		consumer.isCurrentDone = false
		consumer.setConsumerStatus(CONSUME_PROCESSING)
		go func() {
			rollBackCheckpoint := consumer.consumerProcessTask()
			if rollBackCheckpoint != "" {
				consumer.nextFetchCursor = rollBackCheckpoint
				level.Info(consumer.logger).Log("msg", "Checkpoints set for users have been reset", "shardWorkerId", consumer.shardId, "rollBackCheckpoint", rollBackCheckpoint)
			}
			consumer.lastFetchLogGroupList = nil
			consumer.setConsumerStatus(CONSUME_PROCESSING_DONE)
			consumer.isCurrentDone = true
		}()
	}

}

func (consumer *ShardConsumerWorker) consumerShutDown() {
	consumer.consumerShutDownFlag = true
	// If the shutdown task is executing, return directly without calling the shutdown function
	if consumer.isShutDowning == true {
		if consumer.startShutDownTime != 0 && time.Now().Unix()-consumer.startShutDownTime >= 30 {
			consumer.setConsumerStatus(SHUTDOWN_COMPLETE)
			level.Warn(consumer.logger).Log("msg", "shards close timeout, forced shutdown")
		}
		return
	}
	if !consumer.isShutDownComplete() {
		consumer.startShutDownTime = time.Now().Unix()
		consumer.consume()
	}
}

func (consumer *ShardConsumerWorker) isShutDownComplete() bool {
	return consumer.getConsumerStatus() == SHUTDOWN_COMPLETE
}
