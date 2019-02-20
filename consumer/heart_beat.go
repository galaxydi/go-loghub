package consumerLibrary

import (
	"fmt"
	"time"
)

type ConsumerHeatBeat struct{
	*ConsumerClient
	*ConsumerCheckpointTracker
	HeartShutDownFlag	bool
	HeldShard		[]int
	HeartShard 		[]int
}



func InitConsumerHeatBeat(consumerClient *ConsumerClient)*ConsumerHeatBeat{
	consumerHeatBeat := &ConsumerHeatBeat{
		ConsumerClient:consumerClient,
		ConsumerCheckpointTracker:InitConsumerCheckpointTracker(consumerClient),
		HeartShutDownFlag:false,
		HeldShard:[]int{},
		HeartShard:[]int{},
	}
	return consumerHeatBeat
}




func (consumerHeatBeat *ConsumerHeatBeat)GetHeldShards()[]int{
	return consumerHeatBeat.HeartShard
}

func (consumerHeatBeat *ConsumerHeatBeat) ShutDownHeart(){
	Info.Println("try to stop heart beat")
	consumerHeatBeat.HeartShutDownFlag = true
}

func (consumerHeatBeat *ConsumerHeatBeat) RemoveHeartShard(shardId int){
	for i,x:= range consumerHeatBeat.HeartShard{
		if shardId == x{
			consumerHeatBeat.HeartShard = append(consumerHeatBeat.HeartShard[:i],consumerHeatBeat.HeartShard[i+1:]...)
		}
	}
	for i,x:= range consumerHeatBeat.HeldShard{
		if shardId == x{
			consumerHeatBeat.HeldShard = append(consumerHeatBeat.HeldShard[:i],consumerHeatBeat.HeldShard[i+1:]...)
		}
	}
}

func (consumerHeatBeat *ConsumerHeatBeat) HeartBeatRun(){
	for !consumerHeatBeat.HeartShutDownFlag{
		last_heatbeat_time := time.Now().Unix()
		fmt.Println("1111")
		if consumerHeatBeat.HeartShard == nil{
			fmt.Println("caonima")
		}
		fmt.Println(consumerHeatBeat.HeartShard)
		fmt.Println(consumerHeatBeat.Project)
		response_shards := consumerHeatBeat.MheartBeat(consumerHeatBeat.HeartShard)
		fmt.Println("2222")
		Info.Println("heart beat result: %v,get:%v",consumerHeatBeat.HeartShard,response_shards)
		// TODO 这为什么报错说不相等,想起来了，golang ，列表没办法判断是否相等
		Info.Println(consumerHeatBeat.HeartShard,consumerHeatBeat.HeldShard)
		if !IntSliceReflectEqual(consumerHeatBeat.HeartShard,consumerHeatBeat.HeldShard) {
			current_set := Set(consumerHeatBeat.HeartShard)
			response_set := Set(consumerHeatBeat.HeldShard)
			// 获得的减去当前的，等于应该增加的
			add := Subtract(current_set,response_set)
			// 当前的减去获得，等于应该放弃的
			remove := Subtract(response_set,current_set)
			Info.Printf("shard reorganize, adding: %v, removing: %v",add,remove)
		}

		consumerHeatBeat.HeldShard = response_shards
		consumerHeatBeat.HeartShard = consumerHeatBeat.HeartShard[:]

		time_to_sleep := int64(consumerHeatBeat.HeartbeatInterval) - (time.Now().Unix() - last_heatbeat_time)
		for time_to_sleep > 0 && !consumerHeatBeat.HeartShutDownFlag{
			time.Sleep(time.Duration(Min(time_to_sleep,1))*time.Second)
			time_to_sleep = int64(consumerHeatBeat.HeartbeatInterval) - (time.Now().Unix() - last_heatbeat_time)
		}
	}
	Info.Println("heart beat exit")
}
