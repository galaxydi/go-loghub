package consumerLibrary

import (
	"strings"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// CheckPointTracker
// Generally, you just need SaveCheckPoint, if you use more funcs, make sure you understand these
type CheckPointTracker interface {
	// GetCheckPoint get lastest saved check point
	GetCheckPoint() string
	// SaveCheckPoint, save next cursor to checkpoint
	SaveCheckPoint(force bool) error
	// GetCurrentCursor get current fetched data cursor
	GetCurrentCursor() string
	// GetNextCursor get next fetched data cursor(this is also the next checkpoint to be saved)
	GetNextCursor() string
	// GetShardId, return the id of shard tracked
	GetShardId() int
}

type DefaultCheckPointTracker struct {
	client            *ConsumerClient
	heartBeat         *ConsumerHeartBeat
	nextCursor        string // cursor for already pulled data
	currentCursor     string // cursor for data processed, but may not be saved to server
	pendingCheckPoint string // pending cursor to saved
	savedCheckPoint   string // already saved
	shardId           int
	logger            log.Logger
}

func initConsumerCheckpointTracker(shardId int, consumerClient *ConsumerClient, consumerHeatBeat *ConsumerHeartBeat, logger log.Logger) *DefaultCheckPointTracker {
	checkpointTracker := &DefaultCheckPointTracker{
		client:    consumerClient,
		heartBeat: consumerHeatBeat,
		shardId:   shardId,
		logger:    logger,
	}
	return checkpointTracker
}

func (tracker *DefaultCheckPointTracker) initCheckPoint(cursor string) {
	tracker.savedCheckPoint = cursor
}

func (tracker *DefaultCheckPointTracker) SaveCheckPoint(force bool) error {
	tracker.pendingCheckPoint = tracker.nextCursor
	if force {
		return tracker.flushCheckPoint()
	}

	return nil
}

func (tracker *DefaultCheckPointTracker) GetCheckPoint() string {
	return tracker.savedCheckPoint
}

func (tracker *DefaultCheckPointTracker) GetCurrentCursor() string {
	return tracker.currentCursor
}

func (tracker *DefaultCheckPointTracker) setCurrentCursor(cursor string) {
	tracker.currentCursor = cursor
}

func (tracker *DefaultCheckPointTracker) GetNextCursor() string {
	return tracker.nextCursor
}

func (tracker *DefaultCheckPointTracker) setNextCursor(cursor string) {
	tracker.nextCursor = cursor
}

func (tracker *DefaultCheckPointTracker) GetShardId() int {
	return tracker.shardId
}

func (tracker *DefaultCheckPointTracker) flushCheckPoint() error {
	if tracker.pendingCheckPoint == "" || tracker.pendingCheckPoint == tracker.savedCheckPoint {
		return nil
	}
	for i := 0; ; i++ {
		err := tracker.client.updateCheckPoint(tracker.shardId, tracker.pendingCheckPoint, true)
		if err == nil {
			break
		}
		slsErr, ok := err.(*sls.Error)
		if ok {
			if strings.EqualFold(slsErr.Code, "ConsumerNotExsit") || strings.EqualFold(slsErr.Code, "ConsumerNotMatch") {
				tracker.heartBeat.removeHeartShard(tracker.shardId)
				level.Warn(tracker.logger).Log("msg", "consumer has been removed or shard has been reassigned", "shard", tracker.shardId, "err", slsErr)
				break
			} else if strings.EqualFold(slsErr.Code, "ShardNotExsit") {
				tracker.heartBeat.removeHeartShard(tracker.shardId)
				level.Warn(tracker.logger).Log("msg", "shard does not exist", "shard", tracker.shardId)
				break
			}
		}
		if i >= 2 {
			level.Error(tracker.logger).Log(
				"msg", "failed to save checkpoint",
				"consumer", tracker.client.option.ConsumerName,
				"shard", tracker.shardId,
				"checkpoint", tracker.pendingCheckPoint,
			)
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}

	tracker.savedCheckPoint = tracker.pendingCheckPoint
	return nil
}
