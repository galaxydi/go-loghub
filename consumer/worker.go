package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"os"
	"os/signal"
	"time"
)

type ConsumerWorker struct {
	*ConsumerHeatBeat
	*ConsumerClient
	WorkerShutDownFlag bool
	ShardConsumer      map[int]*ShardConsumerWorker
	Do                 func(shard int, logGroup *sls.LogGroupList)
}

func InitConsumerWorker(option LogHubConfig, do func(int, *sls.LogGroupList)) *ConsumerWorker {

	consumerClient := InitConsumerClient(option)
	consumerHeatBeat := InitConsumerHeatBeat(consumerClient)
	consumerWorker := &ConsumerWorker{
		consumerHeatBeat,
		consumerClient,
		false,
		make(map[int]*ShardConsumerWorker),
		do,
	}
	consumerClient.McreateConsumerGroup()
	return consumerWorker
}

func (consumerWorker *ConsumerWorker) Worker() {
	ch := make(chan os.Signal)
	signal.Notify(ch)
	go consumerWorker.run()
	if _, ok := <-ch; ok {
		Info.Printf("get stop signal, start to stop consumer worker:%v", consumerWorker.ConsumerName)
		consumerWorker.WorkerShutDown()
	}
}

func (consumerWorker *ConsumerWorker) WorkerShutDown() {
	Info.Println("*** try to exit ***")
	consumerWorker.WorkerShutDownFlag = true
	consumerWorker.ShutDownHeart()
	for {
		// Used to wait for all shardWorkers to close, otherwise sometimes they will die.
		time.Sleep(1 * time.Second)
		if consumerWorker.ShardConsumer == nil {
			break
		}
	}
	Info.Printf("consumer worker %v stopped", consumerWorker.ConsumerName)
}

func (consumerWorker *ConsumerWorker) run() {
	go consumerWorker.HeartBeatRun()

	for !consumerWorker.WorkerShutDownFlag {
		heldShards := consumerWorker.GetHeldShards()
		lastFetchTime := time.Now().Unix()
		sh := make(chan bool)
		go func(sh chan bool) {
			for _, shard := range heldShards {
				if consumerWorker.WorkerShutDownFlag {
					break
				}
				mShardConsumer := consumerWorker.getShardConsumer(shard)
				go mShardConsumer.consume()
			}
			sh <- true
		}(sh)
		<-sh
		consumerWorker.cleanShardConsumer(heldShards)
		timeToSleep := consumerWorker.DataFetchInterval - (time.Now().Unix() - lastFetchTime)
		for timeToSleep > 0 && !consumerWorker.HeartShutDownFlag {
			time.Sleep(time.Duration(Min(timeToSleep, 1)) * time.Second)
			timeToSleep = consumerWorker.DataFetchInterval - (time.Now().Unix() - lastFetchTime)
		}
	}
	Info.Printf("consumer worker %v try to cleanup consumers", consumerWorker.ConsumerName)
	consumerWorker.ShutDownAndWait()
}

func (consumerWorker *ConsumerWorker) ShutDownAndWait() {
	for _, consumer := range consumerWorker.ShardConsumer {
		if !consumer.IsShutDown() {
			consumer.ConsumerShutDown()
		}
	}
	consumerWorker.ShardConsumer = nil
}

func (consumerWorker *ConsumerWorker) getShardConsumer(shardId int) *ShardConsumerWorker {
	consumer := consumerWorker.ShardConsumer[shardId]
	if consumer != nil {
		return consumer
	}
	consumer = InitShardConsumerWorker(shardId, consumerWorker.ConsumerClient, consumerWorker.Do)
	consumerWorker.ShardConsumer[shardId] = consumer
	return consumer

}

func (consumerWorker *ConsumerWorker) cleanShardConsumer(owned_shards []int) {
	for shard, consumer := range consumerWorker.ShardConsumer {
		if !Contain(shard, owned_shards) {
			Info.Printf("try to call shut down for unassigned consumer shard: %v", shard)
			consumer.ConsumerShutDown()
			Info.Printf("Complete call shut down for unassigned consumer shard: %v", shard)
		}
		if consumer.IsShutDown() {

			consumerWorker.RemoveHeartShard(shard)
			Info.Printf("Remove an unassigned consumer shard: %v", shard)
			delete(consumerWorker.ShardConsumer, shard)
		}
	}

}
