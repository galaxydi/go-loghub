package consumerLibrary

import (
	"time"
)

type ConsumerHeatBeat struct {
	*ConsumerClient
	HeartShutDownFlag bool
	HeldShard         []int
	HeartShard        []int
}

func initConsumerHeatBeat(consumerClient *ConsumerClient) *ConsumerHeatBeat {
	consumerHeatBeat := &ConsumerHeatBeat{
		ConsumerClient:    consumerClient,
		HeartShutDownFlag: false,
		HeldShard:         []int{},
		HeartShard:        []int{},
	}
	return consumerHeatBeat
}

func (consumerHeatBeat *ConsumerHeatBeat) getHeldShards() []int {
	return consumerHeatBeat.HeartShard
}

func (consumerHeatBeat *ConsumerHeatBeat) shutDownHeart() {
	Info.Println("try to stop heart beat")
	consumerHeatBeat.HeartShutDownFlag = true
}

func (consumerHeatBeat *ConsumerHeatBeat) removeHeartShard(shardId int) {
	for i, x := range consumerHeatBeat.HeartShard {
		if shardId == x {
			consumerHeatBeat.HeartShard = append(consumerHeatBeat.HeartShard[:i], consumerHeatBeat.HeartShard[i+1:]...)
		}
	}
	for i, x := range consumerHeatBeat.HeldShard {
		if shardId == x {
			consumerHeatBeat.HeldShard = append(consumerHeatBeat.HeldShard[:i], consumerHeatBeat.HeldShard[i+1:]...)
		}
	}
}

func (consumerHeatBeat *ConsumerHeatBeat) heartBeatRun() {
	for !consumerHeatBeat.HeartShutDownFlag {
		lastHeatbeatTime := time.Now().Unix()
		responseShards := consumerHeatBeat.mHeartBeat(consumerHeatBeat.HeartShard)
		Info.Printf("heart beat result: %v,get:%v", consumerHeatBeat.HeartShard, responseShards)

		if !IntSliceReflectEqual(consumerHeatBeat.HeartShard, consumerHeatBeat.HeldShard) {
			currentSet := Set(consumerHeatBeat.HeartShard)
			responseSet := Set(consumerHeatBeat.HeldShard)
			add := Subtract(currentSet, responseSet)
			remove := Subtract(responseSet, currentSet)
			Info.Printf("shard reorganize, adding: %v, removing: %v", add, remove)
		}

		consumerHeatBeat.HeldShard = responseShards

		consumerHeatBeat.HeartShard = consumerHeatBeat.HeldShard[:]

		timeToSleep := int64(consumerHeatBeat.HeartbeatInterval) - (time.Now().Unix() - lastHeatbeatTime)
		for timeToSleep > 0 && !consumerHeatBeat.HeartShutDownFlag {
			time.Sleep(time.Duration(Min(timeToSleep, 1)) * time.Second)
			timeToSleep = int64(consumerHeatBeat.HeartbeatInterval) - (time.Now().Unix() - lastHeatbeatTime)
		}
	}
	Info.Println("heart beat exit")
}
