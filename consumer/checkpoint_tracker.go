package consumerLibrary

import (
	"time"
)

type ConsumerCheckPointTracker struct {
	*ConsumerClient
	DefaultFlushCheckPointInterval int64
	TempCheckPoint                 string
	LastPersistentCheckPoint       string
	TrackerShardId                 int
	LastCheckTime                  int64
}

func InitConsumerCheckpointTracker(shardId int, consumerClient *ConsumerClient) *ConsumerCheckPointTracker {
	checkpointTracker := &ConsumerCheckPointTracker{
		DefaultFlushCheckPointInterval: 60,
		ConsumerClient:                 consumerClient,
		TrackerShardId:                 shardId,
	}
	return checkpointTracker
}

func (checkPointTracker *ConsumerCheckPointTracker) SetMemoryCheckPoint(cursor string) {
	checkPointTracker.TempCheckPoint = cursor
}

func (checkPointTracker *ConsumerCheckPointTracker) SetPersistentCheckPoint(cursor string) {
	checkPointTracker.LastPersistentCheckPoint = cursor
}

func (checkPointTracker *ConsumerCheckPointTracker) MflushCheckPoint() {
	if checkPointTracker.TempCheckPoint != "" && checkPointTracker.TempCheckPoint != checkPointTracker.LastPersistentCheckPoint {
		checkPointTracker.MupdateCheckPoint(checkPointTracker.TrackerShardId, checkPointTracker.TempCheckPoint, true)
		checkPointTracker.LastPersistentCheckPoint = checkPointTracker.TempCheckPoint
	}
}

func (checkPointTracker *ConsumerCheckPointTracker) FlushCheck() {
	current_time := time.Now().Unix()
	if current_time > checkPointTracker.LastCheckTime+checkPointTracker.DefaultFlushCheckPointInterval {
		checkPointTracker.MflushCheckPoint()
		checkPointTracker.LastCheckTime = current_time
	}
}

func (checkPointTracker *ConsumerCheckPointTracker) GetCheckPoint() string {
	return checkPointTracker.TempCheckPoint
}
