package utils

import (
	"reflect"
)

// SearchArray .
// not found return -1
func SearchArray(arr []interface{}, target interface{}) (index int) {
	index = -1
	for i, v := range arr {
		if v == target {
			return i
		}
	}
	return
}

// GetUniqueArraySubSequences 获取两个unique的数组之间的 所有同序子组 (length >= 2)
// 时间复杂度 O(n^2)
func GetUniqueArraySubSequences(arr1, arr2 []interface{}) (result [][]interface{}) {
	length1 := len(arr1)
	length2 := len(arr2)

	for i := 0; i < length1; i++ {
		var sub []interface{}
		for j := 0; j < length2; j++ {
			if arr1[i] == arr2[j] {
				sub = append(sub, arr1[i])
				if i == length1-1 || j == length2-1 {
					if len(sub) > 1 {
						result = append(result, sub)
					}
					break
				}
				i++
			} else if len(sub) >= 1 {
				if len(sub) > 1 {
					result = append(result, sub)
				}
				i-- // recover i index
				break
			}
		}
	}

	return result
}

// SliceValueToInterface   []T -> []interface{}
func SliceValueToInterface(arr interface{}) (slice []interface{}) {
	slice = []interface{}{}
	val := reflect.ValueOf(arr)
	if val.Kind() != reflect.Slice {
		return
	}
	if val.Len() == 0 {
		return
	}
	slice = make([]interface{}, 0, val.Len())
	for i := 0; i < val.Len(); i++ {
		slice = append(slice, val.Index(i).Interface())
	}
	return slice
}
