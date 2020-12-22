package utils

import "reflect"

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

// Set .
type Set map[interface{}]struct{}

// NewSet .
func NewSet(arr []interface{}) Set {
	set := make(Set, len(arr))
	for _, v := range arr {
		set[v] = struct{}{}
	}
	return set
}

// Subset 判断子集
func (s Set) Subset(arr []interface{}) bool {
	for _, v := range arr {
		if _, has := s[v]; !has {
			return false
		}
	}
	return true
}

// Intersect 取交集
func (s Set) Intersect(arr []interface{}) []interface{} {
	result := make([]interface{}, 0, len(arr))
	for _, v := range arr {
		if _, has := s[v]; has {
			result = append(result, v)
		}
	}
	return result
}

// Union 并集
func (s Set) Union(arr []interface{}) []interface{} {
	result := make([]interface{}, 0, len(arr)+len(s))
	for k := range s {
		result = append(result, k)
	}
	for _, v := range arr {
		if _, has := s[v]; !has {
			result = append(result, v)
		}
	}
	return result
}

// Diff 差集
// default return source - target
// isMinuend optional parameters. true return target - source
func (s Set) Diff(target []interface{}, isMinuend ...bool) (diff []interface{}) {
	arrSet := NewSet(target)
	if len(isMinuend) > 0 && isMinuend[0] {
		diff = make([]interface{}, 0, len(arrSet)/2)
		for k := range arrSet {
			if _, has := s[k]; !has {
				diff = append(diff, k)
			}
		}
	} else {
		diff = make([]interface{}, 0, len(s)/2)
		for k := range s {
			if _, has := arrSet[k]; !has {
				diff = append(diff, k)
			}
		}
	}
	return diff
}

// DiffBoth .
// subtrahend  source - target
// minuend  target - source
func (s Set) DiffBoth(target []interface{}) (subtrahend []interface{}, minuend []interface{}) {
	arrSet := NewSet(target)
	subtrahend = make([]interface{}, 0, len(s)/2)
	for k := range s {
		if _, has := arrSet[k]; !has {
			subtrahend = append(subtrahend, k)
		}
	}
	minuend = make([]interface{}, 0, len(arrSet)/2)
	for k := range arrSet {
		if _, has := s[k]; !has {
			minuend = append(minuend, k)
		}
	}
	return subtrahend, minuend
}
