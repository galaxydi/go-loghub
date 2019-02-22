package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"time"
)

type ShardConsumerWorker struct {
	*ConsumerClient
	*ConsumerCheckPointTracker
	ConsumerShutDownFlag bool
	LastFetchLogGroup    *sls.LogGroupList
	NextFetchCursor      string
	LastFetchGroupCount  int
	LastFetchtime        int64
	ConsumerStatus       string
	Process              func(shard int, logGroup *sls.LogGroupList)
	ShardId              int
	RollBackCheckPoint   string
}

func InitShardConsumerWorker(shardId int, consumerClient *ConsumerClient, do func(shard int, logGroup *sls.LogGroupList)) *ShardConsumerWorker {
	shardConsumeWorker := &ShardConsumerWorker{
		ConsumerShutDownFlag:      false,
		Process:                   do,
		ConsumerCheckPointTracker: initConsumerCheckpointTracker(shardId, consumerClient),
		ConsumerClient:            consumerClient,
		ConsumerStatus:            INITIALIZ,
		ShardId:                   shardId,
		LastFetchtime:             0,
	}
	return shardConsumeWorker
}

func (consumer *ShardConsumerWorker) consume() {
	a := make(chan int)
	b := make(chan int)
	c := make(chan int)
	d := make(chan int)
	if consumer.ConsumerShutDownFlag == true {
		consumer.ConsumerStatus = SHUTTING_DOWN
	}
	if consumer.ConsumerStatus == SHUTTING_DOWN {
		go func() {
			d <- channelD
		}()
	}
	if consumer.ConsumerStatus == INITIALIZ {
		go func() {
			a <- channelA
		}()
	}
	if consumer.ConsumerStatus == PROCESS && consumer.LastFetchLogGroup == nil {
		go func() {
			b <- channelB
		}()
	}
	if consumer.ConsumerStatus == PROCESS && consumer.LastFetchLogGroup != nil {
		go func() {
			c <- channelC
		}()
	}
	// event loop，When the signal is obtained, the corresponding task is put into groutine to execute each time.
	select {
	case _, ok := <-a:
		if ok {
			consumer.NextFetchCursor = consumer.consumerInitializeTask()
			consumer.ConsumerStatus = PROCESS
		}
	case _, ok := <-b:
		if ok {

			var isGenerateFetchTask = true
			// throttling control, similar as Java's SDK
			if consumer.LastFetchGroupCount < 100 {
				// The time used here is in milliseconds.
				isGenerateFetchTask = (time.Now().UnixNano()/1e6 - consumer.LastFetchtime) > 500
			}
			if consumer.LastFetchGroupCount < 500 {
				isGenerateFetchTask = (time.Now().UnixNano()/1e6 - consumer.LastFetchtime) > 200
			}
			if consumer.LastFetchGroupCount < 1000 {
				isGenerateFetchTask = (time.Now().UnixNano()/1e6 - consumer.LastFetchtime) > 50
			}
			if isGenerateFetchTask {
				consumer.LastFetchtime = time.Now().UnixNano() / 1e6
				// Set the logback cursor. If the logs are not consumed, save the logback cursor to the server.
				consumer.RollBackCheckPoint = consumer.NextFetchCursor

				consumer.LastFetchLogGroup, consumer.NextFetchCursor = consumer.consumerFetchTask()
				consumer.setMemoryCheckPoint(consumer.NextFetchCursor)
				consumer.LastFetchGroupCount = GetLogCount(consumer.LastFetchLogGroup)
				Info.Printf("shard %v get log conunt %v", consumer.ShardId, consumer.LastFetchGroupCount)
				if consumer.LastFetchGroupCount == 0 {
					consumer.LastFetchLogGroup = nil
				}
			}

		}
	case _, ok := <-c:
		if ok {
			consumer.consumerProcessTask()
			consumer.LastFetchLogGroup = nil
		}
	case _, ok := <-d:
		if ok {
			// If the data is not consumed, save the RollBackCheckPoint to the server
			if consumer.LastFetchLogGroup != nil && consumer.LastFetchGroupCount != 0 {
				consumer.TempCheckPoint = consumer.RollBackCheckPoint
			}
			consumer.mFlushCheckPoint()
			consumer.ConsumerStatus = SHUTDOWN_COMPLETE
			Info.Printf("shardworker %v are shut down complete", consumer.ShardId)
		}
	}

}

func (consumer *ShardConsumerWorker) ConsumerShutDown() {
	consumer.ConsumerShutDownFlag = true
	if !consumer.IsShutDown() {
		consumer.consume()
	}
}

func (consumer *ShardConsumerWorker) IsShutDown() bool {
	return consumer.ConsumerStatus == SHUTDOWN_COMPLETE
}
