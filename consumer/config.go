package consumer



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
	DataFetchInterval	int
	MaxFetchLogGroupSize int
	CursorStarttime string
	InOrder 		bool    // TODO 是否按序消费，我暂时没用到这个参数
	// security_token 还有这个参数我没有用
}