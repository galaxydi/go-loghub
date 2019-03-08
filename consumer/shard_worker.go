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
	process                   func(shard int, logGroup *sls.LogGroupList)
	shardId                   int
	rollBackCheckPoint        string
	isCurrentDone             bool
	isShutDowning             bool
	logger                    log.Logger
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

func initShardConsumerWorker(shardId int, consumerClient *ConsumerClient, do func(shard int, logGroup *sls.LogGroupList), logger log.Logger) *ShardConsumerWorker {
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
	}
	return shardConsumeWorker
}

func (consumer *ShardConsumerWorker) consume() {

	if consumer.consumerShutDownFlag {
		consumer.isShutDowning = true
		go func() {
			// If the data is not consumed, save the RollBackCheckPoint to the server
			if consumer.getConsumerStatus() == PULL_PROCESSING_DONE {
				consumer.consumerCheckPointTracker.tempCheckPoint = consumer.rollBackCheckPoint
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
				consumer.rollBackCheckPoint = consumer.nextFetchCursor

				logGroupList, nextCursor := consumer.consumerFetchTask()
				if nextCursor == "PullLogFailed" {
					consumer.setConsumerStatus(INITIALIZING_DONE)
				} else {
					consumer.lastFetchLogGroupList = logGroupList
					consumer.nextFetchCursor = nextCursor
					consumer.consumerCheckPointTracker.setMemoryCheckPoint(consumer.nextFetchCursor)
					consumer.lastFetchGroupCount = GetLogCount(consumer.lastFetchLogGroupList)
					if consumer.lastFetchGroupCount == 0 {
						consumer.lastFetchLogGroupList = nil
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
			consumer.consumerProcessTask()
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
		return
	}
	if !consumer.isShutDownComplete() {
		consumer.consume()
	}
}

func (consumer *ShardConsumerWorker) isShutDownComplete() bool {
	return consumer.getConsumerStatus() == SHUTDOWN_COMPLETE
}
