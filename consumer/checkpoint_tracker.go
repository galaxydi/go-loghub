package consumerLibrary

import (
	"time"
)

type ConsumerCheckpointTracker struct{
	*ConsumerClient
	DefaultFlushCheckPointInterval int64
	TempCheckPoint string
	LastPersistentCheckPoint string
	ShardId int
	LastCheckTime int64
}


func InitConsumerCheckpointTracker(consumerClient *ConsumerClient)*ConsumerCheckpointTracker{
	checkpointTracker:= &ConsumerCheckpointTracker{
		DefaultFlushCheckPointInterval:60,
		ConsumerClient:consumerClient,
	}
	return checkpointTracker
}


func (checkPointTracker *ConsumerCheckpointTracker) SetMemoryCheckPoint(cursor string){
	checkPointTracker.TempCheckPoint = cursor
}


func (checkPointTracker *ConsumerCheckpointTracker) SetPersistentCheckPoint(cursor string){
	checkPointTracker.LastPersistentCheckPoint = cursor
}



func (checkPointTracker *ConsumerCheckpointTracker) MflushCheckPoint(){
	if checkPointTracker.TempCheckPoint != "" && checkPointTracker.TempCheckPoint != checkPointTracker.LastPersistentCheckPoint{
		checkPointTracker.MupdateCheckPoint(checkPointTracker.ShardId,checkPointTracker.TempCheckPoint,true)
		checkPointTracker.LastPersistentCheckPoint = checkPointTracker.TempCheckPoint
	}
}


func (checkPointTracker *ConsumerCheckpointTracker) FlushCheck(){
	current_time := time.Now().Unix()
	if current_time > checkPointTracker.LastCheckTime + checkPointTracker.DefaultFlushCheckPointInterval{
		checkPointTracker.MflushCheckPoint()
		checkPointTracker.LastCheckTime = current_time
	}
}


func (checkPointTracker *ConsumerCheckpointTracker) GetCheckPoint()string{
	return checkPointTracker.TempCheckPoint
}
