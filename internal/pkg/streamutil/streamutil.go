package streamutil

// Filter 过滤切片。
func Filter[T any](items []T, filter func(T) bool) []T {
	var result []T
	for _, item := range items {
		if filter(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map 映射切片。
func Map[T any, R any](items []T, mapper func(T) R) []R {
	var result []R
	for _, item := range items {
		result = append(result, mapper(item))
	}
	return result
}

// Reduce 归约切片。
func Reduce[T any, R any](items []T, reducer func(R, T) R, initial R) R {
	for _, item := range items {
		initial = reducer(initial, item)
	}
	return initial
}

// 遍历
func ForEach[T any](items []T, action func(T)) {
	for _, item := range items {
		action(item)
	}
}

// Any 判断切片是否存在元素满足条件。
func Any[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if predicate(item) {
			return true
		}
	}
	return false
}

// All 判断切片是否所有元素都满足条件。
func All[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Count 计算切片元素个数。
func Count[T any](items []T) int {
	return len(items)
}

// Sum 计算切片元素之和。
func Sum[T any, R any](items []T, reducer func(acc R, item T) R, init R) R {
	return Reduce(items, reducer, init)
}

// Average 计算切片元素平均值。
func Average[T any](items []T, reducer func(acc float64, item T) float64, init float64) float64 {
	return float64(Sum(items, reducer, init)) / float64(len(items))
}

// Distinct 去除切片重复元素。
func Distinct[T any, K comparable](items []T, selector func(T) K) []T {
	seen := make(map[K]struct{})
	var res []T
	for _, v := range items {
		key := selector(v)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			res = append(res, v)
		}
	}
	return res
}
