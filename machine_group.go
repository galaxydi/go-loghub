package sls

import (
	"encoding/json"
	"fmt"
)

// MachinGroupAttribute defines machine group attribute
type MachinGroupAttribute struct {
	ExternalName string `json:"externalName"`
	TopicName    string `json:"groupTopic"`
}

// MachineGroup defines machine group
type MachineGroup struct {
	Name          string   `json:"groupName"`
	Type          string   `json:"groupType"`
	MachineIDType string   `json:"machineIdentifyType"`
	MachineIDList []string `json:"machineList"`

	Attribute MachinGroupAttribute `json:"groupAttribute"`

	CreateTime     uint32
	LastModifyTime uint32

	project *LogProject
}

// Machine defines machine struct
type Machine struct {
	IP            string
	UniqueID      string `json:"machine-uniqueid"`
	UserdefinedID string `json:"userdefined-id"`
}

// MachineList defines machine list
type MachineList struct {
	Total    int
	Machines []*Machine
}

// ListMachines returns machine list of this machine group.
func (m *MachineGroup) ListMachines() ([]*Machine, int, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	uri := fmt.Sprintf("/machinegroups/%v/machines", m.Name)
	_, buf, err := request(m.project, "GET", uri, h, nil)
	if err != nil {
		return nil, 0, err
	}

	body := &MachineList{}
	err = json.Unmarshal(buf, body)
	if err != nil {
		return nil, 0, err
	}

	ms := body.Machines
	total := body.Total

	return ms, total, nil
}

// GetAppliedConfigs returns applied configs of this machine group.
func (m *MachineGroup) GetAppliedConfigs() (confNames []string, err error) {
	confNames, err = m.project.GetAppliedConfigs(m.Name)
	return
}
