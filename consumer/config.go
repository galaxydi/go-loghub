package consumerLibrary

import "sync"

type LogHubConfig struct {
	//:param endpoint:
	//:param access_key_id:
	//:param access_key:
	//:param project:
	//:param logstore:
	//:param consumer_group_name:
	//:param consumer_name:
	//:param cursor_position: This options is used for initialization, will be ignored once consumer group is created and each shard has beeen started to be consumed.
	//:param heartbeat_interval: default 20, once a client doesn't report to server * heartbeat_interval * 2 interval, server will consider it's offline and re-assign its task to another consumer. thus  don't set the heatbeat interval too small when the network badwidth or performance of consumtion is not so good.
	//:param data_fetch_interval: default 2, don't configure it too small (<1s)
	//:param in_order: default False, during consuption, when shard is splitted, if need to consume the newly splitted shard after its parent shard (read-only) is finished consumption or not. suggest keep it as False (don't care) until you have good reasion for it.
	//:param cursor_start_time: Will be used when cursor_position when could be "begin", "end", "specific time format in time stamp", it's log receiving time.
	//:param security_token:
	//:param max_fetch_log_group_size: default 1000, fetch size in each request, normally use default. maximum is 1000, could be lower. the lower the size the memory efficiency might be better.
	//:param worker_pool_size: default 2. suggest keep the default size (2), use multiple process instead, when you need to have more concurrent processing, launch this consumer for mulitple times and give them different consuer name in same consumer group. will be ignored when shared_executor is passed.

	Endpoint                  string
	AccessKeyID               string
	AccessKeySecret           string
	Project                   string
	Logstore                  string
	ConsumerGroupName         string
	ConsumerName              string
	CursorPosition            string
	HeartbeatIntervalInSecond int
	DataFetchInterval         int64
	MaxFetchLogGroupCount     int
	CursorStartTime           int64 // Unix time stamp
	InOrder                   bool
	// SecurityToken        string
}

const (
	BEGIN_CURSOR            = "BEGIN_CURSOR"
	END_CURSOR              = "END_CURSOR"
	SPECIAL_TIMER_CURSOR    = "SPECIAL_TIMER_CURSOR"
	INITIALIZING            = "INITIALIZING"
	INITIALIZING_DONE       = "INITIALIZING_DONE"
	PULL_PROCESSING         = "PULL_PROCESSING"
	PULL_PROCESSING_DONE    = "PULL_PROCESSING_DONE"
	CONSUME_PROCESSING      = "CONSUME_PROCESSING"
	CONSUME_PROCESSING_DONE = "CONSUME_PROCESSING_DONE"

	SHUTDOWN_COMPLETE = "SHUTDOWN_COMPLETE"
)

var m sync.RWMutex
