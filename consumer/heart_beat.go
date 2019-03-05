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
	return consumerHeatBeat.heldShards
}

func (consumerHeatBeat *ConsumerHeatBeat) setHeldShards(heldShards []int) {
	m.Lock()
	defer m.Unlock()
	consumerHeatBeat.heldShards = heldShards
}

func (consumerHeatBeat *ConsumerHeatBeat) setHeartShards(heartShards []int){
	m.Lock()
	defer m.Unlock()
	consumerHeatBeat.heartShards = heartShards
}


func (consumerHeatBeat *ConsumerHeatBeat) shutDownHeart() {
	Info.Println("try to stop heart beat")
	consumerHeatBeat.shutDownFlag = true
}

func (consumerHeatBeat *ConsumerHeatBeat) removeHeartShard(shardId int) bool {
	var isDeleteShard bool
	for i, heartShard := range consumerHeatBeat.heartShards {
		if shardId == heartShard {
			heartShards := append(consumerHeatBeat.heartShards[:i], consumerHeatBeat.heartShards[i+1:]...)
			consumerHeatBeat.setHeartShards(heartShards)
			isDeleteShard = true
			break
		}
	}
	return isDeleteShard
}

func (consumerHeatBeat *ConsumerHeatBeat) heartBeatRun() {
	var lastHeartBeatTime int64

	for !consumerHeatBeat.shutDownFlag {
		lastHeartBeatTime = time.Now().Unix()
		responseShards, err := consumerHeatBeat.client.heartBeat(consumerHeatBeat.heartShards)
		if err != nil {
			Warning.Println(err)
		} else {
			Info.Printf("heart beat result: %v,get:%v", consumerHeatBeat.heartShards, responseShards)

			if !IntSliceReflectEqual(consumerHeatBeat.heartShards, consumerHeatBeat.heldShards) {
				currentSet := Set(consumerHeatBeat.heartShards)
				responseSet := Set(consumerHeatBeat.heldShards)
				add := Subtract(currentSet, responseSet)
				remove := Subtract(responseSet, currentSet)
				Info.Printf("shard reorganize, adding: %v, removing: %v", add, remove)
			}

			consumerHeatBeat.setHeldShards(responseShards)
			consumerHeatBeat.setHeartShards(consumerHeatBeat.getHeldShards()[:])
		}
		TimeToSleep(int64(consumerHeatBeat.client.option.HeartbeatIntervalInSecond), lastHeartBeatTime, consumerHeatBeat.shutDownFlag)
	}
	Info.Println("heart beat exit")
}
