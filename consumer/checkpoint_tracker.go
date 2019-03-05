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

func (checkPointTracker *ConsumerCheckPointTracker) flushCheckPoint() error {
	if checkPointTracker.tempCheckPoint != "" && checkPointTracker.tempCheckPoint != checkPointTracker.lastPersistentCheckPoint {
		if err := checkPointTracker.client.updateCheckPoint(checkPointTracker.trackerShardId, checkPointTracker.tempCheckPoint, true); err != nil {
			return err
		}
		checkPointTracker.lastPersistentCheckPoint = checkPointTracker.tempCheckPoint
	}
	return nil
}

func (checkPointTracker *ConsumerCheckPointTracker) flushCheck() {
	currentTime := time.Now().Unix()
	if currentTime > checkPointTracker.lastCheckTime+checkPointTracker.defaultFlushCheckPointIntervalSec {
		if err := checkPointTracker.flushCheckPoint(); err != nil {
			checkPointTracker.lastCheckTime = currentTime
		}
	}
}

func (checkPointTracker *ConsumerCheckPointTracker) getCheckPoint() string {
	return checkPointTracker.tempCheckPoint
}
