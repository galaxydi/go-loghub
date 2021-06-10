package sls

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	client := makeClient()
	err := setUp(client)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = client.CreateScheduledSQL("test-scheduled-sql", getScheduleSQL("111"))
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestDelete(t *testing.T) {
	client := makeClient()
	err := client.DeleteScheduledSQL("test-scheduled-sql", "test01")
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestUpdate(t *testing.T) {
	client := makeClient()
	err := client.UpdateScheduledSQL("test-scheduled-sql", getScheduleSQL("222"))
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestGet(t *testing.T) {
	client := makeClient()
	scheduledSQL, err := client.GetScheduledSQL("test-scheduled-sql", "test01")
	if err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Printf("%v\n", scheduledSQL)
}

func TestList(t *testing.T) {
	client := makeClient()
	scheduledSQL, total, count, err := client.ListScheduledSQL("test-scheduled-sql", "", "", 0, 10)
	if err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Printf("%v\n%d\n%d\n", scheduledSQL, total, count)
}

func makeClient() *Client {
	return &Client{
		Endpoint:        "pub-cn-hangzhou-staging-share.log.aliyuncs.com",
		AccessKeyID:     os.Getenv("ALICLOUD_ACCESS_KEY"),
		AccessKeySecret: os.Getenv("ALICLOUD_SECRET_KEY"),
	}
}

func getScheduleSQL(des string) *ScheduledSQL {
	return &ScheduledSQL{
		Name:        "test01",
		DisplayName: "dis001",
		Description: des,
		Status:      ENABLED,
		Configuration: &ScheduledSQLConfiguration{
			SourceLogStore:      "test-source",
			DestProject:         "test-schedulesql",
			DestEndpoint:        "cn-hangzhou-intranet.log.aliyuncs.com",
			DestLogStore:        "test-target",
			Script:              "*|SELECT COUNT(__value__)",
			SqlType:             SEARCH_QUERY,
			ResourcePool:        DEFAULT,
			RoleArn:             os.Getenv("ROLE_ARN"),
			DestRoleArn:         os.Getenv("ROLE_ARN"),
			FromTimeExpr:        "@m-15m",
			ToTimeExpr:          "@m",
			MaxRunTimeInSeconds: 60,
			MaxRetries:          20,
			FromTime:            1621828800,
			ToTime:              1623311901,
			DataFormat:          LOG_TO_LOG,
			Parameters:          nil,
		},
		Schedule: &Schedule{
			Type:      "FixedRate",
			Interval:  "15m",
			Delay:     30,
			DayOfWeek: 0,
			Hour:      0,
		},
		CreateTime:       0,
		LastModifiedTime: 0,
		Type:             SCHEDULED_SQL_JOB,
	}
}

func setUp(c *Client) error {
	if ok, err := c.CheckProjectExist("test-scheduled-sql"); err != nil {
		return err
	} else if ok {
		err := c.DeleteProject("test-scheduled-sql")
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 30)
		_, err = c.CreateProject("test-scheduled-sql", "test scheduled sql")
		if err != nil {
			return err
		} else {
			time.Sleep(time.Second * 60)
		}
	}
	err1 := c.CreateLogStore("test-scheduled-sql", "test-source", 3, 2, true, 4)
	if err1 != nil {
		return err1
	}
	err2 := c.CreateLogStore("test-scheduled-sql", "test-target", 3, 2, true, 4)
	if err2 != nil {
		return err2
	}
	err3 := c.CreateIndex("test-scheduled-sql", "test-source", Index{
		Keys: map[string]IndexKey{"__labels__": {
			Token:         []string{",", " ", "'"},
			CaseSensitive: true,
			Type:          "text",
			DocValue:      true,
			Chn:           true,
		}},
		Line: nil,
	})
	if err3 != nil {
		return err3
	}
	return nil
}
