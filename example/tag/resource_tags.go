package main

import (
	"fmt"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/example/util"
)

func main() {
	fmt.Println("Tag Resource")
	projectTags := sls.NewProjectTags(util.ProjectName, []sls.ResourceTag{
		{
			Key:   "the-tag",
			Value: "aliyun-log-go-sdk",
		},
		{
			Key:   "the-tag-2",
			Value: "aliyun log go sdk",
		},
	})
	err := util.Client.TagResources(util.ProjectName, projectTags)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Tag Resource success")
	}

	listAllTags()

	fmt.Println("UnTag Resource")
	projectUnTags := sls.NewProjectUnTags(util.ProjectName, []string{"the-tag"})

	err = util.Client.UnTagResources(util.ProjectName, projectUnTags)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("UnTag Resource success")
	}
	listAllTags()
}

func getStringPtr(data string) *string {
	return &data
}

// List all the projects below this region.
func listAllTags() {
	var nextToken string
	fmt.Println("tag list: ")
	for {
		respTags, respNextToken, err := util.Client.ListTagResources(util.ProjectName, "project", []string{util.ProjectName}, []sls.ResourceFilterTag{}, nextToken)
		if err != nil {
			panic(err)
		}
		for _, tag := range respTags {
			fmt.Printf(" resourceType : %s, resourceID : %s, tagKey : %s, tagValue : %s\n",
				tag.ResourceType,
				tag.ResourceID,
				tag.TagKey,
				tag.TagValue)
		}
		nextToken = respNextToken
		if nextToken == "" {
			break
		}
	}
}
