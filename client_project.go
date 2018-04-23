package sls

// ListLogStore returns all logstore names of project p.
func (c *Client) ListLogStore(project string) ([]string, error) {
	proj := convert(c, project)
	return proj.ListLogStore()
}

// GetLogStore returns logstore according by logstore name.
func (c *Client) GetLogStore(project string, logstore string) (*LogStore, error) {
	proj := convert(c, project)
	return proj.GetLogStore(logstore)
}

// CreateLogStore creates a new logstore in SLS,
// where name is logstore name,
// and ttl is time-to-live(in day) of logs,
// and shardCnt is the number of shards.
func (c *Client) CreateLogStore(project string, logstore string, ttl, shardCnt int) error {
	proj := convert(c, project)
	return proj.CreateLogStore(logstore, ttl, shardCnt)
}

// DeleteLogStore deletes a logstore according by logstore name.
func (c *Client) DeleteLogStore(project string, logstore string) (err error) {
	proj := convert(c, project)
	return proj.DeleteLogStore(logstore)
}

// UpdateLogStore updates a logstore according by logstore name,
// obviously we can't modify the logstore name itself.
func (c *Client) UpdateLogStore(project string, logstore string, ttl, shardCnt int) (err error) {
	proj := convert(c, project)
	return proj.UpdateLogStore(logstore, ttl, shardCnt)
}

// ListMachineGroup returns machine group name list and the total number of machine groups.
// The offset starts from 0 and the size is the max number of machine groups could be returned.
func (c *Client) ListMachineGroup(project string, offset, size int) (m []string, total int, err error) {
	proj := convert(c, project)
	return proj.ListMachineGroup(offset, size)
}

// CheckLogstoreExist check logstore exist or not
func (c *Client) CheckLogstoreExist(project string, logstore string) (bool, error) {
	proj := convert(c, project)
	return proj.CheckLogstoreExist(logstore)
}

// CheckMachineGroupExist check machine group exist or not
func (c *Client) CheckMachineGroupExist(project string, machineGroup string) (bool, error) {
	proj := convert(c, project)
	return proj.CheckMachineGroupExist(machineGroup)
}

// GetMachineGroup retruns machine group according by machine group name.
func (c *Client) GetMachineGroup(project string, machineGroup string) (m *MachineGroup, err error) {
	proj := convert(c, project)
	return proj.GetMachineGroup(machineGroup)
}

// CreateMachineGroup creates a new machine group in SLS.
func (c *Client) CreateMachineGroup(project string, m *MachineGroup) error {
	proj := convert(c, project)
	return proj.CreateMachineGroup(m)
}

// UpdateMachineGroup updates a machine group.
func (c *Client) UpdateMachineGroup(project string, m *MachineGroup) (err error) {
	proj := convert(c, project)
	return proj.UpdateMachineGroup(m)
}

// DeleteMachineGroup deletes machine group according machine group name.
func (c *Client) DeleteMachineGroup(project string, machineGroup string) (err error) {
	proj := convert(c, project)
	return proj.DeleteMachineGroup(machineGroup)
}

// ListConfig returns config names list and the total number of configs.
// The offset starts from 0 and the size is the max number of configs could be returned.
func (c *Client) ListConfig(project string, offset, size int) (cfgNames []string, total int, err error) {
	proj := convert(c, project)
	return proj.ListConfig(offset, size)
}

// CheckConfigExist check config exist or not
func (c *Client) CheckConfigExist(project string, config string) (ok bool, err error) {
	proj := convert(c, project)
	return proj.CheckConfigExist(config)
}

// GetConfig returns config according by config name.
func (c *Client) GetConfig(project string, config string) (logConfig *LogConfig, err error) {
	proj := convert(c, project)
	return proj.GetConfig(config)
}

// UpdateConfig updates a config.
func (c *Client) UpdateConfig(project string, config *LogConfig) (err error) {
	proj := convert(c, project)
	return proj.UpdateConfig(config)
}

// CreateConfig creates a new config in SLS.
func (c *Client) CreateConfig(project string, config *LogConfig) (err error) {
	proj := convert(c, project)
	return proj.CreateConfig(config)
}

// DeleteConfig deletes a config according by config name.
func (c *Client) DeleteConfig(project string, config string) (err error) {
	proj := convert(c, project)
	return proj.DeleteConfig(config)
}

// GetAppliedMachineGroups returns applied machine group names list according config name.
func (c *Client) GetAppliedMachineGroups(project string, confName string) (groupNames []string, err error) {
	proj := convert(c, project)
	return proj.GetAppliedMachineGroups(confName)
}

// GetAppliedConfigs returns applied config names list according machine group name groupName.
func (c *Client) GetAppliedConfigs(project string, groupName string) (confNames []string, err error) {
	proj := convert(c, project)
	return proj.GetAppliedConfigs(groupName)
}

// ApplyConfigToMachineGroup applies config to machine group.
func (c *Client) ApplyConfigToMachineGroup(project string, confName, groupName string) (err error) {
	proj := convert(c, project)
	return proj.ApplyConfigToMachineGroup(confName, groupName)
}

// RemoveConfigFromMachineGroup removes config from machine group.
func (c *Client) RemoveConfigFromMachineGroup(project string, confName, groupName string) (err error) {
	proj := convert(c, project)
	return proj.RemoveConfigFromMachineGroup(confName, groupName)
}

func (c *Client) CreateEtlMeta(project string, etlMeta *EtlMeta) (err error) {
	proj := convert(c, project)
	return proj.CreateEtlMeta(etlMeta)
}

func (c *Client) UpdateEtlMeta(project string, etlMeta *EtlMeta) (err error) {
	proj := convert(c, project)
	return proj.UpdateEtlMeta(etlMeta)
}

func (c *Client) DeleteEtlMeta(project string, etlMetaName, etlMetaKey string) (err error) {
	proj := convert(c, project)
	return proj.DeleteEtlMeta(etlMetaName, etlMetaKey)
}

func (c *Client) listEtlMeta(project string, etlMetaName, etlMetaKey, etlMetaTag string, offset, size int) (total int, count int, etlMeta []*EtlMeta, err error) {
	proj := convert(c, project)
	return proj.listEtlMeta(etlMetaName, etlMetaKey, etlMetaTag, offset, size)
}

func (c *Client) GetEtlMeta(project string, etlMetaName, etlMetaKey string) (etlMeta *EtlMeta, err error) {
	proj := convert(c, project)
	return proj.GetEtlMeta(etlMetaName, etlMetaKey)
}

func (c *Client) ListEtlMeta(project string, etlMetaName string, offset, size int) (total int, count int, etlMetaList []*EtlMeta, err error) {
	return c.listEtlMeta(project, etlMetaName, "", EtlMetaAllTagMatch, offset, size)
}

func (c *Client) ListEtlMetaWithTag(project string, etlMetaName, etlMetaTag string, offset, size int) (total int, count int, etlMetaList []*EtlMeta, err error) {
	return c.listEtlMeta(project, etlMetaName, "", etlMetaTag, offset, size)
}

func (c *Client) ListEtlMetaName(project string, offset, size int) (total int, count int, etlMetaNameList []string, err error) {
	proj := convert(c, project)
	return proj.ListEtlMetaName(offset, size)
}
