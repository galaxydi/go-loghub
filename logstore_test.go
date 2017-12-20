package sls

import (
	"fmt"
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
	slsLogstore, err := NewLogStore(s.logstoreName, s.Project)
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
	lg := generateLG()
	err := s.Logstore.PutLogs(lg)
	s.Nil(err)
}

func (s *LogstoreTestSuite) TestProjectNotExist() {
	projectName := "no-exist-project"
	slsProject, err := NewLogProject(projectName, s.endpoint, s.accessKeyID, s.accessKeySecret)
	s.Nil(err)
	slsLogstore, err := NewLogStore(s.logstoreName, slsProject)
	s.Nil(err)

	lg := generateLG()
	err = slsLogstore.PutLogs(lg)
	s.Require().NotNil(err)
	e, ok := err.(*Error)
	s.Require().True(ok)
	s.Require().Equal(e.Code, "ProjectNotExist")
	s.Require().Equal(e.HttpStatus, 404)
	s.Require().Equal(e.Message, fmt.Sprintf("The Project does not exist : %s", projectName))
}

func (s *LogstoreTestSuite) TestLogStoreNotExist() {
	logstoreName := "no-exist-logstore"
	slsLogstore, err := NewLogStore(logstoreName, s.Project)
	s.Nil(err)

	lg := generateLG()
	err = slsLogstore.PutLogs(lg)
	s.Require().NotNil(err)
	e, ok := err.(*Error)
	s.Require().True(ok)
	s.Require().Equal(e.Code, "LogStoreNotExist")
	s.Require().Equal(e.HttpStatus, 404)
	s.Require().Equal(e.Message, fmt.Sprintf("logstore %s not exist", logstoreName))
}

func (s *LogstoreTestSuite) TestAccessIDNotExist() {
	accessID := "no-exist-key"
	slsProject, err := NewLogProject(s.projectName, s.endpoint, accessID, s.accessKeySecret)
	s.Nil(err)
	slsLogstore, err := NewLogStore(s.logstoreName, slsProject)
	s.Nil(err)

	lg := generateLG()
	err = slsLogstore.PutLogs(lg)
	s.Require().NotNil(err)
	e, ok := err.(*Error)
	s.Require().True(ok)
	s.Require().Equal(e.Code, "Unauthorized")
	s.Require().Equal(e.HttpStatus, 401)
	s.Require().Equal(e.Message, fmt.Sprintf("AccessKeyId not found: %s", accessID))
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

	shards, err := s.Logstore.ListShards()
	s.True(len(shards) > 0)

	err = s.Logstore.PutLogs(lg)
	s.Nil(err)

	cursor, err := s.Logstore.GetCursor(0, "begin")
	s.Nil(err)
	endCursor, err := s.Logstore.GetCursor(0, "end")
	s.Nil(err)

	_, _, err = s.Logstore.PullLogs(0, cursor, "", 10)
	s.Nil(err)

	_, _, err = s.Logstore.PullLogs(0, cursor, endCursor, 10)
	s.Nil(err)
}

func (s *LogstoreTestSuite) TestGetLogs() {
	idx, err := s.Logstore.GetIndex()
	if err != nil {
		fmt.Printf("GetIndex fail, err: %v, idx: %v\n", err, idx)
		return
	}
	fmt.Printf("GetIndex success, idx: %v\n", idx)
	idxConf := Index{
		TTL:  7,
		Keys: map[string]IndexKey{},
		Line: &IndexLine{
			Token:         []string{",", ":", " "},
			CaseSensitive: false,
			IncludeKeys:   []string{},
			ExcludeKeys:   []string{},
		},
	}
	s.Logstore.CreateIndex(idxConf)
	time.Sleep(1 * 1000 * time.Millisecond)
	begin_time := uint32(time.Now().Unix())
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

	putErr := s.Logstore.PutLogs(lg)
	s.Nil(putErr)

	time.Sleep(5 * 1000 * time.Millisecond)

	hResp, hErr := s.Logstore.GetHistograms("", int64(begin_time), int64(begin_time+2), "InternalServerError")
	s.Nil(hErr)
	s.Equal(hResp.Count, int64(1))
	lResp, lErr := s.Logstore.GetLogs("", int64(begin_time), int64(begin_time+2), "InternalServerError", 100, 0, false)
	s.Nil(lErr)
	s.Equal(lResp.Count, int64(1))
}

func (s *LogstoreTestSuite) TestLogstore() {
	logstoreName := "github-test"
	err := s.Project.DeleteLogStore(logstoreName)
	time.Sleep(5 * 1000 * time.Millisecond)
	err = s.Project.CreateLogStore(logstoreName, 14, 2)
	s.Nil(err)
	time.Sleep(10 * 1000 * time.Millisecond)
	err = s.Project.UpdateLogStore(logstoreName, 7, 2)
	s.Nil(err)
	time.Sleep(1 * 1000 * time.Millisecond)
	logstores, err := s.Project.ListLogStore()
	s.Nil(err)
	s.True(len(logstores) >= 1)
	configs, configCount, err := s.Project.ListConfig(0, 100)
	s.Nil(err)
	s.True(len(configs) >= 0)
	s.Equal(len(configs), configCount)
	machineGroups, machineGroupCount, err := s.Project.ListMachineGroup(0, 100)
	s.Nil(err)
	s.True(len(machineGroups) >= 0)
	s.Equal(len(machineGroups), machineGroupCount)
	err = s.Project.DeleteLogStore(logstoreName)
	s.Nil(err)
}

func generateLG() *LogGroup {
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
	return lg
}
