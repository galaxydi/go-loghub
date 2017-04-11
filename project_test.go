package sls

import (
	"os"
	"testing"

	"github.com/golang/glog"
	"github.com/stretchr/testify/suite"
)

func TestProject(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
	glog.Flush()
}

type ProjectTestSuite struct {
	suite.Suite
	endpoint        string
	accessKeyID     string
	accessKeySecret string
	client          Client
}

func (s *ProjectTestSuite) SetupTest() {
	s.endpoint = os.Getenv("LOG_TEST_ENDPOINT")
	s.accessKeyID = os.Getenv("LOG_TEST_ACCESS_KEY_ID")
	s.accessKeySecret = os.Getenv("LOG_TEST_ACCESS_KEY_SECRET")
	s.client = Client{
		Endpoint:        s.endpoint,
		AccessKeyID:     s.accessKeyID,
		AccessKeySecret: s.accessKeySecret,
		SecurityToken:   "",
	}
}

func (s *ProjectTestSuite) checkProjectExist(projectName string) (bool, error) {
	_, err := s.client.GetProject(projectName)
	if err != nil {
		switch err.(type) {
		case *Error:
			slsErr := err.(*Error)
			switch slsErr.Code {
			case "ProjectNotExist":
				return false, nil
			default:
				return false, slsErr
			}
		default:
			return false, err
		}
	}
	return true, nil
}

func (s *ProjectTestSuite) TestGetProject() {
	projectName := "not-exist-project"
	exist, err := s.checkProjectExist(projectName)
	s.Nil(err)
	s.False(exist)
}
