package sls

import (
	"fmt"
	"testing"
	"time"
)

func TestCreateDeleteProject(t *testing.T) {
	c := Client{
		Endpoint:        "cn-shanghai.log.aliyuncs.com",
		AccessKeyID:     "",
		AccessKeySecret: "",
	}

	for i := 0; i < 1; i++ {
		var err error
		name := fmt.Sprintf("faint%d", i)
		//proj, err := c.CreateProject(name, "this is faint")
		proj, err := c.GetProject(name)
		if err != nil {
			fmt.Printf("Create project error: %v\n", err)
			return
		}
		time.Sleep(5 * time.Second)

		/*
			err = proj.CreateLogStore("test", 7, 1)
			if err != nil {
				fmt.Printf("Create store error: %v\n", err)
				return
			}
		*/

		store, err := proj.GetLogStore("test")
		if err != nil {
			fmt.Printf("Get store error: %v\n", err)
			return
		}
		err = store.CreateIndex(Index{
			TTL: 7,
			Keys: map[string]IndexKey{
				"function": IndexKey{
					Token:         []string{"\n", "\t", ";", ","},
					CaseSensitive: false,
					Type:          "text",
				},
			},
		})
		if err != nil {
			fmt.Printf("Create index error: %v\n", err)
			return
		}

		err = c.DeleteProject(name)
		if err != nil {
			fmt.Printf("Delete project error: %v\n", err)
			return
		}
	}
}
