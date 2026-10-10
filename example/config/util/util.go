package util

import (
	"github.com/aliyun/aliyun-log-go-sdk/example/util"
)

func CheckConfigExist(confName string) (exist bool, err error) {
	exist, err = util.Client.CheckConfigExist(util.ProjectName, confName)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func DeleteConfig(confName string) {
	err := util.Client.DeleteConfig(util.ProjectName, confName)
	if err != nil {
		panic(err)
	}
}

func GetConfig(configName string) {
	_, err := util.Client.GetConfig(util.ProjectName, configName)
	if err != nil {
		panic(err)
	}
}
