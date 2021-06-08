package sls

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
)

type ScheduledSQL struct {
	Name             string                     `json:"name"`
	DisplayName      string                     `json:"displayName"`
	Description      string                     `json:"description"`
	Status           string                     `json:"status"`
	ScheduleId       string                     `json:"scheduleId"`
	Configuration    *ScheduledSQLConfiguration `json:"configuration"`
	Schedule         *Schedule                  `json:"schedule"`
	CreateTime       int64                      `json:"createTime,omitempty"`
	LastModifiedTime int64                      `json:"lastModifiedTime,omitempty"`
}

type ScheduledSQLConfiguration struct {
	SourceLogStore      string                  `json:"sourceLogstore"`
	DestProject         string                  `json:"destProject"`
	DestEndpoint        string                  `json:"destEndpoint"`
	DestLogStore        string                  `json:"destLogstore"`
	Script              string                  `json:"script"`
	SqlType             string                  `json:"sqlType"`
	ResourcePool        string                  `json:"resourcePool"`
	RoleArn             string                  `json:"roleArn"`
	DestRoleArn         string                  `json:"destRoleArn"`
	FromTimeExpr        string                  `json:"fromTimeExpr"`
	ToTimeExpr          string                  `json:"toTimeExpr"`
	MaxRunTimeInSeconds int32                   `json:"maxRunTimeInSeconds"`
	MaxRetries          int32                   `json:"maxRetries"`
	FromTime            int64                   `json:"fromTime"`
	ToTime              int64                   `json:"toTime"`
	DataFormat          string                  `json:"dataFormat"`
	Parameters          *ScheduledSQLParameters `json:"parameters,omitempty"`
}

func (s *ScheduledSQL) MarshalJSON() ([]byte, error) {
	body := map[string]interface{}{
		"name":             s.Name,
		"displayName":      s.DisplayName,
		"description":      s.Description,
		"status":           s.Status,
		"scheduleId":       s.ScheduleId,
		"configuration":    s.Configuration,
		"schedule":         s.Schedule,
		"createTime":       s.CreateTime,
		"lastModifiedTime": s.LastModifiedTime,
		"type":             "ScheduledSQL",
	}
	return json.Marshal(body)
}

func NewScheduledSQLConfiguration() *ScheduledSQLConfiguration {
	return &ScheduledSQLConfiguration{
		SqlType:      "standard",
		ResourcePool: "default",
		FromTime:     0,
		ToTime:       0,
		DataFormat:   "log2log",
	}
}

type ScheduledSQLParameters struct {
	TimeKey    string `json:"timeKey,omitempty"`
	LabelKeys  string `json:"labelKeys,omitempty"`
	MetricKeys string `json:"metricKeys,omitempty"`
	MetricName string `json:"metricName,omitempty"`
	HashLabels string `json:"hashLabels,omitempty"`
	AddLabels  string `json:"addLabels,omitempty"`
}

func (c *Client) CreateScheduledSQL(project string, scheduledsql *ScheduledSQL) error {
	fromTime := scheduledsql.Configuration.FromTime
	toTime := scheduledsql.Configuration.ToTime
	timeRange := fromTime > 1451577600 && toTime > fromTime
	sustained := fromTime > 1451577600 && toTime == 0
	if !timeRange && !sustained {
		return fmt.Errorf("invalid fromTime: %d toTime: %d, please ensure fromTime more than 1451577600", fromTime, toTime)
	}
	body, err := scheduledsql.MarshalJSON()
	if err != nil {
		return NewClientError(err)
	}
	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
	}

	uri := "/jobs"
	r, err := c.request(project, "POST", uri, h, body)
	if err != nil {
		return err
	}
	r.Body.Close()
	return nil
}

func (c *Client) DeleteScheduledSQL(project string, name string) error {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Content-Type":      "application/json",
	}

	uri := "/jobs/" + name
	r, err := c.request(project, "DELETE", uri, h, nil)
	if err != nil {
		return err
	}
	r.Body.Close()
	return nil
}

func (c *Client) UpdateScheduledSQL(project string, scheduledsql *ScheduledSQL) error {
	body, err := scheduledsql.MarshalJSON()
	if err != nil {
		return NewClientError(err)
	}
	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
	}

	uri := "/jobs/" + scheduledsql.Name
	r, err := c.request(project, "PUT", uri, h, body)
	if err != nil {
		return err
	}
	r.Body.Close()
	return nil
}

func (c *Client) GetScheduledSQL(project string, name string) (*ScheduledSQL, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Content-Type":      "application/json",
	}
	uri := "/jobs/" + name
	r, err := c.request(project, "GET", uri, h, nil)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	buf, _ := ioutil.ReadAll(r.Body)
	scheduledSQL := &ScheduledSQL{}
	if err = json.Unmarshal(buf, scheduledSQL); err != nil {
		err = NewClientError(err)
	}
	return scheduledSQL, err
}

func (c *Client) ListScheduledSQL(project, name, displayName string, offset, size int) (scheduledsqls []*ScheduledSQL, total, count int, error error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Content-Type":      "application/json",
	}
	v := url.Values{}
	v.Add("jobName", name)
	if displayName != "" {
		v.Add("displayName", displayName)
	}
	v.Add("jobType", "ScheduledSQL")
	v.Add("offset", fmt.Sprintf("%d", offset))
	v.Add("size", fmt.Sprintf("%d", size))

	uri := "/jobs?" + v.Encode()
	r, err := c.request(project, "GET", uri, h, nil)
	if err != nil {
		return nil, 0, 0, err
	}
	defer r.Body.Close()

	type ScheduledSqlList struct {
		Total   int             `json:"total"`
		Count   int             `json:"count"`
		Results []*ScheduledSQL `json:"results"`
	}
	buf, _ := ioutil.ReadAll(r.Body)
	scheduledSqlList := &ScheduledSqlList{}
	if err = json.Unmarshal(buf, scheduledSqlList); err != nil {
		err = NewClientError(err)
	}
	return scheduledSqlList.Results, scheduledSqlList.Total, scheduledSqlList.Count, err
}
