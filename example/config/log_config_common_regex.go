package main

import (
	"fmt"
	"github.com/aliyun/aliyun-log-go-sdk"
	config "github.com/aliyun/aliyun-log-go-sdk/example/config/util"
	"github.com/aliyun/aliyun-log-go-sdk/example/util"
)

func main() {
	// log config sample
	testConf := "test-conf5"
	exist, err := config.CheckConfigExist(testConf)
	fmt.Println(exist)
	if err != nil {
		fmt.Println("check conf exist fail:", err)
		return
	}
	if exist {
		config.DeleteConfig(testConf)
	}
	err = createConfig(testConf, util.ProjectName, util.LogStoreName)
	if err != nil {
		fmt.Println("create config fail:", err)
		return
	}
	fmt.Println("create common regex logtail config sucessed")

	updateConfig(testConf)
	config.GetConfig(testConf)

	exist, err = config.CheckConfigExist(testConf)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !exist {
		fmt.Println("config:" + testConf + " should be exist")
		return
	}

	config.DeleteConfig(testConf)
	fmt.Println("delete common regex logtail config sucessed")
	exist, err = config.CheckConfigExist(testConf)
	if err != nil {
		fmt.Println(err)
		return
	}
	if exist {
		fmt.Println("config:" + testConf + " should not be exist")
		return
	}
	fmt.Println("log config sample end")

}

func createConfig(configName string, projectName string, logstore string) (err error) {
	regexConfig := new(sls.RegexConfigInputDetail)
	sls.InitRegexConfigInputDetail(regexConfig)
	regexConfig.DiscardUnmatch = false
	regexConfig.Key = []string{"logger", "time", "cluster", "hostname", "sr", "app", "workdir", "exe", "corepath", "signature", "backtrace"}
	regexConfig.Regex = "\\S*\\s+(\\S*)\\s+(\\S*\\s+\\S*)\\s+\\S*\\s+(\\S*)\\s+(\\S*)\\s+(\\S*)\\s+(\\S*)\\s+(\\S*)\\s+(\\S*)\\s+(\\S*)\\s+\\S*\\s+(\\S*)\\s*([^$]+)"
	regexConfig.TimeFormat = "%Y/%m/%d %H:%M:%S"
	regexConfig.LogBeginRegex = `INFO core_dump_info_data .*`
	regexConfig.LogPath = "/cloud/log/tianji/TianjiClient#/core_dump_manager"
	regexConfig.FilePattern = "core_dump_info_data.log*"
	regexConfig.MaxDepth = 0
	outputDetail := sls.OutputDetail{
		ProjectName:  projectName,
		LogStoreName: logstore,
	}
	logConfig := &sls.LogConfig{
		Name:         configName,
		InputType:    "file",
		OutputType:   "LogService", // Now only supports LogService
		InputDetail:  regexConfig,
		OutputDetail: outputDetail,
	}
	err = util.Client.CreateConfig(projectName, logConfig)
	if err != nil {
		return err
	}
	return nil
}

func updateConfig(configName string) {
	logtailConfig, _ := util.Client.GetConfig(util.ProjectName, configName)
	inputDetail, _ := sls.ConvertToRegexConfigInputDetail(logtailConfig.InputDetail)
	inputDetail.FilePattern = "*.log"
	err := util.Client.UpdateConfig(util.ProjectName, logtailConfig)
	if err != nil {
		panic(err)
	}
}
