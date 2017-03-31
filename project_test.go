package sls

import (
	"fmt"
	"testing"
	"time"
)

func TestCreateDeleteProject(t *testing.T) {
	c := Client{
		Endpoint:        "cn-shanghai.log.aliyuncs.com",
		AccessKeyID:     "LTAIJ4gv93WLYyA4",
		AccessKeySecret: "xYv3JoZ8jeFmRIVzIiRUghEwVHnVdQ",
	}

	for i := 0; i < 2; i++ {
		var err error
		name := fmt.Sprintf("faint%d", i)
		_, err = c.CreateProject(name, "this is faint")
		if err != nil {
			fmt.Printf("Create project error: %v\n", err)
			return
		}

		time.Sleep(5 * time.Second)

		err = c.DeleteProject(name)
		if err != nil {
			fmt.Printf("Delete project error: %v\n", err)
			return
		}
	}
}
