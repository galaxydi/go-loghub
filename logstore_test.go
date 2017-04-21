package sls

import (
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/glog"
	"github.com/stretchr/testify/suite"
)

func TestLogStore(t *testing.T) {
	suite.Run(t, new(LogstoreTestSuite))
	glog.Flush()
}

type LogstoreTestSuite struct {
	suite.Suite
	endpoint        string
	projectName     string
	logstoreName    string
	accessKeyID     string
	accessKeySecret string
	Project         *LogProject
	Logstore        *LogStore
}

func (s *LogstoreTestSuite) SetupTest() {
	s.endpoint = os.Getenv("LOG_TEST_ENDPOINT")
	s.projectName = os.Getenv("LOG_TEST_PROJECT")
	s.logstoreName = os.Getenv("LOG_TEST_LOGSTORE")
	s.accessKeyID = os.Getenv("LOG_TEST_ACCESS_KEY_ID")
	s.accessKeySecret = os.Getenv("LOG_TEST_ACCESS_KEY_SECRET")
	slsProject, err := NewLogProject(s.projectName, s.endpoint, s.accessKeyID, s.accessKeySecret)
	s.Nil(err)
	s.NotNil(slsProject)
	s.Project = slsProject
	slsLogstore, err := s.Project.GetLogStore(s.logstoreName)
	s.Nil(err)
	s.NotNil(slsLogstore)
	s.Logstore = slsLogstore
}

func (s *LogstoreTestSuite) TestCheckLogstoreExist() {
	exist, err := s.Project.CheckLogstoreExist("not-exist-logstore")
	s.Nil(err)
	s.False(exist)
}

func (s *LogstoreTestSuite) TestCheckMachineGroupExist() {
	exist, err := s.Project.CheckMachineGroupExist("not-exist-group")
	s.Nil(err)
	s.False(exist)
}

func (s *LogstoreTestSuite) TestCheckConfigExist() {
	exist, err := s.Project.CheckConfigExist("not-exist-config")
	s.Nil(err)
	s.False(exist)
}

func (s *LogstoreTestSuite) TestPutLogs() {
	content := &LogContent{
		Key:   proto.String("demo_key"),
		Value: proto.String("demo_value"),
	}
	logRecord := &Log{
		Time:     proto.Uint32(uint32(time.Now().Unix())),
		Contents: []*LogContent{content},
	}
	lg := &LogGroup{
		Topic:  proto.String("test"),
		Source: proto.String("10.168.122.110"),
		Logs:   []*Log{logRecord},
	}
	err := s.Logstore.PutLogs(lg)
	s.Nil(err)
}

func (s *LogstoreTestSuite) TestEmptyLogGroup() {
	lg := &LogGroup{
		Topic:  proto.String("test"),
		Source: proto.String("10.168.122.110"),
		Logs:   []*Log{},
	}
	err := s.Logstore.PutLogs(lg)
	s.Nil(err)
}

func (s *LogstoreTestSuite) TestPullLogs() {
	c := &LogContent{
		Key:   proto.String("error code"),
		Value: proto.String("InternalServerError"),
	}
	l := &Log{
		Time: proto.Uint32(uint32(time.Now().Unix())),
		Contents: []*LogContent{
			c,
		},
	}
	lg := &LogGroup{
		Topic:  proto.String("demo topic"),
		Source: proto.String("10.230.201.117"),
		Logs: []*Log{
			l,
		},
	}

	err := s.Logstore.PutLogs(lg)
	s.Nil(err)

	cursor, err := s.Logstore.GetCursor(0, "begin")
	s.Nil(err)
	endCursor, err := s.Logstore.GetCursor(0, "end")
	s.Nil(err)

	_, _, err = s.Logstore.PullLogs(0, cursor, endCursor, 10)
	s.Nil(err)
}

func (s *LogstoreTestSuite) TestIndex() {
	logstoreName := "test_index_logstore"
	exist, err := s.Project.CheckLogstoreExist(logstoreName)
	s.Nil(err)
	if exist {
		store, err := s.Project.GetLogStore(logstoreName)
		s.Nil(err)
		s.NotNil(store)
		indexExist, indexErr := store.CheckIndexExist()
		s.Nil(indexErr)
		if indexExist {
			s.Nil(store.DeleteIndex())
		}
		s.Nil(s.Project.DeleteLogStore(logstoreName))
		time.Sleep(3 * time.Second)
	}

	err = s.Project.CreateLogStore(logstoreName, 1, 1)
	s.Nil(err)
	time.Sleep(3 * time.Second)

	store, err := s.Project.GetLogStore(logstoreName)
	s.Nil(err)
	s.NotNil(store)

	exist, err = store.CheckIndexExist()
	s.Nil(err)
	s.False(exist)

	// 创建全文索引
	tokenList := []string{`,`, ` `, `'`, `"`, `;`, `=`, `(`, `)`, `[`, `]`, `{`, `}`, `?`, `@`, `&`, `<`, `>`, `/`, `:`, `\n`, `\t`}
	indexLine := IndexLine{
		Token:         tokenList,
		CaseSensitive: false,
	}
	index := Index{
		TTL:  1,
		Line: &indexLine,
	}
	err = store.CreateIndex(index)
	s.Nil(err)

	time.Sleep(3 * time.Second)
	err = store.DeleteIndex()
	s.Nil(err)

	err = s.Project.DeleteLogStore(logstoreName)
	s.Nil(err)
}
