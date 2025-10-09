package utils

import "slices"

// @en remove duplicate elements from slice
//
// @zh 从切片中移除重复元素
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

// @en check if slice contains element
//
// @zh 检查切片是否包含元素
func Contains[T comparable](slice []T, element T) bool {
	return slices.Contains(slice, element)
}
