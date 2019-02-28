package consumerLibrary

import (
	"time"
)

type ConsumerHeatBeat struct {
	client       *ConsumerClient
	shutDownFlag bool
	heldShards   []int
	heartShards  []int
}

func initConsumerHeatBeat(consumerClient *ConsumerClient) *ConsumerHeatBeat {
	consumerHeatBeat := &ConsumerHeatBeat{
		client:       consumerClient,
		shutDownFlag: false,
		heldShards:   []int{},
		heartShards:  []int{},
	}
	return consumerHeatBeat
}

func (consumerHeatBeat *ConsumerHeatBeat) getHeldShards() []int {
	m.RLock()
	defer m.RUnlock()
	return consumerHeatBeat.heartShards
}

func (consumerHeatBeat *ConsumerHeatBeat) setHeldShards(x []int) {
	m.Lock()
	defer m.Unlock()
	consumerHeatBeat.heldShards = x
}

func (consumerHeatBeat *ConsumerHeatBeat) shutDownHeart() {
	Info.Println("try to stop heart beat")
	consumerHeatBeat.shutDownFlag = true
}

func (consumerHeatBeat *ConsumerHeatBeat) removeHeartShard(shardId int) {
	for i, x := range consumerHeatBeat.heartShards {
		if shardId == x {
			consumerHeatBeat.heartShards = append(consumerHeatBeat.heartShards[:i], consumerHeatBeat.heartShards[i+1:]...)
		}
	}
	for i, x := range consumerHeatBeat.heldShards {
		if shardId == x {
			consumerHeatBeat.heldShards = append(consumerHeatBeat.heldShards[:i], consumerHeatBeat.heldShards[i+1:]...)
		}
	}
}

func (consumerHeatBeat *ConsumerHeatBeat) heartBeatRun() {
	var lastHeartBeatTime int64

	for !consumerHeatBeat.shutDownFlag {
		lastHeartBeatTime = time.Now().Unix()
		responseShards := consumerHeatBeat.client.heartBeat(consumerHeatBeat.heartShards)
		Info.Printf("heart beat result: %v,get:%v", consumerHeatBeat.heartShards, responseShards)

		if !IntSliceReflectEqual(consumerHeatBeat.heartShards, consumerHeatBeat.heldShards) {
			currentSet := Set(consumerHeatBeat.heartShards)
			responseSet := Set(consumerHeatBeat.heldShards)
			add := Subtract(currentSet, responseSet)
			remove := Subtract(responseSet, currentSet)
			Info.Printf("shard reorganize, adding: %v, removing: %v", add, remove)
		}

		consumerHeatBeat.setHeldShards(responseShards)
		consumerHeatBeat.heartShards = consumerHeatBeat.heldShards[:]
		TimeToSleep(int64(consumerHeatBeat.client.option.HeartbeatIntervalInSecond), lastHeartBeatTime, consumerHeatBeat.shutDownFlag)
	}
	Info.Println("heart beat exit")
}
