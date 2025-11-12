package utils

import "slices"

// 从切片中移除重复元素
func RemoveDuplicates[T comparable](slice []T) []T {
	keys := make(map[T]bool)
	var result []T

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// 检查切片是否包含元素
func Contains[T comparable](slice []T, element T) bool {
	return slices.Contains(slice, element)
}
