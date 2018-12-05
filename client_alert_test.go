package sls

import (
	"testing"
	"time"
)

var testProject = "test-project-name"
var testAlertName = "alert-test-name"
var testDashboard = "test-dashboard-name"

func client() *Client {
	return &Client{
		AccessKeyID:     "",
		AccessKeySecret: "",
		Endpoint:        "",
	}
}

func TestClient_CreateAlert(t *testing.T) {
	client().DeleteAlert(testProject, testAlertName)
	err := createAlert()
	if err != nil {
		t.Fatal(err)
	}
	client().DeleteAlert(testProject, testAlertName)
}

func createAlert() error {
	alert := &Alert{
		Name:        testAlertName,
		State:       "Enabled",
		DisplayName: "AlertTest",
		Description: "Description for alert",
		Schedule: &Schedule{
			Type:     "FixedRate",
			Interval: "1h",
		},
		Configuration: &AlertConfiguration{
			QueryList: []*AlertQuery{
				{
					ChartTitle:   "chart-abc",
					Query:        "* | select count(1) as count",
					Start:        "-120s",
					End:          "now",
					TimeSpanType: "Custom",
					LogStore:     "test-logstore",
				},
			},
			Dashboard:  testDashboard,
			MuteUntil:  time.Now().Unix() + 3600,
			Throttling: "5m",
			Condition:  "count > 0",
			NotificationList: []*Notification{
				{
					Type:      NotificationTypeEmail,
					Content:   "${alertName} triggered at ${firetime}",
					EmailList: []string{"test@abc.com"},
				},
				{
					Type:       NotificationTypeSMS,
					Content:    "${alertName} triggered at ${firetime}",
					MobileList: []string{"1234567891"},
				},
			},
			NotifyThreshold: 1,
		},
	}
	return client().CreateAlert(testProject, alert)
}

func TestClient_UpdateAlert(t *testing.T) {
	createAlert()
	alert, err := client().GetAlert(testProject, testAlertName)
	if err != nil {
		t.Fatal(err)
	}
	alert.DisplayName = "new display name"
	alert.CreateTime = 0
	alert.LastModifiedTime = 0
	err = client().UpdateAlert(testProject, alert)
	if err != nil {
		t.Fatal(err)
	}
	alert, err = client().GetAlert(testProject, testAlertName)
	if alert.DisplayName != "new display name" {
		t.Fatal("update alert failed")
	}
	client().DeleteAlert(testProject, testAlertName)
}

func TestClient_DeleteAlert(t *testing.T) {
	createAlert()
	_, err := client().GetAlert(testProject, testAlertName)
	if err != nil {
		t.Fatal(err)
	}
	err = client().DeleteAlert(testProject, testAlertName)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client().GetAlert(testProject, testAlertName)
	if err == nil {
		t.Fatal(err)
	}
}

func TestClient_DisableAndEnableAlert(t *testing.T) {
	createAlert()
	err := client().DisableAlert(testProject, testAlertName)
	if err != nil {
		t.Fatal(err)
	}
	alert, err := client().GetAlert(testProject, testAlertName)
	if err != nil {
		t.Fatal(err)
	}
	if alert.State != "Disabled" {
		t.Fatal("disable alert failed")
	}
	err = client().EnableAlert(testProject, testAlertName)
	if err != nil {
		t.Fatal(err)
	}
	alert, err = client().GetAlert(testProject, testAlertName)
	if err != nil {
		t.Fatal(err)
	}
	if alert.State != "Enabled" {
		t.Fatal("enable alert failed")
	}
	client().DeleteAlert(testProject, testAlertName)
}

func TestClient_GetAlert(t *testing.T) {
	getAlert, err := client().GetAlert("project-to-test-alert", "alert-test-name")
	if err != nil {
		t.Fatal(err)
	}
	if getAlert.Name != "alert-test-name" {
		t.Fail()
	}
}

func TestClient_ListAlert(t *testing.T) {
	err := createAlert()
	if err != nil {
		t.Fatal(err)
	}
	alerts, total, count, err := client().ListAlert(testProject, "", "", 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || count != 1 {
		t.Log(total)
		t.Log(count)
		t.Fatal("list alert failed")
	}
	if len(alerts) != 1 {
		t.Fatal("there should be only one alert")
	}
	alert := alerts[0]
	if alert.Name != testAlertName {
		t.Fatal("list alert failed")
	}
	client().DeleteAlert(testProject, testAlertName)
}
