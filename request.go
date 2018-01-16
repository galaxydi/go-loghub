package sls

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

// timeout configs
var (
	requestTimeout = 10 * time.Second
	retryTimeout   = 30 * time.Second
)

func retryReadErrorCheck(ctx context.Context, err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	switch e := err.(type) {
	case *url.Error:
		return true, e
	case *Error:
		if e.HttpStatus >= 500 && e.HttpStatus <= 599 {
			return true, e
		}
	case *BadResponseError:
		if e.HttpStatus >= 500 && e.HttpStatus <= 599 {
			return true, e
		}
	default:
		return false, e
	}

	return false, err
}

func retryWriteErrorCheck(ctx context.Context, err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	switch e := err.(type) {
	case *Error:
		if e.HttpStatus == 502 || e.HttpStatus == 503 {
			return true, e
		}
	case *BadResponseError:
		if e.HttpStatus == 502 || e.HttpStatus == 503 {
			return true, e
		}
	default:
		return false, e
	}

	return false, err
}

// request sends a request to SLS.
// mock param only for test, default is []
func request(project *LogProject, method, uri string, headers map[string]string,
	body []byte, mock ...interface{}) (http.Header, []byte, error) {

	var respHeader http.Header
	var respBody []byte
	var slsErr error
	var err error
	var mockErr *mockErrorRetry

	cctx, cancel := context.WithTimeout(context.Background(), retryTimeout)
	defer cancel()

	// all GET method is read function
	if method == "GET" {
		err = RetryWithCondition(cctx, backoff.NewExponentialBackOff(), func() (bool, error) {
			if len(mock) == 0 {
				respHeader, respBody, slsErr = realRequest(project, method, uri, headers, body)
			} else {
				respHeader, respBody, mockErr = nil, nil, mock[0].(*mockErrorRetry)
				mockErr.RetryCnt--
				if mockErr.RetryCnt <= 0 {
					slsErr = nil
					return false, nil
				}
				slsErr = &mockErr.Err
			}
			return retryReadErrorCheck(cctx, slsErr)
		})

	} else {
		err = RetryWithCondition(cctx, backoff.NewExponentialBackOff(), func() (bool, error) {
			if len(mock) == 0 {
				respHeader, respBody, slsErr = realRequest(project, method, uri, headers, body)
			} else {
				respHeader, respBody, mockErr = nil, nil, mock[0].(*mockErrorRetry)
				mockErr.RetryCnt--
				if mockErr.RetryCnt <= 0 {
					slsErr = nil
					return false, nil
				}
				slsErr = &mockErr.Err
			}
			return retryWriteErrorCheck(cctx, slsErr)
		})
	}

	if err != nil {
		return respHeader, respBody, err
	} else {
		return respHeader, respBody, slsErr
	}
}

// request sends a request to SLS.
func realRequest(project *LogProject, method, uri string, headers map[string]string,
	body []byte) (http.Header, []byte, error) {

	// The caller should provide 'x-log-bodyrawsize' header
	if _, ok := headers["x-log-bodyrawsize"]; !ok {
		return nil, nil, NewClientError("Can't find 'x-log-bodyrawsize' header")
	}

	// SLS public request headers
	headers["Host"] = project.Name + "." + project.Endpoint
	headers["Date"] = nowRFC1123()
	headers["x-log-apiversion"] = version
	headers["x-log-signaturemethod"] = signatureMethod

	// Access with token
	if project.SecurityToken != "" {
		headers["x-acs-security-token"] = project.SecurityToken
	}

	if body != nil {
		bodyMD5 := fmt.Sprintf("%X", md5.Sum(body))
		headers["Content-MD5"] = bodyMD5
		if _, ok := headers["Content-Type"]; !ok {
			return nil, nil, NewClientError("Can't find 'Content-Type' header")
		}
	}

	// Calc Authorization
	// Authorization = "SLS <AccessKeyId>:<Signature>"
	digest, err := signature(project, method, uri, headers)
	if err != nil {
		return nil, nil, NewClientError(err.Error())
	}
	auth := fmt.Sprintf("SLS %v:%v", project.AccessKeyID, digest)
	headers["Authorization"] = auth

	// Initialize http request
	reader := bytes.NewReader(body)

	// Handle the endpoint
	urlStr := fmt.Sprintf("%s%s", project.baseURL, uri)
	req, err := http.NewRequest(method, urlStr, reader)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	req = req.WithContext(ctx)
	if err != nil {
		return nil, nil, NewClientError(err.Error())
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if glog.V(1) {
		dump, e := httputil.DumpRequest(req, true)
		if e != nil {
			glog.Info(e)
		}
		glog.Infof("HTTP Request:\n%v", string(dump))
	}

	// Get ready to do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		serverErr := new(Error)
		err := json.Unmarshal(buf, serverErr)
		if err != nil {
			badRespError := new(BadResponseError)
			badRespError.RespHeader = resp.Header
			badRespError.HttpStatus = resp.StatusCode
			badRespError.RespBody = string(buf)
			return nil, nil, badRespError
		}
		serverErr.RequestID = resp.Header.Get(RequestIDHeader)
		serverErr.HttpStatus = resp.StatusCode
		return nil, nil, serverErr
	}

	if glog.V(1) {
		dump, e := httputil.DumpResponse(resp, true)
		if e != nil {
			glog.Info(e)
		}
		glog.Infof("HTTP Response:\n%v", string(dump))
	}
	return resp.Header, buf, nil
}
