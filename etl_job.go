package sls

type ETLJob struct {
	JobName           string          `json:"etlJobName"`
	SourceConfig      *SourceConfig   `json:"sourceConfig"`
	TriggerConfig     *TriggerConfig  `json:"triggerConfig"`
	FunctionConfig    *FunctionConfig `json:"functionConfig"`
	FunctionParameter string          `json:"functionParameter"`
	LogConfig         *JobLogConfig   `json:"logConfig"`
	Enable            bool            `json:"enable"`
}

type SourceConfig struct {
	LogstoreName string `json:"logstoreName"`
}

type TriggerConfig struct {
	MaxRetryTime    int    `json:"maxRetryTime"`
	TriggerInterval int    `json:"triggerInterval"`
	RoleARN         string `json:"roleArn"`
}

type FunctionConfig struct {
	FunctionProvider string `json:"functionProvider"`
	Endpoint         string `json:"endpoint"`
	AccountID        string `json:"accountId"`
	RegionName       string `json:"regionName"`
	ServiceName      string `json:"serviceName"`
	FunctionName     string `json:"functionName"`
}

type JobLogConfig struct {
	Endpoint     string `json:"endpoint"`
	ProjectName  string `json:"projectName"`
	LogstoreName string `json:"logstoreName"`
}
