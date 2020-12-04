package utils

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
// 对象自身作为数据源
func (s Set) Diff(arr []interface{}) []interface{} {
	arrSet := NewSet(arr)
	result := make([]interface{}, 0, len(s))
	for k := range s {
		if _, has := arrSet[k]; !has {
			result = append(result, k)
		}
	}
	return result
}
