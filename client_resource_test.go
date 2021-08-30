package sls

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestResource(t *testing.T) {
	suite.Run(t, new(ResourceTestSuite))
}

type ResourceTestSuite struct {
	suite.Suite
	endpoint        string
	projectName     string
	logstoreName    string
	accessKeyID     string
	accessKeySecret string
	client          Client
	resourceName    string
}

func (s *ResourceTestSuite) SetupSuite() {
	s.endpoint = os.Getenv("LOG_TEST_ENDPOINT")
	s.projectName = os.Getenv("LOG_TEST_PROJECT")
	s.logstoreName = os.Getenv("LOG_TEST_LOGSTORE")
	s.accessKeyID = os.Getenv("LOG_TEST_ACCESS_KEY_ID")
	s.accessKeySecret = os.Getenv("LOG_TEST_ACCESS_KEY_SECRET")
	s.client.AccessKeyID = s.accessKeyID
	s.client.AccessKeySecret = s.accessKeySecret
	s.client.Endpoint = s.endpoint
	s.resourceName = "user.test_resource_1"
}

func (s *ResourceTestSuite) TearDownSuite() {
}

func (s *ResourceTestSuite) createResource() error {
	rs := &ResourceSchema{
		Schema: []*ResourceSchemaItem{
			&ResourceSchemaItem{
				Column:   "col1",
				Desc:     "col1 desc",
				ExtInfo:  map[string]string{},
				Required: true,
				Type:     "string",
			},
			&ResourceSchemaItem{
				Column:   "col2",
				Desc:     "col2 desc",
				ExtInfo:  "optional",
				Required: true,
				Type:     "string",
			},
		},
	}
	customResource := new(Resource)
	customResource.Type = ResourceTypeUserDefine
	customResource.Name = s.resourceName
	customResource.Schema = rs.ToString()
	customResource.Description = "user test resource 1 descc"
	return s.client.CreateResource(customResource)
}

func (s *ResourceTestSuite) TestClient_CreateResource() {
	err := s.createResource()
	s.Require().Nil(err)
	err = s.client.DeleteResource(s.resourceName)
	s.Require().Nil(err)
}

func (s *ResourceTestSuite) TestClient_UpdateResource() {
	err := s.createResource()
	s.Require().Nil(err)
	resource, err := s.client.GetResource(s.resourceName)
	s.Require().Nil(err)
	rs := new(ResourceSchema)
	err = rs.FromJsonString(resource.Schema)
	s.Require().Nil(err)
	rs.Schema[0].Desc = "new desc"
	resource.Schema = rs.ToString()
	err = s.client.UpdateResource(resource)
	s.Require().Nil(err)
	resource, err = s.client.GetResource(s.resourceName)
	s.Require().Nil(err)
	nrs := new(ResourceSchema)
	err = nrs.FromJsonString(resource.Schema)
	s.Require().Nil(err)
	s.Require().Equal("new desc", rs.Schema[0].Desc, "update resource failed")
	err = s.client.DeleteResource(s.resourceName)
	s.Require().Nil(err)
}

func (s *ResourceTestSuite) TestClient_DeleteResource() {
	err := s.createResource()
	s.Require().Nil(err)
	_, err = s.client.GetResource(s.resourceName)
	s.Require().Nil(err)
	err = s.client.DeleteResource(s.resourceName)
	s.Require().Nil(err)
	_, err = s.client.GetResource(s.resourceName)
	s.Require().NotNil(err)
}

func (s *ResourceTestSuite) TestClient_GetResource() {
	err := s.createResource()
	s.Require().Nil(err)
	getResource, err := s.client.GetResource(s.resourceName)
	s.Require().Nil(err)
	s.Require().Equal(getResource.Name, s.resourceName)
	rs := new(ResourceSchema)
	err = rs.FromJsonString(getResource.Schema)
	s.Require().Nil(err)

	s.Require().Equal(len(rs.Schema), 2)
	s.Require().Equal(rs.Schema[0].Desc, "col1 desc")

	err = s.client.DeleteResource(s.resourceName)
	s.Require().Nil(err)
}

func (s *ResourceTestSuite) TestClient_ListResource() {
	err := s.createResource()
	s.Require().Nil(err)
	resources, total, count, err := s.client.ListResource(ResourceTypeUserDefine, s.resourceName, 0, 100)
	s.Require().Nil(err)
	if total != 1 || count != 1 {
		s.Require().Fail("list resource failed")
	}
	s.Require().Equal(1, len(resources), "there should be only one resource")
	resource := resources[0]
	s.Require().Equal(s.resourceName, resource.Name, "list resource failed")
	err = s.client.DeleteResource(s.resourceName)
	s.Require().Nil(err)
}
