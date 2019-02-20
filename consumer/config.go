package consumerLibrary



type LogHubConfig struct {
	Endpoint string
	AccessKeyID string
	AccessKeySecret string
	Project string
	Logstore string
	MConsumerGroupName string
	ConsumerName string
	CursorPosition string
	HeartbeatInterval  int
	DataFetchInterval	int64
	MaxFetchLogGroupSize int
	CursorStarttime string
	InOrder 		bool    // TODO 是否按序消费，我暂时没用到这个参数
	// security_token 还有这个参数我没有用
	SecurityToken string
}

const (
	BEGIN_CURSOR = "begin"
    END_CURSOR = "end"
    SPECIAL_TIMER_CURSOR = "SPECIAL_TIMER_CURSOR"
	INITIALIZ = "INITIALIZ"
    PROCESS = "PROCESS"
    SHUTTING_DOWN = "SHUTTING_DOWN"
    SHUTDOWN_COMPLETE = "SHUTDOWN_COMPLETE"
)