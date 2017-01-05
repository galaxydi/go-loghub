package main

import (
	"fmt"
	"time"

	sls "github.com/galaxydi/go-loghub"
	"github.com/gogo/protobuf/proto"
)

const (
	project         = "test-project"
	logstoreName    = "test-logstore"
	endpoint        = "cn-hangzhou.log.aliyuncs.com"
	accessKeyID     = "xxx"
	accessKeySecret = "xxx"
	tmpAK           = "xxx"
	tmpSecret       = "xxx"
	token           = "xxx"
)

// PutLogsSample ...
func PutLogsSample() {
	p1, _ := sls.NewLogProject(project, endpoint, accessKeyID, accessKeySecret)
	s1, _ := p1.GetLogStore(logstoreName)
	content := &sls.LogContent{
		Key:   proto.String("demo_key"),
		Value: proto.String("demo_value"),
	}
	logRecord := &sls.Log{
		Time:     proto.Uint32(uint32(time.Now().Unix())),
		Contents: []*sls.LogContent{content},
	}
	lg := &sls.LogGroup{
		Topic:  proto.String("test"),
		Source: proto.String("10.168.122.110"),
		Logs:   []*sls.Log{logRecord},
	}
	fmt.Println("access with AK")
	err := s1.PutLogs(lg)
	if err != nil {
		fmt.Println("PutLogs to " + s1.Name + " fail:")
		fmt.Println(err)
	} else {
		fmt.Println("PutLogs to " + s1.Name + " success")
	}

	fmt.Println("access with token")
	p2, _ := sls.NewLogProject(projectName, endpoint, tmpAK, tmpSecret)
	p2.WithToken(token)
	s2, err := p2.GetLogStore(logstoreName)
	if err != nil {
		fmt.Println("GetLogstore fail:", err)
	}
	lg = &sls.LogGroup{
		Topic:  proto.String("token-test"),
		Source: proto.String("10.168.122.110"),
		Logs:   []*sls.Log{logRecord},
	}
	err = s2.PutLogs(lg)
	if err != nil {
		fmt.Println("PutLogs to " + s2.Name + " fail:")
		fmt.Println(err)
	} else {
		fmt.Println("PutLogs to " + s2.Name + " success")
	}
}
