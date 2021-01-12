package sls

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	// MetricAggRulesSQL sql type
	MetricAggRulesSQL = "SQL"
	// MetricAggRulesPromQL promql type
	MetricAggRulesPromQL = "PromQL"
)

type MetricAggRules struct {
	ID   string
	Name string
	Desc string

	SrcStore           string
	SrcAccessKeyID     string // ETL_STS_DEFAULT
	SrcAccessKeySecret string // acs:ram::${aliuid}:role/aliyunlogetlrole

	DestEndpoint        string // same region, inner endpoint; different region, public endpoint
	DestProject         string
	DestStore           string
	DestAccessKeyID     string // ETL_STS_DEFAULT
	DestAccessKeySecret string // acs:ram::${aliuid}:role/aliyunlogetlrole

	AggRules []MetricAggRuleItem
}

type MetricAggRuleItem struct {
	Name        string
	QueryType   string
	Query       string
	TimeName    string
	MetricNames []string
	LabelNames  map[string]string

	BeginUnixTime int64
	EndUnixTime   int64
	Interval      int64
	DelaySeconds  int64
}

func (c *Client) getScheduledSQLParams(aggRules []MetricAggRuleItem) map[string]string {
	params := make(map[string]string)
	params["sls.config.job_mode"] = `{"type":"ml","source":"ScheduledSQL"}`

	var aggRuleJsons []interface{}

	for _, aggRule := range aggRules {
		aggRuleMap := make(map[string]interface{})

		aggRuleMap["rule_name"] = aggRule.Name

		advancedQueryMap := make(map[string]interface{})
		advancedQueryMap["type"] = aggRule.QueryType
		advancedQueryMap["query"] = aggRule.Query
		advancedQueryMap["time_name"] = aggRule.TimeName
		advancedQueryMap["metric_names"] = aggRule.MetricNames
		advancedQueryMap["labels"] = aggRule.LabelNames
		aggRuleMap["advanced_query"] = advancedQueryMap

		scheduleControlMap := make(map[string]interface{})
		scheduleControlMap["from_unixtime"] = aggRule.BeginUnixTime
		scheduleControlMap["to_unixtime"] = aggRule.EndUnixTime
		scheduleControlMap["granularity"] = aggRule.Interval
		scheduleControlMap["delay"] = aggRule.DelaySeconds
		aggRuleMap["schedule_control"] = scheduleControlMap

		aggRuleJsons = append(aggRuleJsons, aggRuleMap)
	}

	scheduledSql := make(map[string]interface{})
	scheduledSql["agg_rules"] = aggRuleJsons
	scheduledSqlJson, err := json.Marshal(scheduledSql)
	if err != nil {
		fmt.Printf("Marshal scheduledSql error, %s \n %s\n", err.Error(), scheduledSqlJson)
	}
	params["config.ml.scheduled_sql"] = string(scheduledSqlJson)

	return params
}

func (c *Client) createMetricAggRulesConfig(aggRules *MetricAggRules) *ETL {
	etl := new(ETL)

	etl.Name = aggRules.ID
	etl.DisplayName = aggRules.Name
	etl.Description = aggRules.Desc
	etl.Type = "ETL"

	etl.Configuration.AccessKeyId = aggRules.SrcAccessKeyID
	etl.Configuration.AccessKeySecret = aggRules.SrcAccessKeySecret
	etl.Configuration.Script = ""
	etl.Configuration.Logstore = aggRules.SrcStore
	etl.Configuration.Parameters = c.getScheduledSQLParams(aggRules.AggRules)
	etl.Configuration.FromTime = time.Now().Unix()

	var sink ETLSink
	sink.Endpoint = aggRules.DestEndpoint
	sink.Name = "sls-convert-metric"
	sink.AccessKeyId = aggRules.DestAccessKeyID
	sink.AccessKeySecret = aggRules.DestAccessKeySecret
	sink.Project = aggRules.DestProject
	sink.Logstore = aggRules.DestStore
	etl.Configuration.ETLSinks = append(etl.Configuration.ETLSinks, sink)

	etl.Schedule.Type = ScheduleTypeResident

	return etl
}

