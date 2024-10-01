package reviewService

import "sort"

func searchInsert(nums []int, target int) int {
	return sort.SearchInts(nums, target)
}
