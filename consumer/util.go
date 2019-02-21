package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"reflect"
)

// List removal of duplicate elements
func Set(slc []int) []int {
	result := []int{}
	tempMap := map[int]byte{}
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l {
			result = append(result, e)
		}
	}
	return result
}

// Get the difference between the two lists
func Subtract(a []int, b []int) (diffSlice []int) {

	lengthA := len(a)

	if len(a) == 0 {
		return b
	}

	for _, valueB := range b {

		temp := valueB
		for j := 0; j < lengthA; j++ {
			if temp == a[j] {
				break
			} else {
				if lengthA == (j + 1) {
					diffSlice = append(diffSlice, temp)

				}
			}
		}
	}

	return diffSlice
}

// Returns the smallest of two numbers
func Min(a, b int64) int64 {
	if a > b {
		return b
	}
	if a < b {
		return a
	}
	return 0
}

// Determine whether two lists are equal
func IntSliceReflectEqual(a, b []int) bool {
	return reflect.DeepEqual(a, b)
}

// Determine whether obj is in target object
func Contain(obj interface{}, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}

	return false
}

// Get the total number of logs
func GetLogCount(logGroupList *sls.LogGroupList) int {
	a := 0
	for _, x := range logGroupList.LogGroups {
		a = a + len(x.Logs)
	}
	return a
}
