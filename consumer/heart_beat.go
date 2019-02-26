package consumerLibrary

import (
	"sync"
	"time"
)

type ConsumerHeatBeat struct {
	client       *ConsumerClient
	shutDownFlag bool
	heldShards   []int
	heartShard   []int
}

func initConsumerHeatBeat(consumerClient *ConsumerClient) *ConsumerHeatBeat {
	consumerHeatBeat := &ConsumerHeatBeat{
		client:       consumerClient,
		shutDownFlag: false,
		heldShards:   []int{},
		heartShard:   []int{},
	}
	return consumerHeatBeat
}

func (consumerHeatBeat *ConsumerHeatBeat) getHeldShards() []int {
	return consumerHeatBeat.heartShard
}

func (consumerHeatBeat *ConsumerHeatBeat) shutDownHeart() {
	Info.Println("try to stop heart beat")
	consumerHeatBeat.shutDownFlag = true
}

func (consumerHeatBeat *ConsumerHeatBeat) removeHeartShard(shardId int) {
	for i, x := range consumerHeatBeat.heartShard {
		if shardId == x {
			consumerHeatBeat.heartShard = append(consumerHeatBeat.heartShard[:i], consumerHeatBeat.heartShard[i+1:]...)
		}
	}
	for i, x := range consumerHeatBeat.heldShards {
		if shardId == x {
			consumerHeatBeat.heldShards = append(consumerHeatBeat.heldShards[:i], consumerHeatBeat.heldShards[i+1:]...)
		}
	}
}

//heartBeatRun运行的时候，其它线程get会有线程安全问题吗？
func (consumerHeatBeat *ConsumerHeatBeat) heartBeatRun() {
	var lastHeartBeatTime int64
	var lock sync.Mutex
	for !consumerHeatBeat.shutDownFlag {
		lastHeartBeatTime = time.Now().Unix()
		responseShards := consumerHeatBeat.client.heartBeat(consumerHeatBeat.heartShard)
		Info.Printf("heart beat result: %v,get:%v", consumerHeatBeat.heartShard, responseShards)

		if !IntSliceReflectEqual(consumerHeatBeat.heartShard, consumerHeatBeat.heldShards) {
			currentSet := Set(consumerHeatBeat.heartShard)
			responseSet := Set(consumerHeatBeat.heldShards)
			add := Subtract(currentSet, responseSet)
			remove := Subtract(responseSet, currentSet)
			Info.Printf("shard reorganize, adding: %v, removing: %v", add, remove)
		}
		lock.Lock() // Adding locks to modify HeldShards to keep threads safe
		consumerHeatBeat.heldShards = responseShards
		consumerHeatBeat.heartShard = consumerHeatBeat.heldShards[:]
		lock.Unlock()
		timeToSleep := int64(consumerHeatBeat.client.option.HeartbeatIntervalInSecond)*1000 - (time.Now().Unix()-lastHeartBeatTime)*1000
		for timeToSleep > 0 && !consumerHeatBeat.shutDownFlag {
			time.Sleep(time.Duration(Min(timeToSleep, 1000)) * time.Millisecond)
			timeToSleep = int64(consumerHeatBeat.client.option.HeartbeatIntervalInSecond)*1000 - (time.Now().Unix()-lastHeartBeatTime)*1000
		}
	}
	Info.Println("heart beat exit")
}
