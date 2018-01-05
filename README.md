This is a Golang SDK for Alibaba Cloud [Log Service](https://sls.console.aliyun.com/).

API Reference :

* [Chinese](https://help.aliyun.com/document_detail/29007.html)
* [English](https://intl.aliyun.com/help/doc-detail/29007.htm)

[![Build Status](https://travis-ci.org/aliyun-fc/go-loghub.svg?branch=master)](https://travis-ci.org/aliyun-fc/go-loghub)
[![Coverage Status](https://coveralls.io/repos/github/aliyun-fc/go-loghub/badge.svg?branch=master&foo=bar)](https://coveralls.io/github/aliyun-fc/go-loghub?branch=master&foo=bar)


# Install Instruction

### Third Dependencies

```
go get github.com/cloudflare/golz4
go get github.com/golang/glog
go get github.com/gogo/protobuf/proto
go get github.com/stretchr/testify/suite
go get github.com/cenkalti/backoff
go get github.com/pkg/errors
go get golang.org/x/net/context
```

### LogHub Golang SDK

```
go get github.com/aliyun-fc/go-loghub 
```

# Example 

### Write and Read LogHub

[loghub_sample.go](example/loghub/loghub_sample.go)

### Use Index on LogHub (SLS)

[index_sample.go](example/index/index_sample.go)

### Create Config for Logtail

[log_config_sample.go](example/config/log_config_sample.go)

### Create Machine Group for Logtail

[machine_group_sample.go](example/machine_group/machine_group_sample.go)
