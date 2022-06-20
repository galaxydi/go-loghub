package sls

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

func TestMetricStore(t *testing.T) {
	suite.Run(t, new(MetricStoreTestSuite))
}

type MetricStoreTestSuite struct {
	suite.Suite
	endpoint        string
	projectName     string
	metricStoreName string
	accessKeyID     string
	accessKeySecret string
	ttl             int
	shardCnt        int
	client          *Client
}

func (m *MetricStoreTestSuite) SetupSuite() {
	m.endpoint = os.Getenv("LOG_TEST_ENDPOINT")
	m.accessKeyID = os.Getenv("LOG_TEST_ACCESS_KEY_ID")
	m.accessKeySecret = os.Getenv("LOG_TEST_ACCESS_KEY_SECRET")
	suffix := time.Now().Unix()
	m.projectName = fmt.Sprintf("test-metric-store-%d", suffix)
	m.metricStoreName = "test"
	m.ttl = 30
	m.shardCnt = 2
	m.client = &Client{
		Endpoint:        m.endpoint,
		AccessKeyID:     m.accessKeyID,
		AccessKeySecret: m.accessKeySecret,
	}
	_, err := m.client.CreateProject(m.projectName, "test metric store")
	m.Require().Nil(err)
	time.Sleep(time.Minute)
}

func (m *MetricStoreTestSuite) TearDownSuite() {
	err := m.client.DeleteProject(m.projectName)
	m.Require().Nil(err)
}

func (m *MetricStoreTestSuite) TestClient_CreateAndDeleteMetricStore() {
	metricStore := LogStore{
		Name:       m.metricStoreName,
		TTL:        m.ttl,
		ShardCount: m.shardCnt,
	}
	ce := m.client.CreateMetricStore(m.projectName, metricStore)
	m.Require().Nil(ce)
	de := m.client.DeleteMetricStore(m.projectName, m.metricStoreName)
	m.Require().Nil(de)
}

func (m *MetricStoreTestSuite) TestClient_UpdateAndGetMetricStore() {
	metricStore1 := LogStore{
		Name:       m.metricStoreName,
		TTL:        m.ttl,
		ShardCount: m.shardCnt,
	}
	ce := m.client.CreateMetricStore(m.projectName, metricStore1)
	m.Require().Nil(ce)
	metricStore, ge := m.client.GetMetricStore(m.projectName, m.metricStoreName)
	m.Require().Nil(ge)
	m.Require().Equal(m.metricStoreName, metricStore.Name)
	m.Require().Equal(m.ttl, metricStore.TTL)
	m.Require().Equal(m.shardCnt, metricStore.ShardCount)
	m.Require().Equal("Metrics", metricStore.TelemetryType)

	metricStore1.TTL = 15
	ue := m.client.UpdateMetricStore(m.projectName, metricStore1)
	m.Require().Nil(ue)
	metricStore2, ge2 := m.client.GetMetricStore(m.projectName, m.metricStoreName)
	m.Require().Nil(ge2)
	m.Require().Equal(m.metricStoreName, metricStore2.Name)
	m.Require().Equal(15, metricStore2.TTL)
	m.Require().Equal("Metrics", metricStore.TelemetryType)
	de := m.client.DeleteMetricStore(m.projectName, m.metricStoreName)
	m.Require().Nil(de)
}
