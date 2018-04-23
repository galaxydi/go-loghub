package sls

import "encoding/json"

// InputDetail defines log_config input
// @note : deprecated and no maintenance
type InputDetail struct {
	LogType       string   `json:"logType"`
	LogPath       string   `json:"logPath"`
	FilePattern   string   `json:"filePattern"`
	LocalStorage  bool     `json:"localStorage"`
	TimeKey       string   `json:"timeKey"`
	TimeFormat    string   `json:"timeFormat"`
	LogBeginRegex string   `json:"logBeginRegex"`
	Regex         string   `json:"regex"`
	Keys          []string `json:"key"`
	FilterKeys    []string `json:"filterKey"`
	FilterRegex   []string `json:"filterRegex"`
	TopicFormat   string   `json:"topicFormat"`
}

func ConvertToInputDetail(detail InputDetailInterface) (*InputDetail, bool) {
	// ConvertToPluginLogConfigInputDetail need a plugin
	if mapVal, ok := detail.(map[string]interface{}); ok {
		if logType, ok := mapVal["logType"]; !ok || logType != "common_reg_log" {
			return nil, false
		}
	} else {
		return nil, false
	}
	buf, err := json.Marshal(detail)
	if err != nil {
		return nil, false
	}
	destDetail := &InputDetail{}
	err = json.Unmarshal(buf, destDetail)
	return destDetail, err == nil
}

type SensitiveKey struct {
	Key          string `json:"key"`
	Type         string `json:"type"`
	RegexBegin   string `json:"regex_begin"`
	RegexContent string `json:"regex_content"`
	All          bool   `json:"all"`
	ConstString  string `json:"const"`
}

// ApsaraLogConfigInputDetail apsara log config
type ApsaraLogConfigInputDetail struct {
	LocalFileConfigInputDetail
	LogBeginRegex string `json:"logBeginRegex"`
}

// InitApsaraLogConfigInputDetail ...
func InitApsaraLogConfigInputDetail(detail *ApsaraLogConfigInputDetail) {
	InitLocalFileConfigInputDetail(&detail.LocalFileConfigInputDetail)
	detail.LogBeginRegex = ".*"
	detail.LogType = "apsara_log"
}

func ConvertToApsaraLogConfigInputDetail(detail InputDetailInterface) (*ApsaraLogConfigInputDetail, bool) {
	// ConvertToPluginLogConfigInputDetail need a plugin
	if mapVal, ok := detail.(map[string]interface{}); ok {
		if logType, ok := mapVal["logType"]; !ok || logType != "apsara_log" {
			return nil, false
		}
	} else {
		return nil, false
	}
	buf, err := json.Marshal(detail)
	if err != nil {
		return nil, false
	}
	destDetail := &ApsaraLogConfigInputDetail{}
	err = json.Unmarshal(buf, destDetail)
	return destDetail, err == nil
}

// RegexConfigInputDetail regex log config
type RegexConfigInputDetail struct {
	LocalFileConfigInputDetail
	Key           []string `json:"key"`
	LogBeginRegex string   `json:"logBeginRegex"`
	Regex         string   `json:"regex"`
}

// InitRegexConfigInputDetail ...
func InitRegexConfigInputDetail(detail *RegexConfigInputDetail) {
	InitLocalFileConfigInputDetail(&detail.LocalFileConfigInputDetail)
	detail.LogBeginRegex = ".*"
	detail.LogType = "common_reg_log"
}

func ConvertToRegexConfigInputDetail(detail InputDetailInterface) (*RegexConfigInputDetail, bool) {
	// ConvertToPluginLogConfigInputDetail need a plugin
	if mapVal, ok := detail.(map[string]interface{}); ok {
		if logType, ok := mapVal["logType"]; !ok || logType != "common_reg_log" {
			return nil, false
		}
	} else {
		return nil, false
	}
	buf, err := json.Marshal(detail)
	if err != nil {
		return nil, false
	}
	destDetail := &RegexConfigInputDetail{}
	err = json.Unmarshal(buf, destDetail)
	return destDetail, err == nil
}

