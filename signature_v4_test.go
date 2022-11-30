package sls

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SignerV4Suite struct {
	suite.Suite
	mockAKID  string
	mockAKSec string
	uri       string
	method    string
	region    string
	body      string
	dateTime  string
	urlParams map[string]string
	headers   map[string]string
	signer    Signer
}

func toUriWithQuery(uri string, urlParams map[string]string) string {
	vals := url.Values{}
	for k, v := range urlParams {
		vals.Add(k, v)
	}
	return fmt.Sprintf("%s?%s", uri, vals.Encode())
}

func (s *SignerV4Suite) SetupTest() {
	s.mockAKID = "acsddda21dsd"
	s.mockAKSec = "zxasdasdasw2"
	s.uri = "/logstores"
	s.method = "POST"
	s.region = "cn-hangzhou"
	s.body = "adasd= -asd zcas"
	s.headers = map[string]string{
		"hello":      "world",
		"hello-Text": "a12X- ",
		" Ko ":       "",
		"":           "AA",
		"x-log-test": "het123",
		"x-acs-ppp":  "dds",
	}
	s.urlParams = map[string]string{
		" abc":  "efg",
		" agc ": "",
		"":      "efg",
		"A-bc":  "eFg",
	}
	s.dateTime = "20220808T032330Z"
	// Set dateTime for debugging
	s.headers[HTTPHeaderLogDate] = s.dateTime
	s.signer = &SignerV4{
		accessKeyID:     s.mockAKID,
		accessKeySecret: s.mockAKSec,
		region:          s.region,
	}
}

func (s *SignerV4Suite) TestSignV4Case1() {
	assert.Nil(s.T(), s.signer.Sign(s.method, toUriWithQuery(s.uri, s.urlParams), s.headers, []byte(s.body)))
	auth := s.headers[HTTPHeaderAuthorization]
	exp := "SLS4-HMAC-SHA256 " +
		"Credential=acsddda21dsd/20220808/cn-hangzhou/sls/aliyun_v4_request," +
		"Signature=bcdc405707a79dd61b1407a31613e36cbec25d3bbeecf7101add56aacadbdf1e"
	assert.Equal(s.T(), exp, auth)
}

// Empty urlParams, empty headers, region cn-shanghai
func (s *SignerV4Suite) TestSignV4Case2() {
	s.region = "cn-shanghai"
	s.signer = &SignerV4{
		accessKeyID:     s.mockAKID,
		accessKeySecret: s.mockAKSec,
		region:          s.region,
	}
	s.headers = make(map[string]string)
	s.headers[HTTPHeaderLogDate] = s.dateTime

	assert.Nil(s.T(), s.signer.Sign(s.method, s.uri, s.headers, []byte(s.body)))
	auth := s.headers[HTTPHeaderAuthorization]
	exp := "SLS4-HMAC-SHA256 " +
		"Credential=acsddda21dsd/20220808/cn-shanghai/sls/aliyun_v4_request," +
		"Signature=8a10a5e723cb2e75964816de660b2c16a58af8bc0261f7f0722d832468c76ce8"
	assert.Equal(s.T(), exp, auth)
}

// Empty body
func (s *SignerV4Suite) TestSignV4Case3() {
	s.body = ""
	assert.Nil(s.T(), s.signer.Sign(s.method, toUriWithQuery(s.uri, s.urlParams), s.headers, []byte(s.body)))
	auth := s.headers[HTTPHeaderAuthorization]
	exp := "SLS4-HMAC-SHA256 " +
		"Credential=acsddda21dsd/20220808/cn-hangzhou/sls/aliyun_v4_request," +
		"Signature=b657145686c93047f9c71444e1f2d4bed5ed02f6f24a996ef5067676221de732"
	assert.Equal(s.T(), exp, auth)
}

