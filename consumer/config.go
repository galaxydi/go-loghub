package consumerLibrary

type LogHubConfig struct {
	Endpoint             string
	AccessKeyID          string
	AccessKeySecret      string
	Project              string
	Logstore             string
	MConsumerGroupName   string
	ConsumerName         string
	CursorPosition       string
	HeartbeatInterval    int
	DataFetchInterval    int64
	MaxFetchLogGroupSize int
	CursorStarttime      string
	InOrder              bool
	SecurityToken        string // TODO need security_token ?
}

const (
	BEGIN_CURSOR         = "begin"
	END_CURSOR           = "end"
	SPECIAL_TIMER_CURSOR = "SPECIAL_TIMER_CURSOR"
	INITIALIZ            = "INITIALIZ"
	PROCESS              = "PROCESS"
	SHUTTING_DOWN        = "SHUTTING_DOWN"
	SHUTDOWN_COMPLETE    = "SHUTDOWN_COMPLETE"
)
