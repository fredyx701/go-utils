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
// 问题：无法匹配出有重合部分的子组
// e.g.  {9, 11, 14, 12, 15, 1, 2, 3}, {0, 11, 14, 12, 7, 14, 12, 15, 1, 2, 3}
// 输出 {11, 14, 12}, {15, 1, 2, 3}   实际上 应该输出  {14, 12, 15, 1, 2, 3}  应使用下面的 GetLongestSubSequence 方法
func GetUniqueArraySubSequences(arr1, arr2 []interface{}) (result [][]interface{}) {
	len1 := len(arr1)
	len2 := len(arr2)
	for i := 0; i < len1-1; i++ { // 不处理 i = len1 - 1, 只剩一个元素
		var sub []interface{}
		for j := 0; j < len2; j++ {
			if arr1[i] == arr2[j] { // 拼接子组
				sub = append(sub, arr1[i])
				if i == len1-1 || j == len2-1 { // 末尾
					if len(sub) > 1 {
						result = append(result, sub)
					}
					break
				}
				i++ // i , j 同时递增
			} else if len(sub) >= 1 {
				if len(sub) > 1 {
					result = append(result, sub) // 只返回 len >= 2 的子组
				}
				i-- // 回退一位
				break
			}
		}
	}

	return result
}

// GetLongestSubSequence 获取最大公共子序列
// 二维动态规划  dp[i][j] 表示 arr1 第 i 个元素和 arr2 第 j 个元素为最后一个元素所构成的最长公共子序列，
// 时间复杂度 O(n²)、空间复杂度 O(n²)
func GetLongestSubSequence(arr1, arr2 []interface{}) (result []interface{}) {
	len1, len2 := len(arr1), len(arr2)
	dp := make([][]int, len1+1)
	for i := 0; i < len1+1; i++ {
		dp[i] = make([]int, len2+1)
	}
	maxLen, index := 0, 0
	for i := 1; i < len1+1; i++ {
		for j := 1; j < len2+1; j++ {
			if arr1[i-1] == arr2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
				if maxLen < dp[i][j] {
					maxLen = dp[i][j]
					index = i
				}
			} // else  dp[i][j] = 0.  初始化时 已为 0
		}
	}
	return arr1[index-maxLen : index]
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