// JSONConfigInputDetail pure json log config
type JSONConfigInputDetail struct {
	LocalFileConfigInputDetail
	TimeKey string `json:"timeKey"`
}

// InitJSONConfigInputDetail ...
func InitJSONConfigInputDetail(detail *JSONConfigInputDetail) {
	InitLocalFileConfigInputDetail(&detail.LocalFileConfigInputDetail)
	detail.LogType = "json_log"
}

func ConvertToJSONConfigInputDetail(detail InputDetailInterface) (*JSONConfigInputDetail, bool) {
	// ConvertToPluginLogConfigInputDetail need a plugin
	if mapVal, ok := detail.(map[string]interface{}); ok {
		if logType, ok := mapVal["logType"]; !ok || logType != "json_log" {
			return nil, false
		}
	} else {
		return nil, false
	}
	buf, err := json.Marshal(detail)
	if err != nil {
		return nil, false
	}
	destDetail := &JSONConfigInputDetail{}
	err = json.Unmarshal(buf, destDetail)
	return destDetail, err == nil
}

// DelimiterConfigInputDetail delimiter log config
type DelimiterConfigInputDetail struct {
	LocalFileConfigInputDetail
	Separator  string   `json:"separator"`
	Quote      string   `json:"quote"`
	Key        []string `json:"key"`
	TimeKey    string   `json:"timeKey"`
	AutoExtend bool     `json:"autoExtend"`
}

// InitDelimiterConfigInputDetail ...
func InitDelimiterConfigInputDetail(detail *DelimiterConfigInputDetail) {
	InitLocalFileConfigInputDetail(&detail.LocalFileConfigInputDetail)
	detail.Quote = `\u001`
	detail.AutoExtend = true
	detail.LogType = "delimiter_log"
}

func ConvertToDelimiterConfigInputDetail(detail InputDetailInterface) (*DelimiterConfigInputDetail, bool) {
	// ConvertToPluginLogConfigInputDetail need a plugin
	if mapVal, ok := detail.(map[string]interface{}); ok {
		if logType, ok := mapVal["logType"]; !ok || logType != "delimiter_log" {
			return nil, false
		}
	} else {
		return nil, false
	}
	buf, err := json.Marshal(detail)
	if err != nil {
		return nil, false
	}
	destDetail := &DelimiterConfigInputDetail{}
	err = json.Unmarshal(buf, destDetail)
	return destDetail, err == nil
}

// LocalFileConfigInputDetail all file input detail's basic config
type LocalFileConfigInputDetail struct {
	CommonConfigInputDetail
	LogType            string            `json:"logType"`
	LogPath            string            `json:"logPath"`
	FilePattern        string            `json:"filePattern"`
	TimeFormat         string            `json:"timeFormat"`
	TopicFormat        string            `json:"topicFormat"`
	Preserve           bool              `json:"preserve"`
	PreserveDepth      int               `json:"preserveDepth"`
	FileEncoding       string            `json:"fileEncoding"`
	DiscardUnmatch     bool              `json:"discardUnmatch"`
	MaxDepth           int               `json:"maxDepth"`
	TailExisted        bool              `json:"tailExisted"`
	DiscardNonUtf8     bool              `json:"discardNonUtf8"`
	DelaySkipBytes     int               `json:"delaySkipBytes"`
	IsDockerFile       bool              `json:"dockerFile"`
	DockerIncludeLabel map[string]string `json:"dockerIncludeLabel"`
	DockerExcludeLabel map[string]string `json:"dockerExcludeLabel"`
	DockerIncludeEnv   map[string]string `json:"dockerIncludeEnv"`
	DockerExcludeEnv   map[string]string `json:"dockerExcludeEnv"`
}

// InitLocalFileConfigInputDetail ...
func InitLocalFileConfigInputDetail(detail *LocalFileConfigInputDetail) {
	InitCommonConfigInputDetail(&detail.CommonConfigInputDetail)
	detail.FileEncoding = "utf8"
	detail.MaxDepth = 100
	detail.TopicFormat = "none"
	detail.Preserve = true
}

