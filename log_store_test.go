package sls

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Netflix/go-env"
	"github.com/stretchr/testify/suite"
)

func TestGetLogsTestSuite(t *testing.T) {
	suite.Run(t, new(GetLogsTestSuite))
}

type GetLogsTestSuite struct {
	suite.Suite
	env    TestEnvInfo
	client *Client
}

func (s *GetLogsTestSuite) SetupSuite() {
	_, err := env.UnmarshalFromEnviron(&s.env)
	s.Require().NoError(err)
	s.client = &Client{
		Endpoint:        s.env.Endpoint,
		AccessKeyID:     s.env.AccessKeyID,
		AccessKeySecret: s.env.AccessKeySecret,
		UserAgent:       DefaultLogUserAgent,
	}
}

func (s *GetLogsTestSuite) TearDownSuite() {

}

func (s *GetLogsTestSuite) TestGetLogsV2() {
	exists, err := s.client.CheckProjectExist(s.env.ProjectName)
	s.Require().NoError(err)
	s.True(exists)
	t := uint32(time.Now().Unix())
	timens := uint32(time.Now().UnixNano() % 1e9)
	key, val := "key1", "val1"
	lg := &LogGroup{
		Logs: []*Log{
			{
				Time:   &t,
				TimeNs: &timens,
				Contents: []*LogContent{
					{
						Key:   &key,
						Value: &val,
					},
					{
						Key:   &key,
						Value: &val,
					},
				},
			},
		},
	}
	// write log
	err = s.client.PostLogStoreLogs(s.env.ProjectName, s.env.LogstoreName, lg, nil)
	s.Require().NoError(err)
	// get logs
	time.Sleep(time.Second * 20)
	req := &GetLogRequest{
		From:  int64(t - 900),
		To:    int64(t + 10),
		Lines: 100,
	}
	// old
	ls := convertLogstore(s.client, s.env.ProjectName, s.env.LogstoreName)
	resp, err := ls.getLogsV2(req)
	s.Require().NoError(err)

	// new
	resp2, err := s.client.GetLogsV2(s.env.ProjectName, s.env.LogstoreName, req)
	s.Require().NoError(err)

	// these headers need not to be compared
	filtered := map[string]bool{
		HTTPHeaderDate:          true,
		ElapsedMillisecond:      true,
		RequestIDHeader:         true,
		HTTPHeaderLogDate:       true,
		HTTPHeaderContentType:   true,
		HTTPHeaderContentLength: true,
		GetLogsQueryInfo:        true, // not support yet
	}
	addFilter := func(key string) {
		filtered[strings.ToLower(key)] = true
	}
	addFilter(HTTPHeaderDate)
	addFilter(ElapsedMillisecond)
	addFilter(RequestIDHeader)
	addFilter(HTTPHeaderLogDate)
	addFilter(HTTPHeaderContentType)
	addFilter(HTTPHeaderContentLength)
	addFilter(GetLogsQueryInfo)

	// compare headers
	for k := range resp.Header {
		key := strings.ToLower(k)
		if _, ok := filtered[key]; !ok {
			s.EqualValuesf(resp.Header.Get(k), resp2.Header.Get(k), "header key %s", k)
		}
	}
	// compare values
	s.EqualValues(resp.Progress, resp2.Progress)
	s.EqualValues(resp.Count, resp2.Count)
	s.EqualValues(resp.HasSQL, resp2.HasSQL)

	fmt.Printf("%#v\n", resp)
	fmt.Printf("%#v\n", resp2)
}
