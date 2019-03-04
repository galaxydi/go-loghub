package consumerLibrary

import (
	"time"
)

type ConsumerCheckPointTracker struct {
	client                            *ConsumerClient
	defaultFlushCheckPointIntervalSec int64
	tempCheckPoint                    string
	lastPersistentCheckPoint          string
	trackerShardId                    int
	lastCheckTime                     int64
}

func initConsumerCheckpointTracker(shardId int, consumerClient *ConsumerClient) *ConsumerCheckPointTracker {
	checkpointTracker := &ConsumerCheckPointTracker{
		defaultFlushCheckPointIntervalSec: 60,
		client:                            consumerClient,
		trackerShardId:                    shardId,
	}
	return checkpointTracker
}

func (checkPointTracker *ConsumerCheckPointTracker) setMemoryCheckPoint(cursor string) {
	checkPointTracker.tempCheckPoint = cursor
}

func (checkPointTracker *ConsumerCheckPointTracker) setPersistentCheckPoint(cursor string) {
	checkPointTracker.lastPersistentCheckPoint = cursor
}

func (checkPointTracker *ConsumerCheckPointTracker) flushCheckPoint() {
	if checkPointTracker.tempCheckPoint != "" && checkPointTracker.tempCheckPoint != checkPointTracker.lastPersistentCheckPoint {
		checkPointTracker.client.updateCheckPoint(checkPointTracker.trackerShardId, checkPointTracker.tempCheckPoint, true)
		checkPointTracker.lastPersistentCheckPoint = checkPointTracker.tempCheckPoint
	}
}

func (checkPointTracker *ConsumerCheckPointTracker) flushCheck() {
	current_time := time.Now().Unix()
	if current_time > checkPointTracker.lastCheckTime+checkPointTracker.defaultFlushCheckPointIntervalSec {
		checkPointTracker.flushCheckPoint()
		checkPointTracker.lastCheckTime = current_time
	}
}

func (checkPointTracker *ConsumerCheckPointTracker) getCheckPoint() string {
	return checkPointTracker.tempCheckPoint
}
