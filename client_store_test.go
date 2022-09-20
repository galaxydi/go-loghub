package sls

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestLostore(t *testing.T) {
	suite.Run(t, new(LostoreTestSuite))
}

type LostoreTestSuite struct {
	suite.Suite
	endpoint        string
	projectName     string
	logstoreName    string
	accessKeyID     string
	accessKeySecret string
	client          Client
	recordId        string
	tagName         string
	resourceName    string
}

func (s *LostoreTestSuite) SetupSuite() {
	s.endpoint = os.Getenv("LOG_TEST_ENDPOINT")
	s.projectName = os.Getenv("LOG_TEST_PROJECT")
	s.logstoreName = os.Getenv("LOG_TEST_LOGSTORE")
	s.accessKeyID = os.Getenv("LOG_TEST_ACCESS_KEY_ID")
	s.accessKeySecret = os.Getenv("LOG_TEST_ACCESS_KEY_SECRET")
	s.client.AccessKeyID = s.accessKeyID
	s.client.AccessKeySecret = s.accessKeySecret
	s.client.Endpoint = s.endpoint

	require.NotEmpty(s.T(), s.endpoint)
	require.NotEmpty(s.T(), s.projectName)
	require.NotEmpty(s.T(), s.logstoreName)
	require.NotEmpty(s.T(), s.accessKeyID)
	require.NotEmpty(s.T(), s.accessKeySecret)
	_, err := s.client.CreateProject(s.projectName, "ProjectAlreadyExist")
	require.True(s.T(), err == nil || strings.Contains(err.Error(), ""))
	time.Sleep(time.Second * 10)
	err = s.client.CreateLogStore(s.projectName, s.logstoreName, 12, 1, false, 64)
	require.NoError(s.T(), err)
	time.Sleep(time.Second * 10)
	err = s.client.CreateIndex(s.projectName, s.logstoreName, Index{
		Keys: map[string]IndexKey{
			"col_0": {
				Token:         []string{" "},
				DocValue:      true,
				CaseSensitive: false,
				Type:          "long",
			},
			"col_1": {
				Token:         []string{",", ":", " "},
				DocValue:      true,
				CaseSensitive: false,
				Type:          "text",
			},
		},
	})
	require.NoError(s.T(), err)
}

func (s *LostoreTestSuite) TearDownSuite() {
	err := s.client.DeleteIndex(s.projectName, s.logstoreName)
	assert.NoError(s.T(), err)
	err = s.client.DeleteLogStore(s.projectName, s.logstoreName)
	assert.NoError(s.T(), err)
	err = s.client.DeleteProject(s.projectName)
	assert.NoError(s.T(), err)
}

func (s *LostoreTestSuite) TestSplitShardDefault() {

	find := func() (num int, id int) {
		shards, err := s.client.ListShards(s.projectName, s.logstoreName)
		assert.NoError(s.T(), err)
		id = -1
		for _, shard := range shards {
			if shard.Status == "readwrite" {
				num++
				id = shard.ShardID
			}
		}
		return
	}

	num, id := find()
	assert.True(s.T(), id != -1)
	assert.Equal(s.T(), num, 1)

	_, err := s.client.SplitShard(s.projectName, s.logstoreName, id, "ef000000000000000000000000000000")
	assert.NoError(s.T(), err)
	time.Sleep(20 * time.Second)
	num2, id2 := find()
	assert.True(s.T(), id2 != -1)
	assert.Equal(s.T(), num2, 2)

	_, err = s.client.SplitNumShard(s.projectName, s.logstoreName, id2, 3)
	assert.NoError(s.T(), err)
	time.Sleep(20 * time.Second)
	num3, id3 := find()
	assert.True(s.T(), id3 != -1)
	assert.Equal(s.T(), 4, num3)
}
