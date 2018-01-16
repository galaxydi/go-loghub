package sls

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// LogProject defines log project
type LogProject struct {
	Name            string // Project name
	Endpoint        string // IP or hostname of SLS endpoint
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string
	baseURL         string
}

// NewLogProject new a SLS project object.
func NewLogProject(name, endpoint, accessKeyID, accessKeySecret string) (p *LogProject, err error) {
	p = &LogProject{
		Name:            name,
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
	}
	p.parseEndpoint()
	return p, nil
}

// WithToken add token parameter
func (p *LogProject) WithToken(token string) (*LogProject, error) {
	p.SecurityToken = token
	return p, nil
}

// ListLogStore returns all logstore names of project p.
func (p *LogProject) ListLogStore() ([]string, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	uri := fmt.Sprintf("/logstores")
	_, buf, err := request(p, "GET", uri, h, nil)
	if err != nil {
		return nil, err
	}

	type Body struct {
		Count     int
		LogStores []string
	}
	body := &Body{}

	err = json.Unmarshal(buf, body)
	if err != nil {
		return nil, err
	}
	storeNames := body.LogStores
	return storeNames, nil
}

// GetLogStore returns logstore according by logstore name.
func (p *LogProject) GetLogStore(name string) (*LogStore, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	_, buf, err := request(p, "GET", "/logstores/"+name, h, nil)
	if err != nil {
		return nil, err
	}

	s := &LogStore{}
	err = json.Unmarshal(buf, s)
	if err != nil {
		return nil, err
	}
	s.Name = name
	s.project = p
	return s, nil
}

// CreateLogStore creates a new logstore in SLS,
// where name is logstore name,
// and ttl is time-to-live(in day) of logs,
// and shardCnt is the number of shards.
func (p *LogProject) CreateLogStore(name string, ttl, shardCnt int) error {
	type Body struct {
		Name       string `json:"logstoreName"`
		TTL        int    `json:"ttl"`
		ShardCount int    `json:"shardCount"`
	}
	store := &Body{
		Name:       name,
		TTL:        ttl,
		ShardCount: shardCnt,
	}
	body, err := json.Marshal(store)
	if err != nil {
		return NewClientError(err.Error())
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}

	_, _, err = request(p, "POST", "/logstores", h, body)

	return err
}

// DeleteLogStore deletes a logstore according by logstore name.
func (p *LogProject) DeleteLogStore(name string) error {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	_, _, err := request(p, "DELETE", "/logstores/"+name, h, nil)
	return err
}

// UpdateLogStore updates a logstore according by logstore name,
// obviously we can't modify the logstore name itself.
func (p *LogProject) UpdateLogStore(name string, ttl, shardCnt int) error {
	type Body struct {
		Name       string `json:"logstoreName"`
		TTL        int    `json:"ttl"`
		ShardCount int    `json:"shardCount"`
	}
	store := &Body{
		Name:       name,
		TTL:        ttl,
		ShardCount: shardCnt,
	}
	body, err := json.Marshal(store)
	if err != nil {
		return NewClientError(err.Error())
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}
	_, _, err = request(p, "PUT", "/logstores/"+name, h, body)

	return err
}

// ListMachineGroup returns machine group name list and the total number of machine groups.
// The offset starts from 0 and the size is the max number of machine groups could be returned.
func (p *LogProject) ListMachineGroup(offset, size int) ([]string, int, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	if size <= 0 {
		size = 500
	}
	uri := fmt.Sprintf("/machinegroups?offset=%v&size=%v", offset, size)
	_, buf, err := request(p, "GET", uri, h, nil)
	if err != nil {
		return nil, 0, err
	}

	type Body struct {
		MachineGroups []string
		Count         int
		Total         int
	}
	body := &Body{}
	err = json.Unmarshal(buf, body)
	if err != nil {
		return nil, 0, err
	}
	m := body.MachineGroups
	total := body.Total
	return m, total, nil
}

// CheckLogstoreExist check logstore exist or not
func (p *LogProject) CheckLogstoreExist(name string) (bool, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	_, _, err := request(p, "GET", "/logstores/"+name, h, nil)
	if err != nil {
		if _, ok := err.(*Error); ok {
			slsErr := err.(*Error)
			if slsErr.Code == "LogStoreNotExist" {
				return false, nil
			}
			return false, slsErr
		}
		return false, err
	}
	return true, nil
}

