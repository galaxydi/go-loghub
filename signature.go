package sls

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Signer interface {
	// Sign modifies @param headers only, adds signature and other http headers
	// that log services authorization requires.
	Sign(method, uriWithQuery string, headers map[string]string, body []byte) error
}

// GMT location
var gmtLoc = time.FixedZone("GMT", 0)

// NowRFC1123 returns now time in RFC1123 format with GMT timezone,
// eg, "Mon, 02 Jan 2006 15:04:05 GMT".
func nowRFC1123() string {
	return time.Now().In(gmtLoc).Format(time.RFC1123)
}

// SignerV1 version v1
type SignerV1 struct {
	accessKeyID     string
	accessKeySecret string
}

func NewSignerV1(accessKeyID, accessKeySecret string) *SignerV1 {
	return &SignerV1{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
	}
}

func (s *SignerV1) Sign(method, uri string, headers map[string]string, body []byte) error {
	var contentMD5, contentType, date, canoHeaders, canoResource, digest string
	if len(body) > 0 {
		contentMD5 = fmt.Sprintf("%X", md5.Sum(body))
		headers["Content-MD5"] = contentMD5
	}

	if val, ok := headers["Content-Type"]; ok {
		contentType = val
	}

	date, ok := headers["Date"]
	if !ok {
		return fmt.Errorf("Can't find 'Date' header")
	}
	headers["x-log-signaturemethod"] = signatureMethod
	var slsHeaderKeys sort.StringSlice

	// Calc CanonicalizedSLSHeaders
	slsHeaders := make(map[string]string, len(headers))
	for k, v := range headers {
		l := strings.TrimSpace(strings.ToLower(k))
		if strings.HasPrefix(l, "x-log-") || strings.HasPrefix(l, "x-acs-") {
			slsHeaders[l] = strings.TrimSpace(v)
			slsHeaderKeys = append(slsHeaderKeys, l)
		}
	}

	sort.Sort(slsHeaderKeys)
	for i, k := range slsHeaderKeys {
		canoHeaders += k + ":" + slsHeaders[k]
		if i+1 < len(slsHeaderKeys) {
			canoHeaders += "\n"
		}
	}

	// Calc CanonicalizedResource
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	canoResource += u.EscapedPath()
	if u.RawQuery != "" {
		var keys sort.StringSlice

		vals := u.Query()
		for k := range vals {
			keys = append(keys, k)
		}

		sort.Sort(keys)
		canoResource += "?"
		for i, k := range keys {
			if i > 0 {
				canoResource += "&"
			}

			for _, v := range vals[k] {
				canoResource += k + "=" + v
			}
		}
	}

	signStr := method + "\n" +
		contentMD5 + "\n" +
		contentType + "\n" +
		date + "\n" +
		canoHeaders + "\n" +
		canoResource

	// Signature = base64(hmac-sha1(UTF8-Encoding-Of(SignString)，AccessKeySecret))
	mac := hmac.New(sha1.New, []byte(s.accessKeySecret))
	_, err = mac.Write([]byte(signStr))
	if err != nil {
		return err
	}
	digest = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	auth := fmt.Sprintf("SLS %v:%v", s.accessKeyID, digest)
	headers["Authorization"] = auth
	return nil
}
