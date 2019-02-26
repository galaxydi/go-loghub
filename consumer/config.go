package consumerLibrary

type LogHubConfig struct {
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
	BEGIN_CURSOR         = "BEGIN_CURSOR"
	END_CURSOR           = "END_CURSOR"
	SPECIAL_TIMER_CURSOR = "SPECIAL_TIMER_CURSOR"
	INITIALIZED          = "INITIALIZED"
	PROCESSING           = "PROCESSING"
	SHUTTING_DOWN        = "SHUTTING_DOWN"
	SHUTDOWN_COMPLETE    = "SHUTDOWN_COMPLETE"
)

const (
	channelA = iota
	channelB
	channelC
)
