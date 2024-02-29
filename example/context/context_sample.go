package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/example/util"
	"github.com/gogo/protobuf/proto"
)

func panicCheck(err error, msg string) {
	if err != nil {
		fmt.Println(err)
		panic(msg)
	}
}

func postLogGroup(logCount int, packID string, logstore *sls.LogStore) {
	lg := &sls.LogGroup{
		Topic:  proto.String(""),
		Source: proto.String("11.11.11.11"),
		LogTags: []*sls.LogTag{
			&sls.LogTag{Key: proto.String("__pack_id__"), Value: proto.String(packID)},
			&sls.LogTag{Key: proto.String("__extra_tag__"), Value: proto.String("extra_tag_value")},
		},
	}
	for idx := 0; idx < logCount; idx++ {
		log := &sls.Log{
			Time: proto.Uint32(uint32(time.Now().Unix())),
			Contents: []*sls.LogContent{
				&sls.LogContent{Key: proto.String("index"), Value: proto.String(fmt.Sprint(idx))},
			},
		}
		lg.Logs = append(lg.Logs, log)
	}
	panicCheck(logstore.PutLogs(lg), "PutLogs")
}

// generatePackIDPrefix generates a PackID by MD5(time, hostname, PID).
func generatePackIDPrefix() string {
	m := md5.New()
	m.Write([]byte(time.Now().String()))
	hostName, _ := os.Hostname()
	m.Write([]byte(hostName))
	m.Write([]byte(fmt.Sprintf("%v", os.Getpid())))
	return fmt.Sprintf("%X", m.Sum(nil))
}

func main() {
	sls.GlobalForceUsingHTTP = true
	fmt.Println(util.AccessKeyID)
	client := sls.CreateNormalInterface(util.Endpoint, util.AccessKeyID, util.AccessKeySecret, "")
	project, err := client.GetProject(util.ProjectName)
	panicCheck(err, "GetProject")
	logstore, err := sls.NewLogStore(util.LogStoreName, project)
	panicCheck(err, "NewLogStore")

	beginTime := time.Now()

	// Write several log groups with same pack prefix. PackID format: {PackPrefix}-{HexSequenceID}.
	packagePrefix := generatePackIDPrefix()
	for seqID := 0; seqID <= 10; seqID++ {
		postLogGroup(20, fmt.Sprintf("%v-%X", packagePrefix, seqID), logstore)
	}
	time.Sleep(time.Second * 5)

	// Use GetLogs to acquire packMeta by querying packID.
	queriedPackID := fmt.Sprintf("%v-%X", packagePrefix, 8)
	from := beginTime.Unix() - 120
	to := time.Now().Unix() + 60
	resp, err := logstore.GetLogs("", from, to,
		fmt.Sprintf("__tag__:__pack_id__:%v|with_pack_meta", queriedPackID),
		20, 0, false)
	panicCheck(err, "GetLogs")
	fmt.Println("GetLogs response", resp.Count)
	middleLog := resp.Logs[resp.Count/2]
	packID := middleLog["__tag__:__pack_id__"]
	packMeta := middleLog["__pack_meta__"]
	fmt.Println(packID, packMeta)

	// Get context logs from both directions.
	contextResp, err := logstore.GetContextLogs(30, 30, packID, packMeta)
	panicCheck(err, "GetContextLogs")
	fmt.Println("GetContextLogs response", contextResp.TotalLines)
	fmt.Println("back lines", contextResp.BackLines)
	fmt.Println("forward lines", contextResp.ForwardLines)
	fmt.Println("oldest context log", contextResp.Logs[0])
	fmt.Println("newest context log", contextResp.Logs[contextResp.TotalLines-1])

	// Use the first log to fetch backward.
	fmt.Println("fetch context logs backward...")
	{
		log := contextResp.Logs[0]
		for loopIdx := 0; loopIdx < 5; loopIdx++ {
			packID := log["__tag__:__pack_id__"]
			packMeta := log["__pack_meta__"]
			fmt.Printf("[Loop %v] ID: %v, meta: %v\n", loopIdx, packID, packMeta)
			resp, err := logstore.GetContextLogs(30, 0, packID, packMeta)
			panicCheck(err, fmt.Sprintf("GetContextLogs backward %v", loopIdx))
			fmt.Printf("[Loop %v] backward total lines: %v, back lines: %v\n",
				loopIdx, resp.TotalLines, resp.BackLines)
			if resp.TotalLines == 0 {
				fmt.Println("No more log backward")
				break
			}
			log = resp.Logs[0]
			fmt.Printf("[Loop %v] log: %v\n", loopIdx, log)
			time.Sleep(time.Second)
		}
	}

	// Use the last log to fetch forward.
	fmt.Println("fetch context logs forward...")
	{
		log := contextResp.Logs[contextResp.TotalLines-1]
		for loopIdx := 0; loopIdx < 5; loopIdx++ {
			packID := log["__tag__:__pack_id__"]
			packMeta := log["__pack_meta__"]
			fmt.Printf("[Loop %v] ID: %v, meta: %v\n", loopIdx, packID, packMeta)
			resp, err := logstore.GetContextLogs(0, 30, packID, packMeta)
			panicCheck(err, fmt.Sprintf("GetContextLogs backward %v", loopIdx))
			fmt.Printf("[Loop %v] forward total lines: %v, lines: %v\n",
				loopIdx, resp.TotalLines, resp.ForwardLines)
			if resp.TotalLines == 0 {
				fmt.Println("No more log forward")
				break
			}
			log = resp.Logs[resp.TotalLines-1]
			fmt.Printf("[Loop %v] log: %v\n", loopIdx, log)
			time.Sleep(time.Second)
		}
	}
}
