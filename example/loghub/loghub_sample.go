package main

import (
	"fmt"
	"time"
	"strconv"
	"math/rand"
	"github.com/aliyun/aliyun-log-go-sdk/example/util"
	"github.com/gogo/protobuf/proto"
	sls "github.com/aliyun/aliyun-log-go-sdk"
)

func main() {

	fmt.Println("loghub sample begin")
	begin_time := uint32(time.Now().Unix())
	rand.Seed(int64(begin_time))
	logstore_name := "test"
	logstore, err := util.Project.GetLogStore(logstore_name)
	if logstore == nil {
		fmt.Printf("GetLogStore fail, err:%v\n", err)
		err = util.Project.CreateLogStore(logstore_name, 1, 2)
		if err != nil {
			fmt.Printf("CreateLogStore fail, err: ", err)
			return
		}
		fmt.Println("CreateLogStore success")
	} else {
		fmt.Printf("GetLogStore success, name: %s, ttl: %d, shardCount: %d, createTime: %d, lastModifyTime: %d\n", logstore.Name, logstore.TTL, logstore.ShardCount, logstore.CreateTime, logstore.LastModifyTime)
	}
	// put logs to logstore
	for loggroupIdx := 0; loggroupIdx < 2; loggroupIdx++ {
		logs := []*sls.Log {}
		for logIdx := 0; logIdx < 100; logIdx++ {
			content := []*sls.LogContent {}
			for colIdx := 0; colIdx < 10; colIdx++ {
				content = append(content, &sls.LogContent {
					Key: proto.String(fmt.Sprintf("col_%d", colIdx)),
					Value: proto.String(fmt.Sprintf("loggroup idx: %d, log idx: %d, col idx: %d, value: %d", loggroupIdx, logIdx, colIdx, rand.Intn(10000000))),
				})
			}
			log := &sls.Log{
				Time: proto.Uint32(uint32(time.Now().Unix())),
				Contents: content, 
			}
			logs = append(logs, log)
		}
		loggroup := &sls.LogGroup {
			Topic: proto.String(""),
			Source: proto.String("10.230.201.117"),
			Logs: logs,
		}
		// PostLogStoreLogs API Ref: https://intl.aliyun.com/help/doc-detail/29026.htm
		err = logstore.PutLogs(loggroup)
		if err == nil {
			fmt.Println("PutLogs success")
		} else {
			fmt.Printf("PutLogs fail, err: %s\n", err)
		}
		time.Sleep(1000 * time.Millisecond)
	}
	// pull logs from logstore 
	shards, err := logstore.ListShards()
	for _, sh := range shards {
		if sh == 0 {
			// GetCursor API Ref: https://intl.aliyun.com/help/doc-detail/29024.htm
			begin_cursor, _ := logstore.GetCursor(sh, "begin")	
			end_cursor, _ := logstore.GetCursor(sh, "end")
			// PullLogs API Ref: https://intl.aliyun.com/help/doc-detail/29025.htm
			loggrouplist, next_cursor, _ := logstore.PullLogs(sh, begin_cursor, end_cursor, 100)
			fmt.Printf("shard: %d, begin_cursor: %s, end_cursor: %s, next_cursor: %s\n", sh, begin_cursor, end_cursor, next_cursor)
			for _, loggroup := range loggrouplist.LogGroups {
				for _, log := range loggroup.Logs {
					for _, content := range log.Contents {
						fmt.Printf("key:%s, value:%s\n", content.GetKey(), content.GetValue())
					}
				}
			}
		} else {
			begin_cursor, _ := logstore.GetCursor(sh, strconv.Itoa(int(begin_time) + 2))
			for {
				loggrouplist, next_cursor, _ := logstore.PullLogs(sh, begin_cursor, "", 2)
				fmt.Printf("shard: %d, begin_cursor: %s, next_cursor: %s, len(loggrouplist.LogGroups): %d\n", sh, begin_cursor, next_cursor, len(loggrouplist.LogGroups))
				if len(loggrouplist.LogGroups) == 0 {
					// means no more data in this shard, you can break out or sleep to wait new data
					break
				} else {
					for _, loggroup := range loggrouplist.LogGroups {
						for _, log := range loggroup.Logs {
							for _, content := range log.Contents {
								fmt.Printf("key:%s, value:%s\n", content.GetKey(), content.GetValue())
							}
						}
					}
					begin_cursor = next_cursor
				}
			}
		}
	}
	fmt.Println("loghub sample end")
}
