package sls

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httputil"

	"encoding/json"
	"io/ioutil"

	"github.com/golang/glog"
)

// request sends a request to SLS.
func request(project *LogProject, method, uri string, headers map[string]string,
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
	httpPrefix := "http://"
	httpsPrefix := "https://"
	defaultPrefix := httpsPrefix
	host := project.Endpoint
	if len(project.Endpoint) >= len(httpPrefix) && project.Endpoint[0:len(httpPrefix)] == httpPrefix {
		host = project.Endpoint[len(httpPrefix):]
		defaultPrefix = httpPrefix
	} else if len(project.Endpoint) >= len(httpsPrefix) && project.Endpoint[0:len(httpsPrefix)] == httpsPrefix {
		host = project.Endpoint[len(httpsPrefix):]
		defaultPrefix = httpsPrefix
	}

	urlStr := fmt.Sprintf("%s%v.%v%v", defaultPrefix, project.Name, host, uri)
	req, err := http.NewRequest(method, urlStr, reader)
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
			return nil, nil, err
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
