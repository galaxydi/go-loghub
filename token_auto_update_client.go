package sls

import (
	"errors"
	"sync"
	"time"

	"github.com/golang/glog"
)

type TokenAutoUpdateClient struct {
	logClient              ClientInterface
	shutdown               chan struct{}
	tokenUpdateFunc        UpdateTokenFunction
	maxTryTimes            int
	waitIntervalMin        time.Duration
	waitIntervalMax        time.Duration
	updateTokenIntervalMin time.Duration
	nextExpire             time.Time

	lock               sync.Mutex
	lastFetch          time.Time
	lastRetryFailCount int
	lastRetryInterval  time.Duration
}

var errSTSFetchHighFrequency = errors.New("sts token fetch frequency is too high")

func (c *TokenAutoUpdateClient) flushSTSToken() {
	for {
		nowTime := time.Now()
		c.lock.Lock()
		sleepTime := c.nextExpire.Sub(nowTime)
		if sleepTime < time.Duration(time.Minute) {
			sleepTime = time.Duration(time.Second * 5)

		} else if sleepTime < time.Duration(time.Minute*10) {
			sleepTime = sleepTime / 10 * 7
		} else if sleepTime < time.Duration(time.Hour) {
			sleepTime = sleepTime / 10 * 6
		} else {
			sleepTime = sleepTime / 10 * 5
		}
		c.lock.Unlock()
		glog.V(1).Info("next fetch sleep interval %s", sleepTime.String())
		trigger := time.After(sleepTime)
		select {
		case <-trigger:
			err := c.fetchSTSToken()
			glog.V(1).Info("fetch sts token done, error", err)
		case <-c.shutdown:
			glog.V(1).Info("receive shutdown signal, exit flushSTSToken")
			return
		}
	}

}

func (c *TokenAutoUpdateClient) fetchSTSToken() error {
	nowTime := time.Now()
	skip := false
	sleepTime := time.Duration(0)
	c.lock.Lock()
	if nowTime.Sub(c.lastFetch) < c.updateTokenIntervalMin {
		skip = true
	} else {
		c.lastFetch = nowTime
	}
	if c.lastRetryFailCount == 0 {
		sleepTime = 0
	} else {
		c.lastRetryInterval *= 2
		if c.lastRetryInterval < c.waitIntervalMin {
			c.lastRetryInterval = c.waitIntervalMin
		}
		if c.lastRetryInterval >= c.waitIntervalMax {
			c.lastRetryInterval = c.waitIntervalMax
		}
		sleepTime = c.lastRetryInterval
	}
	c.lock.Unlock()
	if skip {
		return errSTSFetchHighFrequency
	}
	if sleepTime > time.Duration(0) {
		time.Sleep(sleepTime)
	}

	accessKeyID, accessKeySecret, securityToken, expireTime, err := c.tokenUpdateFunc()
	if err == nil {
		c.lock.Lock()
		c.lastRetryFailCount = 0
		c.lastRetryInterval = time.Duration(0)
		c.nextExpire = expireTime
		c.lock.Unlock()
		c.logClient.ResetAccessKeyToken(accessKeyID, accessKeySecret, securityToken)

	} else {
		c.lock.Lock()
		c.lastRetryFailCount++
		c.lock.Unlock()
		glog.Warning("fetch sts token error", err.Error())
	}
	return err
}

func (c *TokenAutoUpdateClient) processError(err error) (retry bool) {
	if err == nil {
		return false
	}
	if IsTokenError(err) {
		if fetchErr := c.fetchSTSToken(); fetchErr != nil {
			glog.Warning("operation error", err.Error(), "fetch sts token error", fetchErr.Error())
			// if fetch error, return false
			return false
		}
		return true
	}
	return false

}

func (c *TokenAutoUpdateClient) ResetAccessKeyToken(accessKeyID, accessKeySecret, securityToken string) {
	c.logClient.ResetAccessKeyToken(accessKeyID, accessKeySecret, securityToken)
}

func (c *TokenAutoUpdateClient) CreateProject(name, description string) (*LogProject, error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetProject(name string) (*LogProject, error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) ListProject() (projectNames []string, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) CheckProjectExist(name string) (bool, error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) DeleteProject(name string) error {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) ListLogStore(project string) ([]string, error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetLogStore(project string, logstore string) (*LogStore, error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) CreateLogStore(project string, logstore string, ttl, shardCnt int) error {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) DeleteLogStore(project string, logstore string) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) UpdateLogStore(project string, logstore string, ttl, shardCnt int) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) ListMachineGroup(project string, offset, size int) (m []string, total int, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) ListMachines(project, machineGroupName string) (ms []*Machine, total int, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) CheckLogstoreExist(project string, logstore string) (bool, error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) CheckMachineGroupExist(project string, machineGroup string) (bool, error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetMachineGroup(project string, machineGroup string) (m *MachineGroup, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) CreateMachineGroup(project string, m *MachineGroup) error {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) UpdateMachineGroup(project string, m *MachineGroup) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) DeleteMachineGroup(project string, machineGroup string) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) ListConfig(project string, offset, size int) (cfgNames []string, total int, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) CheckConfigExist(project string, config string) (ok bool, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetConfig(project string, config string) (logConfig *LogConfig, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) UpdateConfig(project string, config *LogConfig) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) CreateConfig(project string, config *LogConfig) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) DeleteConfig(project string, config string) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetAppliedMachineGroups(project string, confName string) (groupNames []string, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetAppliedConfigs(project string, groupName string) (confNames []string, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) ApplyConfigToMachineGroup(project string, confName, groupName string) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) RemoveConfigFromMachineGroup(project string, confName, groupName string) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) CreateEtlMeta(project string, etlMeta *EtlMeta) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) UpdateEtlMeta(project string, etlMeta *EtlMeta) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) DeleteEtlMeta(project string, etlMetaName, etlMetaKey string) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) listEtlMeta(project string, etlMetaName, etlMetaKey, etlMetaTag string, offset, size int) (total int, count int, etlMeta []*EtlMeta, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetEtlMeta(project string, etlMetaName, etlMetaKey string) (etlMeta *EtlMeta, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) ListEtlMeta(project string, etlMetaName string, offset, size int) (total int, count int, etlMetaList []*EtlMeta, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) ListEtlMetaWithTag(project string, etlMetaName, etlMetaTag string, offset, size int) (total int, count int, etlMetaList []*EtlMeta, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) ListEtlMetaName(project string, offset, size int) (total int, count int, etlMetaNameList []string, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) ListShards(project, logstore string) (shardIDs []int, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) PutLogs(project, logstore string, lg *LogGroup) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) PutLogsWithCompressType(project, logstore string, lg *LogGroup, compressType int) (err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetCursor(project, logstore string, shardID int, from string) (cursor string, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetLogsBytes(project, logstore string, shardID int, cursor, endCursor string,
	logGroupMaxCount int) (out []byte, nextCursor string, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) PullLogs(project, logstore string, shardID int, cursor, endCursor string,
	logGroupMaxCount int) (gl *LogGroupList, nextCursor string, err error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetHistograms(project, logstore string, topic string, from int64, to int64, queryExp string) (*GetHistogramsResponse, error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetLogs(project, logstore string, topic string, from int64, to int64, queryExp string,
	maxLineNum int64, offset int64, reverse bool) (*GetLogsResponse, error) {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) CreateIndex(project, logstore string, index Index) error {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) UpdateIndex(project, logstore string, index Index) error {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) DeleteIndex(project, logstore string) error {
	panic("implement me")
}

func (c *TokenAutoUpdateClient) GetIndex(project, logstore string) (*Index, error) {
	panic("implement me")
}
