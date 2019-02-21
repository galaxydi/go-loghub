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

func InitConsumerHeatBeat(consumerClient *ConsumerClient) *ConsumerHeatBeat {
	consumerHeatBeat := &ConsumerHeatBeat{
		ConsumerClient:    consumerClient,
		HeartShutDownFlag: false,
		HeldShard:         []int{},
		HeartShard:        []int{},
	}
	return consumerHeatBeat
}

func (consumerHeatBeat *ConsumerHeatBeat) GetHeldShards() []int {
	return consumerHeatBeat.HeartShard
}

func (consumerHeatBeat *ConsumerHeatBeat) ShutDownHeart() {
	Info.Println("try to stop heart beat")
	consumerHeatBeat.HeartShutDownFlag = true
}

func (consumerHeatBeat *ConsumerHeatBeat) RemoveHeartShard(shardId int) {
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

func (consumerHeatBeat *ConsumerHeatBeat) HeartBeatRun() {
	for !consumerHeatBeat.HeartShutDownFlag {
		last_heatbeat_time := time.Now().Unix()
		response_shards := consumerHeatBeat.MheartBeat(consumerHeatBeat.HeartShard)
		Info.Printf("heart beat result: %v,get:%v", consumerHeatBeat.HeartShard, response_shards)

		if !IntSliceReflectEqual(consumerHeatBeat.HeartShard, consumerHeatBeat.HeldShard) {
			current_set := Set(consumerHeatBeat.HeartShard)
			Info.Println(current_set, "current_set")
			response_set := Set(consumerHeatBeat.HeldShard)
			Info.Println(response_set, "response_set")
			add := Subtract(current_set, response_set)
			Info.Println(add)
			remove := Subtract(response_set, current_set)
			Info.Println(remove)
			Info.Printf("shard reorganize, adding: %v, removing: %v", add, remove)
		}

		consumerHeatBeat.HeldShard = response_shards

		consumerHeatBeat.HeartShard = consumerHeatBeat.HeldShard[:]

		time_to_sleep := int64(consumerHeatBeat.HeartbeatInterval) - (time.Now().Unix() - last_heatbeat_time)
		for time_to_sleep > 0 && !consumerHeatBeat.HeartShutDownFlag {
			time.Sleep(time.Duration(Min(time_to_sleep, 1)) * time.Second)
			time_to_sleep = int64(consumerHeatBeat.HeartbeatInterval) - (time.Now().Unix() - last_heatbeat_time)
		}
	}
	Info.Println("heart beat exit")
}
