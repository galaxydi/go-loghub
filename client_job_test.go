package sls

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

func TestIngestion(t *testing.T) {
	suite.Run(t, new(IngestionTestSuite))
}

type IngestionTestSuite struct {
	suite.Suite
	endpoint        string
	projectName     string
	logstoreName    string
	accessKeyID     string
	accessKeySecret string
	client          Client
}

func (i *IngestionTestSuite) SetupSuite() {
	i.endpoint = os.Getenv("LOG_TEST_ENDPOINT")
	i.projectName = os.Getenv("LOG_TEST_PROJECT")
	i.logstoreName = os.Getenv("LOG_TEST_LOGSTORE")
	i.accessKeyID = os.Getenv("LOG_TEST_ACCESS_KEY_ID")
	i.accessKeySecret = os.Getenv("LOG_TEST_ACCESS_KEY_SECRET")
	i.client.AccessKeyID = i.accessKeyID
	i.client.AccessKeySecret = i.accessKeySecret
	i.client.Endpoint = i.endpoint
	if _, err := i.client.CreateProject(i.projectName, ""); err != nil {
		i.FailNowf("create project failed", fmt.Sprintf("%v", err))
	}
	time.Sleep(time.Second * 10)
	if err := i.client.CreateLogStore(i.projectName, i.logstoreName, 3, 2, false, 4); err != nil {
		i.FailNowf("create logstore failed", fmt.Sprintf("%v", err))
	}
	time.Sleep(time.Minute)
}

func (i *IngestionTestSuite) TearDownSuite() {
	i.client.DeleteLogStore(i.projectName, i.logstoreName)
	i.client.DeleteProject(i.projectName)
}

func (i *IngestionTestSuite) TestIngestionOSS_CRUD() {
	ingestion := getOssIngestion(i.logstoreName)
	if err := i.client.CreateIngestion(i.projectName, ingestion); err != nil {
		i.FailNowf("create ingestion failed", fmt.Sprintf("%v", err))
	}
	ingestion.Description = "test"
	if err := i.client.UpdateIngestion(i.projectName, ingestion); err != nil {
		i.FailNowf("update ingestion failed", fmt.Sprintf("%v", err))
	} else if getIngestion, err := i.client.GetIngestion(i.projectName, ingestion.Name); err != nil {
		i.FailNowf("get ingestion failed", fmt.Sprintf("%v", err))
	} else {
		i.Equal(ingestion.Name, getIngestion.Name)
		i.Equal(ingestion.DisplayName, getIngestion.DisplayName)
		i.Equal("test", getIngestion.Description)
		i.Equal(INGESTION_JOB, getIngestion.Type)
		i.Equal(false, getIngestion.Recyclable)
		i.Equal("ENABLED", getIngestion.Status)
		i.Equal("FixedRate", getIngestion.Schedule.Type)
		i.Equal("5m", getIngestion.Schedule.Interval)
		i.Equal(true, getIngestion.Schedule.RunImmediately)
		i.Equal("+0800", getIngestion.Schedule.TimeZone)
		i.Equal("test-logstore", getIngestion.IngestionConfiguration.LogStore)
		source := &AliyunOSSSource{}
		if sourceBytes, err := json.Marshal(getIngestion.IngestionConfiguration.DataSource); err != nil {
			i.FailNowf("marshal datasource failed", fmt.Sprintf("%v", err))
		} else if err = json.Unmarshal(sourceBytes, source); err != nil {
			i.FailNowf("unmarshal AliyunOSSSource failed", fmt.Sprintf("%v", err))
		}
		i.Equal(DataSourceOSS, source.DataSourceType)
		i.Equal("test-bucket", source.Bucket)
		i.Equal("snappy", source.CompressionCodec)
		i.Equal("UTF-8", source.Encoding)
		i.Equal("test-endpoint", source.Endpoint)
		i.Equal("test-pattern", source.Pattern)
		i.Equal("test-prefix", source.Prefix)
		i.Equal(false, source.RestoreObjectEnable)
		i.Equal("test-roleArn", source.RoleArn)
		format := &LineFormat{}
		if formatBytes, err := json.Marshal(source.Format); err != nil {
			i.FailNowf("marshal format failed", fmt.Sprintf("%v", err))
		} else if err = json.Unmarshal(formatBytes, format); err != nil {
			i.FailNowf("marshal LineFormat failed", fmt.Sprintf("%v", err))
		}
		i.Equal(OSSDataFormatTypeLine, format.Type)
		i.Equal("yyyy-MM-dd", format.TimeFormat)
		i.Equal("test-timePattern", format.TimePattern)
		i.Equal("+0800", format.TimeZone)
	}
	if _, total, count, err := i.client.ListIngestion(i.projectName, i.logstoreName, "", "", 0, 10); err != nil {
		i.FailNowf("list ingestion failed", fmt.Sprintf("%v", err))
	} else {
		i.Equal(1, total)
		i.Equal(1, count)
	}
	if err := i.client.DeleteIngestion(i.projectName, ingestion.Name); err != nil {
		i.FailNowf("delete ingestion failed", fmt.Sprintf("%v", err))
	}
}

func getOssIngestion(logstore string) *Ingestion {
	timeUnix := time.Now().Unix()
	ingestion := &Ingestion{
		ScheduledJob: ScheduledJob{
			BaseJob: BaseJob{
				Name:        fmt.Sprintf("test-oss-ingest-%d", timeUnix),
				DisplayName: "test-oss-ingest",
				Description: "",
				Type:        INGESTION_JOB,
			},
			Schedule: &Schedule{
				Type:           "FixedRate",
				Interval:       "5m",
				Delay:          0,
				RunImmediately: true,
				TimeZone:       "+0800",
			},
		},
		IngestionConfiguration: &IngestionConfiguration{
			LogStore: logstore,
			DataSource: AliyunOSSSource{
				DataSource:       DataSource{DataSourceOSS},
				Bucket:           "test-bucket",
				Endpoint:         "test-endpoint",
				RoleArn:          "test-roleArn",
				Prefix:           "test-prefix",
				Pattern:          "test-pattern",
				CompressionCodec: "snappy",
				Encoding:         "UTF-8",
				Format: LineFormat{
					OSSDataFormat: OSSDataFormat{
						Type:       OSSDataFormatTypeLine,
						TimeFormat: "yyyy-MM-dd",
						TimeZone:   "+0800",
					},
					TimePattern: "test-timePattern",
				},
				RestoreObjectEnable:     true,
				LastModifyTimeAsLogTime: true,
			},
		},
	}
	return ingestion
}