// CheckMachineGroupExist check machine group exist or not
func (p *LogProject) CheckMachineGroupExist(name string) (bool, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	_, _, err := request(p, "GET", "/machinegroups/"+name, h, nil)

	if err != nil {
		if _, ok := err.(*Error); ok {
			slsErr := err.(*Error)
			if slsErr.Code == "MachineGroupNotExist" {
				return false, nil
			}
			return false, slsErr
		}
		return false, err
	}
	return true, nil
}

// GetMachineGroup retruns machine group according by machine group name.
func (p *LogProject) GetMachineGroup(name string) (*MachineGroup, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	_, buf, err := request(p, "GET", "/machinegroups/"+name, h, nil)
	if err != nil {
		return nil, err
	}

	m := new(MachineGroup)
	err = json.Unmarshal(buf, m)
	if err != nil {
		return nil, err
	}
	m.project = p
	return m, nil
}

// CreateMachineGroup creates a new machine group in SLS.
func (p *LogProject) CreateMachineGroup(m *MachineGroup) error {
	body, err := json.Marshal(m)
	if err != nil {
		return NewClientError(err.Error())
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}
	_, _, err = request(p, "POST", "/machinegroups", h, body)
	return err
}

// UpdateMachineGroup updates a machine group.
func (p *LogProject) UpdateMachineGroup(m *MachineGroup) error {
	body, err := json.Marshal(m)
	if err != nil {
		return NewClientError(err.Error())
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}
	_, _, err = request(p, "PUT", "/machinegroups/"+m.Name, h, body)
	return err
}

// DeleteMachineGroup deletes machine group according machine group name.
func (p *LogProject) DeleteMachineGroup(name string) error {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	_, _, err := request(p, "DELETE", "/machinegroups/"+name, h, nil)
	return err
}

// ListConfig returns config names list and the total number of configs.
// The offset starts from 0 and the size is the max number of configs could be returned.
func (p *LogProject) ListConfig(offset, size int) ([]string, int, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	if size <= 0 {
		size = 100
	}
	uri := fmt.Sprintf("/configs?offset=%v&size=%v", offset, size)
	_, buf, err := request(p, "GET", uri, h, nil)
	if err != nil {
		return nil, 0, err
	}

	type Body struct {
		Total   int
		Configs []string
	}
	body := &Body{}
	err = json.Unmarshal(buf, body)
	if err != nil {
		return nil, 0, err
	}
	cfgNames := body.Configs
	total := body.Total
	return cfgNames, total, nil
}

// CheckConfigExist check config exist or not
func (p *LogProject) CheckConfigExist(name string) (bool, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	_, _, err := request(p, "GET", "/configs/"+name, h, nil)
	if err != nil {
		if _, ok := err.(*Error); ok {
			slsErr := err.(*Error)
			if slsErr.Code == "ConfigNotExist" {
				return false, nil
			}
			return false, slsErr
		}
		return false, err
	}
	return true, nil
}

// GetConfig returns config according by config name.
func (p *LogProject) GetConfig(name string) (*LogConfig, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	_, buf, err := request(p, "GET", "/configs/"+name, h, nil)
	if err != nil {
		return nil, err
	}

	c := &LogConfig{}
	err = json.Unmarshal(buf, c)
	if err != nil {
		return nil, err
	}
	c.project = p
	return c, nil
}

// UpdateConfig updates a config.
func (p *LogProject) UpdateConfig(c *LogConfig) error {
	body, err := json.Marshal(c)
	if err != nil {
		return NewClientError(err.Error())
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}
	_, _, err = request(p, "PUT", "/configs/"+c.Name, h, body)
	return err
}

// CreateConfig creates a new config in SLS.
func (p *LogProject) CreateConfig(c *LogConfig) error {
	body, err := json.Marshal(c)
	if err != nil {
		return NewClientError(err.Error())
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}
	_, _, err = request(p, "POST", "/configs", h, body)
	return err
}

// DeleteConfig deletes a config according by config name.
func (p *LogProject) DeleteConfig(name string) error {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	_, _, err := request(p, "DELETE", "/configs/"+name, h, nil)
	return err
}