// Empty body and method get
func (s *SignerV4Suite) TestSignV4Case4() {
	s.body = ""
	s.method = "GET"
	assert.Nil(s.T(), s.signer.Sign(s.method, toUriWithQuery(s.uri, s.urlParams), s.headers, []byte(s.body)))
	auth := s.headers[HTTPHeaderAuthorization]
	exp := "SLS4-HMAC-SHA256 " +
		"Credential=acsddda21dsd/20220808/cn-hangzhou/sls/aliyun_v4_request," +
		"Signature=5fb4e9302126de99c05643f8f7469eb6c35b7851a04c495dd90840a741451f1b"
	assert.Equal(s.T(), exp, auth)
}

// Complex uri and urlParams
func (s *SignerV4Suite) TestSignV4Case5() {
	s.uri = "/logstores/hello/a+*~bb/cc"
	s.urlParams["abs-ij*asd/vc"] = "a~js+d ada"
	s.urlParams["a abAas123/vc"] = "a~jdad a2ADFs+d ada"
	assert.Nil(s.T(), s.signer.Sign(s.method, toUriWithQuery(s.uri, s.urlParams), s.headers, []byte(s.body)))
	auth := s.headers[HTTPHeaderAuthorization]
	exp := "SLS4-HMAC-SHA256 " +
		"Credential=acsddda21dsd/20220808/cn-hangzhou/sls/aliyun_v4_request," +
		"Signature=6e3bae51420ade037431836e0b9791a4b750982376fa7e056585af7dcd10eae1"
	assert.Equal(s.T(), exp, auth)
}

func (s *SignerV4Suite) TestSignV1Case1() {
	headers := map[string]string{
		"x-log-apiversion":      "0.6.0",
		"x-log-signaturemethod": "hmac-sha1",
		"x-log-bodyrawsize":     "0",
		"Date":                  "Mon, 3 Jan 2010 08:33:47 GMT",
	}
	mockAKID := "mockAccessKeyID"
	mockAKSec := "mockAccessKeySecret"
	expSign := "Rwm6cTKzoti4HWoe+GKcb6Kv07E="
	expAuth := fmt.Sprintf("SLS %s:%s", mockAKID, expSign)

	v1 := SignerV1{accessKeyID: mockAKID, accessKeySecret: mockAKSec}
	err := v1.Sign("GET", "/logstores", headers, nil)
	assert.Nil(s.T(), err)
	auth := headers[HTTPHeaderAuthorization]
	assert.Equal(s.T(), expAuth, auth)
}

// Protobuf content
func (s *SignerV4Suite) TestSignV1Case2() {
	body := []byte{10, 50, 10, 30, 8, 248, 178, 147,
		158, 5, 18, 22, 10, 7, 84, 101, 115, 116, 75,
		101, 121, 18, 11, 84, 101, 115, 116, 67, 111,
		110, 116, 101, 110, 116, 26, 0, 34, 14, 49,
		48, 46, 50, 51, 48, 46, 50, 48, 49, 46, 49, 49, 55}
	md5Sum := fmt.Sprintf("%X", md5.Sum(body))
	headers := map[string]string{
		"x-log-apiversion":      "0.6.0",
		"x-log-signaturemethod": "hmac-sha1",
		"x-log-bodyrawsize":     "50",
		"Content-MD5":           md5Sum,
		"Content-Type":          "application/x-protobuf",
		"Content-Length":        "50",
		"Date":                  "Mon, 3 Jan 2010 08:33:47 GMT",
	}
	mockAKID := "mockAccessKeyID"
	mockAKSec := "mockAccessKeySecret"
	expSign := "87xQWqFaOSewqRIma8kPjGYlXHc="
	expAuth := fmt.Sprintf("SLS %s:%s", mockAKID, expSign)
	v1 := SignerV1{
		accessKeyID:     mockAKID,
		accessKeySecret: mockAKSec,
	}
	err := v1.Sign("GET", "/logstores/app_log", headers, body)
	assert.Nil(s.T(), err)
	auth := headers[HTTPHeaderAuthorization]
	assert.Equal(s.T(), expAuth, auth)
}

func TestSignerV4Suite(t *testing.T) {
	suite.Run(t, new(SignerV4Suite))
}
