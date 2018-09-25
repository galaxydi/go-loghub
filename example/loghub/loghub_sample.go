package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/example/util"
	"github.com/gogo/protobuf/proto"
)

func main() {

	fmt.Println("loghub sample begin")
	var err error
	offset := 0
	fmt.Println("project list: ")
	for {
		projects, count, total, err := util.Client.ListProjectV2(offset, 100)
		if err != nil {
			panic(err)
		}
		for _, project := range projects {
			fmt.Printf(" name : %s, description : %s, region : %s, ctime : %s, mtime : %s\n",
				project.Name,
				project.Description,
				project.Region,
				project.CreateTime,
				project.LastModifyTime)
		}
		if offset+count >= total {
			break
		}
		offset += count
	}

	beginTime := uint32(time.Now().Unix())
	rand.Seed(int64(beginTime))
	logstoreName := "test-logstore"
	var retryTimes int
	var logstore *sls.LogStore
	for retryTimes = 0; ; retryTimes++ {
		if retryTimes > 5 {
			return
		}
		logstore, err = util.Client.GetLogStore(util.ProjectName, logstoreName)
		if err != nil {
			fmt.Printf("GetLogStore fail, retry:%d, err:%v\n", retryTimes, err)
			if strings.Contains(err.Error(), sls.PROJECT_NOT_EXIST) {
				return
			} else if strings.Contains(err.Error(), sls.LOGSTORE_NOT_EXIST) {
				err = util.Client.CreateLogStore(util.ProjectName, logstoreName, 1, 2, true, 16)
				if err != nil {
					fmt.Printf("CreateLogStore fail, err: %s ", err.Error())
				} else {
					fmt.Println("CreateLogStore success")
				}
			}
		} else {
			fmt.Printf("GetLogStore success, retry:%d, name: %s, ttl: %d, shardCount: %d, createTime: %d, lastModifyTime: %d\n", retryTimes, logstore.Name, logstore.TTL, logstore.ShardCount, logstore.CreateTime, logstore.LastModifyTime)
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	// put logs to logstore
	for loggroupIdx := 0; loggroupIdx < 2; loggroupIdx++ {
		logs := []*sls.Log{}
		for logIdx := 0; logIdx < 100; logIdx++ {
			content := []*sls.LogContent{}
			for colIdx := 0; colIdx < 10; colIdx++ {
				content = append(content, &sls.LogContent{
					Key:   proto.String(fmt.Sprintf("col_%d", colIdx)),
					Value: proto.String(fmt.Sprintf("loggroup idx: %d, log idx: %d, col idx: %d, value: %d", loggroupIdx, logIdx, colIdx, rand.Intn(10000000))),
				})
			}
			log := &sls.Log{
				Time:     proto.Uint32(uint32(time.Now().Unix())),
				Contents: content,
			}
			logs = append(logs, log)
		}
		loggroup := &sls.LogGroup{
			Topic:  proto.String(""),
			Source: proto.String("10.230.201.117"),
			Logs:   logs,
		}
		// PostLogStoreLogs API Ref: https://intl.aliyun.com/help/doc-detail/29026.htm
		for retryTimes = 0; retryTimes < 10; retryTimes++ {
			err := util.Client.PutLogs(util.ProjectName, logstoreName, loggroup)
			if err == nil {
				fmt.Printf("PutLogs success, retry: %d\n", retryTimes)
				break
			} else {
				fmt.Printf("PutLogs fail, retry: %d, err: %s\n", retryTimes, err)
				//handle exception here, you can add retryable erorrCode, set appropriate put_retry
				if strings.Contains(err.Error(), sls.WRITE_QUOTA_EXCEED) || strings.Contains(err.Error(), sls.PROJECT_QUOTA_EXCEED) || strings.Contains(err.Error(), sls.SHARD_WRITE_QUOTA_EXCEED) {
					//mayby you should split shard
					time.Sleep(1000 * time.Millisecond)
				} else if strings.Contains(err.Error(), sls.INTERNAL_SERVER_ERROR) || strings.Contains(err.Error(), sls.SERVER_BUSY) {
					time.Sleep(200 * time.Millisecond)
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	// pull logs from logstore
	var shards []*sls.Shard
	for retryTimes = 0; ; retryTimes++ {
		if retryTimes > 5 {
			return
		}
		shards, err = util.Client.ListShards(util.ProjectName, logstoreName)
		if err != nil {
			fmt.Printf("ListShards fail, retry: %d, err:%v\n", retryTimes, err)
		} else {
			fmt.Printf("ListShards success\n")
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	var beginCursor string
	var endCursor string
	var nextCursor string
	var loggrouplist *sls.LogGroupList
	for _, sh := range shards {
		if sh.ShardID == 0 {
			// sample of pulllogs from begin
			// GetCursor API Ref: https://intl.aliyun.com/help/doc-detail/29024.htm
			for retryTimes = 0; ; retryTimes++ {
				if retryTimes > 5 {
					return
				}
				beginCursor, err = util.Client.GetCursor(util.ProjectName, logstoreName, sh.ShardID, "begin")
				if err != nil {
					fmt.Printf("GetCursor(begin) fail, retry: %d, err:%v\n", retryTimes, err)
				} else {
					fmt.Printf("GetCursor(begin) success\n")
					break
				}
				time.Sleep(200 * time.Millisecond)
			}
			for retryTimes = 0; ; retryTimes++ {
				if retryTimes > 5 {
					return
				}
				endCursor, err = util.Client.GetCursor(util.ProjectName, logstoreName, sh.ShardID, "end")
				if err != nil {
					fmt.Printf("GetCursor(end) fail, retry: %d, err:%v\n", retryTimes, err)
				} else {
					fmt.Printf("GetCursor(end) success\n")
					break
				}
				time.Sleep(200 * time.Millisecond)
			}
			// PullLogs API Ref: https://intl.aliyun.com/help/doc-detail/29025.htm
			for retryTimes = 0; ; retryTimes++ {
				if retryTimes > 100 {
					return
				}
				loggrouplist, nextCursor, err = util.Client.PullLogs(util.ProjectName, logstoreName, sh.ShardID, beginCursor, endCursor, 100)
				if err == nil {
					fmt.Printf("PullLogs success, retry: %d\n", retryTimes)
					break
				} else {
					fmt.Printf("PullLogs fail, retry: %d, err: %s\n", retryTimes, err)
					//handle exception here, you can add retryable erorrCode, set appropriate put_retry
					if strings.Contains(err.Error(), sls.READ_QUOTA_EXCEED) || strings.Contains(err.Error(), sls.PROJECT_QUOTA_EXCEED) || strings.Contains(err.Error(), sls.SHARD_READ_QUOTA_EXCEED) {
						//mayby you should split shard
						time.Sleep(1000 * time.Millisecond)
					} else if strings.Contains(err.Error(), sls.INTERNAL_SERVER_ERROR) || strings.Contains(err.Error(), sls.SERVER_BUSY) {
						time.Sleep(200 * time.Millisecond)
					}
				}
			}
			fmt.Printf("shard: %d, begin_cursor: %s, end_cursor: %s, next_cursor: %s\n", sh.ShardID, beginCursor, endCursor, nextCursor)
			for _, loggroup := range loggrouplist.LogGroups {
				for _, log := range loggroup.Logs {
					for _, content := range log.Contents {
						fmt.Printf("key:%s, value:%s\n", content.GetKey(), content.GetValue())
					}
				}
			}
		} else {
			// sample of pulllogs from setted time
			for retryTimes = 0; ; retryTimes++ {
				if retryTimes > 5 {
					return
				}
				beginCursor, err = util.Client.GetCursor(util.ProjectName, logstoreName, sh.ShardID, strconv.Itoa(int(beginTime)+2))
				if err != nil {
					fmt.Printf("GetCursor fail, retry: %d, err:%v\n", retryTimes, err)
				} else {
					fmt.Printf("GetCursor success\n")
					break
				}
				time.Sleep(200 * time.Millisecond)
			}
			for {
				for retryTimes = 0; ; retryTimes++ {
					if retryTimes > 100 {
						return
					}
					loggrouplist, nextCursor, err = util.Client.PullLogs(util.ProjectName, logstoreName, sh.ShardID, beginCursor, "", 2)
					if err == nil {
						fmt.Printf("PullLogs success, retry: %d\n", retryTimes)
						break
					} else {
						fmt.Printf("PullLogs fail, retry: %d, err: %s\n", retryTimes, err)
						//handle exception here, you can add retryable erorrCode, set appropriate put_retry
						if strings.Contains(err.Error(), sls.READ_QUOTA_EXCEED) || strings.Contains(err.Error(), sls.PROJECT_QUOTA_EXCEED) || strings.Contains(err.Error(), sls.SHARD_READ_QUOTA_EXCEED) {
							//mayby you should split shard
							time.Sleep(1000 * time.Millisecond)
						} else if strings.Contains(err.Error(), sls.INTERNAL_SERVER_ERROR) || strings.Contains(err.Error(), sls.SERVER_BUSY) {
							time.Sleep(200 * time.Millisecond)
						}
					}
				}
				fmt.Printf("shard: %d, begin_cursor: %s, next_cursor: %s, len(loggrouplist.LogGroups): %d\n", sh.ShardID, beginCursor, nextCursor, len(loggrouplist.LogGroups))
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
					beginCursor = nextCursor
				}
			}
		}
	}
	fmt.Println("loghub sample end")
}