func (c *Client) CreateMetricAggRules(project string, aggRules *MetricAggRules) error {
	etl := c.createMetricAggRulesConfig(aggRules)
	if err := c.CreateETL(project, *etl); err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateMetricAggRules(project string, aggRules *MetricAggRules) error {
	etl := c.createMetricAggRulesConfig(aggRules)
	if err := c.UpdateETL(project, *etl); err != nil {
		return err
	}
	return nil
}

func (c *Client) castInterfaceArrayToStringArray(inter map[string]interface{}, key string) []string {
	t, ok := inter[key].([]interface{})
	if !ok {
		fmt.Printf("castInterfaceArrayToStringArray is not ok, key: %s, value: %v\n", key, inter[key])
		return []string{}
	}
	s := make([]string, len(t))
	for i, v := range t {
		s[i] = fmt.Sprint(v)
	}
	return s
}

func (c *Client) castInterfaceMapToStringMap(inter map[string]interface{}, key string) map[string]string {
	t, ok := inter[key].(map[string]interface{})
	if !ok {
		fmt.Printf("castInterfaceMapToStringMap is not ok, key: %s, value: %v\n", key, inter[key])
		return map[string]string{}
	}
	s := make(map[string]string, len(t))
	for k, v := range t {
		s[k] = fmt.Sprint(v)
	}
	return s
}

func (c *Client) castInterfaceToInt(inter map[string]interface{}, key string) int64 {
	t, ok := inter[key].(float64)
	if !ok {
		fmt.Printf("castInterfaceToInt is not ok, key: %s, value: %v\n", key, inter[key])
	}
	return int64(t)
}

func (c *Client) castInterfaceToString(inter map[string]interface{}, key string) string {
	t, ok := inter[key].(string)
	if !ok {
		fmt.Printf("castInterfaceToString is not ok, key: %s, value: %v\n", key, inter[key])
	}
	return t
}

func (c *Client) castInterfaceToMap(inter map[string]interface{}, key string) (map[string]interface{}, bool) {
	t, ok := inter[key].(map[string]interface{})
	if !ok {
		fmt.Printf("castInterfaceToMap is not ok, key: %s, value: %v\n", key, inter[key])
	}
	return t, ok
}

func (c *Client) GetMetricAggRules(project string, ruleID string) (*MetricAggRules, error) {
	etl, err := c.GetETL(project, ruleID)
	if err != nil {
		return nil, err
	}

	aggRules := new(MetricAggRules)
	aggRules.ID = etl.Name
	aggRules.Name = etl.DisplayName
	aggRules.Desc = etl.Description
	aggRules.SrcAccessKeyID = etl.Configuration.AccessKeyId
	aggRules.SrcAccessKeySecret = etl.Configuration.AccessKeySecret
	aggRules.SrcStore = etl.Configuration.Logstore

	scheduledSqlJson := etl.Configuration.Parameters["config.ml.scheduled_sql"]
	aggRuleJson := make(map[string][]map[string]interface{})
	err = json.Unmarshal([]byte(scheduledSqlJson), &aggRuleJson)
	if err != nil {
		fmt.Printf("Unmarshal scheduledSqlJson error, %s \n %s\n", err.Error(), scheduledSqlJson)
		panic(err)
	}
	aggRuleMaps := aggRuleJson["agg_rules"]

	var aggRuleItems []MetricAggRuleItem
	for _, aggRuleMap := range aggRuleMaps {
		aggRuleItem := new(MetricAggRuleItem)

		aggRuleItem.Name = c.castInterfaceToString(aggRuleMap, "rule_name")

		advancedQuery, ok := c.castInterfaceToMap(aggRuleMap, "advanced_query")
		if ok {
			aggRuleItem.QueryType = c.castInterfaceToString(advancedQuery, "type")
			aggRuleItem.Query = c.castInterfaceToString(advancedQuery, "query")
			aggRuleItem.TimeName = c.castInterfaceToString(advancedQuery, "time_name")
			aggRuleItem.MetricNames = c.castInterfaceArrayToStringArray(advancedQuery, "metric_names")
			aggRuleItem.LabelNames = c.castInterfaceMapToStringMap(advancedQuery, "labels")
		}

		scheduleControl, ok := c.castInterfaceToMap(aggRuleMap, "schedule_control")
		if ok {
			aggRuleItem.BeginUnixTime = c.castInterfaceToInt(scheduleControl, "from_unixtime")
			aggRuleItem.EndUnixTime = c.castInterfaceToInt(scheduleControl, "to_unixtime")
			aggRuleItem.Interval = c.castInterfaceToInt(scheduleControl, "granularity")
			aggRuleItem.DelaySeconds = c.castInterfaceToInt(scheduleControl, "delay")
		}
		aggRuleItems = append(aggRuleItems, *aggRuleItem)
	}
	aggRules.AggRules = aggRuleItems
	for _, sink := range etl.Configuration.ETLSinks {
		aggRules.DestEndpoint = sink.Endpoint
		aggRules.DestAccessKeyID = sink.AccessKeyId
		aggRules.DestAccessKeySecret = sink.AccessKeySecret
		aggRules.DestProject = sink.Project
		aggRules.DestStore = sink.Logstore
	}

	return aggRules, nil

}

func (c *Client) DeleteMetricAggRules(project string, ruleID string) error {
	if err := c.DeleteETL(project, ruleID); err != nil {
		return err
	}
	return nil
}
