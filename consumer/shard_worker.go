package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
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
}

func initShardConsumerWorker(shardId int, consumerClient *ConsumerClient, do func(shard int, logGroup *sls.LogGroupList)) *ShardConsumerWorker {
	shardConsumeWorker := &ShardConsumerWorker{
		consumerShutDownFlag:      false,
		process:                   do,
		consumerCheckPointTracker: initConsumerCheckpointTracker(shardId, consumerClient),
		client:                    consumerClient,
		consumerStatus:            INITIALIZED,
		shardId:                   shardId,
		lastFetchtime:             0,
		isCurrentDone:             true,
	}
	return shardConsumeWorker
}

func (consumer *ShardConsumerWorker) consume() {
	ch := make(chan int)

	if consumer.consumerShutDownFlag == true {
		consumer.consumerStatus = SHUTTING_DOWN
	}
	if consumer.consumerStatus == SHUTTING_DOWN {
		consumer.isCurrentDone = false
		go func() {
			// If the data is not consumed, save the RollBackCheckPoint to the server
			if consumer.lastFetchLogGroupList != nil && consumer.lastFetchGroupCount != 0 {
				consumer.consumerCheckPointTracker.tempCheckPoint = consumer.rollBackCheckPoint
			}
			consumer.consumerCheckPointTracker.flushCheckPoint()
			ch <- channelC
		}()
	} else if consumer.consumerStatus == INITIALIZED {
		consumer.isCurrentDone = false
		go func() {
			consumer.nextFetchCursor = consumer.consumerInitializeTask()
			ch <- channelA
		}()
	} else if consumer.consumerStatus == PROCESSING {
		if consumer.lastFetchLogGroupList == nil {
			consumer.isCurrentDone = false
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

					consumer.lastFetchLogGroupList, consumer.nextFetchCursor = consumer.consumerFetchTask()
					consumer.consumerCheckPointTracker.setMemoryCheckPoint(consumer.nextFetchCursor)
					consumer.lastFetchGroupCount = GetLogCount(consumer.lastFetchLogGroupList)
					Info.Printf("shard %v get log conunt %v", consumer.shardId, consumer.lastFetchGroupCount)
					if consumer.lastFetchGroupCount == 0 {
						consumer.lastFetchLogGroupList = nil
					}
				}
				ch <- channelB
			}()
		} else if consumer.lastFetchLogGroupList != nil {
			consumer.isCurrentDone = false
			go func() {
				consumer.consumerProcessTask()
				consumer.lastFetchLogGroupList = nil
				ch <- channelB
			}()
		}
	}
	// event loop，When the signal is obtained, the corresponding task is put into groutine to execute each time.
	select {
	case a, ok := <-ch:
		if ok && a == channelA {
			consumer.consumerStatus = PROCESSING
			consumer.isCurrentDone = true
		} else if ok && a == channelB {
			consumer.isCurrentDone = true
		} else if ok && a == channelC {
			consumer.isCurrentDone = true
			consumer.consumerStatus = SHUTDOWN_COMPLETE
			Info.Printf("shardworker %v are shut down complete", consumer.shardId)
		}
	}

}

func (consumer *ShardConsumerWorker) consumerShutDown() {
	consumer.consumerShutDownFlag = true
	if !consumer.isShutDownComplete() {
		consumer.consume()
	}
}

func (consumer *ShardConsumerWorker) isShutDownComplete() bool {
	return consumer.consumerStatus == SHUTDOWN_COMPLETE
}
