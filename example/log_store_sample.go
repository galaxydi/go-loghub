package main

import (
	"fmt"
	"os"
	"time"

	sls "github.com/galaxydi/go-loghub"
	"github.com/galaxydi/go-loghub/example/util"

	"github.com/gogo/protobuf/proto"
)

// LogstoreSample ...
func LogstoreSample() {
	list, err := util.Project.ListLogStore()
	for _, v := range list {
		_, err := util.Project.GetLogStore(v)
		if err != nil {
			fmt.Println("GetLogStore fail:" + v)
			fmt.Println(err)
			os.Exit(1)
		}
	}

	teststore := "store5"
	err = util.Project.CreateLogStore(teststore, 1, 1)
	if err != nil {
		if !(err.(*sls.Error).Code == "SLSLogStoreAlreadyExist") {
			fmt.Println("create logstore:" + teststore + " fail")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	util.Project.UpdateLogStore(teststore, 2, 1)
	err = util.Project.DeleteLogStore(teststore)
	if err != nil {
		if !(err.(*sls.Error).Code == "SLSLogStoreNotExist") {
			fmt.Println("delete log store:" + teststore + " fail")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	store, err := util.Project.GetLogStore(logstore)
	_, err = store.ListShards()
	if err != nil {
		fmt.Println("ListShards fail:")
		fmt.Println(err)
		os.Exit(1)
	}

	// Construct a LogGroup
	c := &sls.LogContent{
		Key:   proto.String("errorCode"),
		Value: proto.String("InternalServerError"),
	}
	l := &sls.Log{
		Time: proto.Uint32(uint32(time.Now().Unix())),
		Contents: []*sls.LogContent{
			c,
		},
	}
	lg := &sls.LogGroup{
		Topic:  proto.String(""),
		Source: proto.String("10.230.201.117"),
		Logs: []*sls.Log{
			l,
		},
	}
	err = store.PutLogs(lg)
	if err != nil {
		fmt.Println("PutLogs to " + store.Name + " fail:")
		fmt.Println(err)
		os.Exit(1)
	}
	from := []string{
		"begin",
		"end",
		fmt.Sprintf("%v", time.Now().Unix()),
	}
	cursor := ""
	for _, f := range from {
		c, err := store.GetCursor(0, f)
		if err != nil {
			fmt.Println("GetCursor fail:")
			fmt.Println(err)
			os.Exit(1)
		}
		cursor = c
		break
	}

	endCursor, _ := store.GetCursor(0, "end")
	for {
		gl, next, err := store.PullLogs(0, cursor, endCursor, 100)
		if err != nil {
			fmt.Println("GetLogs from:" + store.Name + " fail:")
			fmt.Println(err)
			os.Exit(1)
		}
		for _, lg := range gl.LogGroups {
			var s string
			for _, l := range lg.Logs {
				for _, c := range l.Contents {
					s += fmt.Sprintf("%v:%v\n", *c.Key, *c.Value)
				}
			}
		}
		if next == cursor {
			break
		}
		cursor = next
	}
	fmt.Println("logstore sample end")
}
