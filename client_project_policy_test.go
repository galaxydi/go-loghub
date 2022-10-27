package sls

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestProjectPolicy(t *testing.T) {
	suite.Run(t, new(ProjectPolicyTestSuite))
}

type ProjectPolicyTestSuite struct {
	suite.Suite
	endpoint        string
	projectName     string
	accessKeyID     string
	accessKeySecret string
	client          Client
	policy          string
	newPolicy       string
}

func (s *ProjectPolicyTestSuite) SetupSuite() {
	s.endpoint = os.Getenv("LOG_TEST_ENDPOINT")
	s.projectName = os.Getenv("LOG_TEST_PROJECT")
	s.accessKeyID = os.Getenv("LOG_TEST_ACCESS_KEY_ID")
	s.accessKeySecret = os.Getenv("LOG_TEST_ACCESS_KEY_SECRET")
	s.client.AccessKeyID = s.accessKeyID
	s.client.AccessKeySecret = s.accessKeySecret
	s.client.Endpoint = s.endpoint
	s.policy = `
{
  "Statement": [
    {
      "Action": [
        "log:Post*"
      ],
      "Effect": "Deny",
      "Resource": "acs:log:*:*:project/test-project-policy/*"
    }
  ],
  "Version": "1"
}`
	s.newPolicy = `
{
  "Statement": [
    {
      "Action": [
        "log:Post*"
      ],
      "Effect": "Allow",
      "Resource": "acs:log:*:*:project/test-project-policy/*"
    }
  ],
  "Version": "1"
}`
}

func (s *ProjectPolicyTestSuite) TearDownSuite() {
}

func (s *ProjectPolicyTestSuite) TestClient_CURDProjectPolicy() {
	err := s.client.UpdateProjectPolicy(s.projectName, s.policy)
	s.Require().Nil(err)
	policy, err := s.client.GetProjectPolicy(s.projectName)
	s.Require().Nil(err)
	s.Require().Equal(policy, s.policy)
	err = s.client.UpdateProjectPolicy(s.projectName, s.newPolicy)
	s.Require().Nil(err)
	newPolicy, err := s.client.GetProjectPolicy(s.projectName)
	s.Require().Nil(err)
	s.Require().Equal(newPolicy, s.newPolicy)
	err = s.client.DeleteProjectPolicy(s.projectName)
	s.Require().Nil(err)
	policy, err = s.client.GetProjectPolicy(s.projectName)
	s.Require().Nil(err)
	s.Require().Empty(policy)
}
