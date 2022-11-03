package sls

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	emptyStringSha256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	HttpHeaderLogDate = "x-log-date"
)

var (
	ErrSignerV4MissingRegion = errors.New("sign version v4 require a valid region")
)

var defaultSignedHeaders = map[string]bool{
	"host":         true,
	"content-type": true,
}

// SignerV4 sign version v4, a non-empty region is required
type SignerV4 struct {
	accessKeyID     string
	accessKeySecret string
	region          string
}

func NewSignerV4(accessKeyID, accessKeySecret, region string) *SignerV4 {
	return &SignerV4{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		region:          region,
	}
}

func (s *SignerV4) Sign(method, uri string, headers map[string]string, body []byte) error {
	if s.region == "" {
		return ErrSignerV4MissingRegion
	}

	uri, urlParams, err := s.parseUri(uri)
	if err != nil {
		return err
	}

	dateTime, ok := headers[HttpHeaderLogDate]
	if !ok {
		return fmt.Errorf("can't find '%s' header", HttpHeaderLogDate)
	}
	date := dateTime[:8]

	// If content-type value is empty string, server will ignore it.
	// So we add a default value here.
	if contentType, ok := headers["Content-Type"]; ok && len(contentType) == 0 {
		headers["Content-Type"] = "application/json"
	}

	// Host should not contain schema here.
	if host, ok := headers["Host"]; ok {
		if strings.HasPrefix(host, "http://") {
			headers["Host"] = host[len("http://"):]
		} else if strings.HasPrefix(host, "https://") {
			headers["Host"] = host[len("https://"):]
		}
	}

	contentLength := len(body)
	var sha256Payload string
	if contentLength != 0 {
		sha256Payload = fmt.Sprintf("%x", sha256.Sum256(body))
	} else {
		sha256Payload = emptyStringSha256
	}
	headers["x-log-content-sha256"] = sha256Payload
	headers["Content-Length"] = strconv.Itoa(contentLength)

	// Canonical header & signedHeaderStr
	canonHeaders := s.buildCanonicalHeader(headers)
	signedHeaderStr := s.buildSignedHeaderStr(canonHeaders)

	// CanonicalRequest
	canonReq := s.buildCanonicalRequest(method, uri, signedHeaderStr, sha256Payload, urlParams, canonHeaders)
	scope := s.buildScope(date, s.region)

	// SignKey + signMessage => signature
	msg := s.buildSignMessage(canonReq, dateTime, scope)
	key, err := s.buildSignKey(s.accessKeySecret, s.region, date)
	if err != nil {
		return err
	}
	hash, err := s.hmacSha256([]byte(msg), key)
	if err != nil {
		return err
	}
	signature := fmt.Sprintf("%x", hash)
	auth := s.buildAuthorization(s.accessKeyID, signature, scope)
	headers["Authorization"] = auth
	return nil
}

func (s *SignerV4) parseUri(uriWithQuery string) (string, map[string]string, error) {
	u, err := url.Parse(uriWithQuery)
	if err != nil {
		return "", nil, err
	}
	urlParams := make(map[string]string)
	for k, vals := range u.Query() {
		if len(vals) == 0 {
			urlParams[k] = ""
		} else {
			urlParams[k] = vals[0] // param val should at most one value
		}
	}
	return u.Path, urlParams, nil
}

func dateTimeISO8601() string {
	return time.Now().In(gmtLoc).Format("20060102T150405Z")
}

func (s *SignerV4) buildCanonicalHeader(headers map[string]string) map[string]string {
	res := make(map[string]string)
	for k, v := range headers {
		lower := strings.ToLower(k)
		_, ok := defaultSignedHeaders[lower]
		if ok || strings.HasPrefix(lower, "x-log-") || strings.HasPrefix(lower, "x-acs-") {
			res[lower] = v
		}
	}
	return res
}

func (s *SignerV4) buildSignedHeaderStr(canonicalHeaders map[string]string) string {
	res, sep := "", ""
	s.forEachSorted(canonicalHeaders, func(k, v string) {
		res += sep + k
		sep = ";"
	})
	return res
}

// Iterate over m in sorted order, and apply func f
func (s *SignerV4) forEachSorted(m map[string]string, f func(k, v string)) {
	var ss sort.StringSlice
	for k := range m {
		ss = append(ss, k)
	}
	ss.Sort()
	for _, k := range ss {
		f(k, m[k])
	}
}

func (s *SignerV4) buildCanonicalRequest(method, uri, signedHeaderStr, sha256Payload string, urlParams, canonicalHeaders map[string]string) string {
	res := ""

	res += method + "\n"
	res += s.urlEncode(uri, true) + "\n"

	// Url params
	canonParams := make(map[string]string)
	for k, v := range urlParams {
		ck := s.urlEncode(strings.TrimSpace(k), false)
		cv := s.urlEncode(strings.TrimSpace(v), false)
		canonParams[ck] = cv
	}

	sep := ""
	s.forEachSorted(canonParams, func(k, v string) {
		res += sep + k
		sep = "&"
		if len(v) != 0 {
			res += "=" + v
		}
	})
	res += "\n"

	// Canonical headers
	s.forEachSorted(canonicalHeaders, func(k, v string) {
		res += k + ":" + strings.TrimSpace(v) + "\n"
	})
	res += "\n"

	res += signedHeaderStr + "\n"
	res += sha256Payload
	return res
}

func (s *SignerV4) urlEncode(uri string, ignoreSlash bool) string {
	u := url.QueryEscape(uri)
	u = strings.ReplaceAll(u, "+", "%20")
	u = strings.ReplaceAll(u, "*", "%2A")
	if ignoreSlash {
		u = strings.ReplaceAll(u, "%2F", "/")
	}
	return u
}

func (s *SignerV4) buildScope(date, region string) string {
	return date + "/" + region + "/sls/aliyun_v4_request"
}

func (s *SignerV4) buildSignMessage(canonReq, dateTime, scope string) string {
	return "SLS4-HMAC-SHA256" + "\n" + dateTime + "\n" + scope + "\n" + fmt.Sprintf("%x", sha256.Sum256([]byte(canonReq)))
}

func (s *SignerV4) hmacSha256(message, key []byte) ([]byte, error) {
	hmacHasher := hmac.New(sha256.New, key)
	_, err := hmacHasher.Write(message)
	if err != nil {
		return nil, errors.Wrap(err, "hmac-sha256")
	}
	return hmacHasher.Sum(nil), nil
}

func (s *SignerV4) buildSignKey(accessKeySecret, region, date string) ([]byte, error) {
	signDate, err := s.hmacSha256([]byte(date), []byte("aliyun_v4"+accessKeySecret))
	if err != nil {
		return nil, errors.Wrap(err, "signDate")
	}
	signRegion, err := s.hmacSha256([]byte(region), signDate)
	if err != nil {
		return nil, errors.Wrap(err, "signRegion")
	}
	signService, err := s.hmacSha256([]byte("sls"), signRegion)
	if err != nil {
		return nil, errors.Wrap(err, "signProductName")
	}
	signAll, err := s.hmacSha256([]byte("aliyun_v4_request"), signService)
	if err != nil {
		return nil, errors.Wrap(err, "signTerminator")
	}
	return signAll, nil
}

func (s *SignerV4) buildAuthorization(accessKeyID, signature, scope string) string {
	return fmt.Sprintf("SLS4-HMAC-SHA256 Credential=%s/%s,Signature=%s",
		accessKeyID, scope, signature)
}
