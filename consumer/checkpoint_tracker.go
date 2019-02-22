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

func initConsumerCheckpointTracker(shardId int, consumerClient *ConsumerClient) *ConsumerCheckPointTracker {
	checkpointTracker := &ConsumerCheckPointTracker{
		DefaultFlushCheckPointInterval: 60,
		ConsumerClient:                 consumerClient,
		TrackerShardId:                 shardId,
	}
	return checkpointTracker
}

func (checkPointTracker *ConsumerCheckPointTracker) setMemoryCheckPoint(cursor string) {
	checkPointTracker.TempCheckPoint = cursor
}

func (checkPointTracker *ConsumerCheckPointTracker) setPersistentCheckPoint(cursor string) {
	checkPointTracker.LastPersistentCheckPoint = cursor
}

func (checkPointTracker *ConsumerCheckPointTracker) mFlushCheckPoint() {
	if checkPointTracker.TempCheckPoint != "" && checkPointTracker.TempCheckPoint != checkPointTracker.LastPersistentCheckPoint {
		checkPointTracker.mUpdateCheckPoint(checkPointTracker.TrackerShardId, checkPointTracker.TempCheckPoint, true)
		checkPointTracker.LastPersistentCheckPoint = checkPointTracker.TempCheckPoint
	}
}

func (checkPointTracker *ConsumerCheckPointTracker) flushCheck() {
	current_time := time.Now().Unix()
	if current_time > checkPointTracker.LastCheckTime+checkPointTracker.DefaultFlushCheckPointInterval {
		checkPointTracker.mFlushCheckPoint()
		checkPointTracker.LastCheckTime = current_time
	}
}

func (checkPointTracker *ConsumerCheckPointTracker) getCheckPoint() string {
	return checkPointTracker.TempCheckPoint
}
