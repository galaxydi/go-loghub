package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"reflect"
)

// just like python function set
func Set(slc []int) []int {
	result := []int{}
	tempMap := map[int]byte{}  // 存放不重复主键
	for _, e := range slc{
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l{  // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// 检查后面的列表减去前面的列表，求差 TODO 这个有问题
func Subtract(a []int, b []int) (diffSlice []int) {

	lengthA := len(a)

	if len(a) == 0{
		return b
	}

	for _, valueB := range b {

		temp := valueB //遍历取出B中的元素
		for j := 0; j < lengthA; j++ {
			if temp == a[j] { //如果相同 比较下一个
				break
			} else {
				if lengthA == (j + 1) { //如果不同 查看a的元素个数及当前比较元素的位置 将不同的元素添加到返回slice中
					diffSlice = append(diffSlice, temp)

				}
			}
		}
	}


	return diffSlice
}

func Min(a,b int64)int64{
	if a > b {
		return b
	}
	if a <b {
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

func GetLogCount(logGroupList *sls.LogGroupList) int {
	a:=0
	for _,x:= range logGroupList.LogGroups{
		a = a + len(x.Logs)
	}
	return a
}











