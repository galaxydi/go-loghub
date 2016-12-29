package sls

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/golang/glog"
)

// request sends a request to SLS.
func request(project *LogProject, method, uri string, headers map[string]string,
	body []byte) (resp *http.Response, err error) {

	// The caller should provide 'x-sls-bodyrawsize' header
	if _, ok := headers["x-sls-bodyrawsize"]; !ok {
		err = fmt.Errorf("Can't find 'x-sls-bodyrawsize' header")
		return
	}

	// SLS public request headers
	headers["Host"] = project.Name + "." + project.Endpoint
	headers["Date"] = nowRFC1123()
	headers["x-sls-apiversion"] = version
	headers["x-sls-signaturemethod"] = signatureMethod

	// Access with token
	if project.SessionToken != "" {
		headers["x-acs-security-token"] = project.SessionToken
	}

	if body != nil {
		bodyMD5 := fmt.Sprintf("%X", md5.Sum(body))
		headers["Content-MD5"] = bodyMD5
		if _, ok := headers["Content-Type"]; !ok {
			err = fmt.Errorf("Can't find 'Content-Type' header")
			return
		}
	}

	// Calc Authorization
	// Authorization = "SLS <AccessKeyId>:<Signature>"
	digest, err := signature(project, method, uri, headers)
	if err != nil {
		return
	}
	auth := fmt.Sprintf("SLS %v:%v", project.AccessKeyID, digest)
	headers["Authorization"] = auth

	// Initialize http request
	reader := bytes.NewReader(body)
	urlStr := fmt.Sprintf("http://%v.%v%v", project.Name, project.Endpoint, uri)
	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return
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
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if glog.V(1) {
		dump, e := httputil.DumpResponse(resp, true)
		if e != nil {
			glog.Info(e)
		}
		glog.Infof("HTTP Response:\n%v", string(dump))
	}
	return
}
