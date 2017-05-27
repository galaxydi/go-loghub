package main

import (
	"fmt"
	"os"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/example/util"
)

var projectName = "another-project"
var logstore = "demo-store"

func main() {
	// log config sample
	testConf := "test-conf"
	testService := "demo-service"
	exist, err := checkConfigExist(testConf)
	if err != nil {
		fmt.Println("check conf exist fail:", err)
		os.Exit(1)
	}
	if exist {
		deleteConfig(testConf)
	}
	err = createConfig(testConf, projectName, logstore, testService)
	if err != nil {
		fmt.Println("create config fail:", err)
		os.Exit(1)
	}
	fmt.Println("create config success")

	updateConfig(testConf)
	getConfig(testConf)

	exist, err = checkConfigExist(testConf)
	if err != nil {
		os.Exit(1)
	}
	if !exist {
		fmt.Println("config:" + testConf + " should be exist")
		os.Exit(1)
	}

	deleteConfig(testConf)

	exist, err = checkConfigExist(testConf)
	if err != nil {
		os.Exit(1)
	}
	if exist {
		fmt.Println("config:" + testConf + " should not be exist")
		os.Exit(1)
	}
	fmt.Println("log config sample end")

}

func checkConfigExist(confName string) (exist bool, err error) {
	exist, err = util.Project.CheckConfigExist(confName)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func deleteConfig(confName string) (err error) {
	err = util.Project.DeleteConfig(confName)
	if err != nil {
		return err
	}
	return nil
}

func updateConfig(configName string) (err error) {
	config, _ := util.Project.GetConfig(configName)
	config.InputDetail.FilePattern = "*.log"
	err = util.Project.UpdateConfig(config)
	if err != nil {
		return err
	}
	return nil
}
func getConfig(configName string) (err error) {
	_, err = util.Project.GetConfig(configName)
	if err != nil {
		return err
	}
	return nil
}
func createConfig(configName string, projectName string, logstore string, serviceName string) (err error) {
	keys := []string{"message"}
	inputDetail := sls.InputDetail{
		LogType:       "common_reg_log",
		LogPath:       "/var/log/lambda/" + serviceName,
		FilePattern:   "*.LOG",
		TopicFormat:   "/var/log/lambda/([^/]*)/.*",
		LocalStorage:  true,
		TimeFormat:    "",
		LogBeginRegex: ".*",   // 日志首行特征
		Regex:         "(.*)", // 日志对提取正则表达式
		Keys:          keys,
		FilterKeys:    make([]string, 1),
		FilterRegex:   make([]string, 1),
	}
	outputDetail := sls.OutputDetail{
		ProjectName:  projectName,
		LogStoreName: logstore,
	}
	config := &sls.LogConfig{
		Name:         configName,
		InputType:    "file",       //现在只支持file
		OutputType:   "LogService", //现在只支持LogService
		InputDetail:  inputDetail,
		OutputDetail: outputDetail,
	}
	err = util.Project.CreateConfig(config)
	if err != nil {
		return err
	}
	return nil
}
