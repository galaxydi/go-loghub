package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"os"
	"os/signal"
	"time"
)

type ConsumerWorker struct {
	consumerHeatBeat   *ConsumerHeatBeat
	client             *ConsumerClient
	workerShutDownFlag bool
	shardConsumer      map[int]*ShardConsumerWorker
	do                 func(shard int, logGroup *sls.LogGroupList)
}

func InitConsumerWorker(option LogHubConfig, do func(int, *sls.LogGroupList)) *ConsumerWorker {

	consumerClient := initConsumerClient(option)
	consumerHeatBeat := initConsumerHeatBeat(consumerClient)
	consumerWorker := &ConsumerWorker{
		consumerHeatBeat,
		consumerClient,
		false,
		make(map[int]*ShardConsumerWorker),
		do,
	}
	consumerClient.createConsumerGroup()
	return consumerWorker
}

func (consumerWorker *ConsumerWorker) Start() {
	ch := make(chan os.Signal)
	signal.Notify(ch)
	go consumerWorker.run()
	if _, ok := <-ch; ok {
		Info.Printf("get stop signal, start to stop consumer worker:%v", consumerWorker.client.option.ConsumerName)
		consumerWorker.workerShutDown()
	}
}

func (consumerWorker *ConsumerWorker) workerShutDown() {
	Info.Println("*** try to exit ***")
	consumerWorker.workerShutDownFlag = true
	consumerWorker.consumerHeatBeat.shutDownHeart()
	for {
		// Used to wait for all shardWorkers to close, otherwise sometimes they will die.
		time.Sleep(1 * time.Second)
		if consumerWorker.shardConsumer == nil {
			break
		}
	}
	Info.Printf("consumer worker %v stopped", consumerWorker.client.option.ConsumerName)
}

func (consumerWorker *ConsumerWorker) run() {
	Info.Printf("consumer worker %v start", consumerWorker.client.option.ConsumerName)
	go consumerWorker.consumerHeatBeat.heartBeatRun()

	for !consumerWorker.workerShutDownFlag {
		heldShards := consumerWorker.consumerHeatBeat.getHeldShards()
		lastFetchTime := time.Now().Unix()

		for _, shard := range heldShards {
			if consumerWorker.workerShutDownFlag {
				break
			}
			shardConsumer := consumerWorker.getShardConsumer(shard)
			if shardConsumer.isCurrentDone {
				shardConsumer.consume()
			} else {
				continue
			}
		}
		consumerWorker.cleanShardConsumer(heldShards)
		TimeToSleep(consumerWorker.client.option.DataFetchInterval, lastFetchTime, consumerWorker.workerShutDownFlag)

	}
	Info.Printf("consumer worker %v try to cleanup consumers", consumerWorker.client.option.ConsumerName)
	consumerWorker.shutDownAndWait()
}

func (consumerWorker *ConsumerWorker) shutDownAndWait() {
	for {
		time.Sleep(500 * time.Millisecond)
		for shard, consumer := range consumerWorker.shardConsumer {
			if !consumer.isShutDownComplete() {
				if consumer.isShutDowning {
					continue
				} else {
					consumer.consumerShutDown()
				}

			} else if consumer.isShutDownComplete() {
				delete(consumerWorker.shardConsumer, shard)
			}
		}
		if len(consumerWorker.shardConsumer) == 0 {
			consumerWorker.shardConsumer = nil
			break
		}
	}

}

func (consumerWorker *ConsumerWorker) getShardConsumer(shardId int) *ShardConsumerWorker {
	consumer := consumerWorker.shardConsumer[shardId]
	if consumer != nil {
		return consumer
	}
	consumer = initShardConsumerWorker(shardId, consumerWorker.client, consumerWorker.do)
	consumerWorker.shardConsumer[shardId] = consumer
	return consumer

}

func (consumerWorker *ConsumerWorker) cleanShardConsumer(owned_shards []int) {
	for shard, consumer := range consumerWorker.shardConsumer {
		if !Contain(shard, owned_shards) {
			Info.Printf("try to call shut down for unassigned consumer shard: %v", shard)
			consumer.consumerShutDown()
			Info.Printf("Complete call shut down for unassigned consumer shard: %v", shard)
		}
		if consumer.isShutDownComplete() {
			consumerWorker.consumerHeatBeat.removeHeartShard(shard)
			Info.Printf("Remove an unassigned consumer shard: %v", shard)
			delete(consumerWorker.shardConsumer, shard)
		}
	}

}
