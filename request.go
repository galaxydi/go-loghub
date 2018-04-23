package sls

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"encoding/json"
	"io/ioutil"

	"github.com/golang/glog"
)

// request sends a request to SLS.
func request_v2(project, endpoint, accessKeyID, accessKeySecret, stsToken, method, uri string, headers map[string]string, body []byte) (*http.Response, error) {
	// The caller should provide 'x-log-bodyrawsize' header
	if _, ok := headers["x-log-bodyrawsize"]; !ok {
		return nil, fmt.Errorf("Can't find 'x-log-bodyrawsize' header")
	}

	var usingHttps bool
	if strings.HasPrefix(endpoint, "https://") {
		endpoint = endpoint[8:]
		usingHttps = true
	} else if strings.HasPrefix(endpoint, "http://") {
		endpoint = endpoint[7:]
	}

	// SLS public request headers
	var hostStr string
	if len(project) == 0 {
		hostStr = project
	} else {
		hostStr = project + "." + endpoint
	}
	headers["Host"] = hostStr
	headers["Date"] = nowRFC1123()
	headers["x-log-apiversion"] = version
	headers["x-log-signaturemethod"] = signatureMethod

	// Access with token
	if stsToken != "" {
		headers["x-acs-security-token"] = stsToken
	}

	if body != nil {
		bodyMD5 := fmt.Sprintf("%X", md5.Sum(body))
		headers["Content-MD5"] = bodyMD5
		if _, ok := headers["Content-Type"]; !ok {
			return nil, fmt.Errorf("Can't find 'Content-Type' header")
		}
	}

	// Calc Authorization
	// Authorization = "SLS <AccessKeyId>:<Signature>"
	digest, err := signature(accessKeySecret, method, uri, headers)
	if err != nil {
		return nil, err
	}
	auth := fmt.Sprintf("SLS %v:%v", accessKeyID, digest)
	headers["Authorization"] = auth

	// Initialize http request
	reader := bytes.NewReader(body)
	var urlStr string
	// using http as default
	if !GlobalForceUsingHTTP && usingHttps {
		urlStr = "https://"
	} else {
		urlStr = "http://"
	}
	urlStr += hostStr + uri
	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Parse the sls error from body.
	if resp.StatusCode != http.StatusOK {
		err := &Error{}
		err.HTTPCode = (int32)(resp.StatusCode)
		defer resp.Body.Close()
		buf, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(buf, err)
		err.RequestID = resp.Header.Get("x-log-requestid")
		return nil, err
	}

	if glog.V(4) {
		dump, e := httputil.DumpResponse(resp, true)
		if e != nil {
			glog.Info(e)
		}
		glog.Infof("HTTP Response:\n%v", string(dump))
	}
	return resp, nil
}

// request sends a request to SLS.
func request(project *LogProject, method, uri string, headers map[string]string,
	body []byte) (*http.Response, error) {

	// The caller should provide 'x-log-bodyrawsize' header
	if _, ok := headers["x-log-bodyrawsize"]; !ok {
		return nil, fmt.Errorf("Can't find 'x-log-bodyrawsize' header")
	}

	// SLS public request headers
	var hostStr string
	if len(project.Name) == 0 {
		hostStr = project.Endpoint
	} else {
		hostStr = project.Name + "." + project.Endpoint
	}
	headers["Host"] = hostStr
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
			return nil, fmt.Errorf("Can't find 'Content-Type' header")
		}
	}

	// Calc Authorization
	// Authorization = "SLS <AccessKeyId>:<Signature>"
	digest, err := signature(project.AccessKeySecret, method, uri, headers)
	if err != nil {
		return nil, err
	}
	auth := fmt.Sprintf("SLS %v:%v", project.AccessKeyID, digest)
	headers["Authorization"] = auth

	// Initialize http request
	reader := bytes.NewReader(body)
	var urlStr string
	if GlobalForceUsingHTTP || project.UsingHTTP {
		urlStr = "http://"
	} else {
		urlStr = "https://"
	}
	urlStr += hostStr + uri
	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Parse the sls error from body.
	if resp.StatusCode != http.StatusOK {
		err := &Error{}
		err.HTTPCode = (int32)(resp.StatusCode)
		defer resp.Body.Close()
		buf, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(buf, err)
		err.RequestID = resp.Header.Get("x-log-requestid")
		return nil, err
	}

	if glog.V(4) {
		dump, e := httputil.DumpResponse(resp, true)
		if e != nil {
			glog.Info(e)
		}
		glog.Infof("HTTP Response:\n%v", string(dump))
	}
	return resp, nil
}