// PluginLogConfigInputDetail plugin log config, eg: docker stdout, binlog, mysql, http...
type PluginLogConfigInputDetail struct {
	CommonConfigInputDetail
	PluginDetail string `json:"plugin"`
}

// InitPluginLogConfigInputDetail ...
func InitPluginLogConfigInputDetail(detail *PluginLogConfigInputDetail) {
	InitCommonConfigInputDetail(&detail.CommonConfigInputDetail)
}

func ConvertToPluginLogConfigInputDetail(detail InputDetailInterface) (*PluginLogConfigInputDetail, bool) {
	// ConvertToPluginLogConfigInputDetail need a plugin
	if mapVal, ok := detail.(map[string]interface{}); ok {
		if _, ok := mapVal["plugin"]; !ok {
			return nil, false
		}
	} else {
		return nil, false
	}
	buf, err := json.Marshal(detail)
	if err != nil {
		return nil, false
	}
	destDetail := &PluginLogConfigInputDetail{}
	err = json.Unmarshal(buf, destDetail)
	return destDetail, err == nil
}

// StreamLogConfigInputDetail syslog config
type StreamLogConfigInputDetail struct {
	CommonConfigInputDetail
	Tag string `json:"tag"`
}

// InitStreamLogConfigInputDetail ...
func InitStreamLogConfigInputDetail(detail *StreamLogConfigInputDetail) {
	InitCommonConfigInputDetail(&detail.CommonConfigInputDetail)
}

func ConvertToStreamLogConfigInputDetail(detail InputDetailInterface) (*StreamLogConfigInputDetail, bool) {
	// ConvertToStreamLogConfigInputDetail need a tag
	if mapVal, ok := detail.(map[string]interface{}); ok {
		if _, ok := mapVal["tag"]; !ok {
			return nil, false
		}
	} else {
		return nil, false
	}
	buf, err := json.Marshal(detail)
	if err != nil {
		return nil, false
	}
	destDetail := &StreamLogConfigInputDetail{}
	err = json.Unmarshal(buf, destDetail)
	return destDetail, err == nil
}

// CommonConfigInputDetail is all input detail's basic config
type CommonConfigInputDetail struct {
	LocalStorage    bool           `json:"localStorage"`
	FilterKeys      []string       `json:"filterKey"`
	FilterRegex     []string       `json:"filterRegex"`
	ShardHashKey    []string       `json:"shardHashKey"`
	EnableTag       bool           `json:"enableTag"`
	EnableRawLog    bool           `json:"enableRawLog"`
	MaxSendRate     int            `json:"maxSendRate"`
	SendRateExpire  int            `json:"sendRateExpire"`
	SensitiveKeys   []SensitiveKey `json:"sensitive_keys"`
	MergeType       string         `json:"mergeType"`
	DelayAlarmBytes int            `json:"delayAlarmBytes"`
	AdjustTimeZone  bool           `json:"adjustTimezone"`
	LogTimeZone     string         `json:"logTimezone"`
	Priority        int            `json:"priority"`
}

// InitCommonConfigInputDetail ...
func InitCommonConfigInputDetail(detail *CommonConfigInputDetail) {
	detail.LocalStorage = true
	detail.EnableTag = true
	detail.MaxSendRate = -1
}

// OutputDetail defines output
type OutputDetail struct {
	ProjectName  string `json:"projectName"`
	LogStoreName string `json:"logstoreName"`
}

// InputDetailInterface all input detail's interface
type InputDetailInterface interface {
}

// LogConfig defines log config
type LogConfig struct {
	Name         string               `json:"configName"`
	LogSample    string               `json:"logSample"`
	InputType    string               `json:"inputType"` // syslog plugin file
	InputDetail  InputDetailInterface `json:"inputDetail"`
	OutputType   string               `json:"outputType"`
	OutputDetail OutputDetail         `json:"outputDetail"`

	CreateTime     uint32 `json:"createTime"`
	LastModifyTime uint32 `json:"lastModifyTime"`
}
