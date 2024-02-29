package consumerLibrary

import sls "github.com/aliyun/aliyun-log-go-sdk"

type Processor interface {
	Process(int, *sls.LogGroupList, CheckPointTracker) (string, error)
	Shutdown(CheckPointTracker) error
}

type ProcessFunc func(int, *sls.LogGroupList, CheckPointTracker) (string, error)

func (processor ProcessFunc) Process(shard int, lgList *sls.LogGroupList, checkpointTracker CheckPointTracker) (string, error) {
	return processor(shard, lgList, checkpointTracker)
}

func (processor ProcessFunc) Shutdown(checkpointTracker CheckPointTracker) error {
	// Do nothing
	return nil
}