// GetAppliedMachineGroups returns applied machine group names list according config name.
func (p *LogProject) GetAppliedMachineGroups(confName string) ([]string, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	uri := fmt.Sprintf("/configs/%v/machinegroups", confName)
	_, buf, err := request(p, "GET", uri, h, nil)
	if err != nil {
		return nil, err
	}

	type Body struct {
		Count         int
		Machinegroups []string
	}
	body := &Body{}
	err = json.Unmarshal(buf, body)
	if err != nil {
		return nil, err
	}
	groupNames := body.Machinegroups
	return groupNames, nil
}

// GetAppliedConfigs returns applied config names list according machine group name groupName.
func (p *LogProject) GetAppliedConfigs(groupName string) ([]string, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	uri := fmt.Sprintf("/machinegroups/%v/configs", groupName)
	_, buf, err := request(p, "GET", uri, h, nil)
	if err != nil {
		return nil, err
	}

	type Cfg struct {
		Count   int      `json:"count"`
		Configs []string `json:"configs"`
	}
	body := &Cfg{}
	err = json.Unmarshal(buf, body)
	if err != nil {
		return nil, err
	}
	confNames := body.Configs
	return confNames, nil
}

// ApplyConfigToMachineGroup applies config to machine group.
func (p *LogProject) ApplyConfigToMachineGroup(confName, groupName string) error {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	uri := fmt.Sprintf("/machinegroups/%v/configs/%v", groupName, confName)
	_, _, err := request(p, "PUT", uri, h, nil)
	return err
}

// RemoveConfigFromMachineGroup removes config from machine group.
func (p *LogProject) RemoveConfigFromMachineGroup(confName, groupName string) error {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	uri := fmt.Sprintf("/machinegroups/%v/configs/%v", groupName, confName)
	_, _, err := request(p, "DELETE", uri, h, nil)

	return err
}

// CreateETLJob creates a new ETL job in SLS.
func (p *LogProject) CreateETLJob(j *ETLJob) error {
	body, err := json.Marshal(j)
	if err != nil {
		return NewClientError(err.Error())
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}
	_, _, err = request(p, "POST", "/etljobs", h, body)
	return err
}

// GetETLJob returns ETL job according to job name.
func (p *LogProject) GetETLJob(name string) (*ETLJob, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	_, buf, err := request(p, "GET", "/etljobs/"+name, h, nil)
	if err != nil {
		return nil, err
	}
	job := &ETLJob{}
	err = json.Unmarshal(buf, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// UpdateETLJob updates an ETL job according to job name,
// Not all fields of ETLJob can be updated
func (p *LogProject) UpdateETLJob(name string, job *ETLJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return NewClientError(err.Error())
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}
	_, _, err = request(p, "PUT", "/etljobs/"+name, h, body)

	return err
}

// DeleteETLJob deletes a job according to job name.
func (p *LogProject) DeleteETLJob(name string) error {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	_, _, err := request(p, "DELETE", "/etljobs/"+name, h, nil)
	return err
}

// ListETLJobs returns all job names of project.
func (p *LogProject) ListETLJobs() ([]string, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	uri := fmt.Sprintf("/etljobs")
	_, buf, err := request(p, "GET", uri, h, nil)
	if err != nil {
		return nil, err
	}

	type Body struct {
		Count   int
		ETLJobs []string `json:"etlJobNameList"`
		Total   int
	}
	body := &Body{}

	err = json.Unmarshal(buf, body)
	if err != nil {
		return nil, err
	}
	return body.ETLJobs, nil
}

const (
	httpScheme  = "http://"
	httpsScheme = "https://"
	ipRegex     = `\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}.*`
)

func (p *LogProject) parseEndpoint() {
	scheme := httpScheme // default to http scheme
	if strings.HasPrefix(p.Endpoint, httpScheme) {
		scheme = httpScheme
	} else if strings.HasPrefix(p.Endpoint, httpsScheme) {
		scheme = httpsScheme
	}
	host := strings.TrimPrefix(p.Endpoint, scheme)
	reg := regexp.MustCompile(ipRegex)
	if reg.MatchString(host) { // ip format
		p.baseURL = fmt.Sprintf("%s%s/%s", scheme, host, p.Name)
	} else {
		p.baseURL = fmt.Sprintf("%s%s.%s", scheme, p.Name, host)
	}
}
